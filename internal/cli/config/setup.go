package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// SetupConfig inicializa o Viper para ler variáveis de ambiente e arquivo de configuração do usuário
func SetupConfig() {
	viper.AutomaticEnv()

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Erro ao encontrar o diretório HOME:", err)
		os.Exit(1)
	}

	// Compatível com repo-scanner: ~/.config/repo-scanner/config.yaml
	viper.AddConfigPath(filepath.Join(home, ".config", "repo-scanner"))
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Defaults
	viper.SetDefault("GITLAB_BASE_URL", "")
	viper.SetDefault("SNYK_TOKEN", "")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Nenhum arquivo de configuração encontrado: %v\n", err)
	}
}
