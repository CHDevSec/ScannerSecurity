package application

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/caio-henrique/secscan/internal/ports"
)

// LocalScanUseCase é responsável por orquestrar scans em diretórios locais.
type LocalScanUseCase struct {
	snykScanner ports.Scanner
	Logger      *slog.Logger
}

func NewLocalScanUseCase(scanner ports.Scanner, logger *slog.Logger) *LocalScanUseCase {
	return &LocalScanUseCase{
		snykScanner: scanner,
		Logger:      logger,
	}
}

// ExecuteOnDir escaneia um único diretório e, opcionalmente, executa snyk monitor.
func (s *LocalScanUseCase) ExecuteOnDir(path string, monitor bool) (*ScanResult, error) {
	s.Logger.Info("Executando snyk test", "path", path)
	jsonOut, err := s.snykScanner.Scan(path)
	if err != nil {
		// Se o binário não tiver o comando 'test', seguimos para monitor se solicitado
		if monitor && strings.Contains(err.Error(), "Unknown command \"test\"") {
			s.Logger.Warn("CLI do Snyk não suporta 'test'; seguindo apenas com monitor", "path", path)
		} else {
			s.Logger.Error("Falha no snyk test", "path", path, "error", err)
			return nil, fmt.Errorf("falha durante o 'snyk test' em '%s': %w", path, err)
		}
	}

	result := &ScanResult{Vulnerabilities: nil, RawJSON: nil}
	if err == nil {
		var parsed snykTestOutput
		if err := json.Unmarshal(jsonOut, &parsed); err != nil {
			return nil, fmt.Errorf("falha ao processar JSON do Snyk para '%s': %w", path, err)
		}
		result.Vulnerabilities = parsed.Vulnerabilities
		result.RawJSON = jsonOut
		s.Logger.Info("Scan do diretório concluído", "path", path, "vulnerabilities", len(result.Vulnerabilities))
	}

	if monitor {
		s.Logger.Info("Executando snyk monitor", "path", path)
		monOut, err := s.snykScanner.Monitor(path)
		if err != nil {
			s.Logger.Warn("Falha no snyk monitor", "path", path, "error", err)
		} else {
			result.MonitorOutput = string(monOut)
			s.Logger.Debug("Monitor output registrado")
		}
	}
	return result, nil
}

// FindGoProjects percorre um diretório raiz e encontra todos os subdiretórios que contêm um `go.mod`.
func (s *LocalScanUseCase) FindGoProjects(rootDir string) ([]string, error) {
	var projects []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Verifica se o arquivo é um go.mod
		if !info.IsDir() && info.Name() == "go.mod" {
			// Adiciona o diretório que contém o go.mod
			projectDir := filepath.Dir(path)
			s.Logger.Debug("Projeto Go encontrado", "path", projectDir)
			projects = append(projects, projectDir)
			// Impede que o Walk entre mais fundo neste diretório, já que encontramos o projeto.
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao procurar por projetos Go em '%s': %w", rootDir, err)
	}
	return projects, nil
}
