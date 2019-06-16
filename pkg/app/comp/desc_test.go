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
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/rs/zerolog"
	"reflect"
	"testing"
)

func NewDesc(id, name, version string) *comp.Desc {
	// define some app options
	type Greeter func() string
	type ProvideGreeter func() Greeter
	type LogGreeting func(Greeter)

	var (
		// option descriptors
		GreeterDesc     = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc = option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))
	)

	return comp.MustNewDesc(
		comp.ID(id),
		comp.Name(name),
		comp.Version(version),
		Package,
		// Specify the component's option descriptors
		GreeterDesc,
		LogGreetingDesc,
	)
}

func TestDesc(t *testing.T) {
	// define some app options
	type Greeter func() string
	type ProvideGreeter func() Greeter

	type LogGreeting func(Greeter)

	var (
		// option descriptors
		GreeterDesc     = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc = option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))

		// component descriptor
		Foo = comp.MustNewDesc(
			comp.ID("01DCVEMPTQRNED7XDKGB30V2CF"),
			comp.Name("Foo"),
			comp.Version("0.1.0"),
			Package,
			// Specify the component's option descriptors
			GreeterDesc,
			LogGreetingDesc)
	)

	t.Log(Foo)
	if Foo.ID != ulid.MustParse("01DCVEMPTQRNED7XDKGB30V2CF") {
		t.Error("ID did not match")
	}

	if Foo.Name != "Foo" {
		t.Error("Name did not match")
	}

	if !Foo.Version.Equal(semver.MustParse("0.1.0")) {
		t.Error("Version did not match")
	}

	if Foo.Package != Package {
		t.Error("Package did not match")
	}

	if len(Foo.OptionDescs) != 2 {
		t.Error("Option count did not match")
	}

}

func TestMustNewDesc(t *testing.T) {
	t.Run("no option descriptors defined", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("comp.MustNewDesc() should have panicked because no option descriptors were specified")
			} else {
				t.Log(e)
				e2 := e.(*err.Instance)
				if e2.SrcID != comp.OptionsRequiredErr.SrcID {
					t.Errorf("unexpected error: %v", e2.SrcID)
				}
			}
		}()
		comp.MustNewDesc(
			comp.ID("01DCVEMPTQRNED7XDKGB30V2CF"),
			comp.Name("Foo"),
			comp.Version("0.1.0"),
			Package,
		)
	})

	t.Run("duplicate option descriptor type", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("comp.MustNewDesc() should have panicked because duplicate descriptors were specified")
			} else {
				t.Log(e)
				e2 := e.(*err.Instance)
				if e2.SrcID != comp.UniqueOptionTypeConstraintErr.SrcID {
					t.Errorf("unexpected error: %v", e2.SrcID)
				}
			}
		}()

		// define some app options
		type Greeter func() string
		type ProvideGreeter func() Greeter
		type LogGreeting func(Greeter)

		GreeterDesc := option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc := option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))

		comp.MustNewDesc(
			comp.ID("01DCVEMPTQRNED7XDKGB30V2CF"),
			comp.Name("Foo"),
			comp.Version("0.1.0"),
			Package,
			GreeterDesc,
			GreeterDesc,
			LogGreetingDesc,
		)
	})
}

func TestDesc_MustNewComp(t *testing.T) {
	t.Run("no options specified", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("should have panicked because no options were specified")
			} else {
				t.Log(e)
				e2 := e.(*err.Instance)
				if e2.SrcID != comp.OptionCountDoesNotMatchErr.SrcID {
					t.Errorf("unexpected error: %v", e2.SrcID)
				}
			}
		}()
		desc := NewDesc(ulidgen.MustNew().String(), "foo", "0.1.0")
		desc.MustNewComp()
	})

	t.Run("matching option missing", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("should have panicked because no options were specified")
			} else {
				t.Log(e)
				e2 := e.(*err.Instance)
				if e2.SrcID != comp.OptionDescTypeNotMatchedErr.SrcID {
					t.Errorf("unexpected error: %v", e2.SrcID)
				}
			}
		}()
		desc := NewDesc(ulidgen.MustNew().String(), "foo", "0.1.0")

		type F func()
		invalidOption := option.NewDesc(option.Provide, reflect.TypeOf(F(nil))).NewOption(func() {})
		invalidOptions := make([]option.Option, len(desc.OptionDescs))
		for i := 0; i < len(invalidOptions); i++ {
			invalidOptions[i] = invalidOption
		}
		desc.MustNewComp(invalidOptions...)
	})
}

func TestDesc_EventRegistry(t *testing.T) {
	event1 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)
	event2 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)

	type Foo func()
	optionDesc := option.NewDesc(option.Invoke, reflect.TypeOf(Foo(nil)))

	compDesc := comp.MustNewDesc(
		comp.ID(ulidgen.MustNew().String()),
		comp.Name("foo"),
		comp.Version("0.1.0"),
		Package,
		optionDesc,
	)
	compDesc.EventRegistry.Register(
		event1,
		event2,
		// dup events will get deduped
		event1,
		event2,
	)

	if compDesc.EventRegistry.Count() != 2 {
		t.Errorf("unexpected number of events")
	}
}

func TestDesc_ErrorRegistry(t *testing.T) {
	errDesc1 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc1")
	err1 := err.New(errDesc1, ulidgen.MustNew().String())
	err2 := err.New(errDesc1, ulidgen.MustNew().String())

	errDesc2 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc2")
	err3 := err.New(errDesc2, ulidgen.MustNew().String())

	type Foo func()
	optionDesc := option.NewDesc(option.Invoke, reflect.TypeOf(Foo(nil)))

	compDesc := comp.MustNewDesc(
		comp.ID(ulidgen.MustNew().String()),
		comp.Name("foo"),
		comp.Version("0.1.0"),
		Package,
		optionDesc,
	)
	e := compDesc.ErrorRegistry.Register(
		err1,
		err2,
		err3,
		// dup errs will get deduped
		err1,
	)
	if e != nil {
		t.Fatal(e)
	}

	if compDesc.ErrorRegistry.Count() != 3 {
		t.Errorf("unexpected number of errors: %v", compDesc.ErrorRegistry.Errs())
	}

}

func TestDescBuilder(t *testing.T) {
	type Foo func()
	option1Desc := option.NewDesc(option.Invoke, reflect.TypeOf(Foo(nil)))

	type Bar func()
	option2Desc := option.NewDesc(option.Provide, reflect.TypeOf(Bar(nil)))

	event1 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)
	event2 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)

	errDesc1 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc1")
	err1 := err.New(errDesc1, ulidgen.MustNew().String())
	err2 := err.New(errDesc1, ulidgen.MustNew().String())

	errDesc2 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc2")
	err3 := err.New(errDesc2, ulidgen.MustNew().String())

	options := []option.Desc{option1Desc, option2Desc}
	events := []*logging.Event{event1, event2}
	errs := []*err.Err{err1, err2, err3}

	compID := ulidgen.MustNew()
	desc, e := comp.NewDescBuilder().
		ID(compID.String()).
		Name("foo").
		Version("0.1.0").
		Package(Package).
		Options(options...).
		Events(events...).
		Errors(errs...).
		Build()

	if e != nil {
		t.Fatalf("*** comp desc failed to build: %v", e)
	}
	if desc.ID != compID {
		t.Errorf("*** comp ID did not match: %s != %s", desc.ID, compID)
	}
	if desc.Name != "foo" {
		t.Errorf("*** comp Name did not match: %s", desc.Name)
	}
	if desc.Version.String() != "0.1.0" {
		t.Errorf("*** comp Version did not match: %s", desc.Version)
	}
	if desc.Package != Package {
		t.Errorf("*** comp Package did not match: %s", desc.Package)
	}
	if len(desc.OptionDescs) != 2 && desc.OptionDescs[0] != option1Desc && desc.OptionDescs[1] != option2Desc {
		t.Errorf("comp is missing option desc")
	}
OPTION_LOOP:
	for _, opt := range options {
		for _, registeredOption := range desc.OptionDescs {
			if opt == registeredOption {
				continue OPTION_LOOP
			}
		}
		t.Errorf("*** option not found: %v", opt)
	}
EVENT_LOOP:
	for _, event := range events {
		for _, registeredEvent := range desc.EventRegistry.Events() {
			if event.Equals(registeredEvent) {
				continue EVENT_LOOP
			}
		}
		t.Errorf("*** event not found: %v", event)
	}

ERR_LOOP:
	for _, e := range errs {
		for _, registeredErr := range desc.ErrorRegistry.Errs() {
			if registeredErr.SrcID == e.SrcID {
				continue ERR_LOOP
			}
		}
		t.Errorf("*** err not found: %v", e)
	}
}

func TestDescBuilder_Build(t *testing.T) {
	t.Run("Using invalid ID", func(t *testing.T) {
		_, e := comp.NewDescBuilder().Build()
		if e == nil {
			t.Errorf("*** desc should have failed to build because required options are missing")
		}
	})
}
