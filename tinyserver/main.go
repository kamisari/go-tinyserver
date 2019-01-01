package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	Name        = "tinyserver"
	Version     = "0.2.0"
	DefaultAddr = ":8080"
)

func main() {
	var usageWriter io.Writer = os.Stderr
	flag.Usage = func() {
		// is need?
		tmpw := flag.CommandLine.Output()
		defer flag.CommandLine.SetOutput(tmpw)

		flag.CommandLine.SetOutput(usageWriter)
		fmt.Fprintf(usageWriter, "Usage:\n")
		fmt.Fprintf(usageWriter, "  %s [Options]\n\n", Name)
		fmt.Fprintf(usageWriter, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(usageWriter, "\n")
		fmt.Fprintf(usageWriter, "Examples:\n")
		fmt.Fprintf(usageWriter, "  $ cd \"$srvroot\" && %s     # srve files\n", Name)
		fmt.Fprintf(usageWriter, "  $ %s -addr localhost:8080 # listen only localhost\n", Name)
		fmt.Fprintf(usageWriter, "\n")
	}

	var opt struct {
		help    bool
		version bool

		config string

		// marge to config, priority (opt > config)
		root string
		file string
		addr string
	}
	flag.BoolVar(&opt.help, "help", false, "Display this message")
	flag.BoolVar(&opt.version, "version", false, "Display version")
	flag.StringVar(&opt.config, "config", "", "Specify JSON format configuration file")
	flag.StringVar(&opt.root, "root", "", "Specify serve root")
	flag.StringVar(&opt.file, "file", "", "Specify serve file")
	flag.StringVar(&opt.addr, "addr", DefaultAddr, "Specify listen address")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		log.Fatal("invalid arguments", flag.Args())
	}

	// with exit
	switch {
	case opt.help:
		usageWriter = os.Stdout
		flag.Usage()
		os.Exit(0)
	case opt.version:
		fmt.Printf("%s %s\n", Name, Version)
		os.Exit(0)
	}

	// TODO: define type?
	// for read json, priority (config < opt)
	config := &struct {
		File string
		Root string
		Addr string
	}{
		Addr: DefaultAddr,
	}
	readJSON := func(file string) error {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, config)
	}

	if opt.config != "" {
		if err := readJSON(opt.config); err != nil {
			log.Fatal(err)
		}
	}
	if opt.root != "" {
		config.Root = opt.root
	}
	if opt.addr != DefaultAddr {
		config.Addr = opt.addr
	}

	var (
		// TODO: make? index.html by template
		rootHandler = http.RedirectHandler("/srv/", http.StatusFound)
		fileHandler = http.RedirectHandler("/", http.StatusFound)
		srvHandler  = http.StripPrefix("/srv", http.FileServer(http.Dir(config.Root)))
	)

	// TODO: consider treat case of (opt.file != "" && config.Root != "")
	if opt.file != "" {
		// to os.Lstat?
		f, err := os.Stat(opt.file)
		if err != nil {
			log.Fatal(err)
		}
		if !f.Mode().IsRegular() {
			log.Fatal("not regular file")
		}
		fileHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, opt.file)
		})
		if config.Root == "" {
			srvHandler = fileHandler
		}
	}

	withLog := func(h http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				log.Println(r.Host, r.RemoteAddr, r.RequestURI)
				h.ServeHTTP(w, r)
			})
	}

	http.Handle("/", withLog(rootHandler))
	http.Handle("/srv/", withLog(srvHandler))
	http.Handle("/file", withLog(fileHandler))

	log.Printf("Listen on %s\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}
