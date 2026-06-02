package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	aic "github.com/mtraver/awsiotcore"
)

type DeviceConfig struct {
	Endpoint               string `json:"endpoint"`
	DeviceID               string `json:"device_id"`
	TelemetryTopicOverride string `json:"telemetry_topic"`
	// CACerts must contain the path to a .pem file containing Amazon's trusted root certs. See the README for more info.
	CACerts     string `json:"ca_certs_path"`
	CertPath    string `json:"cert_path"`
	PrivKeyPath string `json:"priv_key_path"`
}

func parseDeviceFile(path string) (*aic.Device, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config DeviceConfig
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	// Load CA certs.
	pemCerts, err := os.ReadFile(config.CACerts)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certs: %w", err)
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(pemCerts) {
		return nil, errors.New("no certs were parsed")
	}

	// Load device certificate/private key pair.
	cert, err := tls.LoadX509KeyPair(config.CertPath, config.PrivKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load device cert/key pair: %w", err)
	}

	// If the config doesn't have a device ID set then use the cert's Common Name (CN).
	if config.DeviceID == "" {
		commonName := cert.Leaf.Subject.CommonName
		if commonName == "" {
			return nil, errors.New("config has no device ID set and cert Common Name (CN) is empty")
		}

		config.DeviceID = commonName
	}

	device := &aic.Device{
		Endpoint:               config.Endpoint,
		DeviceID:               config.DeviceID,
		TelemetryTopicOverride: config.TelemetryTopicOverride,
		CACerts:                certpool,
		Cert:                   cert,
	}

	return device, nil
}
