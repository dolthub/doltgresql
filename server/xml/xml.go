// Copyright 2026 Dolthub, Inc.
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

package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// declarationRegex matches the XML declaration that may begin an xml value, capturing its version and standalone
// values.
var declarationRegex = regexp.MustCompile(`^<\?xml\s+version\s*=\s*["']([^"']*)["'](?:\s+encoding\s*=\s*["'][^"']*["'])?(?:\s+standalone\s*=\s*["'](yes|no)["'])?\s*\?>`)

// CheckWellFormed returns an error if `input` is not well-formed XML content, or not a well-formed XML document when
// `document` is set.
func CheckWellFormed(input string, document bool) error {
	kind, code := "content", pgcode.InvalidXMLContent
	if document {
		kind, code = "document", pgcode.InvalidXMLDocument
	}
	_, _, input = SplitDeclaration(input)
	decoder := xml.NewDecoder(strings.NewReader(input))
	depth := 0
	roots := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return pgerror.Newf(code, "invalid XML %s: %s", kind, err)
		}
		switch token := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if document && depth == 0 && len(bytes.TrimSpace(token)) > 0 {
				return pgerror.New(pgcode.InvalidXMLDocument, "invalid XML document")
			}
		}
	}
	if document && roots != 1 {
		return pgerror.New(pgcode.InvalidXMLDocument, "invalid XML document")
	}
	return nil
}

// SplitDeclaration splits a leading XML declaration off `str`, returning its version and standalone values (both empty
// when there is no declaration) along with the remaining text.
func SplitDeclaration(str string) (version string, standalone string, rest string) {
	match := declarationRegex.FindStringSubmatch(str)
	if match == nil {
		return "", "", str
	}
	return match[1], match[2], str[len(match[0]):]
}

// Declaration returns the XML declaration that PostgreSQL prints for the given version and standalone values, which is
// empty when the version is 1.0 (or unknown) and standalone is absent.
func Declaration(version string, standalone string) string {
	if version == "" {
		version = "1.0"
	}
	if version == "1.0" && standalone == "" {
		return ""
	} else if standalone == "" {
		return fmt.Sprintf(`<?xml version="%s"?>`, version)
	}
	return fmt.Sprintf(`<?xml version="%s" standalone="%s"?>`, version, standalone)
}

// Output returns the text form of an xml value, dropping or trimming a leading XML declaration the way PostgreSQL's
// `xml_out` does.
func Output(str string) string {
	version, standalone, rest := SplitDeclaration(str)
	if version == "" {
		return str
	} else if declaration := Declaration(version, standalone); declaration != "" {
		return declaration + rest
	}
	return strings.TrimPrefix(rest, "\n")
}

// Concat concatenates the xml values `values`, merging their XML declarations the way PostgreSQL's `xmlconcat` does.
func Concat(values ...string) string {
	sb := strings.Builder{}
	globalVersion := ""
	versionsDiffer := false
	globalStandalone := "yes"
	for _, value := range values {
		version, standalone, rest := SplitDeclaration(value)
		if standalone == "" {
			globalStandalone = ""
		} else if standalone == "no" && globalStandalone == "yes" {
			globalStandalone = "no"
		}
		if version == "" || (globalVersion != "" && version != globalVersion) {
			versionsDiffer = true
		} else {
			globalVersion = version
		}
		sb.WriteString(rest)
	}
	if versionsDiffer {
		globalVersion = ""
	}
	if versionsDiffer && globalStandalone == "" {
		return sb.String()
	}
	return Declaration(globalVersion, globalStandalone) + sb.String()
}
