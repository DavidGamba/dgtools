package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/DavidGamba/go-getoptions"
)

var logger = log.New(ioutil.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	var port int
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("debug", false, opt.GetEnv("DEBUG"))
	opt.IntVar(&port, "port", 8080)
	opt.HelpSynopsisArgs("<dir>")
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("debug") {
		logger.SetOutput(os.Stderr)
	}
	if len(remaining) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: missing dir to serve!\n")
		fmt.Println(opt.Help())
		return 1
	}
	dir := remaining[0]

	err = realMain(port, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func realMain(port int, dir string) error {
	logger.Printf("Serve: %s contents on :%d\n", dir, port)
	fs := http.FileServer(http.Dir(dir))
	err := http.ListenAndServe(":"+strconv.Itoa(port), http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// TODO: Add flag to control this
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	}))
	if err != nil {
		return err
	}
	return nil
}
