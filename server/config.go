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
	"os"

	"gopkg.in/yaml.v2"
)

/* type DoltgresServerConfig struct {
	*sqlserver.YAMLConfig
	PostgresReplicationConfig *PostgresReplicationConfig `yaml:"postgres_replication,omitempty"`
}*/

type PostgresReplicationConfig struct {
	PostgresServerAddress string `yaml:"postgres_server_address"`
	PostgresUser          string `yaml:"postgres_user"`
	PostgresPassword      string `yaml:"postgres_password"`
	PostgresDatabase      string `yaml:"postgres_database"`
	PostgresPort          int    `yaml:"postgres_port"`
	SlotName              string `yaml:"slot_name"`
}

type Config struct {
	PostgresReplicationConfig *PostgresReplicationConfig `yaml:"postgres_replication,omitempty" minver:"TBD"`
}

func (cfg *Config) ToSqlServerConfig() sqlserver.ServerConfig {
	return &sqlserver.YAMLConfig{}
}

func ReadConfigFromYamlFile(configFilePath string) (*Config, error) {
	configFileData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file '%s': %w", configFilePath, err)
	}
	return ConfigFromYamlData(configFileData)
}

func ConfigFromYamlData(configFileData []byte) (*Config, error) {
	var cfg Config
	err := yaml.UnmarshalStrict(configFileData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config data: %w", err)
	}
	return &cfg, err
}
