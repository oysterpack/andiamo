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
	"sync"
)

// EventRegistry is used to register application events
type EventRegistry struct {
	m      sync.RWMutex
	events []*Event
}

// NewEventRegistry is used to construct a new EventRegistry
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{}
}

// Register is used to register application events
func (r *EventRegistry) Register(event ...*Event) {
	r.m.Lock()
	defer r.m.Unlock()
	for _, event := range event {
		if !r.registered(event) {
			r.events = append(r.events, event)
		}
	}
}

// Registered returns true is the event is already registered.
func (r *EventRegistry) Registered(event *Event) bool {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.registered(event)
}

func (r *EventRegistry) registered(event *Event) bool {
	for i := 0; i < len(r.events); i++ {
		if r.events[i].Equals(event) {
			return true
		}
	}
	return false
}

// Events returns all registered events
func (r *EventRegistry) Events() []*Event {
	r.m.RLock()
	defer r.m.RUnlock()
	events := make([]*Event, len(r.events))
	copy(events, r.events)
	return events
}

// Filter returns all Events that match the specified filter
func (r *EventRegistry) Filter(filter func(event *Event) bool) []*Event {
	r.m.RLock()
	defer r.m.RUnlock()
	var events []*Event
	for _, event := range r.events {
		if filter(event) {
			events = append(events, event)
		}
	}
	return events
}
