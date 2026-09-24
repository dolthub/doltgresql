// Copyright 2026 Dolthub, Inc.
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

package core

import (
	"context"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/sessionstate"
)

// settingValue retains existence separately from value. A custom parameter
// that has been reset exists with an empty value; an unseen name does not.
type settingValue struct {
	value  any
	exists bool
}

// sessionSetting journals each parameter independently so a session change
// to one setting cannot persist a LOCAL change to another setting.
type sessionSetting struct {
	journal  sessionstate.Journal[settingValue]
	postgres bool
}

// Setting returns a scoped PostgreSQL value. The bool distinguishes an
// unknown custom setting from a custom setting reset to the empty string.
func Setting(ctx *sql.Context, name string) (any, bool, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, false, err
	}
	setting, ok := cv.settings[strings.ToLower(name)]
	if !ok {
		return nil, false, nil
	}
	value := setting.journal.Current()
	return value.value, value.exists, nil
}

// SettingNames returns the known parameters introduced through PostgreSQL
// SET in this session so RESET ALL can restore only changed settings.
func SettingNames(ctx *sql.Context) ([]string, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cv.settings))
	for name, setting := range cv.settings {
		if setting.journal.Current().exists {
			names = append(names, name)
		}
	}
	return names, nil
}

// SetSetting records a custom parameter in its session journal. Callers own
// parameter validation before reaching this API.
func SetSetting(ctx *sql.Context, name string, value any, local bool) error {
	return setSetting(ctx, name, value, local, false)
}

// SetPostgresSetting journals a validated built-in and writes its effective
// value to the session store used by SHOW, the planner, and the executor.
func SetPostgresSetting(ctx *sql.Context, name string, value any, local bool) error {
	return setSetting(ctx, name, value, local, true)
}

// setSetting initializes a parameter journal at the current transaction and
// savepoint scope, including parameters first changed after a savepoint.
func setSetting(ctx *sql.Context, name string, value any, local, postgres bool) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	if local && !cv.identity.InTransaction() {
		return nil
	}
	name = strings.ToLower(name)
	setting, ok := cv.settings[name]
	if !ok {
		initial := settingValue{}
		if postgres {
			initial.value, err = ctx.GetSessionVariable(ctx, name)
			if err != nil {
				return err
			}
			initial.exists = true
		}
		setting = &sessionSetting{journal: sessionstate.NewJournal(initial), postgres: postgres}
		if cv.identity.InTransaction() {
			setting.journal.Begin()
			for _, savepoint := range cv.settingsSavepoints {
				setting.journal.Savepoint(savepoint)
			}
		}
		if cv.settings == nil {
			cv.settings = make(map[string]*sessionSetting)
		}
		cv.settings[name] = setting
	}
	if postgres {
		if err := ctx.SetSessionVariable(ctx, name, value); err != nil {
			return err
		}
		cv.dateOutputFormat = ""
	}
	next := settingValue{value: value, exists: true}
	if local {
		setting.journal.SetLocal(next)
	} else {
		setting.journal.SetSession(next)
	}
	return nil
}

// restoreSettings projects restored built-in values back into the session.
// Lifecycle callbacks cannot return errors, so retain any restoration error
// and report it at the next access to the session payload.
func (cv *contextValues) restoreSettings() {
	ctx := sql.NewContext(context.Background(), sql.WithSession(cv.session))
	for name, setting := range cv.settings {
		if setting.postgres {
			if err := cv.session.SetSessionVariable(ctx, name, setting.journal.Current().value); err != nil {
				cv.settingsRestoreErr = err
				return
			}
		}
	}
	cv.dateOutputFormat = ""
}

// trimSettingsSavepoints mirrors PostgreSQL's savepoint shadowing and release
// rules for journals first initialized later in the transaction.
func (cv *contextValues) trimSettingsSavepoints(name string, release bool) {
	for i := len(cv.settingsSavepoints) - 1; i >= 0; i-- {
		if cv.settingsSavepoints[i] == name {
			if !release {
				i++
			}
			cv.settingsSavepoints = cv.settingsSavepoints[:i]
			return
		}
	}
}
