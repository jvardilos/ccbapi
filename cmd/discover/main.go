// discover is a fuzzy-search tool over CCB's public OpenAPI spec, for
// finding endpoint paths, methods, and summaries during development without
// needing a live MCP connection.
//
// Usage: go run ./cmd/discover <search terms>
package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const specURL = "https://api.ccbchurch.com/documentation.yaml"

type endpoint struct {
	Method  string
	Path    string
	Summary string
}

type spec struct {
	Paths map[string]map[string]yaml.Node `yaml:"paths"`
}

var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true,
	"delete": true, "patch": true, "head": true, "options": true,
}

func fetchEndpoints() ([]endpoint, error) {
	resp, err := http.Get(specURL)
	if err != nil {
		return nil, fmt.Errorf("fetching spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching spec: status %s", resp.Status)
	}

	var doc spec
	if err := yaml.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decoding spec: %w", err)
	}

	var endpoints []endpoint
	for path, methods := range doc.Paths {
		for method, node := range methods {
			if !httpMethods[strings.ToLower(method)] {
				continue
			}

			var op struct {
				Summary string `yaml:"summary"`
			}
			_ = node.Decode(&op)

			endpoints = append(endpoints, endpoint{
				Method:  strings.ToUpper(method),
				Path:    path,
				Summary: op.Summary,
			})
		}
	}

	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].Path != endpoints[j].Path {
			return endpoints[i].Path < endpoints[j].Path
		}
		return endpoints[i].Method < endpoints[j].Method
	})

	return endpoints, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: discover <search terms>")
		os.Exit(1)
	}
	pattern := strings.ToLower(strings.Join(os.Args[1:], " "))

	endpoints, err := fetchEndpoints()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	matches := 0
	for _, e := range endpoints {
		haystack := strings.ToLower(e.Method + " " + e.Path + " " + e.Summary)
		if strings.Contains(haystack, pattern) {
			fmt.Printf("%-6s %-55s %s\n", e.Method, e.Path, e.Summary)
			matches++
		}
	}

	if matches == 0 {
		fmt.Println("no matches")
	}
}
