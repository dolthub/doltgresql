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

package _go

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/madflojo/testcerts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

type SSLListener struct {
	*dserver.Listener
}

func NewSslListener(_ server.Config, listenerCfg mysql.ListenerConfig, sel server.ServerEventListener) (server.ProtocolListener, error) {
	// Since this is intended for testing, we'll configure a test certificate so that we can test for SSL support
	cert, key, err := testcerts.GenerateCerts()
	if err != nil {
		panic(err)
	}

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}

	listener, err := dserver.NewListenerWithOpts(listenerCfg, sel, dserver.WithCertificate(certificate))
	if err != nil {
		return nil, err
	}

	return &SSLListener{
		listener.(*dserver.Listener),
	}, nil
}

func TestSSL(t *testing.T) {
	port, err := sql.GetEmptyPort()
	require.NoError(t, err)
	controller, err := dserver.RunInMemory(&servercfg.DoltgresConfig{
		DoltgresConfig: cfgdetails.DoltgresConfig{
			ListenerConfig: &cfgdetails.DoltgresListenerConfig{
				PortNumber: &port,
				HostStr:    ptr("127.0.0.1"),
			},
		},
	}, NewSslListener)
	require.NoError(t, err)

	defer func() {
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()

	ctx := context.Background()
	err = func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/?sslmode=require", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}
		return conn.Close(ctx)
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres?sslmode=require", port))
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 int4);")
	require.NoError(t, err)
	_, err = conn.Exec(ctx, "INSERT INTO test VALUES (3645, 37643);")
	require.NoError(t, err)
	rows, err := conn.Query(ctx, "SELECT * FROM test;")
	require.NoError(t, err)
	readRows, _, err := ReadRows(rows, true)
	require.NoError(t, err)
	assert.Equal(t, NormalizeExpectedRow(rows.FieldDescriptions(), []sql.Row{{3645, 37643}}), readRows)
}
