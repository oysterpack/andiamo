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

package logcfg_test

import (
	"encoding/json"
	"errors"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"testing"
	"time"
)

func TestConfigureZerologAndNewLogger(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)
	// Then debug messages should not be logged
	if e := logger.Debug(); e.Enabled() {
		logger.Debug().Msg("debug msg")
		t.Errorf("Default global log level should be INFO")
	}
	// And INFO and above log level messages should be logged
	if e := logger.Info(); !e.Enabled() {
		t.Errorf("Default global log level should be INFO")
	}
	logger.Info().Msg("info msg")
	logEventTime := time.Now()
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent LogEvent
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("Invalid JSON log event: %v", err)
	} else {
		// And JSON log event data matches
		t.Logf("JSON log event: %#v", logEvent)
		t.Logf("now: %v | event timestamp: %v", time.Now(), logEvent.Time())
		if !logEvent.MatchesDesc(&logger.Desc) {
			t.Errorf("app.Desc did not match")
		}
		if logEvent.App.InstanceID != logger.InstanceID.String() {
			t.Errorf("app.InstanceID did not match")
		}
		if logEvent.Level != zerolog.InfoLevel.String() {
			t.Errorf("log level did not match")
		}
		if logEventTime.Unix()-logEvent.Timestamp > 1 {
			t.Errorf("log event Unix time did not match")
		}
		if logEvent.Message != "info msg" {
			t.Errorf("msg did not match")
		}
	}

	// Given and error is logged
	logger.Buf.Reset()
	logger.Error().Err(errors.New("error occurred")).Msg("warning msg")
	logEventMsg = logger.Buf.String()
	t.Log(logEventMsg)

	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Errorf("Invalid JSON log event: %v", err)
	} else {
		// Then the error is logged
		if logEvent.Error != "error occurred" {
			t.Errorf("error did not match")
		}
		// And the log event level is Error
		if logEvent.Level != zerolog.ErrorLevel.String() {
			t.Errorf("log level did not match")
		}
	}
}

type LogEvent struct {
	Level     string  `json:"l"`
	Timestamp int64   `json:"t"`
	Message   string  `json:"m"`
	App       AppDesc `json:"a"`
	Event     string  `json:"n"`
	Error     string  `json:"e"`
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
