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
	"github.com/oysterpack/partire-k8s/pkg/ulids"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"strings"
	"sync"
	"time"
)

type service struct {
	checks []RegisteredCheck

	stop                chan struct{}
	register            chan registerRequest
	getRegisteredChecks chan getRegisteredChecksRequest

	// TODO: refactor into its own service
	// only 1 health check at a time can run
	runLock    sync.Mutex
	results    chan Result
	runResults map[string]Result
}

func newService() *service {
	return &service{
		stop:                make(chan struct{}),
		register:            make(chan registerRequest),
		getRegisteredChecks: make(chan getRegisteredChecksRequest),

		results:    make(chan Result),
		runResults: make(map[string]Result),
	}
}

func (s *service) run() {
Loop:
	for {
		select {
		case <-s.stop:
			s.Stop()
			return
		case req := <-s.register:
			err := s.Register(req)
			if err == nil {
				close(req.reply)
				continue Loop
			}
			go func() {
				defer close(req.reply)
				select {
				case <-s.stop:
				case req.reply <- err:
				}
			}()
		case result := <-s.results:
			s.runResults[result.HealthCheckID] = result
		case req := <-s.getRegisteredChecks:
			s.SendRegisteredChecks(req)
		}
	}
}

func (s *service) TriggerShutdown() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
}

// Stop signals the service to shutdown
func (s *service) Stop() {
	// TODO
}

type registerRequest struct {
	check   Check
	opts    CheckerOpts
	checker Checker

	reply chan error
}

func (s *service) Register(req registerRequest) error {
	TrimSpace := func(check Check) Check {
		check.ID = strings.TrimSpace(check.ID)
		check.Description = strings.TrimSpace(check.Description)
		check.RedImpact = strings.TrimSpace(check.RedImpact)
		check.YellowImpact = strings.TrimSpace(check.YellowImpact)

		for i := 0; i < len(check.Tags); i++ {
			check.Tags[i] = strings.TrimSpace(check.Tags[i])
		}

		return check
	}

	Validate := func(check Check) error {
		_, err := ulids.Parse(check.ID)
		if err != nil {
			err = multierr.Append(ErrIDNotULID, err)
		}
		if check.Description == "" {
			err = multierr.Append(err, ErrBlankDescription)
		}
		if check.RedImpact == "" {
			err = multierr.Append(err, ErrBlankRedImpact)
		}
		for _, tag := range check.Tags {
			if _, err = ulids.Parse(tag); err != nil {
				err = multierr.Append(ErrTagNotULID, err)
				break
			}
		}

		return err
	}

	WithTimeout := func(checker Checker, timeout time.Duration) Checker {
		return func() error {
			reply := make(chan error)
			timer := time.After(timeout)
			go func() {
				select {
				case <-timer:
				case reply <- checker():
				}
			}()

			select {
			case <-timer:
				return ErrTimeout
			case err := <-reply:
				return err
			}
		}
	}

	ToStatus := func(err error) Status {
		if err == nil {
			return Green
		}
		if _, ok := err.(YellowError); ok {
			return Yellow
		}
		return Red
	}

	Schedule := func(id string, check Checker, interval time.Duration) {
		run := func() {
			s.runLock.Lock()
			defer s.runLock.Unlock()
			start := time.Now()
			err := check()
			duration := time.Since(start)
			result := Result{
				HealthCheckID: id,

				Status: ToStatus(err),
				error:  err,

				Time:     start,
				Duration: duration,
			}
			go func() {
				select {
				case <-s.stop:
				case s.results <- result:
				}
			}()
		}

		// run the health check immediately
		run()

		// then run it on its specified interval
		for {
			timer := time.After(interval)
			select {
			case <-s.stop:
				return
			case <-timer:
				run()
			}
		}
	}

	ApplyDefaultOpts := func(opts CheckerOpts) CheckerOpts {
		if opts.Timeout == time.Duration(0) {
			opts.Timeout = DefaultTimeout
		}
		if opts.RunInterval == time.Duration(0) {
			opts.RunInterval = DefaultRunInterval
		}

		return opts
	}

	ValidateOpts := func(opts CheckerOpts) error {
		var err error
		if opts.RunInterval < MinRunInterval {
			err = ErrRunIntervalTooFrequent
		}
		if opts.Timeout > MaxTimeout {
			err = multierr.Append(err, ErrRunTimeoutTooHigh)
		}
		return err
	}

	check := TrimSpace(req.check)
	if err := Validate(check); err != nil {
		return multierr.Append(fmt.Errorf("Invalid health check: %#v", check), err)
	}

	if req.checker == nil {
		return multierr.Append(errors.New(check.ID), ErrNilChecker)
	}

	opts := ApplyDefaultOpts(req.opts)
	if err := ValidateOpts(opts); err != nil {
		return multierr.Append(fmt.Errorf("Invalid health checker opts: %s : %#v", check.ID, opts), err)
	}

	if s.RegisteredCheck(check.ID) != nil {
		return fmt.Errorf("health check is already registered: %s", check.ID)
	}

	registeredCheck := RegisteredCheck{
		Check:       check,
		CheckerOpts: opts,
		Checker:     WithTimeout(req.checker, opts.Timeout),
	}
	s.checks = append(s.checks, registeredCheck)
	go Schedule(registeredCheck.ID, registeredCheck.Checker, registeredCheck.RunInterval)

	return nil
}

func (s *service) RegisteredCheck(id string) *RegisteredCheck {
	for _, c := range s.checks {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

type getRegisteredChecksRequest struct {
	filter func(c Check, opts CheckerOpts) bool
	reply  chan<- []RegisteredCheck
}

func (s *service) SendRegisteredChecks(req getRegisteredChecksRequest) {
	var checks []RegisteredCheck
	if req.filter == nil {
		checks = make([]RegisteredCheck, len(s.checks))
		copy(checks, s.checks)
	} else {
		for _, check := range s.checks {
			if req.filter(check.Check, check.CheckerOpts) {
				checks = append(checks, check)
			}
		}
	}

	go func() {
		select {
		case <-s.stop:
		case req.reply <- checks:
		}
	}()

}
