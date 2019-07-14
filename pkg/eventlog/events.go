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

package eventlog

import (
	"github.com/rs/zerolog"
)

// Logger is used to log events using a consistent and standardized structure.
// Use the `NewLogger()` constructor function to create new Logger functions.
//
// - log event level is encapsulated by the Logger function.
// - event data is optional, i.e., nil may be supplied.
// - events can be tagged
//   - tagging use cases: tracing, grouping related events, tagging source code locations, etc
type Logger func(data zerolog.LogObjectMarshaler, msg string, tags ...string)

// NewLogger creates a new function used to log events using a standardized structure that supports use cases for automated
// monitoring, alerting, querying, and analytics. Having a standardized structure makes it easier to build standardized tools.
// The goal is to get more value out of log events by enabling the log events to be processed programmatically.
//
// The event object data is logged as an event dictionary, using the event name as the key. The event name must be globally
// unique - it is recommended to use ULID as the event name. The event data structure should be designed to be as stable
// as possible. Treat the event data structure as an interface because monitors, tools, queries, and other tools may depend on it.
// Not all events may have event data.
//
// Example application event
//	{
//	  "l": "error", -------------------------------------- event level
//	  "n": "01DE2Z4E07E4T0GJJXCG8NN6A0", ----------------- event name
//	  "d": { --------------------------------------------- event data (optional)
//		"id": "01DE379HHNVHQE5G6NHN2BBKAT", -------------- event data (optional)
//	  }, ------------------------------------------------- event data (optional)
//	  "g": ["tag-a","tag-b"], ---------------------------- event tags (optional)
//	  "m": "health check failed" ------------------------- event short description
//	}
func NewLogger(event string, logger *zerolog.Logger, level zerolog.Level) Logger {
	eventLogger := ForEvent(logger, event)
	return func(eventData zerolog.LogObjectMarshaler, msg string, tags ...string) {
		log(eventLogger.WithLevel(level), eventData, msg, tags...)
	}
}

func log(zerologEvent *zerolog.Event, eventData zerolog.LogObjectMarshaler, msg string, tags ...string) {
	if len(tags) > 0 {
		zerologEvent.Strs("g", tags)
	}

	if eventData != nil {
		data := zerolog.Dict()
		eventData.MarshalZerologObject(data)
		zerologEvent.Dict("d", data)
	}

	zerologEvent.Msg(msg)
}

// Error wraps an underlying error to implement `zerolog.LogObjectMarshaler` interface
type Error struct {
	error
}

// NewError wraps the specified error
func NewError(err error) Error {
	return Error{err}
}

// MarshalZerologObject implements `zerolog.LogObjectMarshaler` interface
func (err Error) MarshalZerologObject(e *zerolog.Event) {
	e.Err(err.error)
}
