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

// Field is used to define log event fields used for structured logging.
type Field string

// Field enum
const (
	// standard field names
	// ID
	ID = Field("i")
	// Name stores app.Name
	Name = Field("n")
	// InstanceID
	InstanceID = Field("x")

	// Timestamp specifies when the log event occurred in Unix time.
	Timestamp = Field("t")
	// Level specifies the log level.
	Level = Field("l")
	// Message specifies the log message.
	Message = Field("m")
	// Error specifies the error message.
	Error = Field("e")
	// Stack is used to log the stack trace.
	Stack = Field("s")

	// Package specifies which package logged the event
	Package = Field("p")
	// EventName is used to specify the event name. All log events should specify the event name.
	EventName = Name
	// Tags is used to tag log events.
	// Tags can be used to further categorize or group related log events, e.g, trace id, application layer (frontend, backend, data, messaging)
	Tags = Field("g")

	// Err is used to group error related fields
	// - f = failure
	Err = Field("f")
	// ErrID stores the unique error ID
	ErrID = ID
	// ErrName stores the human readable name
	ErrName = Name
	// ErrSrcID stores error source ID
	ErrSrcID = Field("s")
	// ErrInstanceID stores error instance ID
	ErrInstanceID = InstanceID

	// App is used to group app related fields
	App = Field("a")
	// AppID stores app.ID as a ULID
	AppID = ID
	// AppReleaseID stores the app.ReleaseID as a ULID
	AppReleaseID = Field("r")
	// AppName stores app.Name
	AppName = Name
	// AppVersion stores app.Version
	AppVersion = Field("v")
	// AppInstanceID stores app.InstanceID
	AppInstanceID = InstanceID

	// Component is used to to specify the application component that logged the event. It stores the component name.
	Component = Field("c")
)
