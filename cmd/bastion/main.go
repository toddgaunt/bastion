package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/toddgaunt/bastion"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/log"
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

func serve(prefixDir string, config configServer) {
	var done chan bool
	var wg = &sync.WaitGroup{}

	logger := log.New()

	dir := path.Clean(prefixDir)
	staticFileServer := http.FileServer(http.Dir(dir + "/static"))

	details := bastion.Details{
		Name:        config.Content.Name,
		Description: config.Content.Description,
		Style:       config.Content.Style,
	}

	scanner := &content.IntervalScanner{
		ScanInterval: config.Content.ScanInterval,
		Logger:       logger,
		WithDetails:  details,
	}

	scanner.Start(dir+"/articles", done, wg)

	r, err := router.New(staticFileServer, scanner, logger)
	if err != nil {
		logger.Fatal("couldn't create router: ", err.Error())
	}

	addr := fmt.Sprintf(":%d", config.Network.Port)

	logger.Info(fmt.Sprintf("Serving on port %d\n", config.Network.Port))

	if !config.Network.TLS.Disable && (config.Network.TLS.Cert != "" && config.Network.TLS.Key != "") {
		// TLS can be used
		logger.Fatal(http.ListenAndServeTLS(addr, config.Network.TLS.Cert, config.Network.TLS.Key, r))
	} else {
		logger.Warn("TLS is disabled")
		// Allow non-TLS for use until a certificate can be acquired
		logger.Fatal(http.ListenAndServe(addr, r))
	}

	// Closing this channel signals all worker threads to stop and cleanup.
	close(done)
	wg.Wait()
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
	flag.BoolVar(&tlsDisable, "tls-disable", false, "Disable TLS even if the config specifies it")
	flag.BoolVar(&exampleConfig, "example-config", false, "Output an example config.json")

	flag.Parse()

	logger := log.New()

	if exampleConfig {
		bytes, err := json.MarshalIndent(DefaultConfig, "", "\t")
		if err != nil {
			logger.Fatal(err.Error())
		}
		fmt.Println(string(bytes))
		os.Exit(0)
	}

	prefixDir := "."

	args := flag.Args()
	if len(args) >= 1 {
		prefixDir = args[0]
	}

	var config configServer
	data, err := ioutil.ReadFile(prefixDir + "/config.json")
	if err != nil {
		logger.Fatal("couldn't load config: ", err.Error())
	}

	if err := json.Unmarshal(data, &config); err != nil {
		logger.Fatal("couldn't load config: ", err.Error())
	}

	// Only flags that are explicitly set from commandline can be visited, so
	// this will skip any that weren't provided and preserve the values from
	// the config file.
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
	serve(prefixDir, config)
}
