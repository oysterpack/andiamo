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
	"github.com/hashicorp/go-retryablehttp"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/rs/zerolog"
	"log"
	"net/http"
)

func ExamplePrometheusHTTPServerRunner() {
	var logger *zerolog.Logger
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(fxapp.PrometheusHTTPServerRunner(
			fxapp.PrometheusHTTPServerOpts{},
		)).
		Populate(&logger).
		Build()

	if err != nil {
		log.Panic(err)
	}

	go app.Run()
	<-app.Started()

	resp, err := retryablehttp.Get("http://:5050/metrics")
	logger.Info().
		Int("StatusCode", resp.StatusCode).
		Msg("")
	switch {
	case err != nil:
		log.Panic(err)
	case resp.StatusCode != http.StatusOK:
		log.Panicf("HTTP request failed: %v : %v", resp.StatusCode, resp.Status)
	default:
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			logger.Info().Msg(line)
		}
	}

	app.Shutdown()
	<-app.Done()

	// Output:
	//
}
