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
	"context"
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"log"
	"strings"
	"time"
)

// Builder is used to construct new Check instances
type Builder interface {
	Description(desription string) Builder

	YellowImpact(impact string) Builder

	RedImpact(impact string) Builder

	Checker(func(ctx context.Context) Failure) Builder

	// Timeout defaults to 5 secs
	Timeout(timeout time.Duration) Builder

	// RunInterval defaults to 10 secs
	RunInterval(interval time.Duration) Builder

	Build() (Check, error)

	MustBuild() Check
}

// Check represents a health check
type Check interface {
	Desc() Desc

	ID() ulid.ULID

	// Description augments the desc description
	Description() string

	// YellowImpact augments the desc yellow impact
	YellowImpact() string

	// RedImpact is required and augments the desc red impact
	RedImpact() string

	// Timeout is used to limit how long the health check is allowed to run.
	// If the health check times out, then it is considered a Red failure.
	Timeout() time.Duration

	// RunInterval is used to schedule the health check to run on a periodic basis.
	// The interval resets after the health check run completes.
	RunInterval() time.Duration

	// Run uses the timeout setting to time limit the health check
	Run() Result

	fmt.Stringer
	json.Marshaler
}

// NewBuilder constructs a new health check Builder.
//
// Defaults:
//  - timeout = 10 secs
//  - run interval = 15 secs
func NewBuilder(desc Desc, healthcheckID ulid.ULID) Builder {
	return &builder{
		check: &check{
			desc:     desc,
			id:       healthcheckID,
			timeout:  10 * time.Second,
			interval: 15 * time.Second,
		},
	}
}

type builder struct {
	check *check
}

func (b *builder) Description(desription string) Builder {
	b.check.description = desription
	return b
}

func (b *builder) YellowImpact(impact string) Builder {
	b.check.yellowImpact = impact
	return b
}

func (b *builder) RedImpact(impact string) Builder {
	b.check.redImpact = impact
	return b
}

func (b *builder) Checker(f func(ctx context.Context) Failure) Builder {
	b.check.run = f
	return b
}

func (b *builder) Timeout(timeout time.Duration) Builder {
	b.check.timeout = timeout
	return b
}

func (b *builder) RunInterval(interval time.Duration) Builder {
	b.check.interval = interval
	return b
}

func (b *builder) Build() (Check, error) {
	b.trimSpace()
	err := b.validate()
	if err != nil {
		return nil, err
	}

	return b.check, nil
}

func (b *builder) trimSpace() {
	b.check.description = strings.TrimSpace(b.check.description)
	b.check.yellowImpact = strings.TrimSpace(b.check.yellowImpact)
	b.check.redImpact = strings.TrimSpace(b.check.redImpact)
}

// health check run constraints
var (
	// Health checks can not be scheduled to run more frequently than once per second.
	// This is used to prevent health checks from overwhelming the application by being run to frequently because
	// of a misconfiguration.
	minRunInterval = time.Second
	// Health checks should be designed to run fast.
	maxRunTimeout = 10 * time.Second
)

func (b *builder) validate() error {
	var err error

	if b.check.description == "" {
		err = errors.New("Description is required and must not be blank")
	}
	if b.check.redImpact == "" {
		err = multierr.Append(err, errors.New("RedImpact is required and must not be blank"))
	}
	if b.check.run == nil {
		err = multierr.Append(err, errors.New("check function is required"))
	}
	// all health checks must be constrained in how long they run
	if b.check.timeout <= time.Duration(0) {
		err = multierr.Append(err, errors.New("timeout cannot be 0"))
	}
	// application health checks should be designed to be fast
	if b.check.timeout > maxRunTimeout {
		err = multierr.Append(err, fmt.Errorf("timeout cannot be more than %s", maxRunTimeout))
	}
	// this is to protect ourselves from accidentally scheduling a health check to run too often
	if b.check.interval < minRunInterval {
		err = multierr.Append(err, fmt.Errorf("run interval cannot be less than %s", minRunInterval))
	}

	return err
}

func (b *builder) MustBuild() Check {
	c, err := b.Build()
	if err != nil {
		log.Panic(err)
	}

	return c
}

type check struct {
	desc         Desc
	id           ulid.ULID
	description  string
	yellowImpact string
	redImpact    string

	run      func(ctx context.Context) Failure
	timeout  time.Duration
	interval time.Duration
}

func (c *check) String() string {
	jsonBytes, err := c.MarshalJSON()
	if err != nil {
		// should never happen
		return fmt.Sprintf("%#v", c)
	}
	return string(jsonBytes)
}

func (c *check) MarshalJSON() (text []byte, err error) {
	type Data struct {
		Desc         Desc
		ID           ulid.ULID
		Description  string
		YellowImpact string
		RedImpact    string
		Timeout      time.Duration
		Interval     time.Duration
	}
	data := Data{
		c.desc,
		c.id,
		c.description,
		c.yellowImpact,
		c.redImpact,
		c.timeout,
		c.interval,
	}
	return json.Marshal(data)
}

func (c *check) ID() ulid.ULID {
	return c.id
}

func (c *check) Description() string {
	return c.description
}

func (c *check) YellowImpact() string {
	return c.yellowImpact
}

func (c *check) RedImpact() string {
	return c.redImpact
}

func (c *check) Desc() Desc {
	return c.desc
}

func (c *check) Run() Result {
	ch := make(chan Failure)
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	result := NewResultBuilder(c.id)
	go func() {
		ch <- c.run(ctx)
	}()

	select {
	case <-ctx.Done():
		return result.Red(TimeoutError{})
	case failure := <-ch:
		if failure == nil {
			return result.Green()
		}
		if failure.Status() == Yellow {
			return result.Yellow(failure)
		}
		return result.Red(failure)
	}
}

func (c *check) Timeout() time.Duration {
	return c.timeout
}

func (c *check) RunInterval() time.Duration {
	return c.interval
}

// Failure represents a health check failure
type Failure interface {
	error
	Status() Status
}

type failure struct {
	error
	status Status
}

func (f failure) Status() Status {
	return f.status
}

// YellowFailure constructs a new Failure with a Yellow status
func YellowFailure(err error) Failure {
	return failure{err, Yellow}
}

// RedFailure constructs a new Failure with a Red status
func RedFailure(err error) Failure {
	return failure{err, Red}
}

// TimeoutError is used for timeout errors.
// Healthcheck timeout errors are flagged as Red.
type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "health check timed out"
}
