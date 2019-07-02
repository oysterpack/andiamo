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
	"fmt"
	"time"
)

// Result for running the health check
type Result interface {
	// Time is when the health check was run
	Time() time.Time
	// Duration is how long it took for the health check to run
	Duration() time.Duration

	Status() Status

	// Error that caused the health check to fail.
	// If the health check passed, i.e., Status == Green, then error will be nil
	Error() error

	fmt.Stringer
}

// ResultBuilder is used to construct new Result instances
type ResultBuilder interface {
	Green() Result

	Yellow(err error) Result

	Red(err error) Result
}

type result struct {
	start    time.Time
	duration time.Duration
	status   Status
	err      error
}

// NewResultBuilder is used to construct new ResultBuilder instances
func NewResultBuilder() ResultBuilder {
	return &result{
		start: time.Now(),
	}
}

func (r *result) String() string {
	if r.err == nil {
		return fmt.Sprintf("Result{Start: %v, Duration: %v, Status: %v}", r.start, r.duration, r.status)
	}
	return fmt.Sprintf("Result{Start: %v, Duration: %v, Status: %v, Error: %q}", r.start, r.duration, r.status, r.err)

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
