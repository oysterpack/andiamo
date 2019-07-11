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
	"github.com/oysterpack/partire-k8s/pkg/ulids"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"strings"
	"time"
)

// Check defines a health check
type Check struct {
	// ID format is ULID
	ID          string
	Description string
	// RedImpact describes what is the application impact of the health check status is red.
	RedImpact string
	// YellowImpact describes what is the application impact of the health check status is yellow.
	// YellowImpact is optional because some health checks may not have a yellow state.
	YellowImpact string // optional
	// Tags are used to categorize related health checks.
	// Tags are ULIDs because naming is hard and we want to avoid accidental collision.
	Tags []string // optional
}

// Checker performs the health check
type Checker func(ctx context.Context) Failure

// checker constraints
const (
	// Health checks should be designed to run fast
	MaxTimeout = 10 * time.Second
	// Health checks should not be scheduled to run more frequently than every second
	MinRunInterval = time.Second
)

// checker defaults
const (
	DefaultTimeout     = 5 * time.Second
	DefaultRunInterval = 15 * time.Second
)

// CheckerOpts is used to configure Checker run options.
// Zero values imply using the system default values.
type CheckerOpts struct {
	// Timeout must not be zero
	Timeout time.Duration
	// Used to schedule health checks to be run on an interval
	RunInterval time.Duration
}

func TrimSpace(check Check) Check {
	check.ID = strings.TrimSpace(check.ID)
	check.Description = strings.TrimSpace(check.Description)
	check.RedImpact = strings.TrimSpace(check.RedImpact)
	check.YellowImpact = strings.TrimSpace(check.YellowImpact)

	for i := 0; i < len(check.Tags); i++ {
		check.Tags[i] = strings.TrimSpace(check.Tags[i])
	}

	return check
}

// Check validation errors
var (
	ErrIDNotULID        = errors.New("ID must be a ULID")
	ErrBlankDescription = errors.New("description must not be blank")
	ErrBlankRedImpact   = errors.New("red impact must not be blank")
	ErrTagNotULID       = errors.New("tags must be ULIDs")
)

func Validate(check Check) error {
	_, err := ulids.Parse(check.ID)
	if err != nil {
		err = multierr.Append(ErrIDNotULID, err)
	}
	if check.Description == "" {
		err = multierr.Append(err, ErrBlankDescription)
	}
	if check.RedImpact == "" {
		err = multierr.Append(err, ErrBlankDescription)
	}
	for _, tag := range check.Tags {
		if _, err = ulids.Parse(tag); err != nil {
			err = multierr.Append(ErrTagNotULID, err)
			break
		}
	}

	return err
}
