package ports

// Scanner define a interface para scanners de segurança
type Scanner interface {
	// Scan executa uma varredura de segurança no diretório especificado
	Scan(path string) ([]byte, error)

	// Monitor registra o projeto no painel de monitoramento
	Monitor(path string) ([]byte, error)
}
