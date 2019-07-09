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

// Event is used as an event type ID.
// It must be globally unique - ULIDs are recommended.
type Event string

func (e Event) String() string {
	return string(e)
}

// Logger is a function used to log events.
type Logger func(eventData zerolog.LogObjectMarshaler, msg string, tags ...string)

// ErrorLogger is a function used to log error events
type ErrorLogger func(eventData zerolog.LogObjectMarshaler, err error, tags ...string)

// NewLogger creates a new function used to log events using a standardized structure that supports use cases for automated
// monitoring, alerting, querying, and analytics. Having a standardized structure makes it easier to build standardized tools.
// The goal is to get more value out of log events by enabling the log events to processed programmatically.
//
// The event object data is logged as an event dictionary, using the event type ID as the key. The event data structure
// should be designed to be as stable as possible. Treat the event data structure as an interface because monitors, tools,
// queries, and other tools may depend on it. Not all events may have event data.
//
// Example application event
//	{
//	  "l": "error", -------------------------------------- event level
//	  "a": "01DE379HHM9Y3QYBDB4MSY7YYQ", ================= app ID
//	  "r": "01DE379HHNRJ4YS4NY4CMJX5YE", ================= app release ID
//	  "x": "01DE379HHN2RRX9YQCG2DN9CHG", ================= app instance ID
//	  "n": "01DE2Z4E07E4T0GJJXCG8NN6A0", ----------------- event type ID
//	  "01DE2Z4E07E4T0GJJXCG8NN6A0": { -------------------- event type ID is used as the event object dictionary key (optional)
//		"id": "01DE379HHNVHQE5G6NHN2BBKAT", -------------- event object data (optional)
//		"e": "failure to connect" ------------------------ event object data (optional)
//	  }, ------------------------------------------------- event object data (optional)
//	  "g": ["tag-a","tag-b"], ---------------------------- event tags (optional)
//	  "z": "01DE379HHNM87XT4PBHXYYBTYS", ================= event instance ID
//	  "t": 1561328928, =================================== event timestamp in Unix time
//	  "m": "health check failed" ------------------------- event short description
//	}
//
//  where
//      ==== means the field was populated by the application logger
//		---- means the field was populated by the event logger
func (e Event) NewLogger(logger *zerolog.Logger, level zerolog.Level) Logger {
	eventLogger := ForEvent(logger, e.String())
	return func(eventObject zerolog.LogObjectMarshaler, msg string, tags ...string) {
		event := eventLogger.WithLevel(level)

		if eventObject != nil {
			data := zerolog.Dict()
			eventObject.MarshalZerologObject(data)
			event.Dict(e.String(), data)
		}

		if len(tags) > 0 {
			event.Strs("g", tags)
		}

		event.Msg(msg)
	}
}

// NewErrorLogger creates a new function used to log errors with contextual data. It uses the same structure as `Logger`
// except that the level is automatically set to `error` and the error is set on the log event.
//
// Example application event
//	{
//	  "l": "error", -------------------------------------- event level
//	  "a": "01DE379HHM9Y3QYBDB4MSY7YYQ", ================= app ID
//	  "r": "01DE379HHNRJ4YS4NY4CMJX5YE", ================= app release ID
//	  "x": "01DE379HHN2RRX9YQCG2DN9CHG", ================= app instance ID
//	  "n": "01DE2Z4E07E4T0GJJXCG8NN6A0", ----------------- event type ID
//    "e": "failure to connect" -------------------------- event error
//	  "01DE2Z4E07E4T0GJJXCG8NN6A0": { -------------------- event type ID is used as the event object dictionary key (optional)
//		"id": "01DE379HHNVHQE5G6NHN2BBKAT", -------------- event object data (optional)
//	  }, ------------------------------------------------- event object data (optional)
//	  "g": ["tag-a","tag-b"], ---------------------------- event tags (optional)
//	  "z": "01DE379HHNM87XT4PBHXYYBTYS", ================= event instance ID
//	  "t": 1561328928, =================================== event timestamp in Unix time
//	  "m": "health check failed" ------------------------- event short description
//	}
//
//  where
//      ==== means the field was populated by the application logger
//		---- means the field was populated by the event logger
func (e Event) NewErrorLogger(logger *zerolog.Logger) ErrorLogger {
	eventLogger := ForEvent(logger, e.String())
	return func(eventObject zerolog.LogObjectMarshaler, err error, tags ...string) {
		event := eventLogger.Error().Err(err)

		if eventObject != nil {
			data := zerolog.Dict()
			eventObject.MarshalZerologObject(data)
			event.Dict(e.String(), data)
		}

		if len(tags) > 0 {
			event.Strs("g", tags)
		}

		event.Msg("")
	}
}
