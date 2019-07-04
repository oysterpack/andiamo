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
	"github.com/pkg/errors"
	"sync"
)

// Registry is used as a health registry
type Registry interface {
	// Register is used to register health checks.
	// An error will be returned if a health check with the same ID is already registered.
	Register(check Check) error

	// HealthChecks returns health checks that match against the filter.
	// If the filter is nil, then all health checks are returned
	HealthChecks(filter func(Check) bool) []Check

	// Subscribe is used to be notified when a health check has been registered
	Subscribe() <-chan Check
}

// NewRegistry creates a new Registry
func NewRegistry() Registry {
	return &registry{}
}

type registry struct {
	sync.RWMutex
	checks        []Check
	subscriptions []chan<- Check
}

func (r *registry) Register(check Check) error {
	if check == nil {
		return errors.New("check was nil")
	}

	r.Lock()
	defer r.Unlock()
	for _, c := range r.checks {
		if c.ID() == check.ID() {
			return fmt.Errorf("health check is already registered using same ID : %v", c)
		}
	}

	r.checks = append(r.checks, check)
	for _, ch := range r.subscriptions {
		subscriber := ch
		go func() {
			subscriber <- check
		}()
	}
	return nil
}

func (r *registry) HealthChecks(filter func(c Check) bool) []Check {
	r.RLock()
	defer r.RUnlock()

	if filter == nil {
		checks := make([]Check, len(r.checks))
		copy(checks, r.checks)
		return checks
	}

	var checks []Check
	for _, c := range r.checks {
		if filter(c) {
			checks = append(checks, c)
		}
	}
	return checks
}

func (r *registry) Subscribe() <-chan Check {
	r.RLock()
	defer r.RUnlock()
	ch := make(chan Check)
	r.subscriptions = append(r.subscriptions, ch)
	return ch
}
