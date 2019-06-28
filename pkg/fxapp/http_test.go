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
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"io"
	"net/http"
	"testing"
)

// The app provides an HTTP server.
// The HTTP server is only started if it is needed, i.e., if endpoint handlers are discovered.
// For example, exposing prometheus metrics via HTTP wil enable the HTTP server.
//
// httpServerOpts can be specified to customize the config.
func TestHTTPServer_WithDefaultOpts(t *testing.T) {
	buf := new(bytes.Buffer)
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
		).
		ExposePrometheusMetricsViaHTTP(nil).
		Invoke(func() {}).
		LogWriter(buf).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Started()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()

		// Then the HTTP server is running
		// And the registered endpoints are acccessible
		checkHTTPGetResponseStatusOK(t, "http://:8008/foo")
		checkHTTPGetResponseStatusOK(t, "http://:8008/metrics")

		checkHTTPServerStartedEventLogged(t, buf)
	}
}

func checkHTTPGetResponseStatusOK(t *testing.T, url string) {
	resp, err := http.Get(url)
	switch {
	case err != nil:
		t.Errorf("*** failed to HTTP scrape metrics: %v", err)
	case resp.StatusCode != http.StatusOK:
		t.Errorf("*** request failed: %v", resp.StatusCode)
	}
}

func checkHTTPServerStartedEventLogged(t *testing.T, log io.Reader) {
	type Data struct {
		Addr      string
		Endpoints []string
	}

	type LogEvent struct {
		Name string `json:"n"`
		Data Data   `json:"01DEFM9FFSH58ZGNPSR7Z4C3G2"`
	}

	var logEvent LogEvent
	reader := bufio.NewReader(log)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		err = json.Unmarshal([]byte(line), &logEvent)
		if err != nil {
			t.Errorf("*** failed to parse log event: %v : %v", err, line)
			continue
		}
		if logEvent.Name == fxapp.HTTPServerStarting.String() {
			t.Log(line)
			break
		}
	}
	if logEvent.Name != fxapp.HTTPServerStarting.String() {
		t.Error("*** HTTP server started event was not logged")
	}

}
