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

package health

import (
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid"
	"time"
)

// Result for running the health check
type Result interface {
	HealthCheckID() ulid.ULID

	// Time is when the health check was run
	Time() time.Time
	// Duration is how long it took for the health check to run
	Duration() time.Duration

	Status() Status

	// Error that caused the health check to fail.
	// If the health check passed, i.e., Status == Green, then error will be nil
	Error() error

	fmt.Stringer
	json.Marshaler
}

// ResultBuilder is used to construct new Result instances
type ResultBuilder interface {
	Green() Result

	Yellow(err error) Result

	Red(err error) Result
}

type result struct {
	healthCheckID ulid.ULID
	start         time.Time
	duration      time.Duration
	status        Status
	err           error
}

// NewResultBuilder is used to construct new ResultBuilder instances
func NewResultBuilder(healthCheckID ulid.ULID) ResultBuilder {
	return &result{
		healthCheckID: healthCheckID,
		start:         time.Now(),
	}
}

func (r *result) String() string {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("%#v", r)
	}
	return string(jsonBytes)
}

func (r *result) MarshalJSON() ([]byte, error) {
	type JSON struct {
		HealthCheckID ulid.ULID
		Start         time.Time
		Duration      time.Duration
		Status        Status
		Err           string
	}

	err := ""
	if r.err != nil {
		err = r.err.Error()
	}

	return json.Marshal(JSON{
		r.healthCheckID,
		r.start,
		r.duration,
		r.status,
		err,
	})
}

func (r *result) HealthCheckID() ulid.ULID {
	return r.healthCheckID
}

func (r *result) Time() time.Time {
	return r.start
}

func (r *result) Duration() time.Duration {
	return r.duration
}

func (r *result) Status() Status {
	return r.status
}

func (r *result) Error() error {
	return r.err
}

func (r *result) Green() Result {
	r.duration = time.Since(r.start)
	return r
}

func (r *result) Yellow(err error) Result {
	r.duration = time.Since(r.start)
	r.err = err
	r.status = Yellow
	return r
}

func (r *result) Red(err error) Result {
	r.duration = time.Since(r.start)
	r.err = err
	r.status = Red
	return r
}
