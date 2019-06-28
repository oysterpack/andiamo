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
	"errors"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
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

func TestMetricsContainAppLabels(t *testing.T) {
	type FooCounter prometheus.Counter
	var metricsGatherer prometheus.Gatherer
	var appInstanceID fxapp.InstanceID
	var appDesc fxapp.Desc
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(
			// Given a custom metric is registered
			func(metricRegisterer prometheus.Registerer) (FooCounter, error) {
				counter := prometheus.NewCounter(prometheus.CounterOpts{
					Name: "foo",
					Help: "foo counter",
				})
				err := metricRegisterer.Register(counter)
				if err != nil {
					return nil, err
				}

				return counter, nil
			},
		).
		Populate(&metricsGatherer, &appInstanceID, &appDesc).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** abb build error: %v", err)
	default:
		mfs, err := metricsGatherer.Gather()
		switch {
		case err != nil:
			t.Errorf("*** failed to gather metrics")
		default:
			// Then the custom metric is returned when gathering metrics
			mf := fxapp.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
				return *mf.Name == "foo"
			})
			if mf == nil {
				t.Error("*** foo metric is not registered")
			}

			// And all metrics have app labels
			for _, mf := range mfs {
				labels := mf.GetMetric()[0].GetLabel()
				hasLabel := func(name, value string) bool {
					for _, l := range labels {
						if l.GetName() == name && l.GetValue() == value {
							return true
						}
					}
					return false
				}
				// And the metric has the app labels
				if !hasLabel(fxapp.AppIDLabel, appDesc.ID().String()) {
					t.Errorf("*** app ID label is missing: %s", mf)
				}
				if !hasLabel(fxapp.AppReleaseIDLabel, appDesc.ReleaseID().String()) {
					t.Errorf("*** app ReleaseID label is missing: %s", mf)
				}
				if !hasLabel(fxapp.AppInstanceIDLabel, appInstanceID.String()) {
					t.Errorf("*** app instance ID label is missing: %s", mf)
				}
			}
		}
	}
}

func TestMetricGoCollectorRegistered(t *testing.T) {
	var metricsGatherer prometheus.Gatherer
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func() {}).
		Populate(&metricsGatherer).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** abb build error: %v", err)
	default:
		mfs, err := metricsGatherer.Gather()
		switch {
		case err != nil:
			t.Errorf("*** failed to gather metrics")
		default:
			mfs = fxapp.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
				return strings.HasPrefix(mf.GetName(), "go_")
			})
			if len(mfs) == 0 {
				t.Error("*** go collector metrics are not registered")
			}
		}
	}
}

func TestMetricProcessCollectorRegistered(t *testing.T) {
	var metricsGatherer prometheus.Gatherer
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func() {}).
		Populate(&metricsGatherer).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** abb build error: %v", err)
	default:
		mfs, err := metricsGatherer.Gather()
		switch {
		case err != nil:
			t.Errorf("*** failed to gather metrics")
		default:
			mfs = fxapp.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
				return strings.HasPrefix(mf.GetName(), "process_")
			})
			if len(mfs) == 0 {
				t.Error("*** process metrics are not registered")
			}
		}
	}
}

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
	mf := fxapp.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
		return *mf.Name == "foo1"
	})
	if mf == nil {
		t.Fatal("*** metric was not found")
	}

	mf = fxapp.FindMetricFamily(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
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
		mfs = fxapp.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return *mf.Name == "foo"
		})
		if len(mfs) != 1 {
			t.Fatalf("*** wrong number of metrics were returned: %v", mfs)
		}
	}

	{
		mfs = fxapp.FindMetricFamilies(mfs, func(mf *io_prometheus_client.MetricFamily) bool {
			return false
		})
		if len(mfs) > 0 {
			t.Fatalf("*** no metrics should have been returned: %v", mfs)
		}
	}
}

func TestDescsFromMetricFamilies(t *testing.T) {
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "counter",
		Help: "counter help",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})

	counterNoLabels := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "counterNoLabels",
		Help: "counter help",
	})

	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counterVec",
			Help: "counterVec help",
			ConstLabels: prometheus.Labels{
				"foo": "bar",
			},
		},
		[]string{"x", "y", "z"},
	)
	// metric vecs do not get reported until 1 metric is observed
	counterVec.WithLabelValues("1", "2", "3").Inc()
	counterVec.WithLabelValues("4", "5", "6").Inc()

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gauge",
		Help: "gauge help",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})

	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gaugeVec",
			Help: "gaugeVec help",
			ConstLabels: prometheus.Labels{
				"foo": "bar",
			},
		},
		[]string{"x", "y", "z"},
	)
	gaugeVec.WithLabelValues("1", "2", "3").Inc()

	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "histogram",
		Help: "histogram help",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
		Buckets: []float64{1, 2, 3},
	})

	histogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "histogramVec",
			Help: "histogramVec help",
			ConstLabels: prometheus.Labels{
				"foo": "bar",
			},
			Buckets: []float64{1, 2, 3},
		},
		[]string{"x", "y", "z"},
	)

	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "summary",
		Help: "summary help",
		ConstLabels: prometheus.Labels{
			"foo": "bar",
		},
	})

	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "summaryVec",
			Help: "summaryVec help",
			ConstLabels: prometheus.Labels{
				"foo": "bar",
			},
		},
		[]string{"x", "y", "z"},
	)

	registry.MustRegister(
		counter,
		counterNoLabels,
		counterVec,

		gauge,
		gaugeVec,

		histogram,
		histogramVec,

		summary,
		summaryVec,
	)
	mfs, e := registry.Gather()
	if e != nil {
		t.Fatalf("*** failed to gather metrics: %v", e)
	}

	metrics := fxapp.DescsFromMetricFamilies(nil)
	if len(metrics) > 0 {
		t.Errorf("*** no metrics should have been returned")
	}
	metrics = fxapp.DescsFromMetricFamilies(mfs)
	for _, m := range metrics {
		t.Log(m)
	}

}

func TestExposePrometheusMetricsViaHTTP(t *testing.T) {
	prometheusHTTPServerOpts := fxapp.NewPrometheusHTTPHandlerOpts()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func() {}).
		ExposePrometheusMetricsViaHTTP(prometheusHTTPServerOpts).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failure: %v", err)
	default:
		go app.Run()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()
		<-app.Started()

		// Then the prometheus HTTP server should be running
		resp, err := retryablehttp.Get("http://:8008/metrics")
		switch {
		case err != nil:
			t.Errorf("*** failed to HTTP scrape metrics: %v", err)
		case resp.StatusCode != http.StatusOK:
			t.Errorf("*** /metrics http request failed: %v", resp.Status)
		default:
			metrics, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Errorf("*** failed to read response body")
			} else {
				t.Log(string(metrics))
			}
		}
	}
}

func TestPrometheusHTTPServerRunner_FailOnCollectErrorWithHTTP500(t *testing.T) {
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		ExposePrometheusMetricsViaHTTP(fxapp.NewPrometheusHTTPHandlerOpts()).
		Invoke(
			// register a collector that fails
			func(registerer prometheus.Registerer) error {
				return registerer.Register(FailingMetricCollector{})
			},
		).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failure: %v", err)
	default:
		go app.Run()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()
		<-app.Started()

		// Then the prometheus HTTP server should be running
		resp, err := http.Get("http://:8008/metrics")
		switch {
		case err != nil:
			t.Errorf("*** failed to HTTP scrape metrics: %v", err)
		case resp.StatusCode == http.StatusOK:
			t.Error("request should have failed")
		default:
			t.Logf("status code = %d", resp.StatusCode)
			if resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("*** expected HTTP status 500: %v", resp.StatusCode)
			}
		}
	}
}

func TestPrometheusHTTPServerRunner_ContinueOnCollectError(t *testing.T) {
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		ExposePrometheusMetricsViaHTTP(fxapp.NewPrometheusHTTPHandlerOpts().
			HandleErrorWith(promhttp.ContinueOnError),
		).
		Invoke(
			// register a collector that
			func(registerer prometheus.Registerer) error {
				return registerer.Register(FailingMetricCollector{})
			},
		).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failure: %v", err)
	default:
		go app.Run()
		defer func() {
			app.Shutdown()
			<-app.Done()
		}()
		<-app.Started()

		// Then the prometheus HTTP server should be running
		resp, err := retryablehttp.Get("http://:8008/metrics")
		switch {
		case err != nil:
			t.Errorf("*** failed to HTTP scrape metrics: %v", err)
		case resp.StatusCode != http.StatusOK:
			t.Errorf("*** request failed: %v", resp.StatusCode)
		default:
			metrics, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Errorf("*** failed to read response body")
			} else {
				t.Log(string(metrics))
			}
		}
	}
}

type FailingMetricCollector struct {
}

func (c FailingMetricCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Desc()
}

func (c FailingMetricCollector) Collect(ch chan<- prometheus.Metric) {
	log.Println("FailingMetricCollector.Collect()")
	ch <- c
}

func (c FailingMetricCollector) Desc() *prometheus.Desc {
	return prometheus.NewDesc(
		"clustermanager_oom_crashes_total",
		"Number of OOM crashes.",
		[]string{"host"}, nil,
	)
}

func (c FailingMetricCollector) Write(*dto.Metric) error {
	return errors.New("BOOM!")
}

func TestPrometheusHTTPHandlerOpts_Setters(t *testing.T) {
	opts := fxapp.NewPrometheusHTTPHandlerOpts()
	t.Logf("%#v", opts)
	if opts.Endpoint() != "/metrics" {
		t.Error("*** endpoint option did not match")
	}
	if opts.Timeout() != 5*time.Second {
		t.Error("*** timeout option did not match")
	}
	if opts.ErrorHandling() != promhttp.HTTPErrorOnError {
		t.Error("*** error handling option did not match")
	}

	opts.WithEndpoint("/prometheus/metrics")
	if opts.Endpoint() != "/prometheus/metrics" {
		t.Error("*** endpoint option did not match")
	}

	opts.WithTimeout(time.Minute)
	if opts.Timeout() != time.Minute {
		t.Error("*** timeout option did not match")
	}

	opts.HandleErrorWith(promhttp.ContinueOnError)
	if opts.ErrorHandling() != promhttp.ContinueOnError {
		t.Error("*** error handling option did not match")
	}
}
