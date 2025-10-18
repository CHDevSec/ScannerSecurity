package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/caio-henrique/secscan/application"
	"github.com/caio-henrique/secscan/internal/adapters"
)

// GitLabScanResult representa o resultado de um scan do GitLab
type GitLabScanResult struct {
	ProjectID     int                     `json:"project_id"`
	ProjectName   string                  `json:"project_name"`
	ProjectURL    string                  `json:"project_url"`
	ScanResult    *application.ScanResult `json:"scan_result"`
	MonitorResult string                  `json:"monitor_result,omitempty"`
}

func NewScanGitLabCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scanGitLab",
		Short: "Escaneia um projeto GitLab por ID usando Snyk Monitor",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Coleta e validação de flags
			projectIDStr, _ := cmd.Flags().GetString("project-id")
			monitor, _ := cmd.Flags().GetBool("monitor")
			verbose, _ := cmd.Flags().GetBool("verbose")
			outputFile, _ := cmd.Flags().GetString("output-file")
			outputFormat, _ := cmd.Flags().GetString("output-format")
			snykOrg, _ := cmd.Flags().GetString("snyk-org")
			if snykOrg != "" {
				viper.Set("SNYK_ORG", snykOrg)
			}

			if projectIDStr == "" {
				return fmt.Errorf("--project-id é obrigatório")
			}

			projectID, err := strconv.Atoi(projectIDStr)
			if err != nil {
				return fmt.Errorf("project-id deve ser um número válido: %w", err)
			}

			// Configuração do Logger
			logLevel := slog.LevelInfo
			if verbose {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

			// Inicializa os adapters
			gitlabAdapter := adapters.NewGitLabAdapter()
			gitAdapter := adapters.NewGoGitAdapter(1) // Clone superficial por padrão
			snykAdapter := adapters.NewSnykCLIAdapter()

			// 1. Busca informações do projeto no GitLab
			logger.Info("Buscando projeto no GitLab", "project_id", projectID)
			project, err := gitlabAdapter.GetProjectByID(projectID)
			if err != nil {
				return fmt.Errorf("falha ao buscar projeto no GitLab: %w", err)
			}

			logger.Info("Projeto encontrado",
				"name", project.Name,
				"path", project.PathWithNamespace,
				"url", project.HTTPURLToRepo)

			// 2. Clona o repositório
			logger.Info("Clonando repositório", "url", project.HTTPURLToRepo)
			tempBaseDir, err := os.MkdirTemp("", "secscan-gitlab-*")
			if err != nil {
				return fmt.Errorf("falha ao criar diretório temporário: %w", err)
			}
			defer os.RemoveAll(tempBaseDir)

			// Clona em um subdiretório que ainda não existe
			cloneDir := filepath.Join(tempBaseDir, "repo")

			if err := gitAdapter.Clone(project.HTTPURLToRepo, cloneDir); err != nil {
				return fmt.Errorf("falha ao clonar repositório: %w", err)
			}

			logger.Info("Repositório clonado", "path", cloneDir)

			// 3. Executa scan do Snyk
			logger.Info("Executando Snyk Scan")
			jsonOutput, err := snykAdapter.Scan(cloneDir)
			if err != nil {
				logger.Error("Falha no snyk test", "error", err)
				return fmt.Errorf("falha durante o 'snyk test': %w", err)
			}

			// 4. Processa resultado do scan
			var parsed struct {
				Ok              bool `json:"ok"`
				Vulnerabilities []struct {
					Title      string `json:"title"`
					Severity   string `json:"severity"`
					ModuleName string `json:"moduleName"`
				} `json:"vulnerabilities"`
			}

			if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
				return fmt.Errorf("falha ao processar JSON do Snyk: %w", err)
			}

			scanResult := &application.ScanResult{
				Vulnerabilities: make([]application.SnykVulnerability, len(parsed.Vulnerabilities)),
				RawJSON:         jsonOutput,
			}

			for i, v := range parsed.Vulnerabilities {
				scanResult.Vulnerabilities[i] = application.SnykVulnerability{
					Title:      v.Title,
					Severity:   v.Severity,
					ModuleName: v.ModuleName,
				}
			}

			logger.Info("Scan concluído", "vulnerabilities", len(scanResult.Vulnerabilities))

			// 5. Executa snyk monitor se solicitado
			var monitorResult string
			if monitor {
				logger.Info("Executando Snyk Monitor")
				monitorOutput, err := snykAdapter.Monitor(cloneDir)
				if err != nil {
					logger.Warn("Falha no snyk monitor", "error", err)
				} else {
					monitorResult = string(monitorOutput)
					logger.Info("Snyk Monitor executado com sucesso")
				}
			}

			// 6. Prepara resultado final
			result := &GitLabScanResult{
				ProjectID:     project.ID,
				ProjectName:   project.Name,
				ProjectURL:    project.WebURL,
				ScanResult:    scanResult,
				MonitorResult: monitorResult,
			}

			// 7. Formata e salva resultado
			err = formatAndOutputGitLab(result, outputFormat, outputFile)
			if err != nil {
				return fmt.Errorf("falha ao formatar resultado: %w", err)
			}

			logger.Info("Processo concluído com sucesso")
			return nil
		},
	}

	// Flags para a CLI
	cmd.Flags().StringP("project-id", "p", "", "ID do projeto GitLab a ser escaneado (obrigatório)")
	cmd.Flags().Bool("monitor", true, "Executa snyk monitor após o scan (padrão: true)")
	cmd.Flags().BoolP("verbose", "v", false, "Habilita logs detalhados para depuração")
	cmd.Flags().StringP("output-file", "o", "", "Caminho do arquivo para salvar a saída (padrão: stdout)")
	cmd.Flags().String("output-format", "json", "Formato da saída (text, json)")
	cmd.Flags().String("snyk-org", "", "ID/slug da organização Snyk (override de SNYK_ORG)")

	return cmd
}

// formatAndOutputGitLab formata e salva os resultados do GitLab
func formatAndOutputGitLab(result *GitLabScanResult, format, filePath string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("falha ao formatar saída para JSON: %w", err)
		}
	case "text":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("🔍 Projeto GitLab: %s (ID: %d)\n", result.ProjectName, result.ProjectID))
		sb.WriteString(fmt.Sprintf("📁 URL: %s\n", result.ProjectURL))
		sb.WriteString(fmt.Sprintf("🔒 Vulnerabilidades encontradas: %d\n", len(result.ScanResult.Vulnerabilities)))

		if len(result.ScanResult.Vulnerabilities) > 0 {
			sb.WriteString("\n📋 Detalhes das vulnerabilidades:\n")
			for _, v := range result.ScanResult.Vulnerabilities {
				sb.WriteString(fmt.Sprintf("  - [%s] %s em '%s'\n", v.Severity, v.Title, v.ModuleName))
			}
		} else {
			sb.WriteString("✅ Nenhuma vulnerabilidade encontrada.\n")
		}

		if result.MonitorResult != "" {
			sb.WriteString("\n📊 Snyk Monitor executado com sucesso\n")
		}

		output = []byte(sb.String())
	default:
		return fmt.Errorf("formato de saída desconhecido: %s", format)
	}

	if filePath != "" {
		return os.WriteFile(filePath, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}
