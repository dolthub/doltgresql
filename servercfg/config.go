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

package servercfg

import (
	"github.com/cockroachdb/errors"
	doltservercfg "github.com/dolthub/dolt/go/libraries/doltcore/servercfg"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/go-mysql-server/sql"
	"gopkg.in/yaml.v2"

	"github.com/dolthub/doltgresql/server/analyzer"

	pgsql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

var ConfigHelp = "Supported fields in the config.yaml file, and their default values, " +
	"are as follows:\n\n" + DefaultServerConfig().String()

// DoltgresConfig is the configuration file for Doltgres.
type DoltgresConfig struct {
	cfgdetails.DoltgresConfig
}

var _ doltservercfg.ServerConfig = (*DoltgresConfig)(nil)

// Overrides implements the interface doltservercfg.ServerConfig.
func (*DoltgresConfig) Overrides() sql.EngineOverrides {
	return sql.EngineOverrides{
		Builder: sql.BuilderOverrides{
			ParseTableAsColumn: expression.NewTableToComposite,
			Parser:             pgsql.NewPostgresParser(),
		},
		SchemaFormatter:                 pgsql.NewPostgresSchemaFormatter(),
		CostedIndexScanExpressionFilter: &analyzer.LogicTreeWalker{},
	}
}

// ToSqlServerConfig returns this configuration struct as an implementation of the Dolt interface.
func (cfg *DoltgresConfig) ToSqlServerConfig() doltservercfg.ServerConfig {
	return cfg
}

// DefaultServerConfig creates a *DoltgresConfig that has all of the options set to their default values. Used when no
// config.yaml file is provided.
func DefaultServerConfig() *DoltgresConfig {
	internalCfg := cfgdetails.InternalDefaultServerConfig()
	cfg := &DoltgresConfig{}
	cfg.DoltgresConfig = *internalCfg
	return cfg
}

// ReadConfigFromYamlFile reads the given file from the file system at the specified path.
func ReadConfigFromYamlFile(fs filesys.Filesys, configFilePath string) (*DoltgresConfig, error) {
	configFileData, err := fs.ReadFile(configFilePath)
	if err != nil {
		absPath, absErr := fs.Abs(configFilePath)
		if absErr != nil {
			return nil, errors.Errorf("error reading config file '%s': %w", configFilePath, err)
		} else {
			return nil, errors.Errorf("error reading config file '%s': %w", absPath, err)
		}
	}
	return ConfigFromYamlData(configFileData)
}

// ConfigFromYamlData reads the configuration from the given file's bytes.
func ConfigFromYamlData(configFileData []byte) (*DoltgresConfig, error) {
	var internalCfg cfgdetails.DoltgresConfig
	err := yaml.UnmarshalStrict(configFileData, &internalCfg)
	if err != nil {
		return nil, errors.Errorf("error unmarshalling config data: %w", err)
	}
	cfg := &DoltgresConfig{}
	cfg.DoltgresConfig = internalCfg
	return cfg, err
}
