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

package logging

import "github.com/rs/zerolog"

// Event is used to define application log events.
// This enables application log events to be defined as code and documented.
type Event struct {
	// Name is required
	Name string
	// Level is required
	zerolog.Level
	// Tags are optional - but recommended to help organize and categorize events
	Tags []string
}

// NewEvent constructs a new Event
func NewEvent(name string, level zerolog.Level, tags ...Tag) Event {
	var tagSlice []string
	if len(tags) > 0 {
		tagSlice = make([]string, len(tags))
		for i, tag := range tags {
			tagSlice[i] = tag.String()
		}
	}

	return Event{
		Name:  name,
		Level: level,
		Tags:  tagSlice,
	}
}

// Tag is used to define tags as constants in a type safe manner
type Tag string

func (t Tag) String() string {
	return string(t)
}

// Log starts a new log message.
// - Event.Level is used as the message log level
// - Event.Name is used for the `EventName` log field value
// - Event.Tags are logged, if not empty
//
// NOTE: You must call Msg on the returned event in order to send the event.
func (l *Event) Log(logger *zerolog.Logger) *zerolog.Event {
	event := logger.WithLevel(l.Level).Str(string(EventName), l.Name)
	if len(l.Tags) > 0 {
		event.Strs(string(Tags), l.Tags)
	}
	return event
}
