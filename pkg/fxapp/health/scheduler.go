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
	"log"
	"time"
)

// Scheduler is used to schedule health checks to run.
//
// Design
// - Only 1 health check will be allowed to run at a time to prevent application / system overload.
// - The health check's next run is scheduled when the health check run is complete.
// - As health checks are registered, then they will get scheduled to run.
// - Once the scheduler is stopped, it cannot be restarted
type Scheduler interface {
	Running() <-chan struct{}

	// StopAsync triggers shutdown async
	StopAsync()

	// Stopping returns true is StopAsync has previously been invoked
	Stopping() bool

	// Done returns a channel that is used to signal that the scheduler shutdown has completed
	Done() <-chan struct{}

	// HealthCheckCount returns the number of health checks that are currently scheduled
	HealthCheckCount() uint
}

type scheduler struct {
	Registry

	running, shutdown, done chan struct{}

	healthCheckCount                             uint
	incHealthCheckCounter, decHealthCheckCounter chan struct{}
	getHealthCheckCount                          chan chan uint

	results chan Result
}

func StartScheduler(registry Registry) Scheduler {
	s := &scheduler{
		Registry: registry,

		running:  make(chan struct{}),
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),

		incHealthCheckCounter: make(chan struct{}),
		decHealthCheckCounter: make(chan struct{}),
		getHealthCheckCount:   make(chan chan uint),

		results: make(chan Result),
	}

	schedule := func(check Check) {
		defer func() {
			s.decHealthCheckCounter <- struct{}{}
		}()
		s.incHealthCheckCounter <- struct{}{}
		for {
			timer := time.After(check.RunInterval())
			select {
			case <-s.shutdown:
				return
			case <-timer:
				s.results <- check.Run()
			}
		}
	}

	for _, check := range registry.HealthChecks(nil) {
		go schedule(check)
	}

	go func() {
		close(s.running)
		healthcheckRegistered := make(chan Check)
		s.Registry.Subscribe(healthcheckRegistered)
		for {
			select {
			case check := <-healthcheckRegistered:
				go schedule(check)
			case <-s.incHealthCheckCounter:
				s.healthCheckCount++
			case <-s.decHealthCheckCounter:
				s.healthCheckCount--
				if s.Stopping() && s.healthCheckCount == 0 {
					close(s.done)
					return
				}
			case reply := <-s.getHealthCheckCount:
				reply <- s.healthCheckCount
			case result := <-s.results:
				// TODO: publish result to subscribers
				log.Print(result)
			}
		}
	}()

	return s
}

func (s *scheduler) Running() <-chan struct{} {
	return s.running
}

func (s *scheduler) StopAsync() {
	select {
	case <-s.shutdown:
	default:
		close(s.shutdown)
	}
}

func (s *scheduler) Stopping() bool {
	select {
	case <-s.shutdown:
		return true
	default:
		return false
	}
}

func (s *scheduler) Done() <-chan struct{} {
	return s.done
}

func (s *scheduler) HealthCheckCount() uint {
	select {
	case <-s.done:
		return 0
	default:
		count := make(chan uint)
		select {
		case <-s.done:
			return 0
		case s.getHealthCheckCount <- count: // send request
			select {
			case <-s.done:
				return 0
			case n := <-count: // wait for response
				return n
			}
		}
	}
}
