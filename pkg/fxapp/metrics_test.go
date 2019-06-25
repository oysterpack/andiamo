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
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"testing"
)

// the app provides support for prometheus metrics automatically
// - prometheus Gathererer and Registerer are automatically provided by the app
func TestMetricsRegistryProvided(t *testing.T) {

	type FooCounter prometheus.Counter

	var metricRegisterer prometheus.Registerer
	var metricGatherer prometheus.Gatherer
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(func(registerer prometheus.Registerer) (FooCounter, error) {
			counter := prometheus.NewCounter(prometheus.CounterOpts{
				Name: "foo",
				Help: "foo help",
			})
			err := registerer.Register(counter)
			if err != nil {
				return nil, err
			}
			return FooCounter(counter), nil
		}).
		Invoke(func(counter FooCounter) {
			counter.Inc()
		}).
		Invoke(func(gatherer prometheus.Gatherer, logger *zerolog.Logger) error {
			mfs, err := gatherer.Gather()
			if err != nil {
				return err
			}

			for _, mf := range mfs {
				logger.Info().
					Str("name", *mf.Name).
					Str("help", *mf.Help).
					Str("type", mf.Type.String()).
					Msg("")
			}

			return nil
		}).
		Populate(&metricGatherer, &metricRegisterer).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build error: %v", err)
	case metricRegisterer == nil || metricGatherer == nil:
		if metricRegisterer == nil {
			t.Error("*** prometheus.Registerer target was not populated")
		}
		if metricGatherer == nil {
			t.Error("*** prometheus.Gatherer target was not populated")
		}
	}

}
