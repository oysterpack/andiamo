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

//go:generate -command list ls -larth
package test

//go:generate list
// Foo is used for testing purposes
type Foo struct {
	name string
}

// NewFoo constructs new Foo instances
func NewFoo(name string) Foo {
	return Foo{name}
}

func (f *Foo) String() string {
	return f.name
}

// Name getter
func (f Foo) Name() string {
	return f.name
}
