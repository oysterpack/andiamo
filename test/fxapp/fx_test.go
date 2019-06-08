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

type NewFoo func() *Foo

func NewFoo1() *Foo {
	return &Foo{value: "Foo1"}
}

func NewFoo2() *Foo {
	return &Foo{value: "Foo2"}
}

type BarConfig struct {
	fx.In
	Foos []*Foo `group:"Foo"`
}

func NewBar(in BarConfig) *Bar {
	return &Bar{in.Foos}
}

func NewFoos() fx.Option {
	return fx.Provide(
		fx.Annotated{Group: "Foo", Target: NewFoo1},
		fx.Annotated{Group: "Foo", Target: NewFoo2},
	)
}

func TestFxSimple(t *testing.T) {
	var (
		newFoo1 NewFoo = NewFoo1
		newFoo2 NewFoo = NewFoo2
	)

	app := fxtest.New(t,
		fx.Provide(
			fx.Annotated{Group: "Foo", Target: newFoo1},
			fx.Annotated{Group: "Foo", Target: newFoo2},
			NewBar,
		),
		fx.Invoke(func(bar *Bar, graph fx.DotGraph) {
			log.Println(graph)
			log.Printf("bar: %v", bar)
			for i, foo := range bar.foos {
				log.Printf("bar.foos[%d]: %v", i, foo)
			}
		}),
	)
	app.RequireStart()
}

func TestFxGroup(t *testing.T) {
	app := fxtest.New(t,
		NewFoos(),
		fx.Provide(NewBar),
		fx.Invoke(func(bar *Bar, graph fx.DotGraph) {
			log.Println(graph)
			log.Printf("TestFxGroup: bar: %v", bar)
			log.Printf("TestFxGroup: len(bar.foos): %v", len(bar.foos))
			for i, foo := range bar.foos {
				log.Printf("TestFxGroup: bar.foos[%d]: %v", i, foo)
			}
		}),
	)
	app.RequireStart()
}
