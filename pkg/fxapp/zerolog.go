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

package fxapp

import (
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Applies standard zerolog initialization.
//
// The following global settings are applied for performance reasons:
//   - the following standard logger field names are shortened
//     - Timestamp -> t
//     - Level -> l
//	   - Message -> m
//     - Error -> err
//   - Unix time format is used for performance reasons - seconds granularity is sufficient for log events
//
// An error stack marshaller is configured.
func init() {
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.ErrorFieldName = "e"

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldInteger = true

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

// EventLogger returns a new logger with the event type ID field 'n' set to the specified value.
//
// The event type ID should be unique. To ensure uniqueness, use ULIDs.
func EventLogger(logger *zerolog.Logger, id string) *zerolog.Logger {
	l := logger.With().Str("n", id).Logger()
	return &l
}

// ComponentLogger returns a new logger with the component field 'c' set to the specified value.
func ComponentLogger(logger *zerolog.Logger, id string) *zerolog.Logger {
	l := logger.With().Str("c", id).Logger()
	return &l
}

// LogLevel defines the supported app log levels
type LogLevel uint

// LogLevel enum
const (
	DebugLogLevel LogLevel = iota
	InfoLogLevel
	WarnLogLevel
	ErrorLogLevel
)

// ZerologLevel maps LogLevel to a zerolog.Level
func (level LogLevel) ZerologLevel() zerolog.Level {
	switch level {
	case InfoLogLevel:
		return zerolog.InfoLevel
	case WarnLogLevel:
		return zerolog.WarnLevel
	case ErrorLogLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.DebugLevel
	}
}

// SetEventID injects an event ID field named 'z'. The log event will assigned a ULID event ID.
//
// Use Case: Enables log event to be referenced.
func SetEventID(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Str("z", ulidgen.MustNew().String())
}

// LogEvent logs events using a standardized structure, e.g.,
//
//
func LogEvent(logger *zerolog.Logger, level zerolog.Level, eventTypeID EventTypeID, eventObject zerolog.LogObjectMarshaler, msg string, tags ...string) {
	event := logger.WithLevel(level)

	if eventObject != nil {
		data := zerolog.Dict()
		eventObject.MarshalZerologObject(data)
		event.Dict(string(eventTypeID), data)
	}

	if len(tags) > 0 {
		event.Strs("g", tags)
	}

	event.Msg(msg)
}

// NewLogEventFunc creates a new function used to log events using a standardized structure, e.g., app event
//
//	{
//	  "l": "error", -------------------------------------- event level
//	  "a": "01DE379HHM9Y3QYBDB4MSY7YYQ",
//	  "r": "01DE379HHNRJ4YS4NY4CMJX5YE",
//	  "x": "01DE379HHN2RRX9YQCG2DN9CHG",
//	  "n": "01DE2Z4E07E4T0GJJXCG8NN6A0", ----------------- event type ID
//	  "01DE2Z4E07E4T0GJJXCG8NN6A0": { -------------------- event type ID is used as event object dictionary key (optional)
//		"id": "01DE379HHNVHQE5G6NHN2BBKAT", -------------- event object data (optional)
//		"e": "failure to connect" ------------------------ event object data (optional)
//	  }, ------------------------------------------------- event object data (optional)
//	  "g": ["tag-a","tag-b"], ---------------------------- event tags (optional)
//	  "z": "01DE379HHNM87XT4PBHXYYBTYS",
//	  "t": 1561328928,
//	  "m": "healthcheck failed" -------------------------- event short description
//	}
//
// the event object data is logged as an event dictionary, using the event type ID as the key
func NewLogEventFunc(logger *zerolog.Logger, level zerolog.Level, eventTypeID EventTypeID) func(eventObject zerolog.LogObjectMarshaler, msg string, tags ...string) {
	eventLogger := EventLogger(logger, string(eventTypeID))
	return func(eventObject zerolog.LogObjectMarshaler, msg string, tags ...string) {
		LogEvent(eventLogger, level, eventTypeID, eventObject, msg, tags...)
	}

}
