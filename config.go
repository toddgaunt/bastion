// Copyright 2021, Todd Gaunt <toddgaunt@protonmail.com>
//
// This file is part of Monastery.
//
// Monastery is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Monastery is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Monastery.  If not, see <https://www.gnu.org/licenses/>.

package main

import "github.com/toddgaunt/bastion/internal/router"

var DefaultConfig = Config{
	Router: router.Config{
		Name:         "Example",
		Description:  "This is a simple example website",
		Style:        "default",
		Pinned:       map[string]string{"About": "about", "Contact": "contact"},
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
	Router  router.Config `json:"router"`
	Network ConfigNetwork `json:"network"`
}
