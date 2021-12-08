package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"bastionburrow.com/bastion/internal/router"
)

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	var port int
	var tlsCert string
	var tlsKey string
	var defaultConfig bool

	flag.IntVar(&port, "port", 0, "Specify a port to serve and listen to")
	flag.StringVar(&tlsCert, "tls-cert", "", "Path to TLS Certificate for HTTPS")
	flag.StringVar(&tlsKey, "tls-key", "", "Path to TLS Key for HTTPS")
	flag.BoolVar(&defaultConfig, "default-config", false, "Output default configuration")

	flag.Parse()

	if defaultConfig {
		bytes, err := json.MarshalIndent(DefaultConfig, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bytes))
		os.Exit(0)
	}

	prefixDir := "."

	args := flag.Args()
	if len(args) >= 1 {
		prefixDir = path.Clean(args[0])
	}

	// Load config if present, or use default config
	data, err := ioutil.ReadFile(prefixDir + "/config.json")
	var config Config
	if err != nil {
		log.Print("using default configuration")
		config = DefaultConfig
	} else {
		err := json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalf("couldn't load config: %v", err)
		}
	}

	// Only flags that are explicitly provided from commandline can be visited
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			config.Network.Port = port
		case "tls-cert":
			config.Network.TLSCert = tlsCert
		case "tls-key":
			config.Network.TLSKey = tlsKey
		}
	})

	// Start the server
	r, err := router.New(prefixDir, config.Router)
	if err != nil {
		log.Fatalf("couldn't create router: %v", err)
	}

	addr := fmt.Sprintf(":%d", config.Network.Port)

	if config.Network.TLSCert != "" && config.Network.TLSKey != "" {
		// TLS can be used
		log.Fatal(http.ListenAndServeTLS(addr, tlsCert, tlsKey, r))
	} else {
		// Allow non-TLS for use until a certificate can be acquired
		log.Fatal(http.ListenAndServe(addr, r))
	}
}