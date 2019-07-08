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

package fxapp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/fxapptest"
	"go.uber.org/fx"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Ready is an app lifecycle state. To be ready means the app is ready to serve requests.
// When the app is ready, it logs an event.
func TestReadinessProbe(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	var readinessProbe fxapp.ReadinessWaitGroup
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func(lc fx.Lifecycle, readinessProbe fxapp.ReadinessWaitGroup) {
			readinessProbe.Add(1)
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					readinessProbe.Done()
					return nil
				},
			})
		}).
		Populate(&readinessProbe).
		LogWriter(buf).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed : %v", err)
	default:
		if readinessProbe.Count() < 2 {
			t.Errorf("*** the readiness probe count should be at least 2: %d", readinessProbe.Count())
		}

		go app.Run()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()
		<-app.Started()

		// wait for the app to be ready
		<-readinessProbe.Ready()
		// the app uses the same underying readiness probe
		<-app.Ready()

		// Then the app's readiness HTTP endpoint should pass
		checkHTTPGetResponseStatusOK(t, fmt.Sprintf("http://:8008/%s", fxapp.ReadyEvent))

		app.Shutdown()
		<-app.Done()

		// And the app Ready lifecycle event is logged
		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			t.Log(line)
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v", err)
				break
			}

			if logEvent.Name == fxapp.ReadyEvent.String() {
				break
			}
		}

		if logEvent.Name != fxapp.ReadyEvent.String() {
			t.Error("*** app readiness log event was not logged")
		}
	}
}

func TestReadinessProbeNotReady(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var readinessProbe fxapp.ReadinessWaitGroup
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func(lc fx.Lifecycle, readinessProbe fxapp.ReadinessWaitGroup) {
			readinessProbe.Add(1)
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						wg.Wait()
						readinessProbe.Done()
					}()
					return nil
				},
			})
		}).
		Populate(&readinessProbe).
		LogWriter(buf).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed : %v", err)
	default:
		go app.Run()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()
		<-app.Started()

		checkHTTPGetResponseStatus(t, fmt.Sprintf("http://:8008/%s", fxapp.ReadyEvent), http.StatusServiceUnavailable)
		checkHTTPGetResponse(t, fmt.Sprintf("http://:8008/%s", fxapp.ReadyEvent), func(response *http.Response) {
			t.Log("status ", response.StatusCode)
			if count, err := strconv.ParseUint(response.Header.Get("x-readiness-wait-group-count"), 10, 64); err != nil {
				t.Errorf("*** failed to parse `x-readiness-wait-group-count` header into num: %v", err)
			} else if count == 0 {
				t.Errorf("*** expected count to be > 0: %d", count)
			}

		})
		wg.Done()
		// the app uses the same underying readiness probe
		<-app.Ready()

		// Then the app's readiness HTTP endpoint should pass
		checkHTTPGetResponseStatusOK(t, fmt.Sprintf("http://:8008/%s", fxapp.ReadyEvent))
	}
}

func TestNewReadinessWaitgroup(t *testing.T) {
	readinessGroup := fxapp.NewReadinessWaitgroup(10)
	if readinessGroup.Count() != 10 {
		t.Errorf("*** count should be 10: %d", readinessGroup.Count())
	}
	readinessGroup.Done()
	if readinessGroup.Count() != 9 {
		t.Errorf("*** count should be 9: %d", readinessGroup.Count())
	}

	i := 0
	for ; readinessGroup.Count() > 0; i++ {
		readinessGroup.Done()
	}
	if readinessGroup.Count() != 0 {
		t.Errorf("*** count should be 0: %d", readinessGroup.Count())
	}
	if i != 9 {
		t.Errorf("*** counter should have decremented 9 times: %d", i)
	}
}

func TestNewReadinessWaitgroup_Async(t *testing.T) {
	readinessGroup := fxapp.NewReadinessWaitgroup(10)
	for i := 0; i < 10; i++ {
		go readinessGroup.Done()
	}
	<-readinessGroup.Ready()
	if readinessGroup.Count() != 0 {
		t.Errorf("*** count should be 0: %d", readinessGroup.Count())
	}
}
