package main

import (
	"github.com/toddgaunt/bastion/internal/content"
)

var DefaultConfig = Config{
	Content: content.Config{
		Name:         "Example",
		Description:  "This is a simple example website",
		Style:        "default",
		ScanInterval: 60,
	},
	Network: ConfigNetwork{
		Port: 8080,
		TLS: ConfigTLS{
			Cert:    "",
			Key:     "",
			Disable: true,
		},
	},
}

// ConfigNetwork contains all configuration for a Monastery website's
// networking information, such as which port to bind and the paths to the
// optional TLS certificate and key to serve HTTPS.
type ConfigNetwork struct {
	Port int       `json:"port"`
	TLS  ConfigTLS `json:"tls"`
}

type ConfigTLS struct {
	Cert    string `json:"cert"`
	Key     string `json:"key"`
	Disable bool   `json:"disable"`
}

// Config contains all website configuration options for a Monastery website
type Config struct {
	Content content.Config `json:"content"`
	Network ConfigNetwork  `json:"network"`
}
