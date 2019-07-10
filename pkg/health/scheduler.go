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
	"sync"
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
	// StopAsync triggers shutdown async
	StopAsync()

	// Stopping returns true is StopAsync has previously been invoked
	Stopping() bool

	// Done returns a channel that is used to signal that the scheduler shutdown has completed
	Done() <-chan struct{}

	// HealthCheckCount returns the number of health checks that are currently scheduled
	HealthCheckCount() uint

	// Subscribe is used to subscribe to health check results.
	//
	// If the scheduler has been shutdown, then a closed channel will be returned.
	// As soon as the scheduler shutdown is complete, then no more health check results will be published - even if they
	// are in flight.
	Subscribe(filter func(check Check) bool) <-chan Result

	// Results returns the latest health check results, i.e., from the last time the health checks ran
	// If the scheduler is shutdown, then nil is returned.
	Results(filter func(result Result) bool) <-chan []Result
}

type scheduler struct {
	Registry

	shutdown chan struct{} // used to trigger the scheduler to shutdown
	done     chan struct{} // used to signal when the scheduler shutdown is complete

	healthCheckCount                             uint           // number of health checks that have been scheduled
	incHealthCheckCounter, decHealthCheckCounter chan struct{}  // used to update the health check counter
	getHealthCheckCount                          chan chan uint // used to get the current scheduled health check count

	results          chan runResult        // used to publish health check run results
	subscribe        chan subscribeRequest // used to subscribe to health check run results
	getLatestResults chan getLatestResultsRequest

	runLock sync.Mutex // used to run only 1 health check at a time
}

type runResult struct {
	Check
	Result
}

// StartScheduler starts up a new health check scheduler for the specified registry.
func StartScheduler(registry Registry) Scheduler {
	s := &scheduler{
		Registry: registry,

		shutdown: make(chan struct{}),
		done:     make(chan struct{}),

		incHealthCheckCounter: make(chan struct{}),
		decHealthCheckCounter: make(chan struct{}),
		getHealthCheckCount:   make(chan chan uint),

		results:          make(chan runResult),
		subscribe:        make(chan subscribeRequest),
		getLatestResults: make(chan getLatestResultsRequest),
	}

	subscriptions := make(map[chan Result]func(Check) bool)
	publishResult := func(result runResult) {
		for ch, filter := range subscriptions {
			if filter == nil || filter(result.Check) {
				go func(result Result, ch chan<- Result) {
					select {
					case ch <- result:
					case <-s.shutdown:
					}
				}(result.Result, ch)
			}
		}
	}

	runHealthCheck := func(check Check) Result {
		s.runLock.Lock()
		defer s.runLock.Unlock()
		return check.Run()
	}

	var latestResults []Result
	updateLatestResults := func(newResult Result) {
		for i, result := range latestResults {
			if result.HealthCheckID() == newResult.HealthCheckID() {
				latestResults[i] = newResult
				return
			}
		}
		latestResults = append(latestResults, newResult)
	}

	schedule := func(check Check) {
		defer func() {
			s.decHealthCheckCounter <- struct{}{}
		}()
		s.incHealthCheckCounter <- struct{}{}

		// run the health check immediately
		select {
		case <-s.shutdown:
			return
		case s.results <- runResult{check, runHealthCheck(check)}:
		}

		// schedule the health check to run
		for {
			timer := time.After(check.RunInterval())
			select {
			case <-s.shutdown:
				return
			case <-timer:
				result := runResult{check, runHealthCheck(check)}
				select {
				case <-s.shutdown:
					return
				case s.results <- result:
				}
			}
		}
	}

	for _, check := range registry.HealthChecks(nil) {
		go schedule(check)
	}

	go func() {
		// subscribe to health check registration events
		// - when a health check is registered, then schedule it
		healthcheckRegistered := s.Registry.Subscribe()

		defer close(s.done)

		for {
			select {
			case <-s.shutdown:
				return
			case check := <-healthcheckRegistered:
				go schedule(check)
			case <-s.incHealthCheckCounter:
				// health check has been scheduled
				s.healthCheckCount++
			case <-s.decHealthCheckCounter:
				// health check has been unscheduled
				s.healthCheckCount--
			case reply := <-s.getHealthCheckCount:
				select {
				case <-s.shutdown:
				case reply <- s.healthCheckCount:
				}
			case result := <-s.results:
				updateLatestResults(result.Result)
				publishResult(result)
			case req := <-s.subscribe:
				ch := make(chan Result)
				subscriptions[ch] = req.filter
				select {
				case <-s.shutdown:
				case req.reply <- ch:
				}
			case req := <-s.getLatestResults:
				if req.filter == nil {
					results := make([]Result, len(latestResults))
					copy(results, latestResults)
					select {
					case <-s.shutdown:
					case req.reply <- results:
					}
					continue
				}
				var results []Result
				for _, result := range latestResults {
					if req.filter(result) {
						results = append(results, result)
					}
				}
				select {
				case <-s.shutdown:
				case req.reply <- results:
				}
			}

		}
	}()

	return s
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
	count := make(chan uint)
	select {
	case <-s.shutdown:
		return 0
	case s.getHealthCheckCount <- count: // send request
		select {
		case <-s.shutdown:
			return 0
		case n := <-count: // wait for response
			return n
		}
	}
}

type subscribeRequest struct {
	filter func(Check) bool
	reply  chan chan Result
}

func (s *scheduler) Subscribe(filter func(Check) bool) <-chan Result {
	req := subscribeRequest{
		filter,
		make(chan chan Result),
	}

	closedChan := func() chan Result {
		ch := make(chan Result)
		close(ch)
		return ch
	}

	select {
	case <-s.shutdown:
		return closedChan()
	case s.subscribe <- req:
		select {
		case <-s.shutdown:
			return closedChan()
		case ch := <-req.reply:
			return ch
		}
	}
}

type getLatestResultsRequest struct {
	filter func(result Result) bool
	reply  chan []Result
}

func (s *scheduler) Results(filter func(result Result) bool) <-chan []Result {
	req := getLatestResultsRequest{
		filter,
		make(chan []Result),
	}

	select {
	case <-s.shutdown:
		close(req.reply)
		return req.reply
	case s.getLatestResults <- req:
		return req.reply
	}
}
