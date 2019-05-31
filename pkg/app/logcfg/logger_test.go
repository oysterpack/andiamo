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
	"crypto/rand"
	"encoding/json"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strings"
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

	logger.Buf.Reset()
	logger.Warn().Msg("warning msg")
	logEventMsg = logger.Buf.String()
	t.Log(logEventMsg)
}

func TestUseAsStandardLoggerOutput(t *testing.T) {
	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given an app.Desc and app.InstanceID
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		t.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// And a new zerolog.Logger
	logger := logcfg.NewLogger(instanceID, desc)
	buf := new(strings.Builder)
	logger2 := logger.Output(buf)
	logger = &logger2
	// When the zerolog Logger is used as the std logger
	logcfg.UseAsStandardLoggerOutput(logger)
	// Then std logger will write to zerolog Logger
	log.Printf("this should be logging using zero log: %s", desc)
	logEventMsg := buf.String()
	t.Log(logEventMsg)
}

func TestConfigureZerolog(t *testing.T) {
	t.Run("with invalid log level", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.LOG_GLOBAL_LEVEL, "INVALID")
		if err := logcfg.ConfigureZerolog(); err == nil {
			t.Fatal("should have failed because INVALID log level was set in env")
		} else {
			t.Log(err)
		}
	})
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
