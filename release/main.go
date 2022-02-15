package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
)

type templateArgs struct {
	Version string
	URL     string
	SHA256  string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go version")
		os.Exit(1)
	}
	out, err := processTemplate(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}

func sha256FromURL(url string) (string, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("cannot download: %w", err)
	}
	defer resp.Body.Close()

	hasher := sha256.New()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return "", fmt.Errorf("cannot copy: %w", err)
	}
	hasher.Write(buf.Bytes())
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func processTemplate(version string) (string, error) {
	t, err := template.ParseFiles("cockroach-tmpl.rb")

	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	url := fmt.Sprintf("https://binaries.cockroachdb.com/cockroach-v%s.darwin-10.9-amd64.tgz", version)
	sha256, err := sha256FromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to calculate SHA256: %w", err)
	}
	data := templateArgs{
		Version: version,
		URL:     url,
		SHA256:  sha256,
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("cannot execute template: %w", err)
	}
	return buf.String(), nil
}