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
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/jackc/pgx/v5/pgproto3"
)

var (
	terminate    = &sync.WaitGroup{}                    // terminate is used when a Terminate message has been received.
	messageMutex = &sync.Mutex{}                        // messageMutex guards against both the client and server writing to the message slice.
	allMessages  = make([]pgproto3.Message, 0, 1000000) // allMessages contains all messages exchanged by the client and server.
)

func main() {
	// If no arguments are given, then we'll update the results against the regression files
	if len(os.Args) <= 1 {
		updateResults()
		return
	} else if len(os.Args) != 3 {
		fmt.Println("Expected two arguments, each containing a file name pointing to the tracker files (located in the out directory)")
		os.Exit(1)
	}

	trackersFrom, err := regressionFolder.ReadReplayTrackers("out/" + os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	trackersTo, err := regressionFolder.ReadReplayTrackers("out/" + os.Args[2])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fromTotal := uint32(0)
	fromSuccess := uint32(0)
	fromPartial := uint32(0)
	fromFail := uint32(0)
	for _, tracker := range trackersFrom {
		fromTotal += tracker.Success
		fromTotal += tracker.Failed
		fromSuccess += tracker.Success
		fromPartial += tracker.PartialSuccess
		fromFail += tracker.Failed
	}
	toTotal := uint32(0)
	toSuccess := uint32(0)
	toPartial := uint32(0)
	toFail := uint32(0)
	for _, tracker := range trackersTo {
		toTotal += tracker.Success
		toTotal += tracker.Failed
		toSuccess += tracker.Success
		toPartial += tracker.PartialSuccess
		toFail += tracker.Failed
	}
	sb := strings.Builder{}
	sb.WriteString("|   | Main | PR |\n")
	sb.WriteString("| --- | --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| Total | %d | %d |\n", fromTotal, toTotal))
	sb.WriteString(fmt.Sprintf("| Successful | %d | %d |\n", fromSuccess, toSuccess))
	sb.WriteString(fmt.Sprintf("| Failures | %d | %d |\n", fromFail, toFail))
	sb.WriteString(fmt.Sprintf("| Partial Successes[^1] | %d | %d |\n", fromPartial, toPartial))
	sb.WriteString("\n|   | Main | PR |\n")
	sb.WriteString("| --- | --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| Successful | %.4f%% | %.4f%% |\n",
		(float64(fromSuccess)/float64(fromTotal))*100.0,
		(float64(toSuccess)/float64(toTotal))*100.0))
	sb.WriteString(fmt.Sprintf("| Failures | %.4f%% | %.4f%% |\n",
		(float64(fromFail)/float64(fromTotal))*100.0,
		(float64(toFail)/float64(toTotal))*100.0))
	if len(trackersFrom) == len(trackersTo) {
		foundAnyFailDiff := false
		foundAnySuccessDiff := false
		for trackerIdx := range trackersFrom {
			// They're sorted, so this should always hold true.
			// This will really only fail if the tests were updated.
			if trackersFrom[trackerIdx].File != trackersTo[trackerIdx].File {
				continue
			}
			// Handle regressions (which we'll display first)
			foundFileDiff := false
			fromFailItems := make(map[string]struct{})
			for _, trackerFromItem := range trackersFrom[trackerIdx].FailPartialItems {
				fromFailItems[trackerFromItem.Query] = struct{}{}
			}
			for _, trackerToItem := range trackersTo[trackerIdx].FailPartialItems {
				if _, ok := fromFailItems[trackerToItem.Query]; !ok {
					if !foundAnyFailDiff {
						foundAnyFailDiff = true
						sb.WriteString("\n## Regressions\n")
					}
					if !foundFileDiff {
						foundFileDiff = true
						sb.WriteString(fmt.Sprintf("### %s\n", trackersFrom[trackerIdx].File))
					}
					sb.WriteString(fmt.Sprintf("```\nQUERY:          %s\n", trackerToItem.Query))
					if len(trackerToItem.ExpectedError) != 0 {
						sb.WriteString(fmt.Sprintf("EXPECTED ERROR: %s\n", trackerToItem.ExpectedError))
					}
					if len(trackerToItem.UnexpectedError) != 0 {
						sb.WriteString(fmt.Sprintf("RECEIVED ERROR: %s\n", trackerToItem.UnexpectedError))
					}
					for _, partial := range trackerToItem.PartialSuccess {
						sb.WriteString(fmt.Sprintf("PARTIAL:        %s\n", partial))
					}
					sb.WriteString("```\n")
				}
			}
			// Handle progressions (which we'll display second)
			foundFileDiff = false
			fromSuccessItems := make(map[string]struct{})
			for _, trackerFromItem := range trackersFrom[trackerIdx].SuccessItems {
				fromSuccessItems[trackerFromItem.Query] = struct{}{}
			}
			for _, trackerToItem := range trackersTo[trackerIdx].SuccessItems {
				if _, ok := fromSuccessItems[trackerToItem.Query]; !ok {
					if !foundAnySuccessDiff {
						foundAnySuccessDiff = true
						sb.WriteString("\n## Progressions\n")
					}
					if !foundFileDiff {
						foundFileDiff = true
						sb.WriteString(fmt.Sprintf("### %s\n", trackersFrom[trackerIdx].File))
					}
					sb.WriteString(fmt.Sprintf("```\nQUERY: %s\n```\n", trackerToItem.Query))
				}
			}
		}
	}
	sb.WriteString("[^1]: These are tests that we're marking as `Successful`, however they do not match the expected output in some way. This is due to small differences, such as different wording on the error messages, or the column names being incorrect while the data itself is correct.")
	fmt.Println(sb.String())
}

func updateResults() {
	fmt.Println("Updating results, remember to run the regression tester using our schedule")
	listener, err := server.NewListener("tcp", "127.0.0.1:5431", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	timer := time.NewTimer(5 * time.Second)
	timer.Stop()
	go func() {
		<-timer.C
		_ = listener.Close()
	}()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			break
		}
		timer.Stop()
		terminate = &sync.WaitGroup{}
		terminate.Add(1)
		clientConnBackend := pgproto3.NewBackend(clientConn, clientConn)
		postgresConn, err := (&net.Dialer{}).Dial("tcp", "127.0.0.1:5432")
		if err != nil {
			fmt.Println(err)
			return
		}
		postgresConnFrontend := pgproto3.NewFrontend(postgresConn, postgresConn)

		if err = handleStartup(clientConnBackend, postgresConnFrontend, clientConn); err != nil {
			fmt.Println(err)
			return
		}
		createPassthrough(clientConnBackend, postgresConnFrontend)
		terminate.Wait()
		_ = clientConn.Close()
		_ = postgresConn.Close()
		timer.Reset(5 * time.Second)
	}
	// The first two groups are a part of the setup, so we'll combine them with the first file (which is the setup)
	groups := SplitByTerminate(allMessages)
	groups[2] = CombineGroups(groups[0], groups[1], groups[2])
	groups = groups[2:]
	if len(groups) != len(AllTestResultFilesNames) {
		fmt.Printf("Number of groups: %d\nNumber of test names: %d", len(groups), len(AllTestResultFilesNames))
		return
	}
	for i := 0; i < len(groups); i++ {
		if err = regressionFolder.WriteMessages(AllTestResultFilesNames[i], groups[i], 0644); err != nil {
			fmt.Println(err)
			return
		}
	}
}

// addMessage adds the given message to the message slice.
func addMessage(message pgproto3.Message) {
	messageMutex.Lock()
	defer messageMutex.Unlock()
	allMessages = append(allMessages, message)
	fmt.Printf("%T: %v\n", message, message)
}
