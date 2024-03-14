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
	"regexp"
	"strings"
	"time"

	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/variables"
)

// init initializes or appends to SystemVariables as it functions as a global variable.
// TODO: get rid of me, use an integration point to define new sysvars
func init() {
	InitConfigParameters()
}

// InitConfigParameters resets the postgresConfigVariables singleton in the sql package.
// Currently, we append all of postgres configuration parameters to sql.SystemVariables.
// This means that all of mysql system variables and postgres config parameters will be stored together.
// TODO: issue with this approach is that there are two postgres parameters
//
//	have the same name as mysql variables that will override the mysql variables.
func InitConfigParameters() {
	if sql.SystemVariables == nil {
		// unlikely this would happen since gms package does init()
		variables.InitSystemVariables()
	}
	params := make([]sql.SystemVariable, len(postgresConfigVariables))
	i := 0
	for _, sysVar := range postgresConfigVariables {
		params[i] = sysVar
		i++
	}
	sql.SystemVariables.AddSystemVariables(params)
	//AddNecessaryMySQLSystemVariables()
}

var (
	ErrInvalidValue          = errors.NewKind("ERROR:  invalid value for parameter \"%s\": \"%s\"")
	ErrCannotChangeAtRuntime = errors.NewKind("ERROR:  parameter \"%s\" cannot be changed now")
)

var _ sql.SystemVariable = (*Parameter)(nil)

type Parameter struct {
	Name      string
	Default   any
	Category  string
	ShortDesc string
	Context   ParameterContext
	Type      sql.Type
	Source    ParameterSource
	ResetVal  any
	Scope     sql.SystemVariableScope
	ValueFunc func(any) (any, bool)
}

// GetName implements sql.SystemVariable.
func (p *Parameter) GetName() string {
	return p.Name
}

// SetName implements sql.SystemVariable.
func (s *Parameter) SetName(n string) {
	s.Name = n
}

// GetType implements sql.SystemVariable.
func (p *Parameter) GetType() sql.Type {
	return p.Type
}

// SetDefault implements sql.SystemVariable.
func (p *Parameter) SetDefault(a any) {
	p.Default = a
}

// GetDefault implements sql.SystemVariable.
func (p *Parameter) GetDefault() any {
	return p.Default
}

// GetValue implements sql.SystemVariable.
func (p *Parameter) GetValue(a any) (any, bool) {
	// TODO: might need some validation or conversion for some parameters.
	return a, true
}

// HasDefaultValue implements sql.SystemVariable.
func (p *Parameter) HasDefaultValue(a any) bool {
	return p.Default == a
}

// AssignValue implements sql.SystemVariable.
func (p *Parameter) AssignValue(val any) (sql.SystemVarValue, error) {
	convertedVal, _, err := p.Type.Convert(val)
	if err != nil {
		return sql.SystemVarValue{}, err
	}
	if p.ValueFunc != nil {
		v, ok := p.ValueFunc(convertedVal)
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
	if p.IsReadOnly() {
		return sql.SystemVarValue{}, ErrCannotChangeAtRuntime.New(p.Name)
	}
	// TODO: Maybe do parsing of units for memory and time parameters?
	return p.AssignValue(val)
}

// IsReadOnly implements SystemVariable.
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

// IsGlobalOnly implements SystemVariable.
func (p *Parameter) IsGlobalOnly() bool {
	// TODO: fix - postgres SESSION parameters are considered as GLOBAL in gms for now.
	return true
}

// GetNotifyChanged implements SystemVariable.
func (p *Parameter) GetNotifyChanged() func(sql.SystemVariableScope, sql.SystemVarValue) error {
	// TODO: fix - some parameters might need them, but return nil for now.
	return nil
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

// offsetRegex is a regex for matching MySQL offsets (e.g. +01:00).
var offsetRegex = regexp.MustCompile(`(?m)^([+\-])(\d{2}):(\d{2})$`)

// MySQLOffsetToDuration takes in a MySQL timezone offset (e.g. "+01:00") and returns it as a time.Duration.
// If any problems are encountered, an error is returned.
func MySQLOffsetToDuration(d string) (time.Duration, error) {
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
