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

package fxapp_test

import (
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/rs/zerolog"
	"testing"
)

func BenchmarkEventTypeID_LogEventFunc(b *testing.B) {
	const FooEvent fxapp.EventTypeID = "01DE79DKCAH8DBXZRX9P3THK9G"

	type LogFooEvent func()

	var logEvent LogFooEvent
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Provide(
			func(logger *zerolog.Logger) LogFooEvent {
				logEvent := FooEvent.NewLogEvent(logger, zerolog.InfoLevel)
				return func() {
					logEvent(nil, "foo")
				}
			},
		).
		Populate(&logEvent).
		Build()

	switch {
	case err != nil:
		b.Errorf("*** app build failure: %v", err)
	default:
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logEvent()
		}
	}
}
