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
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"net/http"
	"testing"
)

// The app provides an HTTP server.
// The HTTP server is only started if it is needed, i.e., if endpoint handlers are discovered.
// For example, exposing prometheus metrics via HTTP wil enable the HTTP server.
//
// httpServerOpts can be specified to customize the config.
func TestHTTPServer_WithDefaultOpts(t *testing.T) {
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

		resp, err := http.Get("http://:8008/foo")
		switch {
		case err != nil:
			t.Errorf("*** failed to HTTP scrape metrics: %v", err)
		case resp.StatusCode != http.StatusOK:
			t.Errorf("*** request failed: %v", resp.StatusCode)
		}
	}

}
