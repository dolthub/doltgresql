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

package utils

import (
	"sync"

	"github.com/pkg/profile"
)

// prof is the global profiler that is in use. This should only be interacted with via StartProfiling and StopProfiling.
var prof interface{ Stop() }

// profMutex is the mutex that handles concurrent manipulations of prof.
var profMutex = &sync.Mutex{}

// ProfilingOptions contains all options for starting the profiler.
type ProfilingOptions struct {
	CPU    bool
	Memory bool
	Block  bool
	Trace  bool
	Path   string
}

// StartProfiling starts profiling the CPU.
func StartProfiling(options ProfilingOptions) {
	profMutex.Lock()
	defer profMutex.Unlock()

	if !options.HasOptions() {
		return
	}
	profOpts := []func(p *profile.Profile){profile.NoShutdownHook}
	if options.CPU {
		profOpts = append(profOpts, profile.CPUProfile)
	}
	if options.Memory {
		profOpts = append(profOpts, profile.MemProfile)
	}
	if options.Block {
		profOpts = append(profOpts, profile.BlockProfile)
	}
	if options.Trace {
		profOpts = append(profOpts, profile.TraceProfile)
	}
	if len(options.Path) > 0 {
		profOpts = append(profOpts, profile.ProfilePath(options.Path))
	}
	prof = profile.Start(profOpts...)
}

// StopProfiling stops profiling the CPU.
func StopProfiling() {
	profMutex.Lock()
	defer profMutex.Unlock()

	if prof != nil {
		prof.Stop()
		prof = nil
	}
}

// HasOptions returns true when at least one profiler target has been selected.
func (options ProfilingOptions) HasOptions() bool {
	return options.CPU || options.Memory || options.Block || options.Trace
}
