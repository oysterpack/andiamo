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
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/fxapptest"
	"io"
	"net/http"
	"testing"
)

// The app provides an HTTP server.
// The HTTP server is only started if it is needed, i.e., if endpoint handlers are discovered.
// For example, exposing prometheus metrics via HTTP will enable the HTTP server.
// If an *http.Server is provided, then it will be used. Otherwise a default HTTP server is created automatically by the app.
//
// An event is logged when the HTTP server is starting, containing the address the server is listening on and the list
// of handler endpoints that are registered.
func TestHTTPServer_WithDefaultOpts(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
		).
		Invoke(func() {}).
		LogWriter(buf).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Ready()
		defer func() {
			app.Shutdown()
			<-app.Done()

			checkHTTPServerStartingEventLogged(t, buf, ":8008", []string{"/foo", fxapp.DefaultPrometheusHTTPHandlerOpts().Endpoint})
		}()

		// Then the HTTP server is running
		// And the registered endpoints are acccessible
		checkHTTPGetResponseStatusOK(t, "http://:8008/foo")
		checkHTTPGetResponseStatusOK(t, fmt.Sprintf("http://:8008/%s", fxapp.MetricsEndpoint))
	}
}

func TestHTTPServer_WithProvidedServer(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
			func() *http.Server {
				return &http.Server{
					Addr: ":5050",
				}
			},
		).
		Invoke(func() {}).
		LogWriter(buf).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Ready()
		defer func() {
			app.Shutdown()
			<-app.Done()

			checkHTTPServerStartingEventLogged(t, buf, ":5050", []string{"/foo", fxapp.DefaultPrometheusHTTPHandlerOpts().Endpoint})
		}()

		// Then the HTTP server is running
		// And the registered endpoints are acccessible
		checkHTTPGetResponseStatusOK(t, "http://:5050/foo")
		checkHTTPGetResponseStatusOK(t, fmt.Sprintf("http://:5050/%s", fxapp.MetricsEndpoint))
	}
}

func TestHTTPServer_WithDuplicateEndpoints(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
		).
		Invoke(func() {}).
		LogWriter(buf).
		Build()

	if err == nil {
		t.Error("*** app should have failed to build because there are duplicate endpoints registered")
	} else {
		t.Log(err)
	}
}

func TestHTTPServer_WithNilhandler(t *testing.T) {
	buf := fxapptest.NewSyncLog()
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			},
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/bar", nil)
			},
		).
		Invoke(func() {}).
		LogWriter(buf).
		Build()

	if err == nil {
		t.Error("*** app should have failed to build because the /bar endpoint has a nil handler")
	} else {
		t.Log(err)
	}
}

func TestHTTPServer_HandlerPanic(t *testing.T) {
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func() fxapp.HTTPHandler {
				return fxapp.NewHTTPHandler("/foo", func(writer http.ResponseWriter, request *http.Request) {
					panic("BOOM")
				})
			},
		).
		Invoke(func() {}).
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

		// Then the GET /foo request will fail because the handler panics
		if _, err := http.Get("http://:8008/foo"); err == nil {
			t.Error("HTTP request should have failed with an EOF error")
		} else {
			t.Log(err)
		}
		// And HTTP server should still be able to serve other requests
		checkHTTPGetResponseStatusOK(t, fmt.Sprintf("http://:8008/%s", fxapp.MetricsEndpoint))
	}
}

func checkHTTPGetResponseStatusOK(t *testing.T, url string) {
	t.Log("GET ", url)
	resp, err := retryablehttp.Get(url)
	switch {
	case err != nil:
		t.Errorf("*** %v: failed to HTTP scrape metrics: %v", url, err)
	case resp.StatusCode != http.StatusOK:
		t.Errorf("*** %v: request failed: %v", url, resp.StatusCode)
	}
}

func checkHTTPGetResponseStatus(t *testing.T, url string, expectedStatusCode int) {
	resp, err := http.Get(url)
	switch {
	case err != nil:
		t.Errorf("*** %v: HTTP GET failed: %v", url, err)
	case resp.StatusCode != expectedStatusCode:
		t.Errorf("*** %v: response status did not match: %v", url, resp.StatusCode)
	}
}

func checkHTTPGetResponse(t *testing.T, url string, check func(response *http.Response)) {
	resp, err := http.Get(url)
	switch {
	case err != nil:
		t.Errorf("*** %v: HTTP GET failed: %v", url, err)
	default:
		check(resp)
	}
}

func checkHTTPServerStartingEventLogged(t *testing.T, log io.Reader, addr string, endpoints []string) {
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
	switch {
	case logEvent.Name != fxapp.HTTPServerStarting.String():
		t.Error("*** HTTP server started event was not logged")
	default:
		if logEvent.Data.Addr != addr {
			t.Errorf("*** addr did not match: %v != %v", logEvent.Data.Addr, addr)
		}

	ExpectedEndpoints:
		for _, expectedEndpoint := range endpoints {
			for _, endpoint := range logEvent.Data.Endpoints {
				if endpoint == expectedEndpoint {
					continue ExpectedEndpoints
				}
			}
			t.Errorf("*** endpoint was not logged: %v", expectedEndpoint)
		}
	}

}

// Uses cases for disabling the HTTP server:
// - when using the App for running tests the HTTP server can be disabled to reduce overhead. It also enables tests to be run
//   in parallel
// - for CLI based apps
func TestBuilder_DisableHTTPServer(t *testing.T) {
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func() {}).
		DisableHTTPServer().
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Ready()

		_, err := http.Get("http://:8008/metrics")
		if err == nil {
			t.Error("HTTP GET should have failed because the HTTP server should not be running")
		} else {
			t.Log(err)
		}
	}
}
