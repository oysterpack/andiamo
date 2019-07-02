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
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"log"
	"strings"
)

type Builder interface {
	Description(desription string) Builder

	YellowImpact(impact string) Builder

	RedImpact(impact string) Builder

	Checker(func(ctx context.Context) Failure) Builder

	Build() (Check, error)

	MustBuild() Check
}

type Check interface {
	Desc() Desc

	ID() ulid.ULID

	Description() string

	YellowImpact() string

	RedImpact() string

	Run(ctx context.Context) Result
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
	run          func(ctx context.Context) Failure
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

func (c *check) Run(ctx context.Context) Result {
	ch := make(chan Failure)
	result := NewResultBuilder()
	go func() {
		ch <- c.run(ctx)
	}()

	select {
	case <-ctx.Done():
		return result.Red(TimeoutError{})
	case failure := <-ch:
		switch failure.Status() {
		case Green:
			return result.Green()
		case Yellow:
			return result.Yellow(failure)
		default:
			return result.Red(failure)
		}
	}
}

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

func YellowFailure(err error) Failure {
	return failure{err, Yellow}
}

func RedFailure(err error) Failure {
	return failure{err, Red}
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "health check timed out"
}
