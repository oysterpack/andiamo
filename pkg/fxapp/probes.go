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

package fxapp

import (
	"fmt"
	"net/http"
	"sync"
)

// ReadinessWaitGroup is used by application components to signal when they are ready to service requests
type ReadinessWaitGroup interface {
	Add(delta uint)
	Inc()

	// Count returns the wait group counter value. When the count is zero, it means the wait group is done.
	Count() uint

	// Done decrements the wait group counter by one
	Done()

	// Ready returns a chan that is used to signal when the wait group counter is zero.
	Ready() <-chan struct{}
}

// NewReadinessWaitgroup returns a new ReadinessWaitGroup initialized with the specified count
func NewReadinessWaitgroup(count uint) ReadinessWaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(int(count))
	return &readinessWaitGroup{
		wg,
		&sync.Mutex{},
		count,
	}
}

// ReadinessWaitGroup is used by application components to notify the app when it is ready to service requests.
type readinessWaitGroup struct {
	*sync.WaitGroup
	*sync.Mutex
	count uint
}

func (r *readinessWaitGroup) Add(delta uint) {
	r.WaitGroup.Add(int(delta))
	r.Lock()
	defer r.Unlock()
	r.count += delta
}

func (r *readinessWaitGroup) Inc() {
	r.Add(1)
}

func (r *readinessWaitGroup) Count() uint {
	r.Lock()
	defer r.Unlock()
	return r.count
}

func (r *readinessWaitGroup) Done() {
	r.Lock()
	defer r.Unlock()
	r.count--
	r.WaitGroup.Done()
}

// Ready returns a chan that is used to signal that the application is ready to service requests
func (r *readinessWaitGroup) Ready() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		defer close(c)
		r.Wait()
	}()
	return c
}

func readinessProbeHTTPHandler(readiness ReadinessWaitGroup) HTTPHandler {
	endpoint := fmt.Sprintf("/%s", ReadyEvent)
	return NewHTTPHandler(endpoint, func(writer http.ResponseWriter, request *http.Request) {
		count := readiness.Count()
		switch count {
		case 0:
			writer.WriteHeader(http.StatusOK)
		default:
			writer.Header().Add("x-readiness-wait-group-count", fmt.Sprint(count))
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	})
}
