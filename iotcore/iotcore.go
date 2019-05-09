package iotcore

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

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

func (c *DeviceConfig) publicKey() (*ecdsa.PublicKey, error) {
	priv, err := c.privateKey()
	if err != nil {
		return nil, err
	}

	return &priv.PublicKey, nil
}

func (c *DeviceConfig) privateKey() (*ecdsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(c.PrivKeyPath)
	if err != nil {
		return nil, err
	}

	return jwt.ParseECPrivateKeyFromPEM(keyBytes)
}

func (c *DeviceConfig) VerifyJWT(jwtStr string) (bool, error) {
	token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing algorithm.
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("iotcore: unexpected signing method %v", token.Header["alg"])
		}

		return c.publicKey()
	})

	if err != nil {
		return false, err
	}

	return token.Valid, err
}

func (c *DeviceConfig) NewJWT(exp time.Duration) (string, error) {
	key, err := c.privateKey()
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
