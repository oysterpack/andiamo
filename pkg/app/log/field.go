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

package log

// Field is used to define log event fields used for structured logging.
type Field string

// Field enum
const (
	// TIMESTAMP specifies when the log event occurred in Unix time.
	TIMESTAMP = Field("t")
	// LEVEL specifies the log level.
	LEVEL = Field("l")
	// MESSAGE specifies the log message.
	MESSAGE = Field("m")
	// ERROR specifies the error message.
	ERROR = Field("e")

	// PACKAGE specifies which package logged the event
	PACKAGE = Field("p")
	// EVENT is used to specify the event name. All log events should specify the event name.
	EVENT = Field("n")

	// APP is used to group app related fields
	APP = Field("a")
	// APP_ID stores app.ID as a ULID
	APP_ID = Field("i")
	// APP_RELEASE_ID stores the app.ReleaseID as a ULID
	APP_RELEASE_ID = Field("r")
	// APP_NAME stores app.Name
	APP_NAME = Field("n")
	// APP_VERSION stores app.Version
	APP_VERSION = Field("v")
	// APP_INSTANCE_ID stores app.InstanceID
	APP_INSTANCE_ID = Field("x")

	// COMPONENT is used to to specify the application component that logged the event. It stores the component name.
	COMPONENT = Field("c")
)
