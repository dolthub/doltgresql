package _go

// Temporary instrumentation for the Rust port: when DOLTGRES_RECORD_DIR is set, every pgx connection records the
// bytes it sends, and each script's recording is written to a file named after its test and script. Never committed.

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
)

var recordMutex sync.Mutex
var recordConns []*recordingConn

// recordingConn copies everything written to the connection.
type recordingConn struct {
	net.Conn
	mu  sync.Mutex
	buf []byte
}

// Write implements the interface net.Conn.
func (c *recordingConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	c.buf = append(c.buf, b...)
	c.mu.Unlock()
	return c.Conn.Write(b)
}

// ApplyRecording makes the configuration record its connections when recording is enabled.
func ApplyRecording(config *pgx.ConnConfig) {
	if _, ok := os.LookupEnv("DOLTGRES_RECORD_DIR"); !ok {
		return
	}
	config.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		rc := &recordingConn{Conn: conn}
		recordMutex.Lock()
		recordConns = append(recordConns, rc)
		recordMutex.Unlock()
		return rc, nil
	}
}

// RecordingConnect connects like pgx.Connect, recording the connection when recording is enabled.
func RecordingConnect(ctx context.Context, url string) (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	ApplyRecording(config)
	return pgx.ConnectConfig(ctx, config)
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9]`)
var savedNames = map[string]int{}
var currentSession = 0
var sessionSaved = false

// StartRecording starts recording a new server's connections. A test that never saves its recording through a
// script runner, such as one with custom code, saves it as "custom" when it finishes.
func StartRecording(t *testing.T) {
	if _, ok := os.LookupEnv("DOLTGRES_RECORD_DIR"); !ok {
		return
	}
	recordMutex.Lock()
	recordConns = nil
	recordMutex.Unlock()
	currentSession++
	sessionSaved = false
	session := currentSession
	t.Cleanup(func() {
		if session == currentSession && !sessionSaved {
			SaveRecording(t, "custom")
		}
	})
}

// SaveRecording writes and clears the recording of the current script.
func SaveRecording(t *testing.T, scriptName string) {
	dir, ok := os.LookupEnv("DOLTGRES_RECORD_DIR")
	if !ok {
		return
	}
	if !filepath.IsAbs(dir) {
		_, self, _, _ := runtime.Caller(0)
		dir = filepath.Join(filepath.Dir(self), dir)
	}
	recordMutex.Lock()
	conns := recordConns
	recordConns = nil
	recordMutex.Unlock()
	sessionSaved = true
	test := strings.Split(t.Name(), "/")[0]
	base := test + "__" + unsafeChars.ReplaceAllString(scriptName, "_")
	savedNames[base]++
	name := base
	if savedNames[base] > 1 {
		name = fmt.Sprintf("%s__%d", base, savedNames[base])
	}
	var sb strings.Builder
	for _, c := range conns {
		c.mu.Lock()
		sb.WriteString(fmt.Sprintf("%x\n", c.buf))
		c.mu.Unlock()
	}
	if err := os.WriteFile(filepath.Join(dir, name+".hex"), []byte(sb.String()), 0644); err != nil {
		panic(err)
	}
}
