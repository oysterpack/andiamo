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
	"time"
)

// Opts are being extracted in order to make running health checks on a schedule testable, i.e., make the tests
// run fast. The normal MinRunInterval is 1 sec and the MaxTimeout is 10 secs - too long to use in unit tests.
type Opts struct {
	MinRunInterval time.Duration
	MaxTimeout     time.Duration

	DefaultTimeout     time.Duration
	DefaultRunInterval time.Duration

	MaxCheckParallelism uint8
}

// DefaultOpts
func DefaultOpts() Opts {
	return Opts{
		MinRunInterval: MinRunInterval,
		MaxTimeout:     MaxTimeout,

		DefaultRunInterval: DefaultRunInterval,
		DefaultTimeout:     DefaultTimeout,

		MaxCheckParallelism: MaxCheckParallelism,
	}
}

type service struct {
	Opts

	checks []RegisteredCheck

	stop                chan struct{}
	register            chan registerRequest
	getRegisteredChecks chan getRegisteredChecksRequest

	subscribeForRegisteredChecks     chan subscribeForRegisteredChecksRequest
	subscriptionsForRegisteredChecks map[chan<- RegisteredCheck]struct{}

	subscribeForCheckResults     chan subscribeForCheckResults
	subscriptionsForCheckResults map[chan<- Result]struct{}

	// TODO: refactor into its own service
	// to protect the application and system from the health checks themselves we want to limit the number of health checks
	// that are allowed to run concurrently
	runSemaphore chan struct{}
	results      chan Result
	runResults   map[string]Result
}

func newService(opts Opts) *service {
	runSemaphore := make(chan struct{}, opts.MaxCheckParallelism)
	var i uint8
	for ; i < opts.MaxCheckParallelism; i++ {
		runSemaphore <- struct{}{}
	}

	return &service{
		stop:                make(chan struct{}),
		register:            make(chan registerRequest),
		getRegisteredChecks: make(chan getRegisteredChecksRequest),

		subscribeForRegisteredChecks:     make(chan subscribeForRegisteredChecksRequest),
		subscriptionsForRegisteredChecks: make(map[chan<- RegisteredCheck]struct{}),

		subscribeForCheckResults:     make(chan subscribeForCheckResults),
		subscriptionsForCheckResults: make(map[chan<- Result]struct{}),

		runSemaphore: runSemaphore,
		results:      make(chan Result),
		runResults:   make(map[string]Result),

		Opts: opts,
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
			// send reply
			go func() {
				defer close(req.reply)
				select {
				case <-s.stop:
				case req.reply <- err:
				}
			}()
		case result := <-s.results:
			s.runResults[result.HealthCheckID] = result
			s.publishResult(result)
		case req := <-s.getRegisteredChecks:
			s.SendRegisteredChecks(req)
		case req := <-s.subscribeForRegisteredChecks:
			s.SubscribeForRegisteredChecks(req)
		case req := <-s.subscribeForCheckResults:
			s.SubscribeForCheckResults(req)
		}
	}
}

func (s *service) publishResult(result Result) {
	for ch := range s.subscriptionsForCheckResults {
		go func(ch chan<- Result) {
			select {
			case <-s.stop:
			case ch <- result:
			}
		}(ch)
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
	for ch := range s.subscriptionsForRegisteredChecks {
		close(ch)
	}

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
			<-s.runSemaphore
			defer func() {
				s.runSemaphore <- struct{}{}
			}()
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
			opts.Timeout = s.DefaultTimeout
		}
		if opts.RunInterval == time.Duration(0) {
			opts.RunInterval = s.DefaultRunInterval
		}

		return opts
	}

	ValidateOpts := func(opts CheckerOpts) error {
		var err error
		if opts.RunInterval < s.MinRunInterval {
			err = ErrRunIntervalTooFrequent
		}
		if opts.Timeout > s.MaxTimeout {
			err = multierr.Append(err, ErrRunTimeoutTooHigh)
		}
		return err
	}

	SendRegisteredCheckToSubscribers := func(check RegisteredCheck) {
		for ch := range s.subscriptionsForRegisteredChecks {
			go func(ch chan<- RegisteredCheck) {
				select {
				case <-s.stop:
				case ch <- check:
				}
			}(ch)
		}
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
	SendRegisteredCheckToSubscribers(registeredCheck)

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
		defer close(req.reply)
		select {
		case <-s.stop:
		case req.reply <- checks:
		}
	}()
}

type subscribeForRegisteredChecksRequest struct {
	reply chan chan RegisteredCheck
}

func (s *service) SubscribeForRegisteredChecks(req subscribeForRegisteredChecksRequest) {
	ch := make(chan RegisteredCheck)
	s.subscriptionsForRegisteredChecks[ch] = struct{}{}

	go func() {
		defer close(req.reply)
		select {
		case <-s.stop:
		case req.reply <- ch:
		}
	}()
}

type subscribeForCheckResults struct {
	reply chan chan Result
}

func (s *service) SubscribeForCheckResults(req subscribeForCheckResults) {
	ch := make(chan Result)
	s.subscriptionsForCheckResults[ch] = struct{}{}

	go func() {
		defer close(req.reply)
		select {
		case <-s.stop:
		case req.reply <- ch:
		}
	}()
}
