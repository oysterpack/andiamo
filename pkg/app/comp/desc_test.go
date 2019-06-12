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
	event1 := logging.NewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)
	event2 := logging.NewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)

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

	if len(compDesc.EventRegistry.Events()) != 2 {
		t.Errorf("unexpected number of events")
	}
}
