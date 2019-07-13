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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/oysterpack/andiamo/pkg/eventlog"
	"github.com/rs/zerolog"
	"strings"
	"testing"
)

type FooLogger func(id FooID, tags ...string)

func NewFooLogger(logger *zerolog.Logger) FooLogger {
	logEvent := Foo.NewLogger(logger, zerolog.InfoLevel)
	return func(event FooID, tags ...string) {
		logEvent(event, "foo", tags...)
	}
}

type LogFooFailure func(id FooID, err error, tags ...string)

func NewFooErrorLogger(logger *zerolog.Logger) LogFooFailure {
	log := Foo.NewErrorLogger(logger)
	return func(id FooID, err error, tags ...string) {
		log(id, err, "foo", tags...)
	}
}

const Foo eventlog.Event = "01DE2Z4E07E4T0GJJXCG8NN6A0"

type FooID string

func (id FooID) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", string(id))
}

func TestEvent_Logger(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	logFooEvent := NewFooLogger(&logger)
	logFooEvent(FooID("01DF6P4G2WZ7HKDBES9YPXRHQ0"), "tag-a", "tag-b")

	type Data struct {
		ID string
	}

	type LogEvent struct {
		Level   string   `json:"l"`
		Name    string   `json:"n"`
		Message string   `json:"m"`
		Tags    []string `json:"g"`
		Data    Data     `json:"d"`
	}
	var logEvent LogEvent
	for _, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			break
		}
		err := json.Unmarshal([]byte(line), &logEvent)
		if err != nil {
			t.Errorf("*** failed to parse log event: %v", err)
			break
		}
		if logEvent.Name == string(Foo) {
			t.Log(line)
			break
		}
	}
	switch {
	case logEvent.Name == string(Foo):
		if logEvent.Level != "info" {
			t.Errorf("*** level did not match: %v", logEvent.Level)
		}
		if logEvent.Data.ID != "01DF6P4G2WZ7HKDBES9YPXRHQ0" {
			t.Error("*** event data was not logged")
		}
		if len(logEvent.Tags) != 2 {
			t.Errorf("*** tags were not logged: %v", logEvent.Tags)
		} else {
			if logEvent.Tags[0] != "tag-a" && logEvent.Tags[1] != "tag-b" {
				t.Errorf("*** tags don't match: %v", logEvent.Tags)
			}
		}

	default:
		t.Error("*** event was not logged")
	}

}

func TestEvent_NewErrorLogger(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	logFooFailure := NewFooErrorLogger(&logger)
	logFooFailure(FooID("01DF6P4G2WZ7HKDBES9YPXRHQ0"), errors.New("failure to connect"), "tag-a", "tag-b")

	type Data struct {
		ID string
	}

	type LogEvent struct {
		Level   string   `json:"l"`
		Name    string   `json:"n"`
		Message string   `json:"m"`
		Tags    []string `json:"g"`
		Err     string   `json:"e"`
		Data    Data     `json:"d"`
	}
	var logEvent LogEvent
	for _, line := range strings.Split(buf.String(), "\n") {
		if line == "" {
			break
		}
		err := json.Unmarshal([]byte(line), &logEvent)
		if err != nil {
			t.Errorf("*** failed to parse log event: %v", err)
			break
		}
		if logEvent.Name == string(Foo) {
			t.Log(line)
			break
		}
	}
	switch {
	case logEvent.Name == string(Foo):
		if logEvent.Level != "error" {
			t.Errorf("*** level did not match: %v", logEvent.Level)
		}
		if logEvent.Data.ID != "01DF6P4G2WZ7HKDBES9YPXRHQ0" {
			t.Error("*** event data was not logged")
		}
		if len(logEvent.Tags) != 2 {
			t.Errorf("*** tags were not logged: %v", logEvent.Tags)
		} else {
			if logEvent.Tags[0] != "tag-a" && logEvent.Tags[1] != "tag-b" {
				t.Errorf("*** tags don't match: %v", logEvent.Tags)
			}
		}
		if logEvent.Err != "failure to connect" {
			t.Errorf("*** error did not match: %v", logEvent.Err)
		}
		if logEvent.Message != "foo" {
			t.Errorf("*** message did not match: %v", logEvent.Message)
		}

	default:
		t.Error("*** event was not logged")
	}

}
