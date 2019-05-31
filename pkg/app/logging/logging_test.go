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
	"time"
)

const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/log_test"

func TestLogError(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)
	logger.Error().Err(errors.New("BOOM!!!")).Msg("")
	t.Logf("error log event: %s", logger.Buf.String())
}

func TestLogEvent_Log(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)

	// When a foo event is logged
	FooEvent.Log(logger.Logger).Msg("")
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent LogEvent
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
	}
	logger.Buf.Reset()

	BarEvent.Log(logger.Logger).Msg("")
	logEventMsg = logger.Buf.String()
	t.Log(logEventMsg)
}

type LogEvent struct {
	Level     string  `json:"l"`
	Timestamp int64   `json:"t"`
	Message   string  `json:"m"`
	App       AppDesc `json:"a"`
	Event     string  `json:"n"`
}

func (e *LogEvent) Time() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *LogEvent) MatchesDesc(desc *app.Desc) bool {
	return e.App.ID == desc.ID.String() &&
		e.App.Name == string(desc.Name) &&
		e.App.Version == desc.Version.String() &&
		e.App.ReleaseID == desc.ReleaseID.String()
}

type AppDesc struct {
	ID         string `json:"i"`
	ReleaseID  string `json:"r"`
	Name       string `json:"n"`
	Version    string `json:"v"`
	InstanceID string `json:"x"`
}

var FooEvent = logging.Event{
	Name:  "foo",
	Level: zerolog.WarnLevel,
}
var BarEvent = logging.Event{
	Name:  "bar",
	Level: zerolog.ErrorLevel,
}
