package client

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/dglazkoff/go-metrics/cmd/agent/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client         *resty.Client
	retryIntervals []time.Duration
}

func NewClient(retryIntervals []time.Duration) *Client {
	var client = resty.New()
	return &Client{client: client, retryIntervals: retryIntervals}
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (c *Client) SendMetricsByHTTP(metrics []models.Metrics, cfg *config.Config) {
	c.client.SetBaseURL("http://" + cfg.RunAddr)
	body, err := json.Marshal(metrics)

	if err != nil {
		logger.Log.Debug("Error while marshal data: ", err)
		return
	}

	encryptedBody, err := EncryptBody(body, cfg)

	if err != nil {
		logger.Log.Debug("Error while encrypt body: ", err)
		c.sendBody(body, cfg)
		return
	}

	c.sendBody(encryptedBody, cfg)
}

func EncryptBody(body []byte, cfg *config.Config) ([]byte, error) {
	publicKeyPEM, err := os.ReadFile(cfg.CryptoKey)

	if err != nil {
		logger.Log.Debug("Error while read public key: ", err)
		return nil, err
	}

	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		logger.Log.Debug("Error while parse public key: ", err)
		return nil, err
	}

	var encryptedBuffer bytes.Buffer
	segmentSize := 256
	for i := 0; i < len(body); i += segmentSize {
		j := i + segmentSize
		if j > len(body) {
			j = len(body)
		}
		segmentToEncrypt := body[i:j]

		encryptedSegment, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, segmentToEncrypt)

		if err != nil {
			logger.Log.Debug("Error while encrypt data: ", err)
			return nil, err
		}

		encryptedBuffer.Write(encryptedSegment)
	}

	return encryptedBuffer.Bytes(), nil
}

func (c *Client) sendRequest(body interface{}, hash []byte, retryNumber int) {
	logger.Log.Debug("Do request to /updates/")
	request := c.client.R().SetBody(body).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Real-IP", GetLocalIP())

	if hash != nil {
		request.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}

	res, err := request.Post("/updates/")

	fmt.Println(res.Status())

	if err != nil {
		logger.Log.Debug("Error on request: ", err)

		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			if retryNumber == 3 {
				return
			}

			time.Sleep(c.retryIntervals[retryNumber])
			c.sendRequest(body, hash, retryNumber+1)
		}
	}
}

func (c *Client) sendBody(body []byte, cfg *config.Config) {
	buf := bytes.NewBuffer(body)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write(body)

	if err != nil {
		logger.Log.Debug("Error on write gzip data: ", err)
		return
	}

	err = zb.Close()

	if err != nil {
		logger.Log.Debug("Error on close gzip writer: ", err)
		return
	}

	var hash []byte
	if cfg.SecretKey != "" {
		logger.Log.Debug("Encoding body")
		h := hmac.New(sha256.New, []byte(cfg.SecretKey))
		h.Write(buf.Bytes())
		hash = h.Sum(nil)
	}

	c.sendRequest(buf, hash, 0)
}
