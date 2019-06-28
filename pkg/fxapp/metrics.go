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
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/rs/zerolog"
	"time"
)

// FindMetricFamily returns the first metric family that matches the filter
func FindMetricFamily(mfs []*dto.MetricFamily, accept func(mf *dto.MetricFamily) bool) *dto.MetricFamily {
	for _, mf := range mfs {
		if accept(mf) {
			return mf
		}
	}
	return nil
}

// FindMetricFamilies returns first metric families that match the filter
func FindMetricFamilies(mfs []*dto.MetricFamily, accept func(mf *dto.MetricFamily) bool) []*dto.MetricFamily {
	var result []*dto.MetricFamily
	for _, mf := range mfs {
		if accept(mf) {
			result = append(result, mf)
		}
	}
	return result
}

// MetricDesc is used to describe the metric
type MetricDesc struct {
	Name string
	Help string
	MetricType
	Labels []string
}

// DescsFromMetricFamilies extracts metric descriptors from gathered metrics
func DescsFromMetricFamilies(mfs []*dto.MetricFamily) []*MetricDesc {
	if len(mfs) == 0 {
		return nil
	}

	metrics := make([]*MetricDesc, len(mfs))
	for i, mf := range mfs {
		metrics[i] = NewMetricDesc(mf)
	}

	return metrics
}

// NewMetricDesc extracts the metric descriptor from the gathered metrics
func NewMetricDesc(mf *dto.MetricFamily) *MetricDesc {
	m := &MetricDesc{
		Name:       *mf.Name,
		Help:       *mf.Help,
		MetricType: mapMetricType(*mf.Type),
	}
	if len(mf.Metric) > 0 {
		m.Labels = getLabels(mf.Metric[0])
	}
	return m
}

// MetricType represents a metric type enum
type MetricType uint8

// metric type enum values
const (
	Untyped MetricType = iota
	Counter
	Gauge
	Histogram
	Summary
)

func mapMetricType(t dto.MetricType) MetricType {
	switch t {
	case dto.MetricType_COUNTER:
		return Counter
	case dto.MetricType_GAUGE:
		return Gauge
	case dto.MetricType_HISTOGRAM:
		return Histogram
	case dto.MetricType_SUMMARY:
		return Summary
	default:
		return Untyped
	}
}

func getLabels(m *dto.Metric) []string {
	if len(m.Label) == 0 {
		return nil
	}

	names := make([]string, len(m.Label))
	for i, labelPair := range m.Label {
		names[i] = *labelPair.Name
	}
	return names
}

// PrometheusHTTPHandlerOpts PrometheusHTTPServer options
type PrometheusHTTPHandlerOpts struct {
	Timeout  time.Duration // default = 5 seconds
	Endpoint string        // default = "/metrics"
	promhttp.HandlerErrorHandling
}

// NewHTTPHandler constructs a new HTTPHandler from the PrometheusHTTPHandlerOpts
func (opts PrometheusHTTPHandlerOpts) NewHTTPHandler(gatherer prometheus.Gatherer, registerer prometheus.Registerer, logger *zerolog.Logger) HTTPHandler {
	if opts.Endpoint == "" {
		opts.Endpoint = "/metrics"
	}
	if opts.Timeout == time.Duration(0) {
		opts.Timeout = 5 * time.Second
	}

	errorLog := prometheusHTTPErrorLog(PrometheusHTTPError.NewLogEventer(logger, zerolog.ErrorLevel))
	promhttpHandlerOpts := promhttp.HandlerOpts{
		ErrorLog:            errorLog,
		ErrorHandling:       opts.HandlerErrorHandling,
		Registry:            registerer,
		MaxRequestsInFlight: 3,
		Timeout:             opts.Timeout,
	}
	handler := promhttp.HandlerFor(gatherer, promhttpHandlerOpts)
	return NewHTTPHandler(opts.Endpoint, handler.ServeHTTP)
}

// PrometheusHTTPError indicates an error occurred while handling a metrics scrape HTTP request.
const PrometheusHTTPError EventTypeID = "01DEARG17HNQ606ARQNYFY7PG5"

type prometheusHTTPErrorLog LogEventer

func (errLog prometheusHTTPErrorLog) Println(v ...interface{}) {
	errLog(prometheusHTTPError(fmt.Sprint(v...)), "prometheus HTTP handler error")
}

type prometheusHTTPError string

func (err prometheusHTTPError) MarshalZerologObject(e *zerolog.Event) {
	e.Err(errors.New(string(err)))
}
