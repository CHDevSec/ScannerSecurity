package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/caio-henrique/secscan/application"
	"github.com/caio-henrique/secscan/internal/adapters"
)

func NewScanCommand(usecase *application.ScanUseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scanMonitor",
		Short: "Clona e escaneia um repositório em busca de vulnerabilidades",
		RunE: func(cmd *cobra.Command, args []string) error { // Lógica principal da CLI
			// --- 1. Coleta e validação de flags ---
			repoURLs, err := getRepoURLs(cmd)
			if err != nil {
				return err
			}

			monitor, _ := cmd.Flags().GetBool("monitor")
			depth, _ := cmd.Flags().GetInt("clone-depth")
			verbose, _ := cmd.Flags().GetBool("verbose")
			numWorkers, _ := cmd.Flags().GetInt("workers")
			outputFile, _ := cmd.Flags().GetString("output-file")
			outputFormat, _ := cmd.Flags().GetString("output-format")
			snykOrg, _ := cmd.Flags().GetString("snyk-org")
			// Se o usuário passou a org pela flag, prioriza e injeta no Viper
			if snykOrg != "" {
				viper.Set("SNYK_ORG", snykOrg)
			}

			if len(repoURLs) == 0 {
				return fmt.Errorf("nenhum repositório especificado. Use --repo ou --repo-file")
			}

			// Garante que o número de workers seja pelo menos 1
			if numWorkers < 1 {
				numWorkers = 1
			} else if numWorkers > len(repoURLs) { // Não precisa de mais workers que jobs
				numWorkers = len(repoURLs)
			}

			// Configuração do Logger
			logLevel := slog.LevelInfo
			if verbose {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

			if usecase == nil {
				return fmt.Errorf("usecase não inicializado")
			}

			logger.Info("Iniciando scans", "repos_count", len(repoURLs), "workers", numWorkers)

			// --- 2. Configuração do Worker Pool ---
			jobs := make(chan string, len(repoURLs))
			results := make(chan struct {
				Repo   string
				Result *application.ScanResult
				Err    error
			}, len(repoURLs))

			for w := 1; w <= numWorkers; w++ {
				go func(workerID int) {
					// Cada worker tem seu próprio conjunto de adapters para evitar concorrência
					gitAdapter := adapters.NewGoGitAdapter(depth)
					scanner := adapters.NewSnykCLIAdapter()
					workerUsecase := application.NewScanUseCase(scanner, gitAdapter)

					for repoURL := range jobs {
						logger.Info("Worker iniciando scan", "worker_id", workerID, "repo", repoURL)
						cfg := application.ScanConfig{RepoURL: repoURL, Monitor: monitor, Logger: logger}
						res, err := workerUsecase.Execute(cfg)
						results <- struct {
							Repo   string
							Result *application.ScanResult
							Err    error
						}{Repo: repoURL, Result: res, Err: err}
					}
				}(w)
			}

			// --- 3. Distribuição e Coleta de Resultados ---
			for _, url := range repoURLs {
				jobs <- url
			}
			close(jobs)

			var finalError error
			for i := 0; i < len(repoURLs); i++ {
				res := <-results
				if res.Err != nil {
					logger.Error("Scan falhou", "repo", res.Repo, "error", res.Err)
					finalError = fmt.Errorf("um ou mais scans falharam") // Marca que houve erro
				} else {
					logger.Info("Scan concluído com sucesso", "repo", res.Repo, "vulnerabilities", len(res.Result.Vulnerabilities))
					// Formata e exibe/salva o resultado
					err := formatAndOutput(res.Result, outputFormat, outputFile)
					if err != nil {
						logger.Error("Falha ao salvar/formatar resultado", "repo", res.Repo, "error", err)
						finalError = err
					}
				}
			}

			return finalError
		},
	}

	// Flags para a CLI
	cmd.Flags().StringP("repo", "r", "", "URL de um único repositório a ser escaneado")
	cmd.Flags().String("repo-file", "", "Caminho para um arquivo de texto com uma URL de repositório por linha")
	cmd.Flags().Bool("monitor", false, "Registra o projeto no painel Snyk após o scan (snyk monitor)")
	cmd.Flags().Int("clone-depth", 1, "Profundidade do clone Git (use 0 para clone completo)")
	cmd.Flags().BoolP("verbose", "v", false, "Habilita logs detalhados para depuração")
	cmd.Flags().IntP("workers", "w", 4, "Número de workers para processar repositórios em paralelo")
	cmd.Flags().StringP("output-file", "o", "", "Caminho do arquivo para salvar a saída (padrão: stdout)")
	cmd.Flags().String("output-format", "text", "Formato da saída (text, json)")
	cmd.Flags().String("snyk-org", "", "ID/slug da organização Snyk (override de SNYK_ORG)")

	return cmd
}

// getRepoURLs lê as URLs dos repositórios a partir das flags --repo e --repo-file.
func getRepoURLs(cmd *cobra.Command) ([]string, error) {
	urls := []string{}
	repo, _ := cmd.Flags().GetString("repo")
	if repo != "" {
		urls = append(urls, repo)
	}

	repoFile, _ := cmd.Flags().GetString("repo-file")
	if repoFile != "" {
		file, err := os.Open(repoFile)
		if err != nil {
			return nil, fmt.Errorf("falha ao abrir arquivo de repositórios '%s': %w", repoFile, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") { // Ignora linhas vazias e comentários
				urls = append(urls, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("erro ao ler arquivo de repositórios: %w", err)
		}
	}
	return urls, nil
}

// formatAndOutput é uma função placeholder para formatar e salvar os resultados.
// Em uma implementação real, ela teria lógicas para JSON, SARIF, etc.
func formatAndOutput(result *application.ScanResult, format, filePath string) error {
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
		if len(result.Vulnerabilities) > 0 {
			sb.WriteString(fmt.Sprintf("🚨 Encontradas %d vulnerabilidades\n", len(result.Vulnerabilities)))
			for _, v := range result.Vulnerabilities {
				sb.WriteString(fmt.Sprintf("  - [%s] %s em '%s'\n", v.Severity, v.Title, v.ModuleName))
			}
		} else {
			sb.WriteString("✅ Nenhuma vulnerabilidade encontrada.\n")
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
