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
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// HealthCheckMetricID is used as the prometheus metric name
const HealthCheckMetricID = "U01DF4CVSSF4RT1ZB4EXC44G668"

func registerHealthCheckGauge(done <-chan struct{}, check health.RegisteredCheck, subscribeForCheckResults health.SubscribeForCheckResults, checkResults health.CheckResults, registerer prometheus.Registerer) error {
	healthCheckResult := subscribeForCheckResults(func(result health.Result) bool {
		return result.ID == check.ID
	})

	getResult := make(chan chan health.Result)
	go func() {
		var result health.Result

		// initialize the health check result
		resultsChan := checkResults(func(result health.Result) bool {
			return result.ID == check.ID
		})
		results := <-resultsChan
		if len(results) == 1 {
			result = results[0]
		} else {
			start := time.Now()
			err := check.Checker()
			duration := time.Since(start)
			status := health.Green
			if err != nil {
				switch err.(type) {
				case health.YellowError:
					status = health.Yellow
				default:
					status = health.Red
				}
			}
			result = health.Result{
				ID:       check.ID,
				Status:   status,
				Err:      err,
				Time:     start,
				Duration: duration,
			}
		}

		// event loop
		for {
			select {
			case <-done:
				return
			case result = <-healthCheckResult.Chan(): // update the health check result with the latest result
			case reply := <-getResult: // metrics are being gathered
				go func(result health.Result) {
					reply <- result
				}(result)
			}
		}
	}()

	opts := prometheus.GaugeOpts{
		Name: HealthCheckMetricID,
		ConstLabels: map[string]string{
			"h": check.ID,
		},
		Help: "health check",
	}

	return registerer.Register(prometheus.NewGaugeFunc(opts, func() float64 {
		ch := make(chan health.Result)
		select {
		case <-done:
			return -1
		case getResult <- ch:
			select {
			case <-done:
				return -1
			case result := <-ch:
				return float64(result.Status)
			}
		}
	}))

}
