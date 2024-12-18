package main

import (
	"errors"
	"os"
	"testing"

	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKeyWriter struct {
	mock.Mock
}

func (m *MockKeyWriter) WriteKeyToFile(keyBytes []byte, filePath string) error {
	args := m.Called(keyBytes, filePath)
	return args.Error(0)
}

func TestFileKeyWriter(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
		writer := &FileKeyWriter{}

		tempFilePath := "test_key.pem"
		defer os.Remove(tempFilePath) // Удаляем файл после теста

		keyBytes := []byte("test-key-data")
		err := writer.WriteKeyToFile(keyBytes, tempFilePath)

		assert.NoError(tt, err)

		fileContent, err := os.ReadFile(tempFilePath)
		assert.NoError(tt, err)

		assert.Equal(tt, keyBytes, fileContent)
	})
}

func TestMainWithWriter(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
		mockWriter := new(MockKeyWriter)

		mockWriter.On("WriteKeyToFile", mock.Anything, "keys/private.pem").Return(nil)
		mockWriter.On("WriteKeyToFile", mock.Anything, "keys/public.pem").Return(nil)

		err := mainWithWriter(mockWriter)
		assert.NoError(tt, err, "Expected no error during successful execution")

		mockWriter.AssertCalled(tt, "WriteKeyToFile", mock.Anything, "keys/private.pem")
		mockWriter.AssertCalled(tt, "WriteKeyToFile", mock.Anything, "keys/public.pem")

		mockWriter.AssertNumberOfCalls(tt, "WriteKeyToFile", 2)
	})
}

func TestWriteKeyToFile_FileCreationError(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	writer := &FileKeyWriter{}
	keyBytes := []byte("test-key-content")

	err = writer.WriteKeyToFile(keyBytes, "/invalid/path/to/file.txt")
	assert.Error(t, err, "Expected an error when providing an invalid file path")
}

func TestMainWithWriter_ErrorOnPrivateKeyWrite(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	mockWriter := new(MockKeyWriter)

	mockWriter.On("WriteKeyToFile", mock.Anything, "keys/private.pem").Return(nil)
	mockWriter.On("WriteKeyToFile", mock.Anything, "keys/public.pem").Return(errors.New("mock write error"))

	err = mainWithWriter(mockWriter)
	assert.Error(t, err, "Expected error when writing the public key to the file fails")
}

func TestMainWithWriter_ErrorOnPublicKeyWrite(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	mockWriter := new(MockKeyWriter)

	mockWriter.On("WriteKeyToFile", mock.Anything, "keys/private.pem").Return(errors.New("mock private key write error"))

	err = mainWithWriter(mockWriter)
	assert.Error(t, err, "Expected error when writing the private key to the file fails")
}
