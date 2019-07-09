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
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"io"
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
//   - time.Duration fields are rendered as int instead float because it's more efficiency
func init() {
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.ErrorFieldName = "e"

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldInteger = true
}

// standard top level logger field names
const (
	Name      = "n" // event name - should be a ULID
	Component = "c" // component name - should be a ULID
	ULID      = "z" // event instance ULID
)

var newEventID = ulidgen.MonotonicULIDGenerator()

// ForEvent returns a new logger with the event type ID field 'n' set to the specified value.
//
// The event should be unique. To ensure uniqueness, use ULIDs.
func ForEvent(logger *zerolog.Logger, name string) *zerolog.Logger {
	l := logger.With().Str(Name, name).Logger()
	return &l
}

// ForComponent returns a new logger with the component field 'c' set to the specified value.
// To ensure uniqueness, use ULIDs.
func ForComponent(logger *zerolog.Logger, name string) *zerolog.Logger {
	l := logger.With().Str(Component, name).Logger()
	return &l
}

// WithEventULID augments each log event with an event ULID.
//
// NOTE: The ULID uses a monotonic generator - thus, it's timestamp portion is simply used to construct the ULID
// and does not represent when the ULID was created.
func WithEventULID(logger zerolog.Logger) zerolog.Logger {
	return logger.Hook(zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
		e.Str(ULID, newEventID().String())
	}))
}

// NewZeroLogger constructs a new zerolog.Logger that is configured to add the following fields:
//  - timestamp in UNIX time format
//  - event ULID
//
// Example log message:
//
// {"z":"01DFBGCFD9WD29SGRJPK8KZKQS","t":1562680638,"m":"Hello World"}
//
// where z -> event ULID
//       t -> event timestamp
func NewZeroLogger(w io.Writer) zerolog.Logger {
	return WithEventULID(zerolog.New(w)).
		With().
		Timestamp().
		Logger()
}
