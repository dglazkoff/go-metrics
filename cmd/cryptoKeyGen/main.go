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

func writeKeyToFile(keyBytes []byte, filePath string) error {
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

func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logger.Log.Debug("Error generating private key: ", err)
	}

	var privateKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err != nil {
		logger.Log.Debug("Error encoding private key: ", err)
		return
	}

	err = writeKeyToFile(privateKeyPEM.Bytes(), "keys/private.pem")

	if err != nil {
		logger.Log.Debug("Error writing private key to file: ", err)
		return
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	if err != nil {
		logger.Log.Debug("Error encoding public key: ", err)
		return
	}

	err = writeKeyToFile(publicKeyPEM.Bytes(), "keys/public.pem")

	if err != nil {
		logger.Log.Debug("Error writing public key to file: ", err)
	}
}
