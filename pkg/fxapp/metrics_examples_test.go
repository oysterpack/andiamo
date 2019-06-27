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
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"log"
	"net/http"
)

func ExamplePrometheusHTTPServerRunner() {
	prometheusHTTPServerOpts := fxapp.NewPrometheusHTTPServerOpts()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(fxapp.PrometheusHTTPServerRunner(prometheusHTTPServerOpts)).
		Build()

	if err != nil {
		log.Panic(err)
	}

	go app.Run()
	<-app.Started()
	defer func() {
		app.Shutdown()
		<-app.Done()
	}()

	resp, err := retryablehttp.Get(fmt.Sprintf("http://:%d%s", prometheusHTTPServerOpts.Port(), prometheusHTTPServerOpts.Endpoint()))
	log.Printf("StatusCode: %d", resp.StatusCode)
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
			log.Println(line)
		}
	}

	// Output:
	//
}
