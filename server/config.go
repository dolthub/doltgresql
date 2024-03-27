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

package server

import (
	"fmt"

	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/utils/argparser"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"gopkg.in/yaml.v2"
)

type DoltgresServerConfig struct {
	*sqlserver.YAMLConfig
	PostgresReplicationConfig *PostgresReplicationConfig `yaml:"postgres_replication,omitempty"`
}

type PostgresReplicationConfig struct {
	PostgresServerAddress string `yaml:"postgres_server_address"`
	PostgresUser          string `yaml:"postgres_user"`
	PostgresPassword      string `yaml:"postgres_password"`
	PostgresDatabase      string `yaml:"postgres_database"`
	PostgresPort          int    `yaml:"postgres_port"`
	SlotName              string `yaml:"slot_name"`
}

var _ sqlserver.ServerConfig = (*DoltgresServerConfig)(nil)

type DoltgresConfigReader struct{}

func (d DoltgresConfigReader) ReadConfigFile(fs filesys.Filesys, file string) (sqlserver.ServerConfig, error) {
	data, err := fs.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file '%s'. Error: %w", file, err)
	}

	// TODO: we lose the ability to unmarshal strict here, because our YAMLConfig has fields the Dolt implementation
	//  doesn't recognize. Maybe worthwhile to use a generic map first?
	var cfg sqlserver.YAMLConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse yaml file '%s'. Error: %w", file, err)
	}

	var doltgresCfg DoltgresServerConfig
	err = yaml.Unmarshal(data, &doltgresCfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse yaml file '%s'. Error: %w", file, err)
	}

	doltgresCfg.YAMLConfig = &cfg
	return &doltgresCfg, nil
}

func (d DoltgresConfigReader) ReadConfigArgs(args *argparser.ArgParseResults) (sqlserver.ServerConfig, error) {
	return sqlserver.NewCommandLineConfig(nil, args)
}

var _ sqlserver.ServerConfigReader = (*DoltgresConfigReader)(nil)
