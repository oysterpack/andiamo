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
	"bytes"
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"testing"
)

func TestZerologFieldNames(t *testing.T) {
	type StackFrame struct {
		Func   string
		Line   string
		Source string
	}

	type LogEvent struct {
		Timestamp uint   `json:"t"`
		Message   string `json:"m"`
		Level     string `json:"l"`
		Name      string `json:"n"`
		Component string `json:"c"`
		Error     string `json:"e"`
		Stack     []StackFrame
	}

	buf := new(bytes.Buffer)
	logger := zerolog.New(buf).With().Timestamp().Logger()
	compLogger := fxapp.ComponentLogger(&logger, "foo")
	eventLogger := fxapp.EventLogger(compLogger, "bar")

	eventLogger.Error().
		Stack().
		Err(errors.New("BOOM!!!")).
		Msg("foobar")

	t.Log(buf.String())

	var logEvent LogEvent
	err := json.Unmarshal(buf.Bytes(), &logEvent)
	switch {
	case err != nil:
		t.Errorf("*** failed to parse log event as JSON: %v : %v", err, buf.String())
	default:
		t.Logf("%#v", logEvent)
		if logEvent.Timestamp == 0 {
			t.Errorf("*** timestamp field did not match: %#v", logEvent)
		}
		if logEvent.Message != "foobar" {
			t.Errorf("*** message field did not match: %#v", logEvent)
		}
		if logEvent.Level != "error" {
			t.Errorf("*** level field did not match: %#v", logEvent)
		}
		if logEvent.Component != "foo" {
			t.Errorf("*** component field did not match: %#v", logEvent)
		}
		if logEvent.Name != "bar" {
			t.Errorf("*** name field did not match: %#v", logEvent)
		}
		if logEvent.Error != "BOOM!!!" {
			t.Errorf("*** error field did not match: %#v", logEvent)
		}
		if len(logEvent.Stack) == 0 {
			t.Error("*** error stack was not pasrsed")
		}
	}
}

func TestComponentLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	componentLogger := fxapp.ComponentLogger(&logger, "foo")
	componentLogger.Log().Msg("")

	type LogEvent struct {
		Component string `json:"c"`
	}

	var logEvent LogEvent
	err := json.Unmarshal(buf.Bytes(), &logEvent)
	switch {
	case err != nil:
		t.Errorf("*** failed to parse log event as JSON: %v : %v", err, buf.String())
	default:
		t.Log(logEvent)
		if logEvent.Component != "foo" {
			t.Errorf("*** component field did not match: %#v", logEvent)
		}
	}
}

func TestEventLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	eventLogger := fxapp.EventLogger(&logger, "foo")
	eventLogger.Log().Msg("")

	type LogEvent struct {
		Name string `json:"n"`
	}

	var logEvent LogEvent
	err := json.Unmarshal(buf.Bytes(), &logEvent)
	switch {
	case err != nil:
		t.Errorf("*** failed to parse log event as JSON: %v : %v", err, buf.String())
	default:
		t.Log(logEvent)
		if logEvent.Name != "foo" {
			t.Errorf("*** event name field did not match: %#v", logEvent)
		}
	}
}
