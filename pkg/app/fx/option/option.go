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

// Package option is used to declare application options at compile time.
package option

import (
	"fmt"
	"go.uber.org/fx"
	"reflect"
)

// Type is used to specify the fx.Opion type
type Type uint8

// supported types
const (
	// Provide indicates the fx.Option provides a constructor function
	Provide = Type(iota + 1)
	// Invoke indicates the fx.Option provides a function that is meant to be invoked by the app
	Invoke
)

func (t Type) String() string {
	switch t {
	case Provide:
		return "Provide"
	case Invoke:
		return "Invoke"
	default:
		return fmt.Sprintf("undefined: %v", uint8(t))
	}
}

// Option constructs a new fx.Option based on the Type
//
// Panics if the Type is not defined
func (t Type) Option(f interface{}) fx.Option {
	switch t {
	case Provide:
		return fx.Provide(f)
	case Invoke:
		return fx.Invoke(f)
	default:
		// should never happen
		panic(t)
	}
}

// Desc is used as an fx.Option descriptor
type Desc struct {
	Type
	FuncType reflect.Type
}

// NewDesc constructs a new Desc instance
func NewDesc(t Type, funcType reflect.Type) Desc {
	return Desc{t, funcType}
}

func (d Desc) String() string {
	return fmt.Sprintf("Desc(%s => %v)", d.Type, d.FuncType)
}

// Bind constructs a new Option, binding the specified function. f is type checked against the Desc.FuncType and must be
// assignable to its type.
func (d Desc) Bind(f interface{}) (option Option, e error) {
	if !reflect.TypeOf(f).AssignableTo(d.FuncType) {
		e = UnassignableBindingErr.CausedBy(fmt.Errorf("`%T` is not assignable to `%s`", f, d.FuncType))
		return
	}
	option.Desc = d
	option.Option = d.Type.Option(f)
	return
}

// NewOption constructs a new Option, binding the specified function. f is type checked against the Desc.FuncType and must be
// assignable to its type.
func (d Desc) NewOption(f interface{}) Option {
	opt, e := d.Bind(f)
	if e != nil {
		panic(e)
	}
	return opt
}

// Option binds an option descriptor to the option function
type Option struct {
	Desc
	fx.Option
}
