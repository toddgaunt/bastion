package main

var DefaultConfig = configServer{
	//Credentials: ConfigCredentials{},
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
	Content configContent `json:"content"`
	Network configNetwork `json:"network"`
}

type configCredentials struct {
	Username configVariable `json:"username"`
	Password configVariable `json:"password"`
}

type configVariable struct {
	Location string `json:"location"`
	Value    string `json:"value"`
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
