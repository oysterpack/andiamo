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
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/metric"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.uber.org/fx"
	"math/rand"
	"reflect"
	"testing"
)

func TestDesc_WrapRegisterer(t *testing.T) {
	// metric dependencies should be strongly typed - the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter
	// It is the metric provider's responsibility to register the counter
	type CommandExecutionCounterProvider func(registerer prometheus.Registerer) (CommandExecutionCounter, error)
	commandExecutionCounterOptionDesc := option.NewDesc(option.Provide, reflect.TypeOf(CommandExecutionCounterProvider(nil)))

	type Command func(ctx context.Context) error
	// metrics are injected
	type CommandProvider func(CommandExecutionCounter) Command
	optionDesc := option.NewDesc(option.Provide, reflect.TypeOf(CommandProvider(nil)))

	compDesc, e := comp.NewDescBuilder().
		ID(ulidgen.MustNew().String()).
		Name("foo").
		Version("0.1.0").
		Package(Package).
		Options(optionDesc, commandExecutionCounterOptionDesc).
		Build()

	if e != nil {
		t.Fatalf("*** comp desc failed to build: %v", e)
	}

	metricRegistry := prometheus.NewRegistry()
	// Given a component metric registerer
	metricRegisterer := compDesc.WrapRegisterer(metricRegistry)
	var (
		counterOpts = prometheus.CounterOpts{
			Name: "metric_1",
			Help: "metric_1 help",
			ConstLabels: prometheus.Labels{
				"a": "1",
			},
		}

		counter = prometheus.NewCounter(counterOpts)
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

type GobalErrorCounter prometheus.Counter

func globalErrorCounterProvider(registerer prometheus.Registerer) (GobalErrorCounter, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "err_count",
		Help: "App error count",
	})
	e := registerer.Register(counter)
	if e != nil {
		return nil, e
	}
	return counter, nil
}

func TestCompMetrics(t *testing.T) {
	apptest.InitEnv()

	// metric dependencies should be strongly typed - the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter
	// It is the metric provider's responsibility to register the counter
	type CommandExecutionCounterProvider func(registerer prometheus.Registerer) (CommandExecutionCounter, error)

	type Command func(ctx context.Context) error
	// metrics are injected
	type CommandProvider func(CommandExecutionCounter, GobalErrorCounter) Command
	// used to drive the test
	type FooCommandRunner func(cmd Command)

	// foo comp desc
	var (
		commandExecutionCounterOptionDesc = option.NewDesc(option.Provide, reflect.TypeOf(CommandExecutionCounterProvider(nil)))
		fooCommandRunnerDesc              = option.NewDesc(option.Invoke, reflect.TypeOf(FooCommandRunner(nil)))
		commandOptionDesc                 = option.NewDesc(option.Provide, reflect.TypeOf(CommandProvider(nil)))
	)

	fooCompDesc, e := comp.NewDescBuilder().
		ID(ulidgen.MustNew().String()).
		Name("foo").
		Version("0.1.0").
		Package(Package).
		Options(
			commandOptionDesc,
			fooCommandRunnerDesc,
			commandExecutionCounterOptionDesc,
		).
		Build()

	if e != nil {
		t.Fatalf("*** comp desc failed to build: %v", e)
	}

	fooComp := fooCompDesc.MustNewComp(
		commandOptionDesc.NewOption(func(counter CommandExecutionCounter, errorCounter GobalErrorCounter) Command {
			f := func() error {
				if rand.Int()%2 == 0 {
					return nil
				}
				return errors.New("odd number")
			}
			return func(ctx context.Context) error {
				counter.Inc()
				e := f()
				if e != nil {
					errorCounter.Inc()
				}
				return e
			}
		}),
		fooCommandRunnerDesc.NewOption(func(cmd Command) {
			for i := 0; i < 10; i++ {
				cmd(context.Background())
			}
		}),
		commandExecutionCounterOptionDesc.NewOption(func(registerer prometheus.Registerer) (CommandExecutionCounter, error) {
			registerer = fooCompDesc.WrapRegisterer(registerer)
			var counter CommandExecutionCounter = prometheus.NewCounter(prometheus.CounterOpts{
				Name: "cmd_exec_counter",
				Help: "Command execution counter",
			})
			e := registerer.Register(counter)
			if e != nil {
				return nil, e
			}
			return counter, nil
		}),
	)

	// When the app is initialized the FooCommandRunner is invoked, which depends on metrics being injected
	var metricsGatherer prometheus.Gatherer
	_, e = appfx.NewAppBuilder().
		Comps(fooComp).
		Options(
			fx.Populate(&metricsGatherer),
			fx.Provide(globalErrorCounterProvider),
		).
		Build()
	if e != nil {
		t.Fatalf("*** failed to build app: %v", e)
	}

	mfs, e := metricsGatherer.Gather()
	if e != nil {
		t.Fatalf("*** failed to gather metrics: %v", e)
	}
	mf := metric.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
		// Then the component metric is registered
		if *mf.Name != "cmd_exec_counter" {
			return false
		}
		// And is labeled with the component ID
		for _, m := range mf.Metric {
			for _, labelPair := range m.Label {
				if *labelPair.Name == metric.ComponentID.String() && *labelPair.Value == fooCompDesc.ID.String() {
					return true
				}
			}
		}
		return false
	})
	if mf == nil {
		t.Fatal("*** metric is not registered")
	}
	t.Log(mf)
}

func TestMetricOptionFuncType(t *testing.T) {
	// metric dependencies should be strongly typed - the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter
	// It is the metric provider's responsibility to register the counter
	type CommandExecutionCounterProvider func(prometheus.Registerer) (CommandExecutionCounter, error)

	inTypes := []reflect.Type{reflect.TypeOf((*prometheus.Registerer)(nil)).Elem()}
	outTypes := []reflect.Type{reflect.TypeOf((*CommandExecutionCounter)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()}
	t.Log(inTypes, "->", outTypes)
	var CommandExecutionCounterProviderType = reflect.FuncOf(inTypes, outTypes, false)
	if !CommandExecutionCounterProviderType.ConvertibleTo(reflect.TypeOf(CommandExecutionCounterProvider(nil))) {
		t.Errorf("types don't match: %v != %v", CommandExecutionCounterProviderType, reflect.TypeOf(CommandExecutionCounterProvider(nil)))
	}
}
