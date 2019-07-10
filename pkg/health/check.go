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
	"github.com/rs/zerolog"
	"go.uber.org/multierr"
	"log"
	"strings"
	"time"
)

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
	zerolog.LogObjectMarshaler
}

// CheckOpts is used to construct a new health Check instance.
type CheckOpts struct {
	Desc
	ID           string // ULID
	Description  string
	RedImpact    string
	YellowImpact string // optional
	Checker      func(ctx context.Context) Failure
	Timeout      time.Duration // optional - default = 5 secs
	Interval     time.Duration // optional - default = 15 secs
}

type checkConstraints struct {
	minRunInterval, maxRunTimeout time.Duration
}

// New constructs a new health Check
func (opts CheckOpts) New() (Check, error) {
	return opts.new(checkConstraints{
		minRunInterval: time.Second,
		maxRunTimeout:  10 * time.Second,
	})
}

// MustNew constructs a new health Check and panics if the opts are invalid
func (opts CheckOpts) MustNew() Check {
	check, err := opts.New()
	if err != nil {
		log.Panic(err)
	}
	return check
}

func (opts CheckOpts) mustNew(constraints checkConstraints) Check {
	check, err := opts.new(constraints)
	if err != nil {
		log.Panic(err)
	}
	return check
}

// health check default values
const (
	DefaultTimeout     = 5 * time.Second
	DefaultRunInterval = 15 * time.Second
)

func (opts CheckOpts) new(constraints checkConstraints) (Check, error) {
	opts = opts.normalize()
	id, err := ulid.Parse(opts.ID)
	if err != nil {
		return nil, err
	}
	var zeroULID ulid.ULID
	if id == zeroULID {
		err = errors.New("ID cannot be zero")
	}
	err = multierr.Append(err, opts.validate(constraints))
	if err != nil {
		return nil, err
	}

	check := &check{
		desc:         opts.Desc,
		id:           id,
		description:  opts.Description,
		yellowImpact: opts.YellowImpact,
		redImpact:    opts.RedImpact,

		run:      opts.Checker,
		timeout:  opts.Timeout,
		interval: opts.Interval,
	}

	return check, nil
}

func (opts CheckOpts) normalize() CheckOpts {
	opts.ID = strings.TrimSpace(opts.ID)
	opts.Description = strings.TrimSpace(opts.Description)
	opts.RedImpact = strings.TrimSpace(opts.RedImpact)
	opts.YellowImpact = strings.TrimSpace(opts.YellowImpact)
	if opts.Timeout == time.Duration(0) {
		opts.Timeout = DefaultTimeout
	}
	if opts.Interval == time.Duration(0) {
		opts.Interval = DefaultRunInterval
	}
	return opts
}

func (opts CheckOpts) validate(constraints checkConstraints) error {
	var err error

	if opts.Desc == nil {
		err = errors.New("desc is required")
	}

	if opts.Description == "" {
		err = multierr.Append(err, errors.New("description is required and must not be blank"))
	}
	if opts.RedImpact == "" {
		err = multierr.Append(err, errors.New("red impact is required and must not be blank"))
	}
	if opts.Checker == nil {
		err = multierr.Append(err, errors.New("check function is required"))
	}
	// all health checks must be constrained in how long they run
	if opts.Timeout <= time.Duration(0) {
		err = multierr.Append(err, errors.New("timeout cannot be 0"))
	}
	// application health checks should be designed to be fast
	if opts.Timeout > constraints.maxRunTimeout {
		err = multierr.Append(err, fmt.Errorf("timeout cannot be more than %s", constraints.maxRunTimeout))
	}
	// this is to protect ourselves from accidentally scheduling a health check to run too often
	if opts.Interval < constraints.minRunInterval {
		err = multierr.Append(err, fmt.Errorf("run interval cannot be less than %s", constraints.minRunInterval))
	}

	return err
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
		YellowImpact string `json:",omitempty"`
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

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (c *check) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("id", c.ID().String()).
		Str("desc_id", c.Desc().ID().String()).
		Strs("description", []string{c.Desc().Description(), c.Description()}).
		Strs("red_impact", []string{c.Desc().RedImpact(), c.RedImpact()})

	var yellowImpact []string
	if c.Desc().YellowImpact() != "" {
		yellowImpact = append(yellowImpact, c.Desc().YellowImpact())
	}
	if c.YellowImpact() != "" {
		yellowImpact = append(yellowImpact, c.YellowImpact())
	}
	if len(yellowImpact) != 0 {
		e.Strs("yellow_impact", yellowImpact)
	}

	e.
		Dur("timeout", c.Timeout()).
		Dur("run_interval", c.RunInterval())
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
		return result.Red(ErrTimeout)
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

// ErrTimeout indicates a health check timed out.
// Healthcheck timeout errors are flagged as Red.
var ErrTimeout = errors.New("health check timed out")
