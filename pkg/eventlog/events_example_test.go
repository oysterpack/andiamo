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
	"github.com/oysterpack/partire-k8s/pkg/eventlog"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"os"
)

func ExampleEvent_NewLogger() {

	// Define your application events
	const (
		Foo eventlog.Event = "01DFBM3VV1N11HQB04WZRA4R88"
	)

	// Define your strongly typed logging functions
	FooInfoLogger := func(logger *zerolog.Logger) func(id FooID, msg string, tags ...string) {
		log := Foo.NewLogger(logger, zerolog.InfoLevel)
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

func ExampleEvent_NewErrorLogger() {

	// Define your application events
	const (
		Foo eventlog.Event = "01DFBM3VV1N11HQB04WZRA4R88"
	)

	// Define your strongly typed logging functions
	FooErrorLogger := func(logger *zerolog.Logger) func(id FooID, err error, msg string, tags ...string) {
		log := Foo.NewErrorLogger(logger)
		return func(id FooID, err error, msg string, tags ...string) {
			log(id, err, msg, tags...)
		}
	}

	// Create your application logging functions
	logger := zerolog.New(os.Stdout)
	logFooError := FooErrorLogger(&logger)

	// log some events
	fooID := FooID("01DFBMTSW3J58NG6VGPQ3WTFFZ")
	// tagging events with ULIDs makes it easy to find where the event was logged from in the code
	logFooError(fooID, errors.New("FAILURE#1"), "BOOM!", "01DFBQATWAEJHEZ47V1HNXFJGP")
	logFooError(fooID, errors.New("FAILURE#2"), "BOOM!!", "01DFBQB62ANNQ0YQ240K0XGMQW", "tag-2")

	// Output:
	// {"l":"error","n":"01DFBM3VV1N11HQB04WZRA4R88","e":"FAILURE#1","g":["01DFBQATWAEJHEZ47V1HNXFJGP"],"d":{"id":"01DFBMTSW3J58NG6VGPQ3WTFFZ"},"m":"BOOM!"}
	// {"l":"error","n":"01DFBM3VV1N11HQB04WZRA4R88","e":"FAILURE#2","g":["01DFBQB62ANNQ0YQ240K0XGMQW","tag-2"],"d":{"id":"01DFBMTSW3J58NG6VGPQ3WTFFZ"},"m":"BOOM!!"}
}
