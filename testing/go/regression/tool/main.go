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
	"sync"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/jackc/pgx/v5/pgproto3"
)

var (
	regressionFolder RegressionFolderLocation               // regressionFolder is the disk location of the regression folder
	terminate        = &sync.WaitGroup{}                    // terminate is used when a Terminate message has been received.
	messageMutex     = &sync.Mutex{}                        // messageMutex guards against both the client and server writing to the message slice.
	allMessages      = make([]pgproto3.Message, 0, 1000000) // allMessages contains all messages exchanged by the client and server.
)

func main() {
	var err error
	regressionFolder, err = GetRegressionFolder()
	if err != nil {
		fmt.Println(err)
		return
	}
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
