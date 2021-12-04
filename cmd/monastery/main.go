// Copyright 2020, Todd Gaunt <toddgaunt@protonmail.com>
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

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi"
	"toddgaunt.com/monastery/internal/monastery"
)

func main() {
	var port int
	var tlsCert string
	var tlsKey string
	var defaultConfig bool

	flag.IntVar(&port, "port", 8080, "Specify a port to serve and listen to")
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

	config.Network.Port = port
	config.Network.TLSCert = tlsCert
	config.Network.TLSKey = tlsKey

	indexTemplate, err := template.ParseFiles(prefixDir + "/templates/index.html")
	if err != nil {
		log.Fatalf("couldn't load index template: %v", err)
	}
	articleTemplate, err := template.ParseFiles(prefixDir + "/templates/article.html")
	if err != nil {
		log.Fatalf("couldn't load article template: %v", err)
	}
	problemTemplate, err := template.ParseFiles(prefixDir + "/templates/problem.html")
	if err != nil {
		log.Fatalf("couldn't load problem template: %v", err)
	}

	staticFileServer := http.FileServer(http.Dir(prefixDir + "/static"))

	content := monastery.ScanContent(prefixDir+"/content", config.ScanInterval)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", monastery.GetIndex(indexTemplate, config.Site, content))
		r.With(monastery.ArticlesCtx).Get("/*", monastery.GetArticle(articleTemplate, config.Site, content))
	})

	r.Route("/"+monastery.ProblemPath, func(r chi.Router) {
		r.Route("/{problemID}", func(r chi.Router) {
			r.Use(monastery.ProblemsCtx)
			r.Get("/", monastery.GetProblem(problemTemplate, config.Site))
		})
	})

	r.Handle("/.static/*", http.StripPrefix("/.static/", staticFileServer))

	addr := fmt.Sprintf(":%d", config.Network.Port)

	if config.Network.TLSCert != "" && config.Network.TLSKey != "" {
		// TLS can be used
		log.Fatal(http.ListenAndServeTLS(addr, tlsCert, tlsKey, r))
	} else {
		// Allow non-TLS for use until a certificate can be acquired
		log.Fatal(http.ListenAndServe(addr, r))
	}
}
