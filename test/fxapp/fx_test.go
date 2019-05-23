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
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"log"
	"testing"
)

type Foo struct {
	value string
}

func (f *Foo) String() string {
	return f.value
}

type Bar struct {
	foos []*Foo
}

func NewFoo1() FooInstance {
	return FooInstance{
		Foo: &Foo{value: "Foo1"},
	}
}

func NewFoo2() FooInstance {
	return FooInstance{
		Foo: &Foo{value: "Foo2"},
	}
}

type BarConfig struct {
	fx.In
	Foos []*Foo `group:"Foo"`
}

func NewBar(in BarConfig) *Bar {
	return &Bar{in.Foos}
}

type FooInstance struct {
	fx.Out
	Foo *Foo `group:"Foo"`
}

func TestFxSimple(t *testing.T) {
	fxtest.New(t,
		fx.Provide(
			NewFoo1,
			NewFoo2,
			NewBar,
		),
		fx.Invoke(func(bar *Bar) {
			log.Printf("bar: %v", bar)
			for i, foo := range bar.foos {
				log.Printf("bar.foos[%d]: %v", i, foo)
			}
		}),
	)
}
