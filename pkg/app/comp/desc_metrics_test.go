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

package comp_test

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/metric"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"reflect"
	"testing"
)

func TestDesc_WrapRegisterer(t *testing.T) {
	// metric dependencies should be strongly typed
	// the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter
	type CommandExecutionCounterOpts prometheus.CounterOpts

	type CommandExecutionCounterOptsProvider func(desc app.Desc) CommandExecutionCounterOpts
	type CommandExecutionCounterProvider func(opts CommandExecutionCounterOpts) CommandExecutionCounter

	var newCommandExecutionCounter CommandExecutionCounterProvider = func(opts CommandExecutionCounterOpts) CommandExecutionCounter {
		return CommandExecutionCounter(prometheus.NewCounter(prometheus.CounterOpts(opts)))
	}

	type Command func(ctx context.Context) error
	// metrics are injected
	type CommandProvider func(CommandExecutionCounter) Command
	optionDesc := option.NewDesc(option.Provide, reflect.TypeOf(CommandProvider(nil)))
	counterOptsOptionDesc := option.NewDesc(option.Provide, reflect.TypeOf(CommandExecutionCounterOptsProvider(nil)))

	compDesc := comp.MustNewDesc(
		comp.ID(ulidgen.MustNew().String()),
		comp.Name("foo"),
		comp.Version("0.1.0"),
		Package,
		optionDesc,
		counterOptsOptionDesc,
	)
	metricRegistry := prometheus.NewRegistry()
	// Given a component metric registerer
	metricRegisterer := compDesc.WrapRegisterer(metricRegistry)
	var (
		counterOpts = CommandExecutionCounterOpts{
			Name: "metric_1",
			Help: "metric_1 help",
			ConstLabels: prometheus.Labels{
				"a": "1",
			},
		}

		counter = newCommandExecutionCounter(counterOpts)
	)

	// When metrics are registered
	metricRegisterer.MustRegister(counter)
	counter.Inc()

	mfs, e := metricRegistry.Gather()
	if e != nil {
		t.Fatalf("*** failed to gather metrics: %v", e)
	}
	// Then the component metric will be labeled with the component ID automatically
	mf := metric.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
		for _, m := range mf.Metric {
			for _, labelPair := range m.Label {
				if *labelPair.Name == metric.ComponentID.String() && *labelPair.Value == compDesc.ID.String() {
					return true
				}
			}
		}
		return false
	})
	if mf == nil {
		t.Fatal("*** metric was not found containing component ID label")
	}
	t.Log(mf)

}
