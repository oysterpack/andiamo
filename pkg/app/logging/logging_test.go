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
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"testing"
)

const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/log_test"

func TestLogError(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)
	// in order for the stack to be logged, Stack() needs to be called before Err()
	logger.Error().Stack().Err(errors.New("BOOM!!!")).Msg("")
	t.Logf("error log event: %s", logger.Buf.String())
}

func TestEvent_Log(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)

	// When a foo event is logged
	FooEvent.Log(logger.Logger).Msg("")
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent apptest.LogEvent
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("Invalid JSON log event: %v", err)
	} else {
		t.Logf("JSON log event: %#v", logEvent)
		// Then the log level will match
		if logEvent.Level != FooEvent.Level.String() {
			t.Errorf("log level did not match")
		}
		// And the Event name will match
		if logEvent.Event != FooEvent.Name {
			t.Errorf("msg did not match")
		}

		// And tags are logged
		if len(logEvent.Tags) != len(FooEvent.Tags) {
			t.Errorf("the number of expected tags does not match: %v", len(logEvent.Tags))
		}
		for i := 0; i < len(FooEvent.Tags); i++ {
			if logEvent.Tags[i] != FooEvent.Tags[i] {
				t.Errorf("tag did not match: %v != %v", logEvent.Tags[i], FooEvent.Tags[i])
			}
		}
	}
	logger.Buf.Reset()

	BarEvent.Log(logger.Logger).Msg("")
	logEventMsg = logger.Buf.String()
	t.Log(logEventMsg)

	logEvent = apptest.LogEvent{}
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("Invalid JSON log event: %v", err)
	} else {
		t.Logf("JSON log event: %#v", logEvent)
		// Then bar event should have no tags logged

		// And tags are logged
		if len(logEvent.Tags) != 0 {
			t.Errorf("there should be no tags logged for the BarEvent: %v", len(logEvent.Tags))
		}
	}
}

const (
	DataTag   logging.Tag = "data"
	DGraphTag logging.Tag = "dgraph"
)

var FooEvent = logging.Event{
	Name:  "foo",
	Level: zerolog.WarnLevel,
	Tags:  []string{DataTag.String(), DGraphTag.String()},
}
var BarEvent = logging.Event{
	Name:  "bar",
	Level: zerolog.ErrorLevel,
}
