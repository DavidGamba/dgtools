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
)

var Logger = log.New(ioutil.Discard, "", log.LstdFlags)

// GetURLToFile - Gets the contents of an URL into a file.
// If the file exists and the overwrite flag is not set, it exits.
// TODO: Set file cache timeout, if file is older than timeout re-download file.
// It would replace the overwrite flag and instead we can pass a timeout of 0.
func GetURLToFile(url, fpath string, headers map[string]string, overwrite bool) error {
	Logger.Printf("Downloading %s\n", url)

	// Check if file exist
	_, err := os.Stat(fpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	if err == nil {
		if !overwrite {
			Logger.Printf("File already exists: %s\n", fpath)
			return nil
		}
	}

	// Create dir structure
	_ = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)

	// Ignore SSL cert errors
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Do the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
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
