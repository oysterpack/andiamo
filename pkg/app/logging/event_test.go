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
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"testing"
)

func TestEvent_Equals(t *testing.T) {
	const (
		A logging.Tag = "a"
		B logging.Tag = "b"
		C logging.Tag = "c"
	)

	event1 := logging.MustNewEvent("foo", zerolog.InfoLevel)
	event2 := logging.MustNewEvent("foo", zerolog.WarnLevel)
	event3 := logging.MustNewEvent("foo", zerolog.InfoLevel, A)
	event4 := logging.MustNewEvent("foo", zerolog.InfoLevel, A, B)
	event5 := logging.MustNewEvent("foo", zerolog.InfoLevel, C)
	event6 := logging.MustNewEvent("foo", zerolog.InfoLevel, B, A, B, A)
	event7 := logging.MustNewEvent("foo", zerolog.InfoLevel, A, C)

	if event1.Equals(event2) {
		t.Errorf("*** events should be different because levels are different: %v : %v", event1, event2)
	}

	if event1.Equals(event3) {
		t.Errorf("*** events should be different because tags are different: %v : %v", event1, event3)
	}

	if event1.Equals(event5) {
		t.Errorf("*** events should be different because tags are different: %v : %v", event1, event5)
	}

	if event4.Equals(event7) {
		t.Errorf("*** events should be different because tags are different: %v : %v", event4, event7)
	}

	if event3.Equals(event4) {
		t.Errorf("*** events should be different because number of tags are different: %v : %v", event3, event4)
	}

	if !event4.Equals(event6) {
		t.Errorf("*** events should match bcause tags will be normalized and deduped: %v : %v", event4, event6)
	}
}

func TestEvent_Log(t *testing.T) {
	const (
		DataTag   logging.Tag = "data"
		DGraphTag logging.Tag = "dgraph"
	)

	var (
		// Tag will be normalized and deduped
		FooEvent = logging.MustNewEvent("foo", zerolog.WarnLevel, DGraphTag, DataTag, DataTag, "DATA")
		BarEvent = logging.MustNewEvent("bar", zerolog.ErrorLevel)
	)

	if len(FooEvent.Tags) != 2 {
		t.Fatalf("*** Event tags were not deduped: %v", FooEvent)
	}

	logger := apptest.NewTestLogger(PACKAGE)

	// When a foo event is logged
	FooEvent.Log(logger.Logger).Msg("")
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent apptest.LogEvent
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("*** Invalid JSON log event: %v", err)
	} else {
		t.Logf("JSON log event: %#v", logEvent)
		// Then the log level will match
		if logEvent.Level != FooEvent.Level.String() {
			t.Errorf("*** log level did not match")
		}
		// And the Event name will match
		if logEvent.Event != FooEvent.Name {
			t.Errorf("*** msg did not match")
		}

		// And tags are logged
		if len(logEvent.Tags) != len(FooEvent.Tags) {
			t.Errorf("*** the number of expected tags does not match: %v", len(logEvent.Tags))
		}
		for i := 0; i < len(FooEvent.Tags); i++ {
			if logEvent.Tags[i] != FooEvent.Tags[i] {
				t.Errorf("*** tag did not match: %v != %v", logEvent.Tags[i], FooEvent.Tags[i])
			}
		}
	}
	logger.Buf.Reset()

	BarEvent.Log(logger.Logger).Msg("")
	logEventMsg = logger.Buf.String()
	t.Log(logEventMsg)

	logEvent = apptest.LogEvent{}
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("*** Invalid JSON log event: %v", err)
	} else {
		t.Logf("JSON log event: %#v", logEvent)
		// Then bar event should have no tags logged

		// And tags are logged
		if len(logEvent.Tags) != 0 {
			t.Errorf("*** there should be no tags logged for the BarEvent: %v", len(logEvent.Tags))
		}
	}
}

func TestNewEvent(t *testing.T) {
	t.Run("name is wrapped in whitespace", func(t *testing.T) {
		event, e := logging.NewEvent("   foo  ", zerolog.InfoLevel)
		if e != nil {
			t.Fatal(e)
		}
		if event.Name != "foo" {
			t.Errorf("*** event name should have been trimmed: %q", event.Name)
		}
	})

	t.Run("name is blank", func(t *testing.T) {
		_, e := logging.NewEvent("    ", zerolog.InfoLevel)
		if e == nil {
			t.Error("*** should have failed to create event because name is blank")
		}
		t.Log(e)
	})

	t.Run("tag is blank", func(t *testing.T) {
		_, e := logging.NewEvent("foo", zerolog.InfoLevel, " ")
		if e == nil {
			t.Fatal("*** should have failed to create event because tag is blank")
		}
		t.Log(e)
	})
}

func TestMustNewEvent_PanicsOnError(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Fatal("*** should have failed to create event because name is blank")
		}
		t.Log(e)
	}()
	logging.MustNewEvent("    ", zerolog.InfoLevel)
}
