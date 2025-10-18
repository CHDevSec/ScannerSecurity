package ports

// GitClient define a interface para operações de Git
type GitClient interface {
	// Clone clona um repositório Git para o diretório especificado
	Clone(url, path string) error
}
