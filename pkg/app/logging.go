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
	"github.com/rs/zerolog"
	"os"
)

type LogField string

const (
	TIMESTAMP = LogField("t")
	LEVEL     = LogField("l")
	MESSAGE   = LogField("m")
	ERROR     = LogField("e")

	APP             = LogField("app")
	APP_ID          = LogField("i")
	APP_RELEASE_ID  = LogField("r")
	APP_NAME        = LogField("n")
	APP_VERSION     = LogField("v")
	APP_INSTANCE_ID = LogField("x")
)

// NewLogger constructs a new timestamped Logger with standardized fields.
//
// TODO: show example log statement
func NewLogger(instanceID InstanceID, desc Desc) zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Dict(string(APP), zerolog.Dict().
			Str(string(APP_ID), desc.ID.String()).
			Str(string(APP_RELEASE_ID), desc.ReleaseID.String()).
			Str(string(APP_NAME), string(desc.Name)).
			Str(string(APP_VERSION), desc.Version.String()).
			Str(string(APP_INSTANCE_ID), instanceID.String())).
		Logger()

	return logger
}

// ConfigureZerolog applies the following configurations on zerolog:
// - configures the standard logger field names defined by `LogField`
// - Unix millisecond timestamp format is used for performance reasons
// - applies `LogConfig` settings
func ConfigureZerolog() error {
	configureStandardLogFields()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	var config LogConfig
	err := envconfig.Process(ENV_PREFIX, &config)
	if err != nil {
		return err
	}
	config.Apply()

	return nil
}

func configureStandardLogFields() {
	zerolog.TimestampFieldName = string(TIMESTAMP)
	zerolog.LevelFieldName = string(LEVEL)
	zerolog.MessageFieldName = string(MESSAGE)
	zerolog.ErrorFieldName = string(ERROR)
}

// LogLevel is a type alias for zerolog.Level in order to be able to implement the `envconfig.Decoder` interface on it
type LogLevel zerolog.Level

// Decode implements `envconfig.Decoder` interface
func (l *LogLevel) Decode(value string) error {
	level, err := zerolog.ParseLevel(value)
	if err != nil {
		return err
	}
	*l = LogLevel(level)
	return nil
}

// LogConfig
type LogConfig struct {
	// GlobalLevel specifies the global log level.
	// - default = info
	GlobalLevel     LogLevel `default:"info" envconfig:"log_global_level"`
	DisableSampling bool     `split_words:"true" envconfig:"log_disable_sampling"`
}

func (l *LogConfig) Apply() {
	zerolog.SetGlobalLevel(zerolog.Level(l.GlobalLevel))
	zerolog.DisableSampling(l.DisableSampling)
}

func (c *LogConfig) String() string {
	return fmt.Sprintf("LogConfig{GlobalLevel=%s, DisableSampling=%v}", zerolog.Level(c.GlobalLevel), c.DisableSampling)
}
