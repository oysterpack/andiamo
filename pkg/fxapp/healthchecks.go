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
	"context"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"strings"
	"time"
)

// HealthCheckClass defines a health check class
type HealthCheckClass struct {
	ID           string // should be a ULID
	Description  string // what does the health check do
	YellowImpact string // what's the impact if the health check is yellow - may be blank, if the health check does not have a yellow state
	RedImpact    string // what's the impact if the health check is red
}

// Validate the HealthCheckClass:
//  - ID must parse as a ULID
//  - Description must not be blank
//  - RedImpact must not be blank
func (h *HealthCheckClass) Validate() error {
	var err error

	if h == nil {
		return errors.New("HealthCheckClass is nil")
	}

	if _, e := ulid.Parse(h.ID); e != nil {
		err = e
	}

	if strings.TrimSpace(h.Description) == "" {
		err = multierr.Append(err, errors.New("description must not be blank"))
	}

	if strings.TrimSpace(h.RedImpact) == "" {
		err = multierr.Append(err, errors.New("red impact must not be blank"))
	}

	return err
}

// HealthCheck represents a health check for a HealthCheckClass
type HealthCheck struct {
	Class *HealthCheckClass
	ID    string // should be a ULID

	// Augments the HealthCheckClass info
	Description  string // what does the health check do
	YellowImpact string // what's the impact if the health check is yellow - may be blank, if the health check does not have a yellow state
	RedImpact    string // what's the impact if the health check is red

	Checker func(ctx context.Context) HealthCheckError
}

// Validate the health check:
//  - HealthCheckClass must pass validation
//  - Checker func must not be nil
func (h *HealthCheck) Validate() error {
	var err = h.Class.Validate()
	if _, e := ulid.Parse(h.ID); e != nil {
		err = multierr.Append(err, e)
	}

	if h.Checker == nil {
		err = multierr.Append(err, errors.New("checker func is required"))
	}

	return err
}

// HealthErrorStatus is used to define a health error status
type HealthErrorStatus uint8

// HealthErrorStatus enum
const (
	Yellow HealthErrorStatus = iota
	Red
)

func (e HealthErrorStatus) String() string {
	if e == Yellow {
		return "Yellow"
	}
	return "Red"
}

// HealthCheckError indicates the health check failed.
// It could have failed in a yellow or red state.
type HealthCheckError interface {
	error
	Status() HealthErrorStatus
}

type healthCheckError struct {
	error
	HealthErrorStatus
}

func (err healthCheckError) Error() string {
	return fmt.Sprintf("%s: %v", err.HealthErrorStatus, err.error)
}

func (err healthCheckError) Status() HealthErrorStatus {
	return err.HealthErrorStatus
}

// YellowHealthCheckError creates a HealthCheckError with a Yellow status
func YellowHealthCheckError(err error) HealthCheckError {
	return healthCheckError{err, Yellow}
}

// RedHealthCheckError creates a HealthCheckError with a Red status
func RedHealthCheckError(err error) HealthCheckError {
	return healthCheckError{err, Red}
}

// HealthCheckRegistration is used to register the health check with the app
type HealthCheckRegistration struct {
	fx.Out

	*HealthCheck
	Interval
	Timeout
}

// NewHealthCheckRegistration creates a new HealthCheckRegistration
//  - runInterval must be at least 1 sec
//  - timeout must be greater than 0, at most 10 secs, and must be less than the run interval
//    - health checks must be designed to run fast - even 10 seconds is rather long
func NewHealthCheckRegistration(healthCheck *HealthCheck, runInterval Interval, timeout Timeout) (*HealthCheckRegistration, error) {
	if err := healthCheck.Validate(); err != nil {
		return nil, err
	}

	if time.Duration(runInterval) < time.Second {
		return nil, fmt.Errorf("run interval cannot be less than 1 sec: %v", runInterval)
	}

	if time.Duration(timeout) == time.Duration(0) {
		return nil, errors.New("timeout cannot be 0")
	}

	if time.Duration(timeout) >= time.Duration(runInterval) {
		return nil, fmt.Errorf("timeout must be less than the run interval: %v > %v", timeout, runInterval)
	}

	if time.Duration(timeout) > 10*time.Second {
		return nil, fmt.Errorf("timeout cannot be greater than 10 seconds: %v", timeout)
	}

	return &HealthCheckRegistration{
		HealthCheck: healthCheck,
		Interval:    runInterval,
		Timeout:     timeout,
	}, nil
}
