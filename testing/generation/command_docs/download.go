// Copyright 2023 Dolthub, Inc.
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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const (
	postgresBaseLink = `https://www.postgresql.org/docs/15/`
	commandPageLink  = `sql-commands.html`
)

type Link struct {
	CommandName string
	Link        string
}

// DownloadAllSynopses downloads and writes all synopses from the internet. It is assumed that this will be run from
// within an IDE, and will download to the `commands_docs/synopses` folder. Uses `postgresBaseLink` to determine which
// version of the command page will be downloaded, and `commandPageLink` specifies the exact page that the commands are
// contained on. This function is only necessary if we're upgrading the Postgres version.
func DownloadAllSynopses() {
	commandDocument, err := FetchDocument(postgresBaseLink + commandPageLink)
	if err != nil {
		panic(err)
	}
	var allLinks []Link
	_ = commandDocument.Find(".toc a").Each(func(i int, e *goquery.Selection) {
		href, exists := e.Attr("href")
		if exists {
			allLinks = append(allLinks, Link{
				CommandName: e.Text(),
				Link:        postgresBaseLink + strings.Trim(href, `\/`),
			})
		}
	})
	for _, link := range allLinks {
		linkDocument, err := FetchDocument(link.Link)
		if err != nil {
			panic(err)
		}
		synopsis := linkDocument.Find(".synopsis").First()
		if strings.Contains(synopsis.Text(), "$") {
			panic(fmt.Errorf("Synopsis has a $, which is unexpected:\n\n%s\n\n%s", link, synopsis.Text()))
		}
		synopsis.Find(".replaceable").Each(func(i int, e *goquery.Selection) {
			e.SetText(fmt.Sprintf("$%s$", e.Text()))
		})
		func() {
			fileLocation, err := GetCommandDocsFolder()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fileName := strings.ToLower(strings.ReplaceAll(link.CommandName, " ", "_"))
			data := []byte(synopsis.Text())
			if err = os.WriteFile(fmt.Sprintf("%s/synopses/%s.txt", fileLocation, fileName), data, 0644); err != nil {
				fmt.Println(err.Error())
			}
			fmt.Printf("Downloaded: %s\n", link.Link)
		}()
	}
}

// FetchDocument fetches the document from the given link. The link should be a full HTTP/HTTPS URL.
func FetchDocument(link string) (*goquery.Document, error) {
	res, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	node, err := html.Parse(res.Body)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromNode(node), nil
}

// GetCommandDocsFolder returns the location of this particular Go file: download.go. This is useful to locate relative
// directories, such as the synopses folder. It is assumed that this will always be called from within an IDE.
func GetCommandDocsFolder() (string, error) {
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to fetch the location of the current file")
	}
	return filepath.ToSlash(filepath.Dir(currentFileLocation)), nil
}
