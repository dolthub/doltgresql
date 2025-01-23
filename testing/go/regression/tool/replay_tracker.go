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

package main

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/utils"
)

// ReplayTracker tracks data for a Replay run.
type ReplayTracker struct {
	File             string
	Success          uint32
	PartialSuccess   uint32
	Failed           uint32
	SuccessItems     []ReplayTrackerItem
	FailPartialItems []ReplayTrackerItem
}

// ReplayTrackerItem specifically tracks partial successes and failures for queries.
type ReplayTrackerItem struct {
	Query           string
	PartialSuccess  []string
	UnexpectedError string
	ExpectedError   string
}

// NewReplayTracker returns a new *ReplayTracker.
func NewReplayTracker(file string) *ReplayTracker {
	return &ReplayTracker{
		File:             strings.ReplaceAll(filepath.Base(file), ".results", ""),
		Success:          0,
		PartialSuccess:   0,
		Failed:           0,
		SuccessItems:     nil,
		FailPartialItems: nil,
	}
}

// AddSuccess adds the given ReplayTrackerItem as a Success.
func (rt *ReplayTracker) AddSuccess(item ReplayTrackerItem) {
	rt.SuccessItems = append(rt.SuccessItems, item)
}

// AddFailure adds the given ReplayTrackerItem as a Failure (or Partial Success).
func (rt *ReplayTracker) AddFailure(item ReplayTrackerItem) {
	rt.FailPartialItems = append(rt.FailPartialItems, item)
}

// SerializeTrackers serializes the given trackers.
func SerializeTrackers(trackers ...*ReplayTracker) []byte {
	sort.Slice(trackers, func(i, j int) bool {
		return trackers[i].File < trackers[j].File
	})
	writer := utils.NewWriter(1048576)
	writer.Uint32(2) // Version
	writer.Uint32(uint32(len(trackers)))
	for _, tracker := range trackers {
		writer.String(tracker.File)
		writer.Uint32(tracker.Success)
		writer.Uint32(tracker.PartialSuccess)
		writer.Uint32(tracker.Failed)
		writer.Uint32(uint32(len(tracker.SuccessItems)))
		for _, item := range tracker.SuccessItems {
			writer.String(item.Query)
		}
		writer.Uint32(uint32(len(tracker.FailPartialItems)))
		for _, item := range tracker.FailPartialItems {
			writer.String(item.Query)
			writer.StringSlice(item.PartialSuccess)
			writer.String(item.UnexpectedError)
			writer.String(item.ExpectedError)
		}
	}
	return writer.Data()
}

// DeserializeTrackers deserializes the given data into a sorted list of trackers.
func DeserializeTrackers(data []byte) ([]*ReplayTracker, error) {
	reader := utils.NewReader(data)
	version := reader.Uint32()
	if version != 2 {
		return nil, errors.Errorf("version %d is not supported by this branch", version)
	}
	trackers := make([]*ReplayTracker, reader.Uint32())
	for trackerIdx := 0; trackerIdx < len(trackers); trackerIdx++ {
		trackers[trackerIdx] = &ReplayTracker{}
		trackers[trackerIdx].File = reader.String()
		trackers[trackerIdx].Success = reader.Uint32()
		trackers[trackerIdx].PartialSuccess = reader.Uint32()
		trackers[trackerIdx].Failed = reader.Uint32()
		trackers[trackerIdx].SuccessItems = make([]ReplayTrackerItem, reader.Uint32())
		for itemIdx := 0; itemIdx < len(trackers[trackerIdx].SuccessItems); itemIdx++ {
			trackers[trackerIdx].SuccessItems[itemIdx].Query = reader.String()
		}
		trackers[trackerIdx].FailPartialItems = make([]ReplayTrackerItem, reader.Uint32())
		for itemIdx := 0; itemIdx < len(trackers[trackerIdx].FailPartialItems); itemIdx++ {
			trackers[trackerIdx].FailPartialItems[itemIdx].Query = reader.String()
			trackers[trackerIdx].FailPartialItems[itemIdx].PartialSuccess = reader.StringSlice()
			trackers[trackerIdx].FailPartialItems[itemIdx].UnexpectedError = reader.String()
			trackers[trackerIdx].FailPartialItems[itemIdx].ExpectedError = reader.String()
		}
	}
	sort.Slice(trackers, func(i, j int) bool {
		return trackers[i].File < trackers[j].File
	})
	if !reader.IsEmpty() {
		return trackers, errors.Errorf("additional data remaining after all trackers have been read")
	}
	return trackers, nil
}
