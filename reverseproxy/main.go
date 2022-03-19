package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	"github.com/gorilla/mux"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

func program(args []string) int {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("?"))
	opt.Bool("quiet", false)
	opt.Int("port", 9090)
	opt.StringSlice("target", 1, 1, opt.Required())
	opt.StringSlice("base-path", 1, 1, opt.Required())
	opt.String("cert", "")
	opt.String("key", "")
	remaining, err := opt.Parse(args[1:])
	if opt.Called("help") {
		fmt.Println(opt.Help())
		return 1
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("quiet") {
		Logger.SetOutput(io.Discard)
	}

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = run(ctx, opt, remaining)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	port := opt.Value("port").(int)
	basePaths := opt.Value("base-path").([]string)
	targets := opt.Value("target").([]string)
	cert := opt.Value("cert").(string)
	key := opt.Value("key").(string)

	if cert != "" && key != "" {
		if _, err := os.Stat(cert); os.IsNotExist(err) {
			return fmt.Errorf("not found '%s': %w", cert, err)
		}
		if _, err := os.Stat(key); os.IsNotExist(err) {
			return fmt.Errorf("not found '%s': %w", cert, err)
		}
	}

	if len(targets) != len(basePaths) {
		return fmt.Errorf("targets and basepaths don't match")
	}

	Logger.Printf("Serve: on :%d\n", port)

	proto := "http"
	if cert != "" && key != "" {
		proto = "https"
	}

	r := mux.NewRouter()

	for i, target := range targets {
		turl, err := url.Parse(target)
		if err != nil {
			return err
		}
		Logger.Printf("Target URL: host %s, path %s, scheme %s, %s", turl.Host, turl.Path, turl.Scheme, turl.String())
		basePath := basePaths[i]

		rp := &RP{
			targetURL: turl,
			basePath:  basePath,
			proto:     proto,
		}
		r.PathPrefix(basePath).Handler(rp)
	}

	if cert != "" && key != "" {
		err := http.ListenAndServeTLS(":"+strconv.Itoa(port), cert, key, r)
		if err != nil {
			return err
		}
	} else {
		err := http.ListenAndServe(":"+strconv.Itoa(port), r)
		if err != nil {
			return err
		}
	}
	return nil
}

type RP struct {
	target    string
	targetURL *url.URL
	basePath  string
	proto     string
}

func (rp *RP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			reqHost := req.Host
			reqURL := req.URL.String()
			reqURLPath := req.URL.Path

			req.URL.Host = rp.targetURL.Host
			req.URL.Scheme = rp.targetURL.Scheme
			path := strings.Replace(reqURLPath, rp.basePath, "", 1)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			req.URL.Path = path

			req.Header.Set("X-Forwarded-Host", reqHost)
			req.Header.Set("X-Forwarded-Proto", rp.proto)
			req.Header.Set("X-Forwarded-Url", reqURL)

			Logger.Printf("host: %s://%v%v -> %v - %v\n", rp.proto, reqHost, reqURL, req.Host, req.URL.String())
		},
	}

	proxy.ServeHTTP(w, r)
}
