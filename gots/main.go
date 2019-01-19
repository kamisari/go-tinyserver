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
	Name        = "gots"
	Version     = "0.3.0"
	DefaultAddr = ":8080"
)

type Config struct {
	Addr string `json:"addr"`
	Root string `json:"root"`
	File string `json:"file"`
}

func (c *Config) ReadJSON(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c)
}

func template(w io.Writer) error {
	c := &Config{
		Addr: "localhost:8080",
		Root: "/path/to/srv/root/dir/",
		File: "/path/to/srv/file",
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

const examples = `Examples:
  $ cd "$srvroot" && ` + Name + `     # serve files
  $ ` + Name + ` -addr localhost:8080 # listen only localhost
`

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
		fmt.Fprintf(usageWriter, "%s\n", examples)
	}

	var opt struct {
		help     bool
		version  bool
		template bool

		config string

		// merge to config
		root string
		file string
		addr string
	}
	flag.BoolVar(&opt.help, "help", false, "Display this message")
	flag.BoolVar(&opt.version, "version", false, "Display version")
	flag.BoolVar(&opt.template, "template", false, "Output config template to stdout")
	flag.StringVar(&opt.config, "config", "", "Specify JSON format configuration file")
	flag.StringVar(&opt.root, "root", "", "Specify serve root, defaults to current directory")
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
	case opt.template:
		if err := template(os.Stdout); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	config := &Config{
		Addr: DefaultAddr,
	}
	if opt.config != "" {
		if err := config.ReadJSON(opt.config); err != nil {
			log.Fatal(err)
		}
	}

	// merge from opt
	if opt.addr != DefaultAddr {
		config.Addr = opt.addr
	}
	if opt.root != "" {
		config.Root = opt.root
	}
	if opt.file != "" {
		config.File = opt.file
	}

	if config.Root != "" {
		fi, err := os.Stat(config.Root)
		if err != nil {
			log.Fatal(err)
		}
		if !fi.IsDir() {
			log.Fatal("not a directory: " + config.Root)
		}
	}
	if config.File != "" {
		fi, err := os.Stat(config.File)
		if err != nil {
			log.Fatal(err)
		}
		if !fi.Mode().IsRegular() {
			log.Fatal("not a regular file: " + config.File)
		}
	}

	var (
		// TODO: make? index.html by template
		rootHandler = http.RedirectHandler("/srv/", http.StatusFound)
		fileHandler = http.RedirectHandler("/", http.StatusFound)
		srvHandler  = http.StripPrefix("/srv", http.FileServer(http.Dir(config.Root)))
	)

	// TODO: consider treat case of (config.File != "" && config.Root != "")
	if config.File != "" {
		fileHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, config.File)
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

	log.Printf("Listen on %q\n", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}
