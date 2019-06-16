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

package metric_test

import (
	"github.com/oysterpack/partire-k8s/pkg/app/metric"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"reflect"
	"testing"
)

func TestFindMetricFamily(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo1",
		Help: "foo 1 help",
	}))
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo2",
		Help: "foo 2 help",
	}))

	mfs, e := registry.Gather()
	if e != nil {
		t.Fatalf("*** failed to gather metrics: %v", e)
	}
	mf := metric.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
		return *mf.Name == "foo1"
	})
	if mf == nil {
		t.Fatal("*** metric was not found")
	}

	mf = metric.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
		return false
	})
	if mf != nil {
		t.Fatal("*** metric should not have been found")
	}
}

func TestFindMetricFamilies(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
		Help: "foo help",
		ConstLabels: prometheus.Labels{
			"bar": "1",
		},
	}))
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo",
		Help: "foo help",
		ConstLabels: prometheus.Labels{
			"bar": "2",
		},
	}))
	registry.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bar",
		Help: "foo help",
		ConstLabels: prometheus.Labels{
			"bar": "2",
		},
	}))

	mfs, e := registry.Gather()
	if e != nil {
		t.Fatalf("*** failed to gather metrics: %v", e)
	}

	{
		mfs = metric.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return *mf.Name == "foo"
		})
		if len(mfs) != 1 {
			t.Fatalf("*** wrong number of metrics were returned: %v", mfs)
		}
	}

	{
		mfs = metric.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return false
		})
		if len(mfs) > 0 {
			t.Fatalf("*** no metrics should have been returned: %v", mfs)
		}
	}
}

func TestTypedMetric(t *testing.T) {
	// metric dependencies should be strongly typed
	// the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter

	var (
		FooCounterOpts = prometheus.CounterOpts{
			Name: "foo_counter",
			Help: "foo counter",
		}

		FooCounter                                         = prometheus.NewCounter(FooCounterOpts)
		FooCommandExecutionCounter CommandExecutionCounter = FooCounter
	)

	FooCommandExecutionCounter.Inc()

	if !reflect.TypeOf(FooCommandExecutionCounter).ConvertibleTo(reflect.TypeOf(FooCounter)) {
		t.Error("CommandExecutionCounter is not convertible to prometheus.Counter")
	}

	if !reflect.TypeOf(FooCounter).ConvertibleTo(reflect.TypeOf(FooCommandExecutionCounter)) {
		t.Error("prometheus.Counter is not convertible to CommandExecutionCounter")
	}
}

func TestLabel_String(t *testing.T) {
	if metric.AppInstanceID.String() != string(metric.AppInstanceID) {
		t.Fatal("*** Label.String() should simply return the label as a string")
	}
}
