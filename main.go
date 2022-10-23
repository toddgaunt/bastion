package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/toddgaunt/bastion/internal/router"
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

func Serve(prefixDir string, config Config) {
	// default ScanInterval
	if config.Router.ScanInterval < 1 {
		config.Router.ScanInterval = 5
	}

	r, err := router.New(prefixDir, config.Router)
	if err != nil {
		log.Fatalf("couldn't create router: %v", err)
	}

	addr := fmt.Sprintf(":%d", config.Network.Port)

	log.Printf("Serving on port %d\n", config.Network.Port)

	if !config.Network.TLS.Disable && (config.Network.TLS.Cert != "" && config.Network.TLS.Key != "") {
		// TLS can be used
		log.Fatal(http.ListenAndServeTLS(addr, config.Network.TLS.Cert, config.Network.TLS.Key, r))
	} else {
		log.Print("Warning: not using TLS")
		// Allow non-TLS for use until a certificate can be acquired
		log.Fatal(http.ListenAndServe(addr, r))
	}
}

func main() {
	var port int
	var tlsCert string
	var tlsKey string
	var exampleConfig bool
	var tlsDisable bool

	flag.IntVar(&port, "port", 0, "Specify a port to serve and listen on")
	flag.StringVar(&tlsCert, "tls-cert", "", "Path to server TLS Certificate for HTTPS")
	flag.StringVar(&tlsKey, "tls-key", "", "Path to server TLS Key for HTTPS")
	flag.BoolVar(&exampleConfig, "example-config", false, "Output an example config.json")
	flag.BoolVar(&tlsDisable, "tls-disable", false, "Disable TLS even if the config specifies it")

	flag.Parse()

	if exampleConfig {
		bytes, err := json.MarshalIndent(DefaultConfig, "", "\t")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(bytes))
		os.Exit(0)
	}

	prefixDir := "."

	args := flag.Args()
	if len(args) >= 1 {
		prefixDir = args[0]
	}

	data, err := ioutil.ReadFile(prefixDir + "/config.json")
	var config Config
	if err != nil {
		log.Fatalf("couldn't load config: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("couldn't load config: %v", err)
	}

	// Only flags that are explicitly provided from commandline can be visited
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			config.Network.Port = port
		case "tls-cert":
			config.Network.TLS.Cert = tlsCert
		case "tls-key":
			config.Network.TLS.Key = tlsKey
		case "tls-disable":
			config.Network.TLS.Disable = tlsDisable
		}
	})

	// Start the server
	Serve(prefixDir, config)
}
