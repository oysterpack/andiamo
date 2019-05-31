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

package log

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/rs/zerolog"
	"log"
	"time"
)

// Config is used to load log config settings from env vars
type Config struct {
	// GlobalLevel specifies the global log level.
	// - default = info
	GlobalLevel     Level `default:"info" envconfig:"log_global_level"`
	DisableSampling bool  `split_words:"true" envconfig:"log_disable_sampling"`
}

// Apply will apply the zerolog config settings
func (l *Config) Apply() {
	zerolog.SetGlobalLevel(zerolog.Level(l.GlobalLevel))
	zerolog.DisableSampling(l.DisableSampling)
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{GlobalLevel=%s, DisableSampling=%v}", c.GlobalLevel, c.DisableSampling)
}

// Level is a type alias for zerolog.Level in order to be able to implement the `envconfig.Decoder` interface on it
type Level zerolog.Level

// Decode implements `envconfig.Decoder` interface
func (l *Level) Decode(value string) error {
	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return err
	}
	*l = Level(level)
	return nil
}

func (l Level) String() string {
	return zerolog.Level(l).String()
}

// ConfigureZerolog configures global zerolog settings.
// - configures the standard logger field names defined by `Field`
//   - TIMESTAMP
//   - LEVEL
//	 - MESSAGE
//   - ERROR
// - Unix time format is used for performance reasons - seconds granularity is sufficient for log events
// - duration field unit is set to millisecond
// - loads `Config` from the system env and applies it
func ConfigureZerolog() error {
	configureStandardLogFields := func() {
		zerolog.TimestampFieldName = string(TIMESTAMP)
		zerolog.LevelFieldName = string(LEVEL)
		zerolog.MessageFieldName = string(MESSAGE)
		zerolog.ErrorFieldName = string(ERROR)
	}

	configureTimeRelatedFields := func() {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.DurationFieldUnit = time.Millisecond
		zerolog.DurationFieldInteger = true
	}

	loadLogConfig := func() error {
		var config Config
		err := envconfig.Process(app.ENV_PREFIX, &config)
		if err != nil {
			return err
		}
		config.Apply()
		return nil
	}

	configureStandardLogFields()
	configureTimeRelatedFields()
	return loadLogConfig()
}

// UseAsStandardLoggerOutput uses the specified logger as the go std log output.
func UseAsStandardLoggerOutput(logger *zerolog.Logger) {
	log.SetFlags(0)
	log.SetOutput(logger)
}
