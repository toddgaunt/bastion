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
	"sync"

	"github.com/toddgaunt/bastion/internal/articles"
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
	var done chan bool
	var wg = &sync.WaitGroup{}

	dir := path.Clean(prefixDir)
	staticFileServer := http.FileServer(http.Dir(dir + "/static"))
	articles := articles.IntervalScan(dir+"/articles", config.ScanInterval, done, wg)

	r, err := router.New(staticFileServer, articles, config.Content)
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

	var config Config
	data, err := ioutil.ReadFile(prefixDir + "/config.json")
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
