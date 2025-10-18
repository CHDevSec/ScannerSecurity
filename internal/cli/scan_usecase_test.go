package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/caio-henrique/secscan/application"
	"github.com/caio-henrique/secscan/internal/mocks"
)

func TestScanUseCase_Execute(t *testing.T) {
	repoURL := "https://github.com/user/repo.git"

	t.Run("deve retornar sucesso quando nenhuma vulnerabilidade for encontrada", func(t *testing.T) {
		// Arrange
		gitClientMock := new(mocks.GitClient)
		scannerMock := new(mocks.Scanner)
		useCase := application.NewScanUseCase(scannerMock, gitClientMock)

		// Configura o mock do GitClient para simular o clone com sucesso.
		// Usamos mock.Anything porque o diretório temporário é imprevisível.
		gitClientMock.On("Clone", repoURL, mock.AnythingOfType("string")).Return(nil)

		// Simula uma resposta do Snyk sem vulnerabilidades
		snykOutput := `{"ok": true, "vulnerabilities": []}`
		scannerMock.On("Scan", mock.AnythingOfType("string")).Return([]byte(snykOutput), nil)

		// Act
		err := useCase.Execute(repoURL)

		// Assert
		assert.NoError(t, err)
		gitClientMock.AssertExpectations(t)
		scannerMock.AssertExpectations(t)
	})

	t.Run("deve retornar sucesso e processar vulnerabilidades quando encontradas", func(t *testing.T) {
		// Arrange
		gitClientMock := new(mocks.GitClient)
		scannerMock := new(mocks.Scanner)
		useCase := application.NewScanUseCase(scannerMock, gitClientMock)

		gitClientMock.On("Clone", repoURL, mock.AnythingOfType("string")).Return(nil)

		// Simula uma resposta do Snyk com vulnerabilidades
		snykOutput := `{
			"ok": false,
			"vulnerabilities": [
				{
					"title": "Cross-site Scripting (XSS)",
					"severity": "high",
					"moduleName": "react",
					"from": ["my-app@1.0.0", "react-dom@16.8.0", "react@16.8.0"]
				}
			]
		}`
		scannerMock.On("Scan", mock.AnythingOfType("string")).Return([]byte(snykOutput), nil)

		// Act
		err := useCase.Execute(repoURL)

		// Assert
		assert.NoError(t, err) // O caso de uso não retorna erro se encontrar vulnerabilidades
		gitClientMock.AssertExpectations(t)
		scannerMock.AssertExpectations(t)
	})

	t.Run("deve retornar erro se a clonagem falhar", func(t *testing.T) {
		// Arrange
		gitClientMock := new(mocks.GitClient)
		// O scanner não deve ser chamado, então não o configuramos
		scannerMock := new(mocks.Scanner)
		useCase := application.NewScanUseCase(scannerMock, gitClientMock)

		expectedError := fmt.Errorf("falha de autenticação")
		gitClientMock.On("Clone", repoURL, mock.AnythingOfType("string")).Return(expectedError)

		// Act
		err := useCase.Execute(repoURL)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "falha ao clonar repositório")
		assert.ErrorIs(t, err, expectedError)
		gitClientMock.AssertExpectations(t)
		// Garante que o scanner não foi chamado
		scannerMock.AssertNotCalled(t, "Scan", mock.Anything)
	})

	// Adicione outros casos de teste: falha no scan, JSON inválido, etc.
}
