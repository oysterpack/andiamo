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

package eventlog_test

import (
	"github.com/oysterpack/andiamo/pkg/eventlog"
	"github.com/rs/zerolog"
	"os"
)

func Example() {

	// Define your application events
	const (
		Foo = "01DFBM3VV1N11HQB04WZRA4R88"
	)

	// Define your strongly typed logging functions
	FooInfoLogger := func(logger *zerolog.Logger) func(id FooID, msg string, tags ...string) {
		log := eventlog.NewLogger(Foo, logger, zerolog.InfoLevel)
		return func(id FooID, msg string, tags ...string) {
			log(id, msg, tags...)
		}
	}

	// Create your application logging functions
	logger := zerolog.New(os.Stdout)
	logFooInfo := FooInfoLogger(&logger)

	// log some events
	fooID := FooID("01DFBMTSW3J58NG6VGPQ3WTFFZ")
	// tagging events with ULIDs makes it easy to find where the event was logged from in the code
	logFooInfo(fooID, "MSG#1", "01DFBQA67JKT2HKGDVRN3QN38R")
	logFooInfo(fooID, "MSG#2", "01DFBQ9DWKFBA7KJRBZVGV69NN", "tag-2")

	// Output:
	// {"l":"info","n":"01DFBM3VV1N11HQB04WZRA4R88","g":["01DFBQA67JKT2HKGDVRN3QN38R"],"d":{"id":"01DFBMTSW3J58NG6VGPQ3WTFFZ"},"m":"MSG#1"}
	// {"l":"info","n":"01DFBM3VV1N11HQB04WZRA4R88","g":["01DFBQ9DWKFBA7KJRBZVGV69NN","tag-2"],"d":{"id":"01DFBMTSW3J58NG6VGPQ3WTFFZ"},"m":"MSG#2"}

}
