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

package app

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"time"
)

// Timeouts specifies basic application configuration.
type Timeouts struct {
	// StartTimeout specifies how long to wait for the application to start.
	// - default = 15 seconds
	StartTimeout time.Duration `default:"15s" split_words:"true"`
	// StopTimeout specifies how long to wait for the application to stop.
	// - default = 15 seconds
	StopTimeout time.Duration `default:"15s" split_words:"true"`
}

func (c Timeouts) String() string {
	return fmt.Sprintf("Timeouts{StartTimeout=%s, StopTimeout=%s}", c.StartTimeout, c.StopTimeout)
}

// LoadTimeouts loads the app Timeouts from the system environment. The following env vars are read:
// - APPX12_START_TIMEOUT
// - APPX12_STOP_TIMEOUT
func LoadTimeouts() Timeouts {
	var config Timeouts
	if err := envconfig.Process(ENV_PREFIX, &config); err != nil {
		// an error should never happen because Timeouts has no required fields and defaults are specified
		// if an error does occur, then it's a bug in the underlying `envconfig` package
		panic(err)
	}
	return config
}
