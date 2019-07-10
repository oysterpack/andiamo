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
	"github.com/oysterpack/partire-k8s/pkg/health"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"net/http"
	"sync"
	"time"
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
	r.count += delta
	r.Unlock()
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
	return NewHTTPHandler(fmt.Sprintf("/%s", ReadyEvent), func(writer http.ResponseWriter, request *http.Request) {
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

// LivenessProbe checks if the app is healthy. It returns an error if probe fails, indicating the app is unhealthy.
type LivenessProbe func() error

func livenessProbe(gatherer prometheus.Gatherer) LivenessProbe {
	healthCheckID := func(m *io_prometheus_client.Metric) string {
		for _, label := range m.Label {
			if label.GetName() == "h" {
				return label.GetValue()
			}
		}
		return ""
	}
	return func() error {
		mfs, err := gatherer.Gather()
		if err != nil {
			return err
		}
		healthCheckMetricFamily := FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return mf.GetName() == HealthCheckMetricID
		})
		for _, metric := range healthCheckMetricFamily.GetMetric() {
			if health.Status(metric.Gauge.GetValue()) == health.Red {
				return fmt.Errorf("liveness probe failed because health check is RED: %v", healthCheckID(metric))
			}
		}
		return nil
	}
}

// if any health check status is Red, then the liveness check fails
func livenessProbeHTTPHandler(probe LivenessProbe, logger *zerolog.Logger) HTTPHandler {
	logProbeSuccess := LivenessProbeEvent.NewLogger(logger, zerolog.InfoLevel)
	logProbeFailure := LivenessProbeEvent.NewErrorLogger(logger)
	return NewHTTPHandler(fmt.Sprintf("/%s", LivenessProbeEvent), func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		err := probe()
		probeDuration := duration(time.Since(start))
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			logProbeFailure(probeDuration, err, "liveness probe failed")
			return
		}
		writer.WriteHeader(http.StatusOK)
		logProbeSuccess(probeDuration, "liveness probe success")
	})
}
