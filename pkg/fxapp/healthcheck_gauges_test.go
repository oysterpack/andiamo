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
	"github.com/oysterpack/partire-k8s/pkg/fx/health"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulids"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"testing"
	"time"
)

func TestHealthCheckGauge(t *testing.T) {
	t.Parallel()

	Foo1 := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo1",
		RedImpact:    "Red",
		YellowImpact: "Yellow",
	}

	Foo2 := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo2",
		RedImpact:    "Red",
		YellowImpact: "Yellow",
	}

	var gatherer prometheus.Gatherer
	app, err := fxapp.NewBuilder(fxapp.ID(ulids.MustNew()), fxapp.ReleaseID(ulids.MustNew())).
		Invoke(func(register health.Register) error {
			if err := register(Foo1, health.CheckerOpts{}, func() error {
				return nil
			}); err != nil {
				return err
			}
			if err := register(Foo2, health.CheckerOpts{}, func() error {
				return nil
			}); err != nil {
				return err
			}
			return nil
		}).
		Populate(&gatherer).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** app failed to build: %v", err)
	}

	go app.Run()
	defer func() {
		app.Shutdown()
		<-app.Done()
	}()
	<-app.Ready()

	var healthcheckMetrics *io_prometheus_client.MetricFamily
MetricFamilyLoop:
	for {
		mfs, err := gatherer.Gather()
		if err != nil {
			t.Errorf("*** failed to gather metrics: %v", err)
			return
		}

		healthcheckMetrics = fxapp.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return mf.GetName() == fxapp.HealthCheckMetricID
		})
		if healthcheckMetrics != nil && len(healthcheckMetrics.Metric) >= 2 {
			break MetricFamilyLoop
		}

		time.Sleep(time.Millisecond)
	}

HealthCheckLoop:
	for _, check := range []health.Check{Foo1, Foo2} {
		t.Log(check)
		for _, metric := range healthcheckMetrics.Metric {
			t.Log(metric)
			for _, labelPair := range metric.GetLabel() {
				if labelPair.GetName() == "h" && labelPair.GetValue() == check.ID {
					continue HealthCheckLoop
				}
			}
		}
		t.Errorf("*** health check was not gathered: %v", check)
	}

	app.Shutdown()
	<-app.Done()

MetricFamilyLoop2:
	for {
		mfs, err := gatherer.Gather()
		if err != nil {
			t.Errorf("*** failed to gather metrics: %v", err)
			return
		}

		healthcheckMetrics = fxapp.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return mf.GetName() == fxapp.HealthCheckMetricID
		})
		if healthcheckMetrics != nil && len(healthcheckMetrics.Metric) >= 2 {
			break MetricFamilyLoop2
		}

		time.Sleep(time.Millisecond)
	}

HealthCheckLoop2:
	for _, check := range []health.Check{Foo1, Foo2} {
		t.Log(check)
		for _, metric := range healthcheckMetrics.Metric {
			t.Log(metric)
			for _, labelPair := range metric.GetLabel() {
				if labelPair.GetName() == "h" && metric.Gauge.GetValue() < 0 {
					continue HealthCheckLoop2
				}
			}
		}
		t.Errorf("*** health check was not gathered: %v", check)
	}

}
