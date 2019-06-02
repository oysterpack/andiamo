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
	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogConfig(t *testing.T) {
	apptest.ClearAppEnvSettings()

	t.Run("with default settings", func(t *testing.T) {
		// Given app.Config is loaded from the env
		var config logcfg.Config
		err := envconfig.Process(app.EnvPrefix, &config)
		if err != nil {
			t.Error(err)
		}
		// Then it is loaded with default values
		t.Logf("Config: %s", &config)
		const DEFAULT_LOG_LEVEL = logcfg.Level(zerolog.InfoLevel)
		if config.GlobalLevel != DEFAULT_LOG_LEVEL {
			t.Errorf("Config.GlobalLevel default value did not match: %v", config.GlobalLevel)
		}
		if config.DisableSampling {
			t.Error("Config.DisableSampling default value should be false but was found to be true")
		}
	})

	t.Run("with LOG_GLOBAL_LEVEL warn", func(t *testing.T) {
		// Given app.Config is loaded from the env
		apptest.Setenv(apptest.LogGlobalLevel, "warn")
		var config logcfg.Config
		err := envconfig.Process(app.EnvPrefix, &config)
		if err != nil {
			t.Error(err)
		}
		// Then the global log level is matches the env var setting
		t.Logf("Config: %s", &config)
		const EXPECTED_LOG_LEVEL = logcfg.Level(zerolog.WarnLevel)
		if config.GlobalLevel != EXPECTED_LOG_LEVEL {
			t.Errorf("Config.GlobalLevel did not match: %v", config.GlobalLevel)
		}
	})

	t.Run("with LOG_DISABLE_SAMPLING true", func(t *testing.T) {
		// Given app.Config is loaded from the env
		apptest.Setenv(apptest.LogDisableSampling, "true")
		var config logcfg.Config
		err := envconfig.Process(app.EnvPrefix, &config)
		if err != nil {
			t.Error(err)
		}
		// Then the disable sampling setting matches the env var setting
		t.Logf("Config: %s", &config)
		if !config.DisableSampling {
			t.Errorf("Config.DisableSampling did not match: %v", config.DisableSampling)
		}
	})
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
	t.Run("using default config", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		if err := logcfg.ConfigureZerolog(); err != nil {
			t.Fatalf("logcfg.ConfigureZerolog() failed: %v", err)
		}

		if zerolog.TimestampFieldName != string(logging.Timestamp) {
			t.Errorf("zerolog.TimestampFieldName should be %q, but is %q", zerolog.TimestampFieldName, logging.Timestamp)
		}
		if zerolog.LevelFieldName != string(logging.Level) {
			t.Errorf("zerolog.LevelFieldName should be %q, but is %q", zerolog.LevelFieldName, logging.Level)
		}
		if zerolog.MessageFieldName != string(logging.Message) {
			t.Errorf("zerolog.MessageFieldName should be %q, but is %q", zerolog.MessageFieldName, logging.Message)
		}
		if zerolog.ErrorFieldName != string(logging.Error) {
			t.Errorf("zerolog.ErrorFieldName should be %q, but is %q", zerolog.ErrorFieldName, logging.Error)
		}

		if zerolog.TimeFieldFormat != zerolog.TimeFormatUnix {
			t.Errorf("zerolog.ErrorFieldName should be zerolog.TimeFormatUnix, but is %q", zerolog.TimeFieldFormat)
		}
		if zerolog.DurationFieldUnit != time.Millisecond {
			t.Errorf("zerolog.DurationFieldUnit should be time.Millisecond, but is %s", zerolog.DurationFieldUnit)
		}
		if !zerolog.DurationFieldInteger {
			t.Error("zerolog.DurationFieldInteger should be true")
		}

		if zerolog.GlobalLevel() != zerolog.InfoLevel {
			t.Errorf("zerolog.GlobalLevel() should be Info, but is : %v", zerolog.GlobalLevel())
		}
	})

	t.Run("with invalid log level", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.LogGlobalLevel, "INVALID")
		if err := logcfg.ConfigureZerolog(); err == nil {
			t.Fatal("should have failed because INVALID log level was set in env")
		} else {
			t.Log(err)
		}
	})
}
