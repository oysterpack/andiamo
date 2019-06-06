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

package logging_test

import (
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
	"testing"
)

func TestEventRegistry(t *testing.T) {
	const (
		TagA logging.Tag = "a"
		TagB logging.Tag = "b"
	)

	var (
		FooEvent = logging.NewEvent("foo", zerolog.WarnLevel, TagA, TagB)
		BarEvent = logging.NewEvent("bar", zerolog.ErrorLevel)

		expectedEvents = []logging.Event{FooEvent, BarEvent}
	)

	registry := logging.NewEventRegistry()

	// Registering the same events again is idempotent
	for i := 0; i < 2; i++ {
		registry.Register(FooEvent, BarEvent)
		events := registry.Events()
		t.Log(events)
		if len(events) != len(expectedEvents) {
			t.Fatalf("expected to have 2 events registered, but there were %d", len(expectedEvents))
		}

		for i := 0; i < len(events); i++ {
			matched := false
			for ii := 0; ii < len(events); ii++ {
				if events[i].Equals(expectedEvents[i]) {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("Event was not matched: %v", events[i])
			}
		}

		for _, event := range expectedEvents {
			if !registry.Registered(event) {
				t.Errorf("event was not found in the registry: %v", event)
			}
		}

	}

}
