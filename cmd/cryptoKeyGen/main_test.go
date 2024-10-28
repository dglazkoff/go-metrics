package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockKeyWriter struct {
	WrittenData map[string][]byte
}

func (m *MockKeyWriter) WriteKeyToFile(keyBytes []byte, filePath string) error {
	m.WrittenData[filePath] = keyBytes
	return nil
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
		mockWriter := &MockKeyWriter{WrittenData: make(map[string][]byte)}

		err := mainWithWriter(mockWriter)
		assert.NoError(tt, err)

		if _, ok := mockWriter.WrittenData["keys/private.pem"]; !ok {
			tt.Error("Ожидалась запись закрытого ключа, но она не была выполнена")
		}
		if _, ok := mockWriter.WrittenData["keys/public.pem"]; !ok {
			tt.Error("Ожидалась запись открытого ключа, но она не была выполнена")
		}

		if len(mockWriter.WrittenData["keys/private.pem"]) == 0 {
			tt.Error("Записанный закрытый ключ пуст")
		}
		if len(mockWriter.WrittenData["keys/public.pem"]) == 0 {
			tt.Error("Записанный открытый ключ пуст")
		}
	})
}
