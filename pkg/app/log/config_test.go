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

package log_test

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/apptest"
	"github.com/oysterpack/partire-k8s/pkg/app/log"
	"github.com/rs/zerolog"
	"testing"
)

func TestLogConfig(t *testing.T) {
	apptest.ClearAppEnvSettings()

	t.Run("with default settings", func(t *testing.T) {
		// Given app.Config is loaded from the env
		var config log.Config
		err := envconfig.Process(app.ENV_PREFIX, &config)
		if err != nil {
			t.Error(err)
		}
		// Then it is loaded with default values
		t.Logf("Config: %s", &config)
		const DEFAULT_LOG_LEVEL = log.Level(zerolog.InfoLevel)
		if config.GlobalLevel != DEFAULT_LOG_LEVEL {
			t.Errorf("Config.GlobalLevel default value did not match: %v", config.GlobalLevel)
		}
		if config.DisableSampling {
			t.Errorf("Config.DisableSampling default value did not match: %v", config.DisableSampling)
		}
	})

	t.Run("with LOG_GLOBAL_LEVEL warn", func(t *testing.T) {
		// Given app.Config is loaded from the env
		apptest.Setenv(apptest.LOG_GLOBAL_LEVEL, "warn")
		var config log.Config
		err := envconfig.Process(app.ENV_PREFIX, &config)
		if err != nil {
			t.Error(err)
		}
		// Then the global log level is matches the env var setting
		t.Logf("Config: %s", &config)
		const EXPECTED_LOG_LEVEL = log.Level(zerolog.WarnLevel)
		if config.GlobalLevel != EXPECTED_LOG_LEVEL {
			t.Errorf("Config.GlobalLevel did not match: %v", config.GlobalLevel)
		}
	})

	t.Run("with LOG_DISABLE_SAMPLING true", func(t *testing.T) {
		// Given app.Config is loaded from the env
		apptest.Setenv(apptest.LOG_DISABLE_SAMPLING, "true")
		var config log.Config
		err := envconfig.Process(app.ENV_PREFIX, &config)
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
