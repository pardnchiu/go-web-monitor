package util

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

func CheckExpire(url string) (int, error) {
	if len(url) > 4 && url[:4] != "http" {
		url = "https://" + url
	} else if len(url) <= 4 {
		url = "https://" + url
	}

	host := strings.Split(strings.TrimPrefix(url, "https://"), "/")[0]

	conn, err := tls.Dial("tcp", net.JoinHostPort(host, "443"), nil)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return 0, fmt.Errorf("failed to retrieve certificate")
	}

	expiry := certs[0].NotAfter
	daysLeft := int(time.Until(expiry).Hours() / 24)
	return daysLeft, nil
}
