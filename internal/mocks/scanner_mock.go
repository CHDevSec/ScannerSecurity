package mocks

import (
	"github.com/stretchr/testify/mock"
)

// Scanner é um mock para a interface ports.Scanner
type Scanner struct {
	mock.Mock
}

// Scan simula a execução do scanner.
func (m *Scanner) Scan(path string) ([]byte, error) {
	args := m.Called(path)
	// O primeiro valor de retorno pode ser nil se o segundo (erro) não for.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Caso contrário, fazemos a asserção de tipo.
	return args.Get(0).([]byte), args.Error(1)
}

// Monitor simula a execução do monitor.
func (m *Scanner) Monitor(path string) ([]byte, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
