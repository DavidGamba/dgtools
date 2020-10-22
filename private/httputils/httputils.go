package httputils

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Logger instance
var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

// GetURLToFileOptions - Internal options store
type GetURLToFileOptions struct {
	cacheDuration time.Duration
	headers       map[string]string
	ignoreSSL     bool
}

// GetURLToFileOptionFn - Options type
type GetURLToFileOptionFn func(*GetURLToFileOptions)

// Headers - Set request headers
func Headers(headers map[string]string) GetURLToFileOptionFn {
	return func(options *GetURLToFileOptions) {
		options.headers = headers
	}
}

// CacheDuration - If the file exists and is older than the cacheDuration then re-download the file, otherwise re-use it.
func CacheDuration(duration time.Duration) GetURLToFileOptionFn {
	return func(options *GetURLToFileOptions) {
		options.cacheDuration = duration
	}
}

// InsecureSkipVerify - Skips SSL verification
func InsecureSkipVerify() GetURLToFileOptionFn {
	return func(options *GetURLToFileOptions) {
		options.ignoreSSL = true
	}
}

// GetURLToFile - Gets the contents of an URL into a file.
// If the file exists and the file is older than timeout re-download file.
// Set the timeout to 0 to always download the file.
// NOTE: This function ignores SSL verification errors
func GetURLToFile(url, fpath string, fns ...GetURLToFileOptionFn) error {
	Logger.Printf("Downloading %s\n", url)

	params := GetURLToFileOptions{
		headers:       make(map[string]string),
		ignoreSSL:     false,
		cacheDuration: 0,
	}
	for _, fn := range fns {
		fn(&params)
	}

	// Check if file exist
	fileInfo, err := os.Stat(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		fileInfo = nil
	}
	if fileInfo != nil && params.cacheDuration != 0 {
		if fileInfo.ModTime().After(time.Now().Add(-params.cacheDuration)) {
			Logger.Printf("File already exists and is up to date: %s\n", fpath)
			return nil
		}
	}

	// Create dir structure
	_ = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)

	tr := &http.Transport{}
	if params.ignoreSSL {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Transport: tr}

	// Do the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range params.headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("URL failed with status code: %d", resp.StatusCode)
	}

	// Save output
	out, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
