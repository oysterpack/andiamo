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
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"strings"
	"time"
)

// Opts are used to configure the fx module.
type Opts struct {
	MinRunInterval time.Duration
	MaxTimeout     time.Duration

	DefaultTimeout     time.Duration
	DefaultRunInterval time.Duration

	MaxCheckParallelism uint8
}

// DefaultOpts constructs a new Opts using recommended default values.
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
	getRegisteredChecks chan chan<- []RegisteredCheck
	getCheckResults     chan checkResultsRequest

	subscribeForRegisteredChecks     chan subscribeForRegisteredChecksRequest
	subscriptionsForRegisteredChecks map[chan<- RegisteredCheck]struct{}

	subscribeForCheckResults     chan subscribeForCheckResults
	subscriptionsForCheckResults map[chan<- Result]func(result Result) bool

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
		getRegisteredChecks: make(chan chan<- []RegisteredCheck),
		getCheckResults:     make(chan checkResultsRequest),

		subscribeForRegisteredChecks:     make(chan subscribeForRegisteredChecksRequest),
		subscriptionsForRegisteredChecks: make(map[chan<- RegisteredCheck]struct{}),

		subscribeForCheckResults:     make(chan subscribeForCheckResults),
		subscriptionsForCheckResults: make(map[chan<- Result]func(result Result) bool),

		runSemaphore: runSemaphore,
		results:      make(chan Result),
		runResults:   make(map[string]Result),

		Opts: opts,
	}
}

func (s *service) run() {
	for {
		select {
		case <-s.stop:
			s.Stop()
			return
		case req := <-s.register:
			err := s.Register(req)
			s.sendError(req.reply, err)
		case result := <-s.results:
			s.runResults[result.ID] = result
			s.publishResult(result)
		case replyChan := <-s.getRegisteredChecks:
			s.SendRegisteredChecks(replyChan)
		case replyChan := <-s.getCheckResults:
			s.SendCheckResults(replyChan)
		case req := <-s.subscribeForRegisteredChecks:
			s.SubscribeForRegisteredChecks(req)
		case req := <-s.subscribeForCheckResults:
			s.SubscribeForCheckResults(req)
		}
	}
}

func (s *service) sendError(ch chan<- error, err error) {
	defer close(ch)

	if err == nil {
		return
	}

	select {
	case <-s.stop:
	case ch <- err:
	}
}

func (s *service) publishResult(result Result) {
	for ch, filter := range s.subscriptionsForCheckResults {
		if filter(result) {
			go func(ch chan<- Result) {
				select {
				case <-s.stop:
				case ch <- result:
				}
			}(ch)
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
	for ch := range s.subscriptionsForRegisteredChecks {
		close(ch)
	}

}

type registerRequest struct {
	check   Check
	opts    CheckerOpts
	checker func() (Status, error)

	reply chan<- error
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

	WithTimeout := func(id string, check func() (Status, error), timeout time.Duration) Checker {
		return func() Result {
			reply := make(chan Result, 1)
			timer := time.After(timeout)
			go func() {
				start := time.Now()
				status, err := check()
				duration := time.Since(start)
				reply <- Result{
					ID: id,

					Status: status,
					Err:    err,

					Time:     start,
					Duration: duration,
				}
			}()

			select {
			case <-timer:
				return Result{
					ID: id,

					Status: Red,
					Err:    ErrTimeout,

					Time:     time.Now().Add(timeout * -1),
					Duration: timeout,
				}
			case result := <-reply:
				return result
			}
		}
	}

	Schedule := func(id string, check Checker, interval time.Duration) {
		run := func() {
			<-s.runSemaphore
			defer func() {
				s.runSemaphore <- struct{}{}
			}()
			go func() {
				select {
				case <-s.stop:
				case s.results <- check():
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
		return multierr.Append(fmt.Errorf("invalid health check: %#v", check), err)
	}

	if req.checker == nil {
		return multierr.Append(errors.New(check.ID), ErrNilChecker)
	}

	opts := ApplyDefaultOpts(req.opts)
	if err := ValidateOpts(opts); err != nil {
		return multierr.Append(fmt.Errorf("invalid health checker opts: %s : %#v", check.ID, opts), err)
	}

	if s.RegisteredCheck(check.ID) != nil {
		return fmt.Errorf("health check is already registered: %s", check.ID)
	}

	registeredCheck := RegisteredCheck{
		Check:       check,
		CheckerOpts: opts,
		Checker:     WithTimeout(check.ID, req.checker, opts.Timeout),
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

type checkResultsRequest struct {
	reply  chan []Result
	filter func(result Result) bool
}

func (s *service) SendCheckResults(req checkResultsRequest) {
	var results []Result
	if req.filter == nil {
		results = make([]Result, 0, len(s.runResults))
		for _, result := range s.runResults {
			results = append(results, result)
		}
	} else {
		for _, result := range s.runResults {
			if req.filter(result) {
				results = append(results, result)
			}
		}
	}

	defer close(req.reply)
	req.reply <- results
}

func (s *service) SendRegisteredChecks(reply chan<- []RegisteredCheck) {
	checks := make([]RegisteredCheck, len(s.checks))
	copy(checks, s.checks)

	defer close(reply)
	reply <- checks
}

type subscribeForRegisteredChecksRequest struct {
	reply chan chan RegisteredCheck
}

func (s *service) SubscribeForRegisteredChecks(req subscribeForRegisteredChecksRequest) {
	ch := make(chan RegisteredCheck)
	s.subscriptionsForRegisteredChecks[ch] = struct{}{}

	defer close(req.reply)
	req.reply <- ch
}

type subscribeForCheckResults struct {
	reply  chan chan Result
	filter func(result Result) bool
}

func (s *service) SubscribeForCheckResults(req subscribeForCheckResults) {
	ch := make(chan Result)
	if req.filter != nil {
		s.subscriptionsForCheckResults[ch] = req.filter
	} else {
		s.subscriptionsForCheckResults[ch] = func(Result) bool { return true }
	}

	defer close(req.reply)
	req.reply <- ch
}
