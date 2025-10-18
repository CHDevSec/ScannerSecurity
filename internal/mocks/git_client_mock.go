package mocks

import (
	"github.com/stretchr/testify/mock"
)

// GitClient é um mock para a interface ports.GitClient
type GitClient struct {
	mock.Mock
}

// Clone simula a clonagem de um repositório.
func (m *GitClient) Clone(url, path string) error {
	args := m.Called(url, path)
	return args.Error(0)
}
