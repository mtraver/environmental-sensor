package iotcore

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const certExtension = ".x509"

func NewMQTTOptions(config DeviceConfig, bridge MQTTBridge, caCertsPath string) (*MQTT.ClientOptions, error) {
	// Load CA certs
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(caCertsPath)
	if err != nil {
		return nil, fmt.Errorf("iotcore: failed to read CA certs: %v", err)
	}
	certpool.AppendCertsFromPEM(pemCerts)

	tlsConf := &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{},
		MinVersion:         tls.VersionTLS12,
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(bridge.URL())
	opts.SetClientID(config.ClientID())
	opts.SetTLSConfig(tlsConf)
	opts.SetUsername("unused")

	return opts, nil
}

func DeviceIDFromCert(certPath string) (string, error) {
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("iotcore: cert file does not exist: %v", certPath)
		}

		return "", fmt.Errorf("iotcore: failed to read cert: %v", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("iotcore: failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	return cert.Subject.CommonName, nil
}

type DeviceConfig struct {
	ProjectID   string
	RegistryID  string
	DeviceID    string
	PrivKeyPath string
	Region      string
}

func (c *DeviceConfig) ClientID() string {
	return fmt.Sprintf("projects/%v/locations/%v/registries/%v/devices/%v",
		c.ProjectID, c.Region, c.RegistryID, c.DeviceID)
}

func (c *DeviceConfig) ConfigTopic() string {
	return fmt.Sprintf("/devices/%v/config", c.DeviceID)
}

func (c *DeviceConfig) TelemetryTopic() string {
	return fmt.Sprintf("/devices/%v/events", c.DeviceID)
}

func (c *DeviceConfig) StateTopic() string {
	return fmt.Sprintf("/devices/%v/state", c.DeviceID)
}

func (c *DeviceConfig) CertPath() string {
	ext := path.Ext(c.PrivKeyPath)
	return c.PrivKeyPath[:len(c.PrivKeyPath)-len(ext)] + certExtension
}

func (c *DeviceConfig) NewJWT(keyBytes []byte, exp time.Duration) (string, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return "", fmt.Errorf("iotcore: failed to parse priv key: %v", err)
	}

	token := jwt.New(jwt.SigningMethodES256)
	token.Claims = jwt.StandardClaims{
		Audience:  c.ProjectID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(exp).Unix(),
	}

	return token.SignedString(key)
}

type MQTTBridge struct {
	Host string
	Port int
}

func (b *MQTTBridge) URL() string {
	return fmt.Sprintf("ssl://%v:%v", b.Host, b.Port)
}
