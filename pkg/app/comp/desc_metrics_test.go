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
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/prometheus/client_golang/prometheus"
	"reflect"
	"strings"
	"testing"
)

func TestDesc_RegisterCounter(t *testing.T) {
	// metric dependencies should be strongly typed
	// the type name should convey the metric's purpose
	type CommandExecutionCounter prometheus.Counter

	type Command func(ctx context.Context) error
	// metrics are injected
	type CommandProvider func(CommandExecutionCounter) Command
	optionDesc := option.NewDesc(option.Provide, reflect.TypeOf(CommandProvider(nil)))

	compDesc := comp.MustNewDesc(
		comp.ID(ulidgen.MustNew().String()),
		comp.Name("foo"),
		comp.Version("0.1.0"),
		Package,
		optionDesc,
	)

	if len(compDesc.CounterDescs()) != 0 {
		t.Errorf("*** there should be no counters registered: %v", compDesc.CounterDescs())
	}

	// When the counter is registered
	counter, e := compDesc.RegisterCounter(prometheus.CounterOpts{
		Name: "metric_1",
		Help: "metric_1 help",
		ConstLabels: prometheus.Labels{
			"a": "1",
		},
	})
	if e != nil {
		t.Fatalf("*** failed to register counter: %v", e)
	}
	counter.Inc()

	// Then the comp name is used as the metric subsystem name
	t.Logf("%v", counter.Desc())
	if !strings.Contains(counter.Desc().String(), "foo_metric_1") {
		t.Errorf("*** metric subsystem part should be the comp name: %v", counter.Desc())
	}

	// Then the registered counter desc can retrieved
	counterDescs := compDesc.CounterDescs()
	exists := false
	for _, desc := range counterDescs {
		if desc.Name == "metric_1" {
			exists = true
			break
		}
	}
	if !exists {
		t.Error("*** registered counter was not returned")
	}

	// When a dup counter is registered
	_, e = compDesc.RegisterCounter(prometheus.CounterOpts{
		Name: "metric_1",
		Help: "metric_1 help",
		ConstLabels: prometheus.Labels{
			"a": "1",
		},
	})
	// Then it should fail to register the counter
	if e == nil {
		t.Fatal("*** counter should have failed to register because it is already registered")
	}
	t.Log(e)

	// When a counter is registered using the same name but different labels
	counter2, e := compDesc.RegisterCounter(prometheus.CounterOpts{
		Name: "metric_1",
		Help: "metric_1 help",
		ConstLabels: prometheus.Labels{
			"a": "2",
		},
	})
	// Then it will register successfully
	if e != nil {
		t.Fatal("*** counter should have registered because the labels are different")
	}
	counter2.Inc()
	if len(compDesc.CounterDescs()) != 2 {
		t.Errorf("*** there should be 2 counters registered but found: %d", len(compDesc.CounterDescs()))
	}
}
