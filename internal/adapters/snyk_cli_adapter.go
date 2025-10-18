package adapters

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
	"github.com/caio-henrique/secscan/internal/ports"
)

// SnykCLIAdapter implementa a interface Scanner usando o CLI do Snyk
type SnykCLIAdapter struct{}

// NewSnykCLIAdapter cria uma nova instância do SnykCLIAdapter
func NewSnykCLIAdapter() ports.Scanner {
	return &SnykCLIAdapter{}
}

// Scan executa o comando 'snyk test' no diretório especificado
func (s *SnykCLIAdapter) Scan(path string) ([]byte, error) {
	// Verifica se o diretório existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("diretório não encontrado: %s", path)
	}

	// Verifica token do Snyk
	token := viper.GetString("SNYK_TOKEN")
	if token == "" {
		token = os.Getenv("SNYK_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("SNYK_TOKEN não configurado. Configure seu token com 'export SNYK_TOKEN=...' antes de rodar o scan")
	}

	// Configura as variáveis de ambiente do Snyk
	env := os.Environ()
	env = append(env, "SNYK_TOKEN="+token)

	// Monta comando snyk test com --org após o subcomando
	args := []string{"test"}
	// Resolve org: Viper > Env, ignorando valores booleanos acidentais como "true"
	org := viper.GetString("SNYK_ORG")
	if org == "" || org == "true" {
		org = os.Getenv("SNYK_ORG")
	}
	if org != "" && org != "true" {
		// Use --org=<org> form to avoid argument parsing ambiguity
		args = append(args, "--org="+org)
	}
	// --json is a flag; pass it alone and then the path as positional argument
	args = append(args, "--json")
	args = append(args, path)
	// Debug information to help when org is unexpectedly 'true'
	fmt.Fprintf(os.Stderr, "DEBUG: running snyk %s (SNYK_ORG='%s')\n", strings.Join(args, " "), org)
	cmd := exec.Command("snyk", args...)
	cmd.Dir = path
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		// snyk test retorna código de saída 1 quando encontra vulnerabilidades
		// mas isso não é necessariamente um erro fatal
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Vulnerabilidades encontradas, mas o comando executou com sucesso
			return output, nil
		}
		return nil, fmt.Errorf("falha ao executar snyk test: %s", string(output))
	}

	return output, nil
}

// Monitor executa o comando 'snyk monitor' no diretório especificado
func (s *SnykCLIAdapter) Monitor(path string) ([]byte, error) {
	// Verifica se o diretório existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("diretório não encontrado: %s", path)
	}

	// Verifica token do Snyk
	token := viper.GetString("SNYK_TOKEN")
	if token == "" {
		token = os.Getenv("SNYK_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("SNYK_TOKEN não configurado. Configure seu token com 'export SNYK_TOKEN=...' antes de rodar o monitor")
	}

	// Configura as variáveis de ambiente do Snyk
	env := os.Environ()
	env = append(env, "SNYK_TOKEN="+token)

	// Monta comando snyk monitor com --org após o subcomando
	args := []string{"monitor"}
	orgVal := viper.GetString("SNYK_ORG")
	if orgVal == "" || orgVal == "true" {
		orgVal = os.Getenv("SNYK_ORG")
	}
	if orgVal != "" && orgVal != "true" {
		// Use --org=<org> form to avoid argument parsing ambiguity
		args = append(args, "--org="+orgVal)
	}
	args = append(args, path)
	// Debug info
	fmt.Fprintf(os.Stderr, "DEBUG: running snyk %s (SNYK_ORG='%s')\n", strings.Join(args, " "), orgVal)
	cmd := exec.Command("snyk", args...)
	cmd.Dir = path
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("falha ao executar snyk monitor: %s", string(output))
	}

	return output, nil
}
