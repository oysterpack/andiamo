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
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"time"
)

// EventTypeID is used as an event type ID.
// It must be globally unique - ULIDs are recommended.
type EventTypeID string

func (e EventTypeID) String() string {
	return string(e)
}

// LogEvent is a function used to log events.
type LogEvent func(eventData zerolog.LogObjectMarshaler, msg string, tags ...string)

// NewLogEvent creates a new function used to log events using a standardized structure, e.g., app event
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
func (e EventTypeID) NewLogEvent(logger *zerolog.Logger, level zerolog.Level) LogEvent {
	eventLogger := EventLogger(logger, e.String())
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

// app lifecycle event IDs
const (
	InitializedEventID EventTypeID = "01DE4STZ0S24RG7R08PAY1RQX3"
	InitFailedEventID  EventTypeID = "01DE4SWMZXD1ZB40QRT7RGQVPN"

	StartingEventID    EventTypeID = "01DE4SXMG8W3KSPZ9FNZ8Z17F8"
	StartFailedEventID EventTypeID = "01DE4SY6RYCD0356KYJV7G7THW"

	StartedEventID EventTypeID = "01DE4X10QCV1M8TKRNXDK6AK7C"

	StoppingEventID   EventTypeID = "01DE4SZ1KY60JQTF7XP4DQ8WGC"
	StopFailedEventID EventTypeID = "01DE4T0W35RPD6QMDS42WQXR48"

	StoppedEventID EventTypeID = "01DE4T1V9N50BB67V424S6MG5C"
)

// AppInitialized indicates the application has successfully initialized
type AppInitialized struct {
	App
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppInitialized) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("start_timeout", event.StartTimeout())
	e.Dur("stop_timeout", event.StopTimeout())

	typeNames := func(types []reflect.Type) []string {
		var names []string
		for _, t := range types {
			names = append(names, t.String())
		}
		return names
	}

	e.Strs("provides", typeNames(event.App.ConstructorTypes()))
	e.Strs("invokes", typeNames(event.App.FuncTypes()))
}

// AppStarted indicates the app has successfully been started.
type AppStarted struct {
	time.Duration
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppStarted) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

// AppStopping indicates the app has been triggered to shutdown.
type AppStopping struct {
	os.Signal
}

// AppStopped indicates that the app has been stopped.
// This will always be logged, regardless whether the app failed to shutdown cleanly or not, i.e., if an error occurs
// while shutting down the app, then both the AppStopFailed and AppStopped events will be logged.
type AppStopped struct {
	time.Duration
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppStopped) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

// AppFailed indicates the application failed to be built and initialized
type AppFailed struct {
	Err error
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppFailed) MarshalZerologObject(e *zerolog.Event) {
	e.Err(event.Err)
}
