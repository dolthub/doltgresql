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

// Command postgis-regress runs the core regression tests of PostGIS against a server, the way PostGIS's own
// regress/run_test.pl runs them, and reports which pass. PostGIS's tests are licensed under the GPL, so they are not
// checked in: the release is downloaded, and its checksum verified, the first time the command runs.
package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The PostGIS release whose tests are run, and the checksum of its source archive.
const (
	postgisVersion = "3.6.4"
	postgisURL     = "https://download.osgeo.org/postgis/source/postgis-" + postgisVersion + ".tar.gz"
	postgisSHA256  = "ed8dc6679f1e06f7b113592b04cde2a7e00f1b1e681294c8ca2204058990cec6"
)

// buildFeatures holds the values of the conditions in regress/core/tests.mk.in that select tests. They describe a
// PostGIS built against GEOS 3.11.1 with JSON, SP-GiST, and protobuf support, which the emulated extension follows.
var buildFeatures = map[string]string{
	"POSTGIS_GEOS_VERSION": "31101",
	"HAVE_JSON":            "yes",
	"HAVE_SPGIST":          "yes",
	"HAVE_PROTOBUF":        "yes",
	"INTERRUPTTESTS":       "no",
}

// result is the outcome of one test.
type result struct {
	Name          string  `json:"name"`
	Passed        bool    `json:"passed"`
	Error         string  `json:"error,omitempty"`
	ExpectedLines int     `json:"expectedLines"`
	MatchedLines  int     `json:"matchedLines"`
	Seconds       float64 `json:"seconds"`
}

func main() {
	host := flag.String("host", "127.0.0.1", "host of the server to test")
	port := flag.Int("port", 5432, "port of the server to test")
	user := flag.String("user", "postgres", "user to connect as; the password is read from PGPASSWORD")
	adminDB := flag.String("admin-db", "postgres", "database to connect to while creating each test's database")
	cacheDir := flag.String("cache", defaultCacheDir(), "directory to download PostGIS's source into")
	outDir := flag.String("out", "postgis-regress-results", "directory to write each test's output and diff into")
	run := flag.String("run", "", "regular expression selecting the tests to run")
	timeout := flag.Duration("timeout", 5*time.Minute, "longest that a single test may run")
	scriptDir := flag.String("scriptdir", "", "directory of PostGIS's installed SQL scripts, such as legacy.sql, which a few tests include")
	createOptions := flag.String("create-db-options", "", "options for CREATE DATABASE, such as ENCODING 'UTF8' TEMPLATE template0 LC_COLLATE 'C' when testing PostgreSQL itself, as run_test.pl creates its databases")
	flag.Parse()

	regressDir, err := fetchPostGIS(*cacheDir)
	if err != nil {
		fail(err)
	}
	tests, err := testList(filepath.Join(regressDir, "core", "tests.mk.in"))
	if err != nil {
		fail(err)
	}
	if *run != "" {
		pattern, err := regexp.Compile(*run)
		if err != nil {
			fail(err)
		}
		var selected []string
		for _, test := range tests {
			if pattern.MatchString(test) {
				selected = append(selected, test)
			}
		}
		tests = selected
	}
	if err = os.MkdirAll(*outDir, 0755); err != nil {
		fail(err)
	}
	conn := connection{host: *host, port: *port, user: *user, scriptDir: *scriptDir, createOptions: *createOptions}
	var results []result
	for i, test := range tests {
		fmt.Printf("[%3d/%d] %-40s ", i+1, len(tests), test)
		r := runTest(conn, *adminDB, regressDir, *outDir, test, *timeout)
		results = append(results, r)
		switch {
		case r.Passed:
			fmt.Printf("ok    (%.1fs)\n", r.Seconds)
		case r.Error != "":
			fmt.Printf("ERROR %s\n", r.Error)
		default:
			fmt.Printf("FAIL  %d/%d lines (%.1fs)\n", r.MatchedLines, r.ExpectedLines, r.Seconds)
		}
	}
	if err = writeSummary(*outDir, results); err != nil {
		fail(err)
	}
}

// fail prints the error and exits.
func fail(err error) {
	fmt.Fprintln(os.Stderr, "postgis-regress:", err)
	os.Exit(1)
}

// defaultCacheDir returns the directory that PostGIS's source is downloaded into by default.
func defaultCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "doltgres-postgis-regress")
	}
	return filepath.Join(os.TempDir(), "doltgres-postgis-regress")
}

// fetchPostGIS downloads and extracts the regress directory of PostGIS's source unless it already has been, and returns
// its path.
func fetchPostGIS(cacheDir string) (string, error) {
	root := filepath.Join(cacheDir, "postgis-"+postgisVersion)
	regressDir := filepath.Join(root, "regress")
	if _, err := os.Stat(filepath.Join(regressDir, "core", "tests.mk.in")); err == nil {
		return regressDir, nil
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}
	fmt.Printf("Downloading %s\n", postgisURL)
	response, err := http.Get(postgisURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: %s", postgisURL, response.Status)
	}
	archive, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(archive)
	if hex.EncodeToString(sum[:]) != postgisSHA256 {
		return "", fmt.Errorf("the checksum of %s does not match", postgisURL)
	}
	gz, err := gzip.NewReader(strings.NewReader(string(archive)))
	if err != nil {
		return "", err
	}
	reader := tar.NewReader(gz)
	prefix := "postgis-" + postgisVersion + "/regress/"
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(header.Name, prefix) || header.Typeflag != tar.TypeReg || strings.Contains(header.Name, "..") {
			continue
		}
		path := filepath.Join(cacheDir, filepath.FromSlash(header.Name))
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", err
		}
		file, err := os.Create(path)
		if err != nil {
			return "", err
		}
		if _, err = io.Copy(file, reader); err != nil {
			file.Close()
			return "", err
		}
		if err = file.Close(); err != nil {
			return "", err
		}
	}
	return regressDir, nil
}

// testList returns the core tests that regress/core/tests.mk.in selects for buildFeatures, in its order, leaving out
// the slow tests that it only runs on request.
func testList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	versionCond := regexp.MustCompile(`^ifeq \(\$\(shell expr "\$\(([A-Z_]+)\)" "([<>=]+)" (\d+)\),1\)`)
	flagCond := regexp.MustCompile(`^ifeq \(\$\(([A-Z_]+)\),(\w+)\)`)
	testName := regexp.MustCompile(`\$\(top_srcdir\)/regress/core/([A-Za-z0-9_\-]+)`)
	// active holds, for each enclosing conditional, whether its current branch is taken
	var active []bool
	enabled := func() bool {
		for _, a := range active {
			if !a {
				return false
			}
		}
		return true
	}
	var tests []string
	collecting := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "#") || line == "":
			continue
		case versionCond.MatchString(line):
			m := versionCond.FindStringSubmatch(line)
			have, _ := strconv.Atoi(buildFeatures[m[1]])
			want, _ := strconv.Atoi(m[3])
			taken := map[string]bool{">=": have >= want, "<=": have <= want, ">": have > want, "<": have < want, "=": have == want}[m[2]]
			active = append(active, taken)
		case flagCond.MatchString(line):
			m := flagCond.FindStringSubmatch(line)
			active = append(active, buildFeatures[m[1]] == m[2])
		case line == "else":
			active[len(active)-1] = !active[len(active)-1]
		case line == "endif":
			active = active[:len(active)-1]
		case strings.HasPrefix(line, "TESTS +="):
			collecting = true
		case strings.Contains(line, "=") && !strings.Contains(line, "$(top_srcdir)"):
			collecting = false
		}
		if collecting && enabled() {
			for _, m := range testName.FindAllStringSubmatch(line, -1) {
				tests = append(tests, m[1])
			}
		}
		if !strings.HasSuffix(line, `\`) && !strings.HasPrefix(line, "TESTS +=") {
			collecting = false
		}
	}
	return tests, scanner.Err()
}

// connection holds how to connect to the server being tested.
type connection struct {
	host          string
	port          int
	user          string
	scriptDir     string
	createOptions string
}

// dsn returns the libpq connection string for the given database.
func (c connection) dsn(db string) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s", c.host, c.port, c.user, db)
}

// psql runs psql against the given database with the given arguments, returning its combined output.
func (c connection) psql(ctx context.Context, dir string, db string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "psql", append([]string{"--no-psqlrc", c.dsn(db)}, args...)...)
	cmd.Dir = dir
	// The C locale keeps psql's output, and the locale warnings of some psql wrappers, from depending on the machine
	// psql only skips the byte order mark that starts some test files when the client encoding is UTF-8
	cmd.Env = append(os.Environ(), "PGCONNECT_TIMEOUT=10", "LC_ALL=C", "LANG=C", "LANGUAGE=", "PGCLIENTENCODING=UTF8")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(output), fmt.Errorf("timed out")
	}
	return string(output), err
}

// runTest runs one test in a database of its own, as run_test.pl's run_simple_test does, and compares its output to
// the test's expected output.
func runTest(conn connection, adminDB string, regressDir string, outDir string, test string, timeout time.Duration) (r result) {
	start := time.Now()
	r = result{Name: test}
	defer func() { r.Seconds = time.Since(start).Seconds() }()
	coreDir := filepath.Join(regressDir, "core")
	expectedData, err := os.ReadFile(filepath.Join(coreDir, test+"_expected"))
	if err != nil {
		r.Error = "no expected output"
		return r
	}
	expected := splitLines(string(expectedData))
	r.ExpectedLines = len(expected)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	db := "postgis_reg_" + strings.ReplaceAll(test, "-", "_")
	if out, err := conn.psql(ctx, coreDir, adminDB, "-v", "ON_ERROR_STOP=1", "-c", "DROP DATABASE IF EXISTS "+db, "-c", strings.TrimSpace("CREATE DATABASE "+db+" "+conn.createOptions)); err != nil {
		r.Error = "creating the database: " + firstLine(out, err)
		return r
	}
	defer func() {
		cleanup, cancelCleanup := context.WithTimeout(context.Background(), time.Minute)
		defer cancelCleanup()
		_, _ = conn.psql(cleanup, coreDir, adminDB, "-c", "DROP DATABASE IF EXISTS "+db)
	}()
	if out, err := conn.psql(ctx, coreDir, db, "-v", "ON_ERROR_STOP=1", "-c", "CREATE EXTENSION postgis"); err != nil {
		r.Error = "creating the extension: " + firstLine(out, err)
		return r
	}
	tmpDir, err := os.MkdirTemp("", "postgis-regress")
	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer os.RemoveAll(tmpDir)
	// The server reads and writes the test's temporary file, so it may run as another user, as run_test.pl allows
	if err = os.Chmod(tmpDir, 0777); err != nil {
		r.Error = err.Error()
		return r
	}
	scriptDir := conn.scriptDir
	if scriptDir == "" {
		scriptDir = tmpDir
	}
	output, err := conn.psql(ctx, coreDir, db,
		"-v", "VERBOSITY=terse",
		"-v", "tmpfile='"+filepath.Join(tmpDir, "test_tmp")+"'",
		"-v", "scriptdir="+scriptDir,
		"-v", "regdir="+regressDir,
		"-v", "top_builddir="+filepath.Dir(regressDir),
		"-v", "schema=public.",
		"-c", "SET search_path TO public,public,topology",
		"-tXAq", "-f", test+".sql")
	if err != nil && strings.Contains(err.Error(), "timed out") {
		r.Error = "timed out"
	}
	actual := normalize(output)
	r.MatchedLines = matchedLines(expected, actual)
	r.Passed = r.Error == "" && strings.Join(expected, "\n") == strings.Join(actual, "\n")
	if writeErr := os.WriteFile(filepath.Join(outDir, test+".out"), []byte(strings.Join(actual, "\n")+"\n"), 0644); writeErr != nil {
		r.Error = writeErr.Error()
	}
	if !r.Passed {
		diff := exec.Command("diff", "-u", filepath.Join(coreDir, test+"_expected"), filepath.Join(outDir, test+".out"))
		diffOutput, _ := diff.CombinedOutput()
		_ = os.WriteFile(filepath.Join(outDir, test+".diff"), diffOutput, 0644)
	}
	return r
}

// firstLine returns the first line of a command's output, or its error when it printed nothing.
func firstLine(output string, err error) string {
	if line, _, _ := strings.Cut(strings.TrimSpace(output), "\n"); line != "" {
		return line
	}
	return err.Error()
}

// splitLines returns the lines of the text without their line endings, leaving out lines of only whitespace, as
// run_test.pl does for test output.
func splitLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimRight(line, "\r"))
		}
	}
	return lines
}

// The substitutions that run_test.pl makes on every line of test output.
var (
	exponentPadding = regexp.MustCompile(`[eE]([+-])0+(\d+)`)
	selfIntersect   = regexp.MustCompile(`Self-intersection .*`)
	psqlPrefix      = regexp.MustCompile(`^psql.*(NOTICE|WARNING|ERROR):`)
)

// normalize returns the lines of test output with the substitutions that run_test.pl makes.
func normalize(output string) []string {
	lines := splitLines(output)
	for i, line := range lines {
		line = strings.ReplaceAll(line, "Infinity", "inf")
		line = strings.ReplaceAll(line, "Inf", "inf")
		line = strings.ReplaceAll(line, "1.#INF", "inf")
		line = exponentPadding.ReplaceAllString(line, "e$1$2")
		line = selfIntersect.ReplaceAllString(line, "Self-intersection")
		if strings.HasPrefix(line, "ROLLBACK") {
			line = "COMMIT" + strings.TrimPrefix(line, "ROLLBACK")
		}
		line = psqlPrefix.ReplaceAllString(line, "$1:")
		lines[i] = line
	}
	return lines
}

// matchedLines returns the length of the longest common subsequence of the expected and actual lines, which counts
// the expected lines that the output reproduces in order.
func matchedLines(expected []string, actual []string) int {
	previous := make([]int, len(actual)+1)
	current := make([]int, len(actual)+1)
	for i := 1; i <= len(expected); i++ {
		for j := 1; j <= len(actual); j++ {
			if expected[i-1] == actual[j-1] {
				current[j] = previous[j-1] + 1
			} else {
				current[j] = max(previous[j], current[j-1])
			}
		}
		previous, current = current, previous
	}
	return previous[len(actual)]
}

// writeSummary writes the results as JSON and as text, and prints the totals.
func writeSummary(outDir string, results []result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(outDir, "results.json"), data, 0644); err != nil {
		return err
	}
	passed, errored, expectedLines, matched := 0, 0, 0, 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
		if r.Error != "" {
			errored++
		}
		expectedLines += r.ExpectedLines
		matched += r.MatchedLines
	}
	sorted := append([]result(nil), results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	var sb strings.Builder
	fmt.Fprintf(&sb, "PostGIS %s core regression tests\n", postgisVersion)
	fmt.Fprintf(&sb, "tests passed: %d of %d (%d could not run)\n", passed, len(results), errored)
	if expectedLines > 0 {
		fmt.Fprintf(&sb, "expected output lines reproduced: %d of %d (%.1f%%)\n\n", matched, expectedLines,
			100*float64(matched)/float64(expectedLines))
	}
	for _, r := range sorted {
		status := "FAIL"
		if r.Passed {
			status = "ok"
		} else if r.Error != "" {
			status = "ERROR " + r.Error
		}
		fmt.Fprintf(&sb, "%-40s %5d/%-5d %s\n", r.Name, r.MatchedLines, r.ExpectedLines, status)
	}
	fmt.Print("\n" + strings.SplitN(sb.String(), "\n\n", 2)[0] + "\n")
	return os.WriteFile(filepath.Join(outDir, "summary.txt"), []byte(sb.String()), 0644)
}
