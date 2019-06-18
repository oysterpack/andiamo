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

package err

import (
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"log"
	"strings"
)

var (
	newULID = ulidgen.MonotonicULIDGenerator()
)

// common errors
var (
	InvalidIDErrClass      = MustNewDesc("01DDKT49ZW36AA2ADYD1THB44Y", "InvalidIDErr", "invalid ID")
	InvalidVersionErrClass = MustNewDesc("01DDKTCMCEPQBDM1ADNMFECDCE", "InvalidVersionErr", "invalid version")
)

// Desc is used to define an error
type Desc struct {
	// ID is the globally unique error ID
	ID ulid.ULID
	// Name is the short human readable error name.
	// It should follow go const naming conventions - camel-case with the first letter uppercase.
	//
	// The name should be unique within its scope, i.e., application  or component scope. The scope is determined when the
	// error is logged. If it has a component field, then it is scoped to a component - otherwise it is scoped to the application.
	Name string
	// Message describes the error.
	Message string
	// Tags are used to classify errors, e.g., db, ui, timeout, authc, authz, io, client, server.
	// - tags can be used to organize log events and make it easier to query for events
	Tags []string
	// IncludeStack indicates whether the stacktrace should be logged with the error
	// NOTE: most of the time, the stacktrace does not need to be logged
	IncludeStack bool
}

// MustNewDesc constructs a new Desc
func MustNewDesc(id, name, message string, tags ...Tag) *Desc {
	desc, e := NewDesc(id, name, message, tags...)
	if e != nil {
		log.Panic(e)
	}
	return desc
}

// NewDesc constructs a new Desc
//
// Argument Constraints
// - id must be a ULID (https://github.com/ulid/spec)
// - name and message will be trimmed and cannot be blank
func NewDesc(id, name, message string, tags ...Tag) (*Desc, error) {
	ULID, e := ulid.Parse(id)
	if e != nil {
		return nil, e
	}

	name = strings.TrimSpace(name)
	message = strings.TrimSpace(message)

	if name == "" {
		return nil, errors.New("name cannot be blank")
	}

	if message == "" {
		return nil, errors.New("message cannot be blank")
	}

	var tagSlice []string
	if len(tags) > 0 {
		tagSlice = make([]string, len(tags))
		for i, tag := range tags {
			tagSlice[i] = tag.String()
		}
	}
	desc := &Desc{
		ID:      ULID,
		Name:    name,
		Message: message,
		Tags:    tagSlice,
	}

	return desc, nil
}

// WithStacktrace indicates that the error stacktrace will be logged.
//
// NOTE: logging the error stacktrace is very expensive. Most of the time, it is not needed. Errors are linked to a
// source code location via a ULID - see Err
func (d *Desc) WithStacktrace() *Desc {
	d.IncludeStack = true
	return d
}

func (d *Desc) String() string {
	return fmt.Sprintf("%v", *d)
}

// Tag is used to define tags as constants in a type safe manner
type Tag string

func (t Tag) String() string {
	return string(t)
}

// Common error tags
// - many of which are modeled after HTTP status codes
const (
	// ClientErr means the error was caused by the client, e.g., client submitted an invalid request
	ClientErr Tag = "client"
	// ServerErr means the error was caused by the server, e.g., unexpected server side error, rpc call failed
	ServerErr Tag = "server"
	// RemoteErr means the error was caused remotely, e.g., database error, rpc error
	RemoteErr Tag = "remote"

	// AuthcErr indicates that authentication failed
	AuthcErr Tag = "authc"
	// AuthzErr indicates authorization failed
	AuthzErr Tag = "authz"
	// BadRequestErr indicates a request failed because it was invalid.
	// This can be logged on the client and/or server side.
	BadRequestErr Tag = "bad_req"
	// ConflictErr indicates that the request conflicts with some other request.
	// For example, when using optimistic version control.
	ConflictErr Tag = "conflict"
	// PreconditionFailedErr indicates the request failed because a precondition failed
	PreconditionFailedErr Tag = "precondition_failed"
	// MessageTooLargeErr indicates a message was received that exceeds the max message size supported
	MessageTooLargeErr Tag = "msg_too_large"
	// UnprocessableErr indicates that the server understands the content type of the request entity, and the syntax of
	// the request entity is correct, but it was unable to process the contained instructions.
	// - The client should not repeat this request without modification.
	UnprocessableErr = "unprocessable"
	// RateLimitError indicates the client has sent too many requests in a given amount of time ("rate limiting").
	RateLimitErr Tag = "rate_limit"
	// ResourceQuotaErr indicates that a resource quota constraint would have been violated
	ResourceQuotaErr Tag = "resource_quota"
	// TimeoutErr indicates a timeout has occurred.
	// The error should include more context information:
	// - what timed out
	// - what is the timeout
	TimeoutErr Tag = "timeout"

	// NotImplementedErr indicates that the server does not support the functionality required to fulfill the request.
	NotImplementedErr Tag = "not_implemented"
	// ServiceUnavailableErr indicates that the server is not ready to handle the request.
	ServiceUnavailableErr Tag = "unavailable"

	// UIError indicates an error has occurred in the UI layer
	UIErr Tag = "ui"
	// IOError indicates some type of IO related error has occurred
	IOErr Tag = "io"
	// DatabaseErr indicates the error is database related
	DatabaseErr Tag = "db"
)

// Err is used to define application error instances. It links an error to a source code location via a ULID.
// Most of the time, error stack traces are not required. Getting the stack trace at runtime is expensive.
// Thus, most of the time, just knowing where in the code the error originated from is sufficient.
// Each log event should include the package. Combining the Err.SrcID with the log event package lets us differentiate
// between multiple different errors of the same type that are produced from the same package but from different source
// code locations. This is an interesting metric to monitor.
//
// Err is used to construct new err.Instance(s)
type Err struct {
	*Desc
	SrcID ulid.ULID
}

// New constructs a new Error instance
func New(desc *Desc, srcULID string) *Err {
	return &Err{
		Desc:  desc,
		SrcID: ulid.MustParse(srcULID),
	}
}

// New constructs a new error instance, which is assigned a unique InstanceID.
func (e *Err) New() *Instance {
	return &Instance{
		Err:        e,
		InstanceID: newULID(),
	}
}

// CausedBy constructs a new error instance which wraps the error cause.
func (e *Err) CausedBy(cause error) *Instance {
	return &Instance{
		Err:        e,
		InstanceID: newULID(),
		Cause:      cause,
	}
}

func (e *Err) String() string {
	return fmt.Sprintf("%v", *e)
}

// Instance represents an application error instance.
// All application errors should be wrapped within an Instance.
type Instance struct {
	*Err
	// InstanceID is the unique error instance ID.
	// use case: the InstanceID can be returned back to the client, which can be used to track down the specific error.
	InstanceID ulid.ULID
	// Cause if present, indicates what caused this error.
	Cause error
}

// Error implements the Error interface
func (e *Instance) Error() string {
	if e.Cause == nil {
		return e.Err.Message
	}
	return fmt.Sprintf("%s : %s", e.Err.Message, e.Cause.Error())
}

// Log logs the error using the specified logger
func (e *Instance) Log(logger *zerolog.Logger) *zerolog.Event {
	err := zerolog.Dict().
		Str(string(logging.ErrID), e.ID.String()).
		Str(string(logging.ErrName), e.Name).
		Str(string(logging.ErrSrcID), e.SrcID.String()).
		Str(string(logging.ErrInstanceID), e.InstanceID.String())

	if len(e.Tags) > 0 {
		err = err.Strs(string(logging.Tags), e.Tags)
	}

	event := logger.Error().Dict(string(logging.Err), err)
	if e.IncludeStack {
		event.Stack().Err(errors.WithStack(e))
	} else {
		event.Err(e)
	}

	return event
}
