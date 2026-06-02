package awscerts

import (
	"crypto/x509"
	_ "embed"
)

//go:embed amazon_root_certs.pem
var pemBytes []byte

var CertPool = x509.NewCertPool()

func init() {
	if !CertPool.AppendCertsFromPEM(pemBytes) {
		panic("failed to parse AWS root CA certs")
	}
}
