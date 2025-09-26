// Copyright 2025 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	query         = `extension:sql pg_dump`
	downloadCount = 110
)

// RepoName simply contains the name of the repository.
type RepoName struct {
	FullName string `json:"full_name"`
}

// Item is a SQL file (hopefully) containing a pg_dump.
type Item struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	HtmlURL     string   `json:"html_url"`
	ContentsURL string   `json:"url"`
	Repository  RepoName `json:"repository"`
}

// CodeSearchResult contains the result of a code search.
type CodeSearchResult struct {
	TotalCount        int    `json:"total_count"`
	IncompleteResults bool   `json:"incomplete_results"`
	Items             []Item `json:"items"`
	Message           string `json:"message"` // Only used when there's an error
}

// ContentFile is all of the information about a SQL file, including how to retrieve it.
type ContentFile struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int64  `json:"size"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

func main() {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		fmt.Println("Must provide a GITHUB_TOKEN as an environment variable")
		os.Exit(1)
	}

	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("Unable to find the folder where this file is located")
		os.Exit(1)
	}
	dumpsFolder := filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation), "../sql"))

	var saved int
	page := 1

OuterLoop:
	for {
		remaining := downloadCount - saved
		items, err := SearchCode(ctx, httpClient, token, page, min(50, remaining))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			cf, err := GetContent(ctx, httpClient, token, item.ContentsURL)
			if err != nil {
				fmt.Printf("warn: %s/%s: %v\n", item.Repository.FullName, item.Path, err)
				continue
			}
			if cf.Type != "file" || cf.DownloadURL == "" {
				continue
			}

			dest := filepath.Join(dumpsFolder, SanitizePath(item.Repository.FullName)+filepath.Ext(cf.Path))
			if _, err = os.Stat(dest); err == nil {
				continue
			}
			if err = DownloadFile(ctx, httpClient, item, cf.DownloadURL, dest); err != nil {
				fmt.Printf("download error: %s -> %v\n", dest, err)
				continue
			}
			fmt.Printf("saved: %s (%d bytes)\n", dest, cf.Size)

			saved++
			if saved >= downloadCount {
				break OuterLoop
			}
			time.Sleep(6500 * time.Millisecond) // We sleep to mitigate rate limits
		}
		page++
	}
}

// SearchCode executes the query against the API, returning all items that were found.
func SearchCode(ctx context.Context, hc *http.Client, token string, page int, perPage int) ([]Item, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("per_page", strconv.Itoa(perPage))

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/search/code?"+params.Encode(), nil)
	SetHeaders(req, token)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if HandleRate(resp) {
		return SearchCode(ctx, hc, token, page, perPage)
	}
	var sr CodeSearchResult
	if err = json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		if sr.Message != "" {
			return nil, fmt.Errorf("search error: %s (HTTP %d)", sr.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("search error: HTTP %d", resp.StatusCode)
	}
	return sr.Items, nil
}

// GetContent gets the ContentFile from the given URL.
func GetContent(ctx context.Context, hc *http.Client, token string, contentsURL string) (*ContentFile, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", contentsURL, nil)
	SetHeaders(req, token)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if HandleRate(resp) {
		return GetContent(ctx, hc, token, contentsURL)
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("contents error: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var cf ContentFile
	if err = json.NewDecoder(resp.Body).Decode(&cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

// DownloadFile downloads the given SQL file to the destination.
func DownloadFile(ctx context.Context, hc *http.Client, item Item, rawURL string, dest string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	req.Header.Set("User-Agent", "gh-pg-dump-finder/1.0")
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, _ = io.WriteString(out, fmt.Sprintf("-- Downloaded from: %s\n", item.HtmlURL))
	_, err = io.Copy(out, resp.Body)
	return err
}

// SetHeaders sets the appropriate headers for a request.
func SetHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gh-pg-dump-finder/1.0")
	req.Header.Set("Authorization", "Bearer "+token)
}

// HandleRate handles potential rate limits.
func HandleRate(resp *http.Response) bool {
	if resp.StatusCode == 403 {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, _ := strconv.Atoi(ra); secs > 0 {
				sleepTime := time.Duration(secs) * time.Second
				fmt.Printf("rate limited (%s), retrying\n", sleepTime.String())
				time.Sleep(sleepTime)
				return true
			}
		}
		if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
			if ts, _ := strconv.ParseInt(reset, 10, 64); ts > 0 {
				wait := time.Until(time.Unix(ts+5, 0))
				if wait > 0 && wait < 5*time.Minute {
					fmt.Printf("rate limited (%s), retrying\n", wait.String())
					time.Sleep(wait)
					return true
				}
			}
		}
	}
	return false
}

// SanitizePath removes potentially invalid file system characters.
func SanitizePath(s string) string {
	illegal := `<>:"\|/?*`
	for _, r := range illegal {
		s = strings.ReplaceAll(s, string(r), "_")
	}
	return s
}
