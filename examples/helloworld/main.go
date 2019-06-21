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

package main

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/metric"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
	"reflect"
	"time"
)

type empty struct{}

// counter metric
type HelloCounter prometheus.Counter

// registers the metric and provides it to the app
type HelloCounterProvider func(registerer prometheus.Registerer) (HelloCounter, error)

// Hello simply logs hello
type Hello func()

// HelloProvider is the Hello factory method
type HelloProvider func(logger *zerolog.Logger, counter HelloCounter) Hello

// HelloService says hello every second
type HelloService func(hello Hello, lc fx.Lifecycle, logger *zerolog.Logger, metricsGatherer prometheus.Gatherer)

// Hello Comp
var (
	HelloProviderOptionDesc        = option.NewDesc(option.Provide, reflect.TypeOf(HelloProvider(nil)))
	HelloServiceOptionDesc         = option.NewDesc(option.Invoke, reflect.TypeOf(HelloService(nil)))
	HelloCounterProviderOptionDesc = option.NewDesc(option.Provide, reflect.TypeOf(HelloCounterProvider(nil)))

	HelloCompDesc = comp.NewDescBuilder().
			ID("01DDER5DK2KA7AC0S4YYWFD73V").
			Name("hello").
			Version("0.1.0").
			Package(app.GetPackage(empty{})).
			Options(
			HelloProviderOptionDesc,
			HelloServiceOptionDesc,
			HelloCounterProviderOptionDesc,
		).
		MustBuild()

	HelloComp = HelloCompDesc.MustNewComp(
		HelloProviderOptionDesc.NewOption(func(logger *zerolog.Logger, counter HelloCounter) Hello {
			return func() {
				counter.Inc()
				logger.Info().Msg("hello")
			}
		}),
		HelloServiceOptionDesc.NewOption(func(hello Hello, lc fx.Lifecycle, logger *zerolog.Logger, metricsGatherer prometheus.Gatherer) {
			stop := make(chan struct{})
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						for {
							select {
							case <-stop:
								return
							case <-time.After(time.Second):
								hello()
							}
						}
					}()

					return nil
				},
				OnStop: func(context.Context) error {
					close(stop)
					compLogger := HelloCompDesc.Logger(logger)
					mfs, e := metricsGatherer.Gather()
					if e != nil {
						return e
					}
					mf := metric.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
						return *mf.Name == "hello_count"
					})
					if mf != nil {
						compLogger.Info().Msgf("hello_count = %v", *mf.Metric[0].Counter.Value)
					}
					return nil
				},
			})
			logger.Info().Msg("hello service has been initialized")
		}),
		HelloCounterProviderOptionDesc.NewOption(func(registerer prometheus.Registerer) (HelloCounter, error) {
			registerer = HelloCompDesc.WrapRegisterer(registerer)
			counter := prometheus.NewCounter(prometheus.CounterOpts{
				Name: "hello_count",
				Help: "Hello invocation count",
			})
			e := registerer.Register(counter)
			if e != nil {
				return nil, e
			}
			return counter, nil
		}),
	)
)

func main() {
	fxapp, e := appfx.NewAppBuilder().
		AppDesc(app.Desc{
			ID:        app.ID(ulidgen.MustNew()),
			Name:      "helloworld",
			Version:   app.MustParseVersion("0.1.0"),
			ReleaseID: app.ReleaseID(ulidgen.MustNew()),
		}).
		Comps(HelloComp).
		Build()
	if e != nil {
		log.Panic(e)
	}
	go func() {
		if e := fxapp.Run(); e != nil {
			log.Fatal(e)
		}
	}()
	<-fxapp.Stopped()
}
