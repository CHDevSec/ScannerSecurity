package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/caio-henrique/secscan/application"
)

// ConsolidatedReport estrutura o relatório final com todos os scans.
type ConsolidatedReport struct {
	ProjectScans map[string]*application.ScanResult `json:"project_scans"`
}

func NewScanLocalCommand(usecase *application.LocalScanUseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scanLocal",
		Short: "Escaneia múltiplos repositórios Go em um diretório local",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, _ := cmd.Flags().GetString("dir")
			outputFile, _ := cmd.Flags().GetString("output-file")
			snykOrg, _ := cmd.Flags().GetString("snyk-org")
			if snykOrg != "" {
				viper.Set("SNYK_ORG", snykOrg)
			}
			monitor, _ := cmd.Flags().GetBool("monitor")
			// outputFormat, _ := cmd.Flags().GetString("output-format") // Para CSV no futuro

			if rootDir == "" {
				return fmt.Errorf("a flag --dir é obrigatória")
			}

			// 1. Encontrar todos os projetos Go
			projectPaths, err := usecase.FindGoProjects(rootDir)
			if err != nil {
				return err
			}
			if len(projectPaths) == 0 {
				usecase.Logger.Warn("Nenhum projeto Go (com go.mod) foi encontrado no diretório", "path", rootDir)
				return nil
			}
			usecase.Logger.Info("Projetos Go encontrados", "count", len(projectPaths))

			// 2. Escanear cada projeto e coletar resultados
			report := ConsolidatedReport{
				ProjectScans: make(map[string]*application.ScanResult),
			}

			for _, path := range projectPaths {
				result, err := usecase.ExecuteOnDir(path, monitor)
				if err != nil {
					// Loga o erro mas continua para os outros projetos
					usecase.Logger.Error("Não foi possível escanear o projeto", "path", path, "error", err)
					continue
				}
				report.ProjectScans[path] = result
			}

			// 3. Gerar o relatório consolidado
			reportJSON, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return fmt.Errorf("falha ao gerar relatório JSON consolidado: %w", err)
			}

			if outputFile != "" {
				err = os.WriteFile(outputFile, reportJSON, 0644)
				if err != nil {
					return fmt.Errorf("falha ao salvar relatório em '%s': %w", outputFile, err)
				}
				usecase.Logger.Info("Relatório consolidado salvo", "file", outputFile)
			} else {
				// Imprime no console se nenhum arquivo for especificado
				fmt.Println(string(reportJSON))
			}

			return nil
		},
	}

	cmd.Flags().StringP("dir", "d", "", "Diretório raiz contendo os projetos Go a serem escaneados (obrigatório)")
	cmd.Flags().StringP("output-file", "o", "consolidated_report.json", "Arquivo de saída para o relatório consolidado")
	// cmd.Flags().String("output-format", "json", "Formato do relatório (json, csv)")
	cmd.Flags().String("snyk-org", "", "ID/slug da organização Snyk (override de SNYK_ORG)")
	cmd.Flags().Bool("monitor", true, "Executa snyk monitor após o scan")

	return cmd
}
