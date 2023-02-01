package main

import (
	"os"

	"github.com/toddgaunt/bastion/internal/errors"
)

var defaultConfig = configServer{
	Credentials: configCredentials{
		Username: configVariable{
			Location: "env",
			Value:    "BASTION_USERNAME",
		},
		Password: configVariable{
			Location: "env",
			Value:    "BASTION_PASSWORD",
		},
	},
	Content: configContent{
		Name:         "Example",
		Description:  "This is a simple example website",
		Style:        "default",
		ScanInterval: 60,
	},
	Network: configNetwork{
		Port: 8080,
		TLS: configTLS{
			Cert:    "",
			Key:     "",
			Disable: true,
		},
	},
}

// configServer is the top-level configuration object.
type configServer struct {
	Credentials configCredentials `json:"credentials"`
	Content     configContent     `json:"content"`
	Network     configNetwork     `json:"network"`
}

type configCredentials struct {
	Username configVariable `json:"username"`
	Password configVariable `json:"password"`
}

type configVariable struct {
	Location string `json:"location"`
	Value    string `json:"value"`
}

func (v configVariable) Load() (string, error) {
	switch v.Location {
	case "", "literal":
		return v.Value, nil
	case "env":
		return os.Getenv(v.Value), nil
	}

	return "", errors.Errorf("invalid location %q", v.Location)
}

type configContent struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Style        string `json:"style"`
	ScanInterval int    `json:"scan_interval"`
}

type configNetwork struct {
	Port int       `json:"port"`
	TLS  configTLS `json:"tls"`
}

type configTLS struct {
	Cert    string `json:"cert"`
	Key     string `json:"key"`
	Disable bool   `json:"disable"`
}
