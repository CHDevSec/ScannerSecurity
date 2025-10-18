package adapters

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
	"github.com/caio-henrique/secscan/internal/ports"
)

// GoGitAdapter implementa a interface GitClient usando comandos Git nativos
type GoGitAdapter struct {
	cloneDepth int
}

// NewGoGitAdapter cria uma nova instância do GoGitAdapter
func NewGoGitAdapter(cloneDepth int) ports.GitClient {
	return &GoGitAdapter{
		cloneDepth: cloneDepth,
	}
}

// Clone clona um repositório Git para o diretório especificado
func (g *GoGitAdapter) Clone(repoURL, path string) error {
	// Injeta credenciais HTTP se existirem no Viper (para repositórios privados via HTTPS)
	// GIT_USERNAME + GIT_TOKEN (ou GIT_PASSWORD)
	rawURL := repoURL
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
        // Aviso se não houver credenciais configuradas
        if viper.GetString("GIT_USERNAME") == "" || (viper.GetString("GIT_TOKEN") == "" && viper.GetString("GIT_PASSWORD") == "") {
            fmt.Fprintln(os.Stderr, "[aviso] Nenhum token/username encontrado; o clone pode falhar por permissão.")
        }
		if u, err := url.Parse(rawURL); err == nil {
			user := viper.GetString("GIT_USERNAME")
			pass := viper.GetString("GIT_TOKEN")
			if pass == "" {
				pass = viper.GetString("GIT_PASSWORD")
			}
			if user != "" && pass != "" {
				u.User = url.UserPassword(user, pass)
				rawURL = u.String()
			}
		}
	}
	// Constrói o comando git clone
	args := []string{"clone"}

	// Adiciona --depth se especificado
	if g.cloneDepth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", g.cloneDepth))
	}

	// Adiciona URL e diretório de destino
	args = append(args, rawURL, path)

	// Executa o comando git clone
    cmd := exec.Command("git", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        outStr := string(output)
        if strings.Contains(outStr, "Authentication failed") || strings.Contains(outStr, "Permission denied") {
            return fmt.Errorf("falha de autenticação no Git: %s", outStr)
        }
        return fmt.Errorf("falha ao clonar repositório '%s': %s", repoURL, outStr)
	}

	return nil
}
