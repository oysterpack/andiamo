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

package metric

import (
	dto "github.com/prometheus/client_model/go"
)

// Label is used to define standard metric label names
type Label string

func (l Label) String() string {
	return string(l)
}

// standard application metric labels
const (
	AppID         Label = "a"
	AppReleaseID  Label = "r"
	AppInstanceID Label = "x"

	ComponentID Label = "c"
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

// Desc is used to describe the metric
type Desc struct {
	Name string
	Help string
	Type
	Labels []string
}

// DescsFromMetricFamilies extracts metric descriptors from gathered metrics
func DescsFromMetricFamilies(mfs []*dto.MetricFamily) []*Desc {
	if len(mfs) == 0 {
		return nil
	}

	metrics := make([]*Desc, len(mfs))
	for i, mf := range mfs {
		m := &Desc{
			Name: *mf.Name,
			Help: *mf.Help,
			Type: fromMetricType(*mf.Type),
		}
		if len(mf.Metric) > 0 {
			m.Labels = getLabels(mf.Metric[0])
		}
		metrics[i] = m
	}

	return metrics
}

// Type represents a metric type enum
type Type uint8

// metric type enum values
const (
	Untyped Type = iota
	Counter
	Gauge
	Histogram
	Summary
)

func fromMetricType(t dto.MetricType) Type {
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
