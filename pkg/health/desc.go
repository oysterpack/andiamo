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
	"errors"
	"fmt"
	"github.com/oklog/ulid"
	"go.uber.org/multierr"
	"log"
	"strings"
)

// Desc is used to describe health checks.
//
// The descriptions are meant to be short
type Desc interface {
	ID() ulid.ULID

	// Description briefly describes the health check's purpose
	Description() string

	// YellowImpact describes what it means if the health check error status is Yellow
	YellowImpact() string

	// RedImpact describes what it means if the health check error status is Red
	RedImpact() string

	fmt.Stringer
	json.Marshaler
}

// DescOpts is used to define a health check descriptor
type DescOpts struct {
	ID           string // ULID
	Description  string
	RedImpact    string
	YellowImpact string // optional
}

// New is used to construct a new health check descriptor
func (opts DescOpts) New() (Desc, error) {
	opts = opts.normalize()
	id, err := ulid.Parse(opts.ID)
	err = multierr.Append(err, opts.validate())
	if err != nil {
		return nil, err
	}
	return &desc{
		id:           id,
		description:  opts.Description,
		yellowImpact: opts.YellowImpact,
		redImpact:    opts.RedImpact,
	}, nil
}

func (opts DescOpts) normalize() DescOpts {
	opts.ID = strings.TrimSpace(opts.ID)
	opts.Description = strings.TrimSpace(opts.Description)
	opts.YellowImpact = strings.TrimSpace(opts.YellowImpact)
	opts.RedImpact = strings.TrimSpace(opts.RedImpact)
	return opts
}

func (opts DescOpts) validate() error {
	var err error

	if opts.Description == "" {
		err = errors.New("Description is required and must not be blank")
	}
	if opts.RedImpact == "" {
		err = multierr.Append(err, errors.New("RedImpact is required and must not be blank"))
	}

	return err
}

// MustNew panics if the health check descriptor is not valid
func (opts DescOpts) MustNew() Desc {
	c, err := opts.New()
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

func (d *desc) String() string {
	jsonBytes, err := d.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("%#v", d)
	}
	return string(jsonBytes)
}

func (d *desc) MarshalJSON() (text []byte, err error) {
	type Data struct {
		ID           ulid.ULID
		Description  string
		YellowImpact string `json:",omitempty"`
		RedImpact    string
	}
	return json.Marshal(Data{d.id, d.description, d.yellowImpact, d.redImpact})
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
