package application

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/caio-henrique/secscan/internal/ports"
)

// SnykVulnerability define a estrutura de uma vulnerabilidade no JSON do Snyk.
type SnykVulnerability struct {
	Title      string `json:"title"`
	Severity   string `json:"severity"`
	ModuleName string `json:"moduleName"`
}

// snykOutput define a estrutura do JSON que recebemos do Snyk.
type snykTestOutput struct {
	Ok              bool                `json:"ok"`
	Vulnerabilities []SnykVulnerability `json:"vulnerabilities"`
}

// ScanResult representa o resultado processado de um scan.
type ScanResult struct {
	Vulnerabilities []SnykVulnerability
	RawJSON         []byte
	MonitorOutput   string // Saída do snyk monitor (opcional)
}

type ScanConfig struct {
	RepoURL string
	Monitor bool
	Logger  *slog.Logger
}

type ScanUseCase struct {
	snykScanner ports.Scanner
	gitClient   ports.GitClient
}

func NewScanUseCase(scanner ports.Scanner, gitClient ports.GitClient) *ScanUseCase {
	return &ScanUseCase{
		snykScanner: scanner,
		gitClient:   gitClient,
	}
}

// SetGitAdapter permite reconfigurar o git client, útil para flags como --clone-depth.
func (s *ScanUseCase) SetGitAdapter(gitClient ports.GitClient) {
	s.gitClient = gitClient
}

func (s *ScanUseCase) Execute(config ScanConfig) (*ScanResult, error) {
	logger := config.Logger
	if logger == nil {
		// Fallback para um logger padrão se nenhum for fornecido
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// 1. Clonar o repositório para um diretório temporário
	logger.Info("Passo 1: Clonando repositório", "url", config.RepoURL)
	tempBaseDir, err := os.MkdirTemp("", "secscan-*")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar diretório temporário: %w", err)
	}
	// Garante que o diretório temporário seja removido no final da execução
	defer os.RemoveAll(tempBaseDir)

	// Define um subdiretório alvo que não existe ainda para o git clone criar
	cloneDir := filepath.Join(tempBaseDir, "repo")

	if err := s.gitClient.Clone(config.RepoURL, cloneDir); err != nil {
		return nil, fmt.Errorf("falha ao clonar repositório '%s': %w", config.RepoURL, err)
	}

	logger.Info("Repositório clonado", "path", cloneDir)

	// 1.1. Verifica manifesto para Snyk
	if !hasSnykManifest(cloneDir) {
		return nil, fmt.Errorf("nenhum manifesto suportado encontrado (go.mod, package.json, requirements.txt, pom.xml, etc) em '%s'", cloneDir)
	}

	// 2. Invocar o comando do Snyk de scan através da nossa interface
	logger.Info("Passo 2: Executando Snyk Scan (snyk test)")
	jsonOutput, err := s.snykScanner.Scan(cloneDir)
	if err != nil {
		logger.Error("Falha durante o 'snyk test'", "error", err)
		return nil, fmt.Errorf("falha durante o 'snyk test': %w", err)
	}

	// 3. Processar o resultado e produzir o output
	var parsed snykTestOutput
	if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
		return nil, fmt.Errorf("falha ao processar o JSON do Snyk: %w", err)
	}

	result := &ScanResult{
		Vulnerabilities: parsed.Vulnerabilities,
		RawJSON:         jsonOutput,
	}

	logger.Info("Passo 3: Processamento concluído", "vulnerabilities", len(parsed.Vulnerabilities))

	// 4. (Opcional) Executar snyk monitor
	if config.Monitor {
		logger.Info("Passo 4: Registrando projeto no Snyk (snyk monitor)")
		monitorOutput, err := s.snykScanner.Monitor(cloneDir)
		if err != nil {
			// Trata como aviso: scan foi bem-sucedido mas registro falhou
			logger.Warn("Falha no snyk monitor, registrando aviso", "error", err)
		} else {
			result.MonitorOutput = string(monitorOutput)
			logger.Debug("Saída do monitor", "output", result.MonitorOutput)
		}
	}

	return result, nil
}

// hasSnykManifest verifica se há arquivos de manifesto que permitem o Snyk identificar o projeto
func hasSnykManifest(dir string) bool {
	manifests := []string{
		"go.mod",
		"package.json",
		"requirements.txt",
		"Pipfile",
		"pom.xml",
		"build.gradle",
		"build.gradle.kts",
		"yarn.lock",
		"pnpm-lock.yaml",
		"Cargo.toml",
	}
	for _, f := range manifests {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}
