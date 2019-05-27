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

package app_test

import (
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/apptest"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	defaultTimeout := 15 * time.Second
	t.Run("when env not set, then defaults are used", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != defaultTimeout {
			t.Errorf("StartTimeout did not match the default: %v", config.StartTimeout)
		} else if config.StopTimeout != defaultTimeout {
			t.Errorf("StopTimeout did not match the default: %v", config.StartTimeout)
		}
	})

	t.Run("start timeout is set to 30 secs", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.START_TIMEOUT, "30s")
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != 30*time.Second {
			t.Errorf("StartTimeout did not match 30s: %v", config.StartTimeout)
		} else if config.StopTimeout != defaultTimeout {
			t.Errorf("StopTimeout did not match the default: %v", config.StartTimeout)
		}
	})

	t.Run("stop timeout is set to 30 secs", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.STOP_TIMEOUT, "30s")
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != defaultTimeout {
			t.Errorf("StartTimeout did not match the default: %v", config.StartTimeout)
		} else if config.StopTimeout != 30*time.Second {
			t.Errorf("StopTimeout did not match 30s: %v", config.StartTimeout)
		}
	})
}
