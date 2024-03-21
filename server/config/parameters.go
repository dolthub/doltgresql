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

package config

import (
	"fmt"
	"gopkg.in/src-d/go-errors.v1"
	"regexp"
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/variables"
)

// init initializes or appends to SystemVariables as it functions as a global variable.
// Currently, we append all of postgres configuration parameters to sql.SystemVariables.
// This means that all of mysql system variables and postgres config parameters will be stored together.
// TODO: get rid of me, use an integration point to define new sysvars
func init() {
	// There are two postgres parameters have the same name as mysql variables
	// TODO: issue with this approach is that those parameters will override the mysql variables.
	if sql.SystemVariables == nil {
		// unlikely this would happen since init() in gms package is executed first
		variables.InitSystemVariables()
	}
	params := make([]sql.SystemVariable, len(postgresConfigParameters))
	i := 0
	for _, sysVar := range postgresConfigParameters {
		params[i] = sysVar
		i++
	}
	sql.SystemVariables.AddSystemVariables(params)
}

var (
	ErrInvalidValue          = errors.NewKind("ERROR:  invalid value for parameter \"%s\": \"%s\"")
	ErrCannotChangeAtRuntime = errors.NewKind("ERROR:  parameter \"%s\" cannot be changed now")
)

var _ sql.SystemVariable = (*Parameter)(nil)

type Parameter struct {
	Name         string
	Default      any
	Category     string
	ShortDesc    string
	Context      ParameterContext
	Type         sql.Type
	Source       ParameterSource
	ResetVal     any
	Scope        sql.SystemVariableScope
	ValidateFunc func(any) (any, bool)
}

// GetName implements sql.SystemVariable.
func (p *Parameter) GetName() string {
	return p.Name
}

// GetType implements sql.SystemVariable.
func (p *Parameter) GetType() sql.Type {
	return p.Type
}

// GetSessionScope implements sql.SystemVariable.
func (p *Parameter) GetSessionScope() sql.SystemVariableScope {
	return GetPgsqlScope(PsqlScopeSession)
}

// SetDefault implements sql.SystemVariable.
func (p *Parameter) SetDefault(a any) {
	p.Default = a
}

// GetDefault implements sql.SystemVariable.
func (p *Parameter) GetDefault() any {
	return p.Default
}

// InitValue implements sql.SystemVariable.
func (p *Parameter) InitValue(val any, global bool) (sql.SystemVarValue, error) {
	if global {
		// This shouldn't happen, but sanity check
		return sql.SystemVarValue{}, sql.ErrSystemVariableSessionOnly.New(p.Name)
	}
	convertedVal, _, err := p.Type.Convert(val)
	if err != nil {
		return sql.SystemVarValue{}, err
	}
	if p.ValidateFunc != nil {
		v, ok := p.ValidateFunc(convertedVal)
		if !ok {
			return sql.SystemVarValue{}, ErrInvalidValue.New(p.Name, convertedVal)
		}
		convertedVal = v
	}
	svv := sql.SystemVarValue{
		Var: p,
		Val: convertedVal,
	}
	return svv, nil
}

// SetValue implements sql.SystemVariable.
func (p *Parameter) SetValue(val any, global bool) (sql.SystemVarValue, error) {
	if global {
		// This shouldn't happen, but sanity check
		return sql.SystemVarValue{}, sql.ErrSystemVariableSessionOnly.New(p.Name)
	}
	if p.IsReadOnly() {
		return sql.SystemVarValue{}, ErrCannotChangeAtRuntime.New(p.Name)
	}
	// TODO: Do parsing of units for memory and time parameters
	return p.InitValue(val, global)
}

// IsReadOnly implements sql.SystemVariable.
func (p *Parameter) IsReadOnly() bool {
	switch strings.ToLower(p.Name) {
	case "server_version", "server_encoding", "lc_collate", "lc_ctype", "is_superuser":
		return true
	}
	switch p.Context {
	case ParameterContextInternal, ParameterContextPostmaster, ParameterContextSighup,
		ParameterContextSuperUserBackend, ParameterContextBackend:
		// Read the docs above the ParameterContext
		// TODO: some of above contexts need support, return error for now
		return true
	case ParameterContextSuperUser, ParameterContextUser:
		// TODO: need to check for 'superuser' and appropriate 'SET' privileges.
		// Can be set from `postgresql.conf`, or within a session via the `SET` command.
		return false
	}
	return false
}

// IsGlobalOnly implements sql.SystemVariable.
func (p *Parameter) IsGlobalOnly() bool {
	return false
}

// DisplayString implements sql.SystemVariable.
func (p *Parameter) DisplayString(_ string) string {
	return p.Name
}

// ParameterContext sets level of difficulty of changing the parameter settings.
// For more detailed description on how to change the settings of specific context,
// https://www.postgresql.org/docs/current/view-pg-settings.html
type ParameterContext string

// The following constants are in order of decreasing difficulty of changing the setting.
const (
	ParameterContextInternal         ParameterContext = "internal"
	ParameterContextPostmaster       ParameterContext = "postmaster"
	ParameterContextSighup           ParameterContext = "sighup"
	ParameterContextSuperUserBackend ParameterContext = "superuser-backend"
	ParameterContextBackend          ParameterContext = "backend"
	ParameterContextSuperUser        ParameterContext = "superuser"
	ParameterContextUser             ParameterContext = "user"
)

// ParameterSource sets the source of the current parameter value.
type ParameterSource string

const (
	ParameterSourceClient ParameterSource = "client"
	// ParameterSourceConfigurationFile means that the parameter needs to set
	// its Default and ResetVal to what's defined in the given config file.
	ParameterSourceConfigurationFile ParameterSource = "configuration file"
	ParameterSourceDefault           ParameterSource = "default"
	// ParameterSourceOverride means the default and reset value needs to be set at server start time
	// TODO: currently the default and reset values are dummy values.
	ParameterSourceOverride ParameterSource = "override"
)

var _ sql.SystemVariableScope = (*PgsqlScope)(nil)

// PgsqlScope represents the scope of a PostgreSQL configuration parameter.
type PgsqlScope struct {
	Type PgsqlScopeType
}

func GetPgsqlScope(t PgsqlScopeType) sql.SystemVariableScope {
	return &PgsqlScope{Type: t}
}

// SetValue implements sql.SystemVariableScope.
func (p *PgsqlScope) SetValue(ctx *sql.Context, name string, val any) error {
	switch p.Type {
	case PsqlScopeSession:
		err := ctx.SetSessionVariable(ctx, name, val)
		return err
	case PsqlScopeLocal:
		// TODO: support LOCAL scope
		return nil
	default:
		return fmt.Errorf("unable to set `%s` due to unknown scope `%v`", name, p.Type)
	}
}

// GetValue implements sql.SystemVariableScope.
func (p *PgsqlScope) GetValue(ctx *sql.Context, name string, _ sql.CollationID) (any, error) {
	switch p.Type {
	case PsqlScopeSession:
		val, err := ctx.GetSessionVariable(ctx, name)
		if err != nil {
			return nil, err
		}
		return val, nil
	case PsqlScopeLocal:
		// TODO: support LOCAL scope
		return nil, fmt.Errorf("unsupported scope `%v` on system variable `%s`", p.Type, name)
	default:
		return nil, fmt.Errorf("unknown scope `%v` on system variable `%s`", p.Type, name)
	}
}

// IsGlobalOnly implements sql.SystemVariableScope.
func (p *PgsqlScope) IsGlobalOnly() bool {
	// In Postgres, there is no GLOBAL scope.
	return false
}

// IsSessionOnly implements sql.SystemVariableScope.
func (p *PgsqlScope) IsSessionOnly() bool {
	return p.Type == PsqlScopeSession
}

// PgsqlScopeType represents the scope of a configuration parameter.
type PgsqlScopeType byte

const (
	// PsqlScopeSession is set when the system variable exists only in the session context.
	PsqlScopeSession PgsqlScopeType = iota
	// PsqlScopeLocal is set when the system variable exists only in the local context.
	PsqlScopeLocal
)

// offsetRegex is a regex for matching MySQL offsets (e.g. +01:00).
var offsetRegex = regexp.MustCompile(`(?m)^([+\-])(\d{2}):(\d{2})$`)

// TzOffsetToDuration takes in a timezone offset (e.g. "+01:00") and returns it as a time.Duration.
// If any problems are encountered, an error is returned.
func TzOffsetToDuration(d string) (time.Duration, error) {
	matches := offsetRegex.FindStringSubmatch(d)
	if len(matches) == 4 {
		symbol := matches[1]
		hours := matches[2]
		mins := matches[3]
		return time.ParseDuration(symbol + hours + "h" + mins + "m")
	} else {
		return -1, fmt.Errorf("error: unable to process time")
	}
}
