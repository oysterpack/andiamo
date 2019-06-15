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

package comp

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Desc is the component descriptor.
type Desc struct {
	ID   ulid.ULID
	Name string
	*semver.Version
	app.Package

	// defines component interface, i.e., functionality
	OptionDescs []option.Desc

	EventRegistry *logging.EventRegistry
	ErrorRegistry *err.Registry

	metrics      *prometheus.Registry
	counterDescs []prometheus.CounterOpts
}

// RegisterCounter specifies a new counter metric.
//
// The component name will be used as the subsystem.
// Namespace will be blanked out.
func (d *Desc) RegisterCounter(opts prometheus.CounterOpts) (prometheus.Counter, error) {
	opts.Namespace = ""
	opts.Subsystem = d.Name
	counter := prometheus.NewCounter(opts)
	e := d.metrics.Register(counter)
	if e != nil {
		return nil, e
	}
	d.counterDescs = append(d.counterDescs, opts)
	return counter, nil
}

// CounterDescs returns the set of registered counter descriptors
func (d *Desc) CounterDescs() []prometheus.CounterOpts {
	if len(d.counterDescs) == 0 {
		return nil
	}

	descs := make([]prometheus.CounterOpts, len(d.counterDescs))
	copy(descs, d.counterDescs)
	return descs
}

func (d *Desc) String() string {
	return fmt.Sprintf("Desc{ID=%s, Name=%s, Version=%s, Package=%s, OptionDescs=%v}", d.ID, d.Name, d.Version, d.Package, d.OptionDescs)
}

// Logger adds the comp's package and name to the specified logger
//
// NOTE: if the logger already has the package or component fields, then they will be duplicated.
func (d *Desc) Logger(l *zerolog.Logger) *zerolog.Logger {
	return logging.ComponentLogger(logging.PackageLogger(l, d.Package), d.Name)
}

// MustNewComp builds a new Comp using the specified options.
//
// Panics if the options don't match the options defined by the component descriptor. The order of the options doesn't matter.
// The options must match on the option types declared by the descriptor. They will be sorted according to the order they
// are listed in the descriptor
func (d *Desc) MustNewComp(options ...option.Option) *Comp {
	c, e := d.NewComp(options...)

	if e != nil {
		panic(e)
	}

	return c
}

// NewComp builds a new Comp using the specified options.
//
// Errors
// - OptionCountDoesNotMatchErr
// - OptionDescTypeNotMatchedErr
func (d *Desc) NewComp(options ...option.Option) (*Comp, error) {
	if len(options) != len(d.OptionDescs) {
		return nil, OptionCountDoesNotMatchErr.CausedBy(fmt.Errorf("expected %d options, but only %d were specified", len(d.OptionDescs), len(options)))
	}

	// sort the options in the same order matching Desc.OptionDescs
	compOptions := make([]option.Option, 0, len(options))
OptionDescsLoop:
	for _, optionDesc := range d.OptionDescs {
		for _, opt := range options {
			if opt.Desc == optionDesc {
				compOptions = append(compOptions, opt)
				continue OptionDescsLoop
			}
		}
		return nil, OptionDescTypeNotMatchedErr.CausedBy(fmt.Errorf("no option found for descriptor: %s", optionDesc))
	}

	c := &Comp{
		Desc:    d,
		Options: compOptions,
	}
	return c, nil
}

// MustNewDesc constructs a new component descriptor.
//
// At least 1 option is required - a component without any application options is useless.
func MustNewDesc(id ID, name Name, version Version, pkg app.Package, optionDescs ...option.Desc) *Desc {
	desc, e := NewDesc(id, name, version, pkg, optionDescs...)
	if e != nil {
		panic(e)
	}

	return desc
}

// NewDesc constructs a new component descriptor.
//
// At least 1 option is required - a component without any application options is useless.
//
// Errors
// - OptionsRequiredErr
// - UniqueOptionTypeConstraintErr
func NewDesc(id ID, name Name, version Version, pkg app.Package, optionDescs ...option.Desc) (*Desc, error) {
	if len(optionDescs) == 0 {
		return nil, OptionsRequiredErr.CausedBy(fmt.Errorf("ID: %s, Name: %s, Package: %s", id, name, pkg))
	}

	desc := &Desc{
		ID:            id.MustParse(),
		Name:          name.String(),
		Version:       version.MustParse(),
		Package:       pkg,
		EventRegistry: logging.NewEventRegistry(),
		ErrorRegistry: err.NewRegistry(),
		metrics:       prometheus.NewPedanticRegistry(),
	}

	// verify that option types are unique
	optionType := make(map[option.Desc]bool, len(optionDescs))
	desc.OptionDescs = make([]option.Desc, len(optionDescs))
	for i, optionDesc := range optionDescs {
		if optionType[optionDesc] {
			return nil, UniqueOptionTypeConstraintErr.CausedBy(fmt.Errorf("duplicate desc: %v", optionDesc))
		}
		optionType[optionDesc] = true
		desc.OptionDescs[i] = optionDesc
	}

	return desc, nil
}

// ID is the component ULID ID.
type ID string

// MustParse parses the ID into a ULID.
func (id ID) MustParse() ulid.ULID {
	return ulid.MustParseStrict(string(id))
}

func (id ID) String() string {
	return string(id)
}

// Name is the component name.
type Name string

func (n Name) String() string {
	return string(n)
}

// Version is the component version.
// It must follow semver naming conventions.
type Version string

// MustParse tries to parse the version.
func (v Version) MustParse() *semver.Version {
	return semver.MustParse(string(v))
}
