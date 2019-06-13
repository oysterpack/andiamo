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

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"log"
	"sort"
	"strings"
)

// Event is used to define application log events.
// This enables application log events to be defined as code and documented.
type Event struct {
	// Name is required
	Name string
	// Level is required
	zerolog.Level
	// Tags are used to organize and group related events.
	Tags []string
}

func (e Event) String() string {
	return fmt.Sprintf("Event{name=%s, level=%s, tags=%v}", e.Name, e.Level, e.Tags)
}

// Equals returns true if the 2 events are equal.
func (e Event) Equals(e2 *Event) bool {
	if e.Name != e2.Name {
		return false
	}

	if e.Level != e2.Level {
		return false
	}

	if len(e.Tags) != len(e2.Tags) {
		return false
	}

	for i := 0; i < len(e.Tags); i++ {
		if e.Tags[i] != e2.Tags[i] {
			return false
		}
	}

	return true
}

// MustNewEvent constructs a new Event.
func MustNewEvent(name string, level zerolog.Level, tags ...Tag) *Event {
	event, e := NewEvent(name, level, tags...)
	if e != nil {
		log.Panic(e)
	}
	return event
}

// NewEvent constructs a new Event.
//
// Name will be trimmed.
// Tags will be trimmed, lowercased, deduped, and sorted.
func NewEvent(name string, level zerolog.Level, tags ...Tag) (*Event, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("event name must not be blank")
	}
	normalizedTags, e := normalizeTags(tags...)
	if e != nil {
		return nil, e
	}
	event := &Event{
		Name:  name,
		Level: level,
		Tags:  normalizedTags,
	}

	return event, nil
}

func normalizeTags(tags ...Tag) ([]string, error) {
	var normalizedTags []string
	if len(tags) > 0 {
		// dedupe the tags
		tagSet := make(map[string]bool, len(tags))
		for _, tag := range tags {
			tag = tag.Normalize()
			if tag == "" {
				return nil, errors.New("event tag must not be blank")
			}
			tagSet[tag.String()] = true
		}

		normalizedTags = make([]string, 0, len(tagSet))
		for tag := range tagSet {
			normalizedTags = append(normalizedTags, tag)
		}
		sort.Strings(normalizedTags)
	}
	return normalizedTags, nil
}

// Tag is used to define tags as constants in a type safe manner.
// Tags must be defined as lowercase using snake case.
type Tag string

func (t Tag) String() string {
	return string(t)
}

// Normalize will trim and lowercase the tag, i.e., normalize the tag name.
func (t Tag) Normalize() Tag {
	return Tag(strings.ToLower(strings.TrimSpace(t.String())))
}

// Log starts a new log message.
// - Event.Level is used as the message log level
// - Event.Name is used for the `EventName` log field value
// - Event.Tags are logged, if not empty
//
// NOTE: You must call Msg on the returned event in order to send the event.
func (e *Event) Log(logger *zerolog.Logger) *zerolog.Event {
	event := logger.WithLevel(e.Level).Str(string(EventName), e.Name)
	if len(e.Tags) > 0 {
		event.Strs(string(Tags), e.Tags)
	}
	return event
}
