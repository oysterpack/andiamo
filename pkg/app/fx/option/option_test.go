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

package option_test

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"go.uber.org/fx"
	"reflect"
	"strings"
	"testing"
)

func TestDesc(t *testing.T) {
	type Greeter func() string
	type ProvideGreeting func() Greeter

	type InvokeLogger func(Greeter)

	// Given an option descriptor is defined
	foo := option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeting(nil)))
	t.Logf("%s", foo)

	bar := option.NewDesc(option.Invoke, reflect.TypeOf(InvokeLogger(nil)))
	t.Logf("%s", bar)

	// When the option function is bound
	fooOption, e := foo.Bind(func() Greeter {
		return func() string {
			return "hello"
		}
	})
	if e != nil {
		t.Fatalf("Failed to bind the func to the Desc: %v", e)
	}
	t.Log(fooOption.String())

	barOption := bar.NewOption(func(greeter Greeter) {
		t.Logf("greeting: %s", greeter())
	})

	// Then the option can be used within the fx.App
	fxapp := fx.New(fooOption.FxOption(), barOption.FxOption())
	if e := fxapp.Start(context.Background()); e != nil {
		t.Errorf("failed to start the app")
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Errorf("failed to stop the app")
	}
}

func TestDesc_Bind(t *testing.T) {
	t.Run("unassignable binding", func(t *testing.T) {
		type Greeter func() string
		type ProvideGreeting func() Greeter

		// Given an option descriptor is defined
		foo := &option.Desc{
			Type:     option.Provide,
			FuncType: reflect.TypeOf(ProvideGreeting(nil)),
		}
		t.Logf("%s", foo)

		_, e := foo.Bind(func() {})
		if e == nil {
			t.Fatal("Should have failed to bind func to the Desc")
		}
		t.Log(e)
	})
}

func TestDesc_MustBind(t *testing.T) {
	t.Run("unassignable binding", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("Should have failed to bind func to the Desc")
			} else {
				t.Log(e)
			}
		}()
		type Greeter func() string
		type ProvideGreeting func() Greeter

		// Given an option descriptor is defined
		foo := &option.Desc{
			Type:     option.Provide,
			FuncType: reflect.TypeOf(ProvideGreeting(nil)),
		}
		t.Logf("%s", foo)

		foo.NewOption(func() {})
	})
}

func TestType_String(t *testing.T) {
	t.Run("undefined type", func(t *testing.T) {
		s := option.Type(0)
		if !strings.Contains(s.String(), "undefined") {
			t.Errorf("option.Type(0) should be undefined")
		}
	})
}

func TestType_NewOption(t *testing.T) {
	t.Run("undefined type", func(t *testing.T) {
		defer func() {
			if e := recover(); e == nil {
				t.Error("Type.Option(0) should have panicked because Type(0) is undefined")
			} else {
				t.Log(e)
			}
		}()
		undefined := option.Type(0)
		undefined.Option(func() {})
	})
}
