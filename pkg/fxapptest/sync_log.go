/*
 * Copyright (c) 2019 OysterPack, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fxapptest

import (
	"bytes"
	"sync"
)

// SyncLog is used to to provide a concurrency safe read/write log.
//
// Use Case: used when inspecting logs in unit tests that have multiple go routines writing to the log concurrently
type SyncLog struct {
	sync.Mutex
	buf *bytes.Buffer
}

func NewSyncLog() *SyncLog {
	return &SyncLog{
		buf: new(bytes.Buffer),
	}
}

func (l *SyncLog) Write(data []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.buf.Write(data)
}

func (l *SyncLog) Read(p []byte) (n int, err error) {
	l.Lock()
	defer l.Unlock()
	return l.buf.Read(p)
}

func (l *SyncLog) String() string {
	l.Lock()
	defer l.Unlock()
	return l.buf.String()
}

func (l *SyncLog) Bytes() []byte {
	l.Lock()
	defer l.Unlock()
	return l.buf.Bytes()
}
