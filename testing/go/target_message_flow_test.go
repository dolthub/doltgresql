package _go

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TargetFlowStep keeps a protocol conversation intact while executing exactly
// one explicit expectation variant. Unlike replaying a query for validation,
// this is safe for steps that change session or transaction state.
type TargetFlowStep struct {
	Doltgres FlowStep
	Postgres FlowStep
}

func (s TargetFlowStep) describe() string {
	return s.Doltgres.describe()
}

func (s TargetFlowStep) runStep(r *messageFlowRunner) {
	runTargetPair(r.t, func(t *testing.T, native bool) {
		previous := r.t
		r.t = t
		defer func() { r.t = previous }()

		step := s.Doltgres
		if native {
			step = s.Postgres
		}
		step.runStep(r)
	})
}

type countingFlowStep struct {
	count *int
}

func (s countingFlowStep) describe() string { return "count" }
func (s countingFlowStep) runStep(r *messageFlowRunner) {
	*s.count++
	r.stepIdx++
}

func TestTargetFlowStepSelection(t *testing.T) {
	for _, target := range []string{"doltgres", "postgres", "postgres-fork"} {
		t.Run(target, func(t *testing.T) {
			t.Setenv("DOLTGRES_TEST_TARGET", target)
			var original, native int
			r := &messageFlowRunner{t: t}
			step := TargetFlowStep{
				Doltgres: countingFlowStep{count: &original},
				Postgres: countingFlowStep{count: &native},
			}
			step.runStep(r)
			require.Equal(t, 1, original+native)
			require.Equal(t, target == "doltgres", original == 1)
			require.Equal(t, 1, r.stepIdx)
			require.Same(t, t, r.t)
		})
	}
}
