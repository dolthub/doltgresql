// Copyright 2024 Dolthub, Inc.
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
	"fmt"
	"net"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/jackc/pgx/v5"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
)

// CreateDoltgresServer creates and returns a Doltgres server.
func CreateDoltgresServer() (controller *svcs.Controller, port int, err error) {
	portListener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, 0, err
	}
	port = portListener.Addr().(*net.TCPAddr).Port
	if err = portListener.Close(); err != nil {
		return nil, 0, err
	}
	address := "127.0.0.1"
	logLevel := servercfg.LogLevel_Fatal
	controller, err = dserver.RunInMemory(&servercfg.DoltgresConfig{
		LogLevelStr: &logLevel,
		ListenerConfig: &servercfg.DoltgresListenerConfig{
			PortNumber: &port,
			HostStr:    &address,
		},
	})
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if err != nil {
			controller.Stop()
		}
	}()
	return controller, port, func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, "CREATE DATABASE postgres;")
		return err
	}()
}
