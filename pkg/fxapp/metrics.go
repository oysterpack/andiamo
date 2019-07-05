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

// TODO: log metrics on a scheduled basis

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

// PrometheusHTTPHandlerOpts is used to configure the Prometheus HTTP handler that is used to expose metrics
type PrometheusHTTPHandlerOpts interface {
	WithTimeout(timeout time.Duration) PrometheusHTTPHandlerOpts
	WithEndpoint(path string) PrometheusHTTPHandlerOpts
	HandleErrorWith(errorHandling promhttp.HandlerErrorHandling) PrometheusHTTPHandlerOpts

	// Timeout returns the handler response timeout.
	//
	// If handling a request takes longer than Timeout, it is responded to with 503 ServiceUnavailable and a suitable Message.
	// No timeout is applied if Timeout is 0 or negative. Note that with the current implementation, reaching the timeout
	// simply ends the HTTP requests as described above (and even that only if sending of the body hasn't started yet), while
	// the bulk work of gathering all the metrics keeps running in the background (with the eventual result to be thrown away).
	// Until the implementation is improved, it is recommended to implement a separate timeout in potentially slow Collectors.
	Timeout() time.Duration

	Endpoint() string

	// ErrorHandling defines how errors are handled.
	//
	// // Note that errors are logged regardless of the configured ErrorHandling provided ErrorLog is not nil.
	ErrorHandling() promhttp.HandlerErrorHandling

	// NewHTTPHandler constructs a new HTTPHandler from the PrometheusHTTPHandlerOpts
	//
	// The max requests in flight is limited to 3.
	NewHTTPHandler(gatherer prometheus.Gatherer, registerer prometheus.Registerer, logger *zerolog.Logger) HTTPHandler
}

// PrometheusHTTPHandlerOpts PrometheusHTTPServer options
type prometheusHTTPHandlerOpts struct {
	timeout       time.Duration
	endpoint      string
	errorHandling promhttp.HandlerErrorHandling
}

// NewPrometheusHTTPHandlerOpts constructs a new PrometheusHTTPHandlerOpts with the following options:
// 	- timeout: 5 secs
//	- endpoint: /metrics
//	- error handling: promhttp.HTTPErrorOnError
// 	  - Serve an HTTP status code 500 upon the first error encountered. Report the error message in the body.
func NewPrometheusHTTPHandlerOpts() PrometheusHTTPHandlerOpts {
	return &prometheusHTTPHandlerOpts{
		timeout:  5 * time.Second,
		endpoint: "/metrics",
	}
}

func (opts *prometheusHTTPHandlerOpts) WithTimeout(timeout time.Duration) PrometheusHTTPHandlerOpts {
	opts.timeout = timeout
	return opts
}

// WithEndpoint sets the endpoint path
func (opts *prometheusHTTPHandlerOpts) WithEndpoint(path string) PrometheusHTTPHandlerOpts {
	opts.endpoint = path
	return opts
}

func (opts *prometheusHTTPHandlerOpts) HandleErrorWith(errorHandling promhttp.HandlerErrorHandling) PrometheusHTTPHandlerOpts {
	opts.errorHandling = errorHandling
	return opts
}

func (opts *prometheusHTTPHandlerOpts) Timeout() time.Duration {
	return opts.timeout
}

func (opts *prometheusHTTPHandlerOpts) Endpoint() string {
	return opts.endpoint
}

func (opts *prometheusHTTPHandlerOpts) ErrorHandling() promhttp.HandlerErrorHandling {
	return opts.errorHandling
}

func (opts *prometheusHTTPHandlerOpts) NewHTTPHandler(gatherer prometheus.Gatherer, registerer prometheus.Registerer, logger *zerolog.Logger) HTTPHandler {
	errorLog := prometheusHTTPErrorLog(PrometheusHTTPError.NewLogEventer(logger, zerolog.ErrorLevel))
	promhttpHandlerOpts := promhttp.HandlerOpts{
		ErrorLog:            errorLog,
		ErrorHandling:       opts.errorHandling,
		Registry:            registerer,
		MaxRequestsInFlight: 3,
		Timeout:             opts.timeout,
	}
	handler := promhttp.HandlerFor(gatherer, promhttpHandlerOpts)
	return NewHTTPHandler(opts.endpoint, handler.ServeHTTP)
}

// PrometheusHTTPError indicates an error occurred while handling a metrics scrape HTTP request.
//
// 	type Data struct {
//		Err string `json:"e"`
//	}
const PrometheusHTTPError EventTypeID = "01DEARG17HNQ606ARQNYFY7PG5"

type prometheusHTTPErrorLog LogEventer

// implements promhttp.Logger interface
func (log prometheusHTTPErrorLog) Println(v ...interface{}) {
	log(prometheusHTTPError(fmt.Sprint(v...)), "prometheus HTTP handler error")
}

type prometheusHTTPError string

func (err prometheusHTTPError) MarshalZerologObject(e *zerolog.Event) {
	e.Err(errors.New(string(err)))
}
