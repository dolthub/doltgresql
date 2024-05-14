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

	"github.com/dolthub/dolt/go/cmd/dolt/commands/engine"
	"github.com/dolthub/dolt/go/cmd/dolt/commands/sqlserver"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/cluster"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"gopkg.in/yaml.v2"
)

func Ptr[T any](v T) *T {
	return &v
}

const (
	maxConnectionsKey = "max_connections"
	readTimeoutKey    = "net_read_timeout"
	writeTimeoutKey   = "net_write_timeout"
	eventSchedulerKey = "event_scheduler"

	OverrideDataDirKey = "data_dir"
)

var ConfigHelp = `Place holder. This will be replaced with details on config.yaml options`

type PostgresReplicationConfig struct {
	PostgresServerAddress *string `yaml:"postgres_server_address,omitempty" minver:"0.7.4"`
	PostgresUser          *string `yaml:"postgres_user,omitempty" minver:"0.7.4"`
	PostgresPassword      *string `yaml:"postgres_password,omitempty" minver:"0.7.4"`
	PostgresDatabase      *string `yaml:"postgres_database,omitempty" minver:"0.7.4"`
	PostgresPort          *int    `yaml:"postgres_port,omitempty" minver:"0.7.4"`
	SlotName              *string `yaml:"slot_name,omitempty" minver:"0.7.4"`
}

// BehaviorYAMLConfig contains server configuration regarding how the server should behave
type DoltgresBehaviorConfig struct {
	ReadOnly *bool `yaml:"read_only,omitempty" minver:"0.7.4"`
	// Disable processing CLIENT_MULTI_STATEMENTS support on the
	// sql server.  Dolt's handling of CLIENT_MULTI_STATEMENTS is currently
	// broken. If a client advertises to support it (mysql cli client
	// does), and then sends statements that contain embedded unquoted ';'s
	// (such as a CREATE TRIGGER), then those incoming queries will be
	// misprocessed.
	DisableClientMultiStatements *bool `yaml:"disable_client_multi_statements,omitempty" minver:"0.7.4"`
	// DoltTransactionCommit enables the @@dolt_transaction_commit system variable, which
	// automatically creates a Dolt commit when any SQL transaction is committed.
	DoltTransactionCommit *bool `yaml:"dolt_transaction_commit,omitempty" minver:"0.7.4"`
}

type DoltgresUserConfig struct {
	Name     *string `yaml:"name,omitempty" minver:"0.7.4"`
	Password *string `yaml:"password,omitempty" minver:"0.7.4"`
}

// DoltgresListenerConfig contains information on the network connection that the server will open
type DoltgresListenerConfig struct {
	HostStr            *string `yaml:"host,omitempty" minver:"0.7.4"`
	PortNumber         *int    `yaml:"port,omitempty" minver:"0.7.4"`
	ReadTimeoutMillis  *uint64 `yaml:"read_timeout_millis,omitempty" minver:"0.7.4"`
	WriteTimeoutMillis *uint64 `yaml:"write_timeout_millis,omitempty" minver:"0.7.4"`
	// TLSKey is a file system path to an unencrypted private TLS key in PEM format.
	TLSKey *string `yaml:"tls_key,omitempty" minver:"0.7.4"`
	// TLSCert is a file system path to a TLS certificate chain in PEM format.
	TLSCert *string `yaml:"tls_cert,omitempty" minver:"0.7.4"`
	// RequireSecureTransport can enable a mode where non-TLS connections are turned away.
	RequireSecureTransport *bool `yaml:"require_secure_transport,omitempty" minver:"0.7.4"`
	// AllowCleartextPasswords enables use of cleartext passwords.
	AllowCleartextPasswords *bool `yaml:"allow_cleartext_passwords,omitempty" minver:"0.7.4"`
	// Socket is unix socket file path
	Socket *string `yaml:"socket,omitempty" minver:"0.7.4"`
}

// DoltgresPerformanceConfig contains configuration parameters for performance tweaking
type DoltgresPerformanceConfig struct {
	QueryParallelism *int `yaml:"query_parallelism,omitempty" minver:"0.7.4"`
}

type DoltgesMetricsConfig struct {
	Labels map[string]string `yaml:"labels,omitempty" minver:"0.7.4"`
	Host   *string           `yaml:"host,omitempty" minver:"0.7.4"`
	Port   *int              `yaml:"port,omitempty" minver:"0.7.4"`
}

type DoltgresRemotesapiConfig struct {
	Port     *int  `yaml:"port,omitempty" minver:"0.7.4"`
	ReadOnly *bool `yaml:"read_only,omitempty" minver:"0.7.4"`
}

type DoltgresUserSessionVars struct {
	Name string            `yaml:"name"`
	Vars map[string]string `yaml:"vars,omitempty"`
}

type DoltgresConfig struct {
	LogLevelStr       *string                    `yaml:"log_level,omitempty" minver:"0.7.4"`
	MaxLenInLogs      *int                       `yaml:"max_query_len_in_logs,omitempty" minver:"0.7.4"`
	EncodeLoggedQuery *bool                      `yaml:"encode_logged_query,omitempty" minver:"0.7.4"`
	BehaviorConfig    *DoltgresBehaviorConfig    `yaml:"behavior,omitempty" minver:"0.7.4"`
	UserConfig        *DoltgresUserConfig        `yaml:"user,omitempty" minver:"0.7.4"`
	ListenerConfig    *DoltgresListenerConfig    `yaml:"listener,omitempty" minver:"0.7.4"`
	PerformanceConfig *DoltgresPerformanceConfig `yaml:"performance,omitempty" minver:"0.7.4"`
	DataDirStr        *string                    `yaml:"data_dir,omitempty" minver:"0.7.4"`
	CfgDirStr         *string                    `yaml:"cfg_dir,omitempty" minver:"0.7.4"`
	MetricsConfig     *DoltgesMetricsConfig      `yaml:"metrics,omitempty" minver:"0.7.4"`
	RemotesapiConfig  *DoltgresRemotesapiConfig  `yaml:"remotesapi,omitempty" minver:"0.7.4"`
	PrivilegeFile     *string                    `yaml:"privilege_file,omitempty" minver:"0.7.4"`
	BranchControlFile *string                    `yaml:"branch_control_file,omitempty" minver:"0.7.4"`

	// TODO: Rename to UserVars_
	Vars            []DoltgresUserSessionVars `yaml:"user_session_vars,omitempty" minver:"0.7.4"`
	SystemVariables *engine.SystemVariables   `yaml:"system_variables,omitempty" minver:"0.7.4"`
	Jwks            []engine.JwksConfig       `yaml:"jwks,omitempty" minver:"0.7.4"`
	GoldenMysqlConn *string                   `yaml:"golden_mysql_conn,omitempty" minver:"0.7.4"`

	PostgresReplicationConfig *PostgresReplicationConfig `yaml:"postgres_replication,omitempty" minver:"0.7.4"`
}

func (cfg *DoltgresConfig) AutoCommit() bool {
	return true
}

func (cfg *DoltgresConfig) DoltTransactionCommit() bool {
	if cfg.BehaviorConfig == nil || cfg.BehaviorConfig.DoltTransactionCommit == nil {
		return false
	}

	return *cfg.BehaviorConfig.DoltTransactionCommit
}

func (cfg *DoltgresConfig) DataDir() string {
	if cfg.DataDirStr == nil {
		return ""
	}

	return *cfg.DataDirStr
}

func (cfg *DoltgresConfig) CfgDir() string {
	if cfg.CfgDirStr == nil {
		return ""
	}

	return *cfg.CfgDirStr
}

func (cfg *DoltgresConfig) Host() string {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.HostStr == nil {
		return "localhost"
	}

	return *cfg.ListenerConfig.HostStr
}

func (cfg *DoltgresConfig) Port() int {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.PortNumber == nil {
		return 5432
	}

	return *cfg.ListenerConfig.PortNumber
}

func (cfg *DoltgresConfig) User() string {
	if cfg.UserConfig == nil || cfg.UserConfig.Name == nil {
		return "doltgres"
	}

	return *cfg.UserConfig.Name
}

func (cfg *DoltgresConfig) Password() string {
	if cfg.UserConfig == nil || cfg.UserConfig.Password == nil {
		return ""
	}

	return *cfg.UserConfig.Password
}

func (cfg *DoltgresConfig) ReadTimeout() uint64 {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.ReadTimeoutMillis == nil {
		return 0
	}

	return *cfg.ListenerConfig.ReadTimeoutMillis
}

func (cfg *DoltgresConfig) WriteTimeout() uint64 {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.WriteTimeoutMillis == nil {
		return 0
	}

	return *cfg.ListenerConfig.WriteTimeoutMillis
}

func (cfg *DoltgresConfig) ReadOnly() bool {
	if cfg.BehaviorConfig == nil || cfg.BehaviorConfig.ReadOnly == nil {
		return false
	}

	return *cfg.BehaviorConfig.ReadOnly
}

func (cfg *DoltgresConfig) LogLevel() sqlserver.LogLevel {
	if cfg.LogLevelStr == nil {
		return sqlserver.LogLevel_Info
	}

	switch *cfg.LogLevelStr {
	case "trace":
		return sqlserver.LogLevel_Trace
	case "debug":
		return sqlserver.LogLevel_Debug
	case "info":
		return sqlserver.LogLevel_Info
	case "warn":
		return sqlserver.LogLevel_Warning
	case "error":
		return sqlserver.LogLevel_Error
	case "fatal":
		return sqlserver.LogLevel_Fatal
	default:
		return sqlserver.LogLevel_Info
	}
}

func (cfg *DoltgresConfig) MaxConnections() uint64 {
	return 0
}

func (cfg *DoltgresConfig) QueryParallelism() int {
	if cfg.PerformanceConfig == nil || cfg.PerformanceConfig.QueryParallelism == nil {
		return 1
	}

	return *cfg.PerformanceConfig.QueryParallelism
}

func (cfg *DoltgresConfig) TLSKey() string {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.TLSKey == nil {
		return ""
	}

	return *cfg.ListenerConfig.TLSKey
}

func (cfg *DoltgresConfig) TLSCert() string {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.TLSCert == nil {
		return ""
	}

	return *cfg.ListenerConfig.TLSCert
}

func (cfg *DoltgresConfig) RequireSecureTransport() bool {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.RequireSecureTransport == nil {
		return false
	}

	return *cfg.ListenerConfig.RequireSecureTransport
}

func (cfg *DoltgresConfig) MaxLoggedQueryLen() int {
	if cfg.MaxLenInLogs == nil {
		return 1000
	}

	return *cfg.MaxLenInLogs
}

func (cfg *DoltgresConfig) ShouldEncodeLoggedQuery() bool {
	if cfg.EncodeLoggedQuery == nil {
		return false
	}

	return *cfg.EncodeLoggedQuery
}

func (cfg *DoltgresConfig) PersistenceBehavior() string {
	return "load"
}

func (cfg *DoltgresConfig) DisableClientMultiStatements() bool {
	if cfg.BehaviorConfig == nil || cfg.BehaviorConfig.DisableClientMultiStatements == nil {
		return false
	}

	return *cfg.BehaviorConfig.DisableClientMultiStatements
}

func (cfg *DoltgresConfig) MetricsLabels() map[string]string {
	if cfg.MetricsConfig == nil {
		return nil
	}

	return cfg.MetricsConfig.Labels
}

func (cfg *DoltgresConfig) MetricsHost() string {
	if cfg.MetricsConfig == nil || cfg.MetricsConfig.Host == nil {
		return ""
	}

	return *cfg.MetricsConfig.Host
}

func (cfg *DoltgresConfig) MetricsPort() int {
	if cfg.MetricsConfig == nil || cfg.MetricsConfig.Port == nil {
		return 0
	}

	return *cfg.MetricsConfig.Port
}

func (cfg *DoltgresConfig) PrivilegeFilePath() string {
	if cfg.PrivilegeFile == nil {
		return ""
	}

	return *cfg.PrivilegeFile
}

func (cfg *DoltgresConfig) BranchControlFilePath() string {
	if cfg.BranchControlFile == nil {
		return ""
	}

	return *cfg.BranchControlFile
}

func (cfg *DoltgresConfig) UserVars() []sqlserver.UserSessionVars {
	var userVars []sqlserver.UserSessionVars
	for _, uv := range cfg.Vars {
		userVars = append(userVars, sqlserver.UserSessionVars{
			Name: uv.Name,
			Vars: uv.Vars,
		})
	}

	return userVars
}

func (cfg *DoltgresConfig) SystemVars() engine.SystemVariables {
	if cfg.SystemVariables == nil {
		return engine.SystemVariables{}
	}

	return *cfg.SystemVariables
}

func (cfg *DoltgresConfig) JwksConfig() []engine.JwksConfig {
	return cfg.Jwks
}

func (cfg *DoltgresConfig) AllowCleartextPasswords() bool {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.AllowCleartextPasswords == nil {
		return false
	}

	return *cfg.ListenerConfig.AllowCleartextPasswords
}

func (cfg *DoltgresConfig) Socket() string {
	if cfg.ListenerConfig == nil || cfg.ListenerConfig.Socket == nil {
		return ""
	}

	return *cfg.ListenerConfig.Socket
}

func (cfg *DoltgresConfig) RemotesapiPort() *int {
	if cfg.RemotesapiConfig == nil {
		return nil
	}

	return cfg.RemotesapiConfig.Port
}

func (cfg *DoltgresConfig) RemotesapiReadOnly() *bool {
	if cfg.RemotesapiConfig == nil {
		return nil
	}

	return cfg.RemotesapiConfig.ReadOnly
}

func (cfg *DoltgresConfig) ClusterConfig() cluster.Config {
	return nil
}

func (cfg *DoltgresConfig) EventSchedulerStatus() string {
	return "OFF"
}

func (cfg *DoltgresConfig) ValueSet(value string) bool {
	switch value {
	case readTimeoutKey:
		return cfg.ListenerConfig != nil && cfg.ListenerConfig.ReadTimeoutMillis != nil
	case writeTimeoutKey:
		return cfg.ListenerConfig != nil && cfg.ListenerConfig.WriteTimeoutMillis != nil
	case maxConnectionsKey:
		return false
	case eventSchedulerKey:
		return false
	}

	return false
}

func (cfg *DoltgresConfig) ToSqlServerConfig() sqlserver.ServerConfig {
	return cfg
}

func ReadConfigFromYamlFile(fs filesys.Filesys, configFilePath string, overrides map[string]string) (*DoltgresConfig, error) {
	configFileData, err := fs.ReadFile(configFilePath)
	if err != nil {
		absPath, absErr := fs.Abs(configFilePath)
		if absErr != nil {
			return nil, fmt.Errorf("error reading config file '%s': %w", configFilePath, err)
		} else {
			return nil, fmt.Errorf("error reading config file '%s': %w", absPath, err)
		}
	}
	return ConfigFromYamlData(configFileData, overrides)
}

func ConfigFromYamlData(configFileData []byte, overrides map[string]string) (*DoltgresConfig, error) {
	var cfg DoltgresConfig
	err := yaml.UnmarshalStrict(configFileData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config data: %w", err)
	}

	for k, v := range overrides {
		switch k {
		case OverrideDataDirKey:
			cfg.DataDirStr = &v
		default:
			// this only happens if code to add an override is added but code to handle the override is not.
			panic(fmt.Sprintf("unknown override key: %s", k))
		}
	}

	return &cfg, err
}
