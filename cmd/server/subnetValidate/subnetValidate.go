package subnetvalidate

import (
	"net"
	"net/http"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/internal/logger"
)

type Subnet struct {
	cfg *config.Config
}

func Initialize(cfg *config.Config) *Subnet {
	return &Subnet{cfg}
}

func (subnet *Subnet) Validate(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if subnet.cfg.TrustedSubnet == "" {
			handler.ServeHTTP(writer, request)
			return
		}

		ipHeader := request.Header.Get("X-Real-IP")

		if ipHeader == "" {
			logger.Log.Debug("No IP address in request")
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		_, ipv4Net, _ := net.ParseCIDR(subnet.cfg.TrustedSubnet)
		ipv4 := net.ParseIP(ipHeader)

		if ipv4 == nil || !ipv4Net.Contains(ipv4) {
			logger.Log.Debug("IP address is not in trusted subnet")
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		logger.Log.Debug("IP address is in trusted subnet")

		handler.ServeHTTP(writer, request)
	}
}
