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
	"log"
	"os"
	"time"
)

type LogField string

const (
	TIMESTAMP = LogField("t")
	LEVEL     = LogField("l")
	MESSAGE   = LogField("m")
	ERROR     = LogField("e")
	CALLER    = LogField("c")
	STACK     = LogField("s")

	// all log events should specify the event name
	EVENT = LogField("n")

	// app related fields
	APP             = LogField("a")
	APP_ID          = LogField("i")
	APP_RELEASE_ID  = LogField("r")
	APP_NAME        = LogField("n")
	APP_VERSION     = LogField("v")
	APP_INSTANCE_ID = LogField("x")
)

// App lifecycle events
// - NOTE: they are logged with no level to ensure they are always logged, i.e., regardless of the global log level
var (
	Started = LogEvent{
		Name:  "app_start",
		Level: zerolog.NoLevel,
	}

	Stopped = LogEvent{
		Name:  "app_stop",
		Level: zerolog.NoLevel,
	}
)

// NewLogger constructs a new timestamped Logger with standardized fields.
//
// Example log message:
//
// {"l":"info","a":{"i":"01DBXQXE6WS76C2EXYBC06MSWB","r":"01DBXQXE6WSGB4EZGW8TGEH0PV","n":"foobar","v":"0.0.1","x":"01DBXQXE6WR0H602E09TA96X4D"},"t":1558997547228,"m":"info msg"}
//
// 	l   = level
//	t   = timestamp in Unix time
//  m   = message
//	a   = app
//	a.i = app ID
//	a.r = release ID
//	a.n = app name
//	a.v = app version
//	a.x = app instance ID
func NewLogger(instanceID InstanceID, desc Desc) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Dict(string(APP), zerolog.Dict().
			Str(string(APP_ID), desc.ID.String()).
			Str(string(APP_RELEASE_ID), desc.ReleaseID.String()).
			Str(string(APP_NAME), string(desc.Name)).
			Str(string(APP_VERSION), desc.Version.String()).
			Str(string(APP_INSTANCE_ID), instanceID.String())).
		Logger()

	return &logger
}

// ConfigureZerolog applies the following configurations on zerolog:
// - configures the standard logger field names defined by `LogField`
// - Unix time format is used for performance reasons - seconds granularity is sufficient for log events
// - applies `LogConfig` settings
func ConfigureZerolog() error {
	configureStandardLogFields()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true

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
	zerolog.CallerFieldName = string(CALLER)
	zerolog.ErrorStackFieldName = string(STACK)
}

// UseAsStandardLoggerOutput uses the specified logger as the go std log output.
func UseAsStandardLoggerOutput(logger *zerolog.Logger) {
	log.SetFlags(0)
	log.SetOutput(logger)
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

func (l LogLevel) String() string {
	return zerolog.Level(l).String()
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
	return fmt.Sprintf("LogConfig{GlobalLevel=%s, DisableSampling=%v}", c.GlobalLevel, c.DisableSampling)
}

// LogEvent is used to define application log events.
// This enables application log events to be defined as code and documented.
type LogEvent struct {
	Name string
	zerolog.Level
}

// Log starts a new log message.
// - LogEvent.Level is used as the message log level
// - LogEvent.Name is used for the `EVENT` log field value
//
// NOTE: You must call Msg on the returned event in order to send the event.
func (l *LogEvent) Log(logger *zerolog.Logger) *zerolog.Event {
	return logger.WithLevel(l.Level).Str(string(EVENT), l.Name)
}
