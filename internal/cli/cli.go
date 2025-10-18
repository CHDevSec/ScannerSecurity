package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/caio-henrique/secscan/application"
	"github.com/caio-henrique/secscan/internal/adapters"
	"github.com/caio-henrique/secscan/internal/cli/config"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "secScan",
	Short: "secScan é uma CLI para executar análises de segurança com Snyk",
}

func Execute() {
	// --- Carrega variáveis de ambiente primeiro (para permitir Viper lê-las) ---
	if err := godotenv.Load(); err != nil {
		// Não é um erro fatal se o arquivo .env não existir
		fmt.Println("Aviso: arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// --- Configuração (Viper) ---
	config.SetupConfig()

	// --- Dependências Comuns ---
	snykScanner := adapters.NewSnykCLIAdapter()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// --- Injeção para `scanMonitor` (remoto) ---
	gitClient := adapters.NewGoGitAdapter(1) // Padrão de clone superficial
	scanUseCase := application.NewScanUseCase(snykScanner, gitClient)
	rootCmd.AddCommand(NewScanCommand(scanUseCase))

	// --- Injeção para `scanLocal` (local) ---
	localScanUseCase := application.NewLocalScanUseCase(snykScanner, logger)
	rootCmd.AddCommand(NewScanLocalCommand(localScanUseCase))

	// --- Injeção para `scanGitLab` (GitLab ID) ---
	rootCmd.AddCommand(NewScanGitLabCommand())

	// --- Execução ---
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ERRO: %v\n", err)
		os.Exit(1)
	}
}
