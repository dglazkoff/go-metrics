package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/dglazkoff/go-metrics/internal/logger"
)

type KeyWriter interface {
	WriteKeyToFile(keyBytes []byte, filePath string) error
}

type FileKeyWriter struct{}

func (w *FileKeyWriter) WriteKeyToFile(keyBytes []byte, filePath string) error {
	filePrivate, err := os.Create(filePath)

	if err != nil {
		logger.Log.Debug("Error creating file: ", err)
		return err
	}

	defer filePrivate.Close()

	_, err = filePrivate.Write(keyBytes)
	if err != nil {
		logger.Log.Debug("Error writing file: ", err)
		return err
	}

	return nil
}

func mainWithWriter(writer KeyWriter) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logger.Log.Debug("Error generating private key: ", err)
		return err
	}

	var privateKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err != nil {
		logger.Log.Debug("Error encoding private key: ", err)
		return err
	}

	err = writer.WriteKeyToFile(privateKeyPEM.Bytes(), "keys/private.pem")

	if err != nil {
		logger.Log.Debug("Error writing private key to file: ", err)
		return err
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	if err != nil {
		logger.Log.Debug("Error encoding public key: ", err)
		return err
	}

	err = writer.WriteKeyToFile(publicKeyPEM.Bytes(), "keys/public.pem")

	if err != nil {
		logger.Log.Debug("Error writing public key to file: ", err)
		return err
	}

	return nil
}

func runApp() error {
	err := logger.Initialize()
	if err != nil {
		logger.Log.Debug("Error initializing logger: ", err)
		return err
	}

	writer := &FileKeyWriter{}
	err = mainWithWriter(writer)
	if err != nil {
		logger.Log.Debug("Error in mainWithWriter: ", err)
		return err
	}

	return nil
}

func main() {
	if err := runApp(); err != nil {
		panic(err)
	}
}
