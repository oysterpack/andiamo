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
	"errors"
	"github.com/oklog/ulid"
	"go.uber.org/multierr"
	"log"
	"strings"
)

// Desc is used to describe health checks.
type Desc interface {
	ID() ulid.ULID

	Description() string

	YellowImpact() string

	RedImpact() string
}

// DescBuilder is used to construct a new health check Desc
type DescBuilder interface {
	Description(desription string) DescBuilder

	YellowImpact(impact string) DescBuilder

	RedImpact(impact string) DescBuilder

	Build() (Desc, error)

	MustBuild() Desc
}

type descBuilder struct {
	desc
}

// NewDescBuilder constructs a new DescBuilder instance.
func NewDescBuilder(id ulid.ULID) DescBuilder {
	b := &descBuilder{}
	b.desc.id = id
	return b
}

func (b *descBuilder) Description(description string) DescBuilder {
	b.description = description
	return b
}

func (b *descBuilder) YellowImpact(impact string) DescBuilder {
	b.yellowImpact = impact
	return b
}

func (b *descBuilder) RedImpact(impact string) DescBuilder {
	b.redImpact = impact
	return b
}

func (b *descBuilder) Build() (Desc, error) {
	b.trimSpace()
	err := b.validate()
	if err != nil {
		return nil, err
	}
	return &b.desc, nil
}

func (b *descBuilder) trimSpace() {
	b.description = strings.TrimSpace(b.description)
	b.yellowImpact = strings.TrimSpace(b.yellowImpact)
	b.redImpact = strings.TrimSpace(b.redImpact)
}

func (b *descBuilder) validate() error {
	var err error

	if b.description == "" {
		err = errors.New("Description is required and must not be blank")
	}
	if b.redImpact == "" {
		err = multierr.Append(err, errors.New("RedImpact is required and must not be blank"))
	}

	return err
}

func (b *descBuilder) MustBuild() Desc {
	c, err := b.Build()
	if err != nil {
		log.Panic(err)
	}
	return c
}

type desc struct {
	id           ulid.ULID
	description  string
	yellowImpact string
	redImpact    string
}

func (d *desc) ID() ulid.ULID {
	return d.id
}

func (d *desc) Description() string {
	return d.description
}

func (d *desc) YellowImpact() string {
	return d.yellowImpact
}

func (d *desc) RedImpact() string {
	return d.redImpact
}
