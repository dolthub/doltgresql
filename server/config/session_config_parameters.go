package config

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"strings"
	"sync"
)

var _ sql.SystemVariableRegistry = (*sessionConfigParameters)(nil)

// sessionConfigParameters is the underlying type of SystemVariables.
type sessionConfigParameters struct {
	mutex      *sync.RWMutex
	sysVarVals map[string]sql.SystemVarValue
}

func (s *sessionConfigParameters) AddSystemVariables(sysVars []sql.SystemVariableInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, originalSysVar := range sysVars {
		sysVar := originalSysVar
		lowerName := strings.ToLower(sysVar.VarName())
		sysVar.SetName(lowerName)
		systemVars[lowerName] = sysVar
		s.sysVarVals[lowerName] = sql.SystemVarValue{
			Var: sysVar,
			Val: sysVar.GetDefault(),
		}
	}
}

func (s *sessionConfigParameters) AssignValues(vals map[string]interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for varName, val := range vals {
		varName = strings.ToLower(varName)
		sysVar, ok := systemVars[varName]
		if !ok {
			return sql.ErrUnknownSystemVariable.New(varName)
		}
		convertedVal, _, err := sysVar.VarType().Convert(val)
		if err != nil {
			return err
		}
		svv := sql.SystemVarValue{
			Var: sysVar,
			Val: convertedVal,
		}
		//if sysVar.NotifyChanged != nil {
		//	err := sysVar.NotifyChanged(sql.SystemVariableScope_Global, svv)
		//	if err != nil {
		//		return err
		//	}
		//}
		s.sysVarVals[varName] = svv
	}
	return nil
}

func (s *sessionConfigParameters) NewSessionMap() map[string]sql.SystemVarValue {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	sessionVals := make(map[string]sql.SystemVarValue, len(s.sysVarVals))
	for key, val := range s.sysVarVals {
		sessionVals[key] = val
	}
	return sessionVals
}

func (s *sessionConfigParameters) GetGlobal(name string) (sql.SystemVariableInterface, interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	name = strings.ToLower(name)
	v, ok := systemVars[name]
	if !ok {
		return nil, nil, false
	}

	//if v.ValueFunction != nil {
	//	result, err := v.ValueFunction()
	//	if err != nil {
	//		logrus.StandardLogger().Warnf("unable to get value for system variable %s: %s", name, err.Error())
	//		return v, nil, true
	//	}
	//	return v, result, true
	//}

	// convert any set types to strings
	sysVal := s.sysVarVals[name]
	if sysType, ok := v.VarType().(sql.SetType); ok {
		if sv, ok := sysVal.Val.(uint64); ok {
			var err error
			sysVal.Val, err = sysType.BitsToString(sv)
			if err != nil {
				return nil, nil, false
			}
		}
	}
	return v, sysVal.Val, true
}

// SetGlobal is temporarily used to set the session configuration parameter. There is no GLOBAL parameter in Postgres.
func (s *sessionConfigParameters) SetGlobal(name string, val interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	name = strings.ToLower(name)
	sysVar, ok := systemVars[name]
	if !ok {
		return sql.ErrUnknownSystemVariable.New(name)
	}
	switch sv := sysVar.(type) {
	case *sql.SystemVariable:
		if sv.Scope == sql.SystemVariableScope_Session {
			return sql.ErrSystemVariableSessionOnly.New(name)
		}
		if !sv.Dynamic || sv.ValueFunction != nil {
			return sql.ErrSystemVariableReadOnly.New(name)
		}
		convertedVal, _, err := sv.Type.Convert(val)
		if err != nil {
			return err
		}
		svv := sql.SystemVarValue{Var: sysVar, Val: convertedVal}
		if sv.NotifyChanged != nil {
			err := sv.NotifyChanged(sql.SystemVariableScope_Global, svv)
			if err != nil {
				return err
			}
		}
		s.sysVarVals[name] = svv
	case *Parameter:
		convertedVal, _, err := sv.Type.Convert(val)
		if err != nil {
			return err
		}
		s.sysVarVals[name] = sql.SystemVarValue{Var: sysVar, Val: convertedVal}
	}

	return nil
}

// GetAllGlobalVariables returns all SESSION configuration parameters.
func (s *sessionConfigParameters) GetAllGlobalVariables() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	m := make(map[string]interface{})
	for k, varVal := range s.sysVarVals {
		m[k] = varVal.Val
	}

	return m
}

var _ sql.SystemVariableRegistry = (*sessionConfigParameters)(nil)

// AddNecessaryMySQLSystemVariables adds some of MySQL system variables as they are frequently used. E.g. 'autocommit'
// TODO: support MySQL system parameters to the extent that we have to, but we'll eventually move completely away from them.
func AddNecessaryMySQLSystemVariables() {
	sql.SystemVariables.AddSystemVariables([]sql.SystemVariableInterface{
		// accessed before server starts
		&sql.SystemVariable{
			Name:              sql.AutoCommitSessionVar,
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemBoolType(sql.AutoCommitSessionVar),
			Default:           int8(1),
		},
		&sql.SystemVariable{
			Name:              "max_connections",
			Scope:             sql.SystemVariableScope_Global,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("max_connections", 1, 100000, false),
			Default:           int64(151),
		},
		&sql.SystemVariable{
			Name:              "net_write_timeout",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("net_write_timeout", 1, 9223372036854775807, false),
			Default:           int64(60),
		},
		&sql.SystemVariable{
			Name:              "net_read_timeout",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("net_read_timeout", 1, 9223372036854775807, false),
			Default:           int64(60),
		},
		&sql.SystemVariable{
			Name:              "secure_file_priv",
			Scope:             sql.SystemVariableScope_Global,
			Dynamic:           false,
			SetVarHintApplies: false,
			Type:              types.NewSystemStringType("secure_file_priv"),
			Default:           "",
		},
		// accessed after? server starts
		&sql.SystemVariable{
			Name:              "foreign_key_checks",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: true,
			Type:              types.NewSystemBoolType("foreign_key_checks"),
			Default:           int8(1),
		},
		&sql.SystemVariable{
			Name:              "sql_mode",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: true,
			Type:              types.NewSystemSetType("sql_mode", "ALLOW_INVALID_DATES", "ANSI_QUOTES", "ERROR_FOR_DIVISION_BY_ZERO", "HIGH_NOT_PRECEDENCE", "IGNORE_SPACE", "NO_AUTO_VALUE_ON_ZERO", "NO_BACKSLASH_ESCAPES", "NO_DIR_IN_CREATE", "NO_ENGINE_SUBSTITUTION", "NO_UNSIGNED_SUBTRACTION", "NO_ZERO_DATE", "NO_ZERO_IN_DATE", "ONLY_FULL_GROUP_BY", "PAD_CHAR_TO_FULL_LENGTH", "PIPES_AS_CONCAT", "REAL_AS_FLOAT", "STRICT_ALL_TABLES", "STRICT_TRANS_TABLES", "TIME_TRUNCATE_FRACTIONAL", "TRADITIONAL", "ANSI"),
			Default:           "NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
		},
	})
}
