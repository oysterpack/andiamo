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
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/prometheus/client_golang/prometheus"
)

// HealthCheckMetricID is used as the prometheus metric name
const HealthCheckMetricID = "U01DF4CVSSF4RT1ZB4EXC44G668"

func registerHealthCheckGauge(check health.Check, scheduler health.Scheduler, registerer prometheus.Registerer) error {
	healthCheckResult := scheduler.Subscribe(func(c health.Check) bool {
		return c.ID() == check.ID()
	})

	getResult := make(chan chan health.Result)
	go func() {
		var result health.Result

		// initialize the health check result
		resultsChan := scheduler.Results(func(result health.Result) bool {
			return result.HealthCheckID() == check.ID()
		})
		done := scheduler.Done()
		select {
		case <-done:
			return
		case results := <-resultsChan:
			if len(results) == 1 {
				result = results[0]
			} else {
				result = check.Run()
			}
		}
		// event loop
		for {
			select {
			case <-done:
				return
			case result = <-healthCheckResult: // update the health check result with the latest result
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
			"h": check.ID().String(),
			"d": check.Desc().ID().String(),
		},
		Help: "health check",
	}

	return registerer.Register(prometheus.NewGaugeFunc(opts, func() float64 {
		ch := make(chan health.Result)
		select {
		case <-scheduler.Done():
			return -1
		case getResult <- ch:
			select {
			case <-scheduler.Done():
				return -1
			case result := <-ch:
				return float64(result.Status())
			}
		}
	}))

}
