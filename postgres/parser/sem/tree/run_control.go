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

// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "fmt"

// ControlJobs represents a PAUSE/RESUME/CANCEL JOBS statement.
type ControlJobs struct {
	Jobs    *Select
	Command JobCommand
}

// JobCommand determines which type of action to effect on the selected job(s).
type JobCommand int

// JobCommand values
const (
	PauseJob JobCommand = iota
	CancelJob
	ResumeJob
)

// JobCommandToStatement translates a job command integer to a statement prefix.
var JobCommandToStatement = map[JobCommand]string{
	PauseJob:  "PAUSE",
	CancelJob: "CANCEL",
	ResumeJob: "RESUME",
}

// Format implements the NodeFormatter interface.
func (n *ControlJobs) Format(ctx *FmtCtx) {
	ctx.WriteString(JobCommandToStatement[n.Command])
	ctx.WriteString(" JOBS ")
	ctx.FormatNode(n.Jobs)
}

// CancelQueries represents a CANCEL QUERIES statement.
type CancelQueries struct {
	Queries  *Select
	IfExists bool
}

// Format implements the NodeFormatter interface.
func (node *CancelQueries) Format(ctx *FmtCtx) {
	ctx.WriteString("CANCEL QUERIES ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Queries)
}

// CancelSessions represents a CANCEL SESSIONS statement.
type CancelSessions struct {
	Sessions *Select
	IfExists bool
}

// Format implements the NodeFormatter interface.
func (node *CancelSessions) Format(ctx *FmtCtx) {
	ctx.WriteString("CANCEL SESSIONS ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Sessions)
}

// ScheduleCommand determines which type of action to effect on the selected job(s).
type ScheduleCommand int

// ScheduleCommand values
const (
	PauseSchedule ScheduleCommand = iota
	ResumeSchedule
	DropSchedule
)

func (c ScheduleCommand) String() string {
	switch c {
	case PauseSchedule:
		return "PAUSE"
	case ResumeSchedule:
		return "RESUME"
	case DropSchedule:
		return "DROP"
	default:
		panic("unhandled schedule command")
	}
}

// ControlSchedules represents PAUSE/RESUME SCHEDULE statement.
type ControlSchedules struct {
	Schedules *Select
	Command   ScheduleCommand
}

var _ Statement = &ControlSchedules{}

// Format implements NodeFormatter interface
func (n *ControlSchedules) Format(ctx *FmtCtx) {
	fmt.Fprintf(ctx, "%s SCHEDULES ", n.Command)
	n.Schedules.Format(ctx)
}

// ControlJobsForSchedules represents PAUSE/RESUME/CANCEL clause
// which applies job command to the jobs matching specified schedule(s).
type ControlJobsForSchedules struct {
	Schedules *Select
	Command   JobCommand
}

// Format implements NodeFormatter interface
func (n *ControlJobsForSchedules) Format(ctx *FmtCtx) {
	fmt.Fprintf(ctx, "%s JOBS FOR SCHEDULES %s",
		JobCommandToStatement[n.Command], AsString(n.Schedules))
}

var _ Statement = &ControlJobsForSchedules{}
