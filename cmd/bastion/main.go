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

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/content/scanner"
	"github.com/toddgaunt/bastion/internal/handlers"
	"github.com/toddgaunt/bastion/internal/log"
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

func serve(prefixDir string, config configServer) int {
	var done chan bool
	var wg = &sync.WaitGroup{}

	logger := log.New()

	dir := path.Clean(prefixDir)
	staticFileServer := http.FileServer(http.Dir(dir + "/static"))

	details := content.Details{
		Name:        config.Content.Name,
		Description: config.Content.Description,
		Style:       config.Content.Style,
	}

	store := &scanner.Scanner{
		Interval: config.Content.ScanInterval,
		Logger:   logger,
		Details:  details,
	}

	store.Start(dir+"/content", done, wg)

	username, err := config.Credentials.Username.Load()
	if err != nil {
		logger.Printf(log.Error, "username: %v", err)
		return 1
	}

	password, err := config.Credentials.Password.Load()
	if err != nil {
		logger.Printf(log.Error, "password: %v", err)
		return 1
	}

	simpleAuth, err := auth.NewSimple(username, password)
	if err != nil {
		logger.Printf(log.Error, "new simple auth: %v", err)
		return 1
	}

	env := handlers.Env{
		Store:  store,
		Logger: logger,
		Auth:   simpleAuth,
	}

	r, err := newRouter(staticFileServer, env)
	if err != nil {
		logger.Printf(log.Error, "couldn't create router: %v", err)
		return 1
	}

	addr := fmt.Sprintf(":%d", config.Network.Port)

	logger.Printf(log.Info, "Serving on port %d\n", config.Network.Port)

	if !config.Network.TLS.Disable && (config.Network.TLS.Cert != "" && config.Network.TLS.Key != "") {
		// TLS can be used
		logger.Print(log.Fatal, http.ListenAndServeTLS(addr, config.Network.TLS.Cert, config.Network.TLS.Key, r))
	} else {
		logger.Print(log.Warn, "TLS is disabled")
		// Allow non-TLS for use until a certificate can be acquired
		logger.Print(log.Fatal, http.ListenAndServe(addr, r))
	}

	// Closing this channel signals all worker threads to stop and cleanup.
	close(done)
	wg.Wait()
	return 0
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
		bytes, err := json.MarshalIndent(defaultConfig, "", "\t")
		if err != nil {
			logger.Print(log.Fatal, err)
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
		logger.Printf(log.Fatal, "couldn't load config: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		logger.Printf(log.Fatal, "couldn't load config: %v", err)
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

	// Run the server
	os.Exit(serve(prefixDir, config))
}
