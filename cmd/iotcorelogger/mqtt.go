package main

import (
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	aic "github.com/mtraver/awsiotcore"
)

func fileStoreOpt(dir string) func(*aic.Device, *mqtt.ClientOptions) error {
	return func(device *aic.Device, opts *mqtt.ClientOptions) error {
		opts.SetStore(mqtt.NewFileStore(dir))
		return nil
	}
}

func onConnectOpt(handler func(client mqtt.Client)) func(*aic.Device, *mqtt.ClientOptions) error {
	return func(device *aic.Device, opts *mqtt.ClientOptions) error {
		opts.SetOnConnectHandler(handler)
		return nil
	}
}

func onConnectionLostOpt(handler func(client mqtt.Client, err error)) func(*aic.Device, *mqtt.ClientOptions) error {
	return func(device *aic.Device, opts *mqtt.ClientOptions) error {
		opts.SetConnectionLostHandler(handler)
		return nil
	}
}

func connectTimeoutOpt(timeout time.Duration) func(*aic.Device, *mqtt.ClientOptions) error {
	return func(device *aic.Device, opts *mqtt.ClientOptions) error {
		opts.SetConnectTimeout(timeout)
		return nil
	}
}
