package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

// GitLabAdapter implementa a interface para buscar informações de repositórios do GitLab
type GitLabAdapter struct {
	baseURL string
	token   string
	client  *http.Client
}

// GitLabProject representa um projeto do GitLab
type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	DefaultBranch     string `json:"default_branch"`
}

// NewGitLabAdapter cria uma nova instância do GitLabAdapter
func NewGitLabAdapter() *GitLabAdapter {
	base := viper.GetString("GITLAB_BASE_URL")
	if base == "" {
		base = getEnvOrDefault("GITLAB_BASE_URL", "")
	}
	token := viper.GetString("GITLAB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		token = getEnvOrDefault("GITLAB_PERSONAL_ACCESS_TOKEN", getEnvOrDefault("GITLAB_TOKEN", ""))
	}
	
	// Valida se as configurações necessárias estão presentes
	if base == "" {
		fmt.Fprintf(os.Stderr, "Aviso: GITLAB_BASE_URL não configurado. Configure com 'export GITLAB_BASE_URL=https://gitlab.example.com/api/v4'\n")
	}
	
	return &GitLabAdapter{
		baseURL: base,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProjectByID busca um projeto do GitLab pelo ID
func (g *GitLabAdapter) GetProjectByID(projectID int) (*GitLabProject, error) {
	if g.token == "" {
		return nil, fmt.Errorf("GITLAB_TOKEN não configurado")
	}

	// g.baseURL já deve incluir /api/v4 quando vindo do Viper por padrão
	url := fmt.Sprintf("%s/projects/%d", g.baseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição: %w", err)
	}

    // Para PAT do GitLab, o cabeçalho recomendado é PRIVATE-TOKEN
    req.Header.Set("PRIVATE-TOKEN", g.token)
    req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro na API do GitLab (status %d): %s", resp.StatusCode, string(body))
	}

	var project GitLabProject
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("falha ao decodificar resposta: %w", err)
	}

	return &project, nil
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
