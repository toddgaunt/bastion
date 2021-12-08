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

import "toddgaunt.com/monastery/internal/monastery"

var DefaultConfig = Config{
	Site: monastery.SiteConfig{
		Name:        "Monastery",
		Description: "Monastery is a simple content management server",
		Style:       "default",
		Pinned:      map[string]string{"About": "about", "Contact": "contact"},
	},
	Network: ConfigNetwork{
		Port:    8080,
		TLSCert: "",
		TLSKey:  "",
	},
	ScanInterval: 60,
}

// ConfigNetwork contains all configuration for a Monastery website's
// networking information, such as which port to bind and the paths to the
// optional TLS certificate and key to serve HTTPS.
type ConfigNetwork struct {
	Port    int    `json:"port"`
	TLSCert string `json:"tls_cert"`
	TLSKey  string `json:"tls_key"`
}

// Config contains all website configuration options for a Monastery website
type Config struct {
	Site         monastery.SiteConfig `json:"site"`
	Network      ConfigNetwork        `json:"network"`
	ScanInterval int                  `json:"scan_interval"`
}