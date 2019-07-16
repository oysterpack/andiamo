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
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

// Module provides the fx Module for the health module
func Module(opts Opts) fx.Option {
	options := []fx.Option{
		fx.Provide(
			startService(opts),

			provideRegisterFunc,

			provideRegisteredChecksFunc,
			provideCheckResultsFunc,

			provideSubscribeForRegisteredChecks,
			provideSubscribeForCheckResults,

			provideOverallHealth,
		),
	}
	if opts.FailFastOnStartup {
		options = append(options, fx.Invoke(checkHealthOnStart))
	}
	return fx.Options(options...)
}

func startService(svcOpts Opts) func(lc fx.Lifecycle) *service {
	s := newService(svcOpts)
	return func(lc fx.Lifecycle) *service {
		go s.run()
		lc.Append(fx.Hook{
			OnStop: func(context.Context) error {
				s.TriggerShutdown()
				return nil
			},
		})
		return s
	}
}

func provideRegisterFunc(s *service) Register {
	return func(check Check, opts CheckerOpts, checker func() (Status, error)) error {
		reply := make(chan error, 1) // a chan buf size 1 decouples the producer from the consumer
		req := registerRequest{
			check:   check,
			opts:    opts,
			checker: checker,
			reply:   reply,
		}
		// send service request
		select {
		case <-s.stop:
			return ErrServiceNotRunning
		case s.register <- req:
		}

		// receive service reply
		select {
		case <-s.stop:
			return ErrServiceNotRunning
		case err := <-reply:
			return err
		}
	}
}

func provideRegisteredChecksFunc(s *service) RegisteredChecks {
	return func() <-chan []RegisteredCheck {
		reply := make(chan []RegisteredCheck, 1) // a chan buf size 1 decouples the producer from the consumer
		go func() {
			select {
			case <-s.stop:
				close(reply)
			case s.getRegisteredChecks <- reply:
			}
		}()
		return reply
	}
}

func provideCheckResultsFunc(s *service) CheckResults {
	return func(filter func(result Result) bool) <-chan []Result {
		req := checkResultsRequest{
			reply:  make(chan []Result, 1), // a chan buf size 1 decouples the producer from the consumer
			filter: filter,
		}
		go func() {
			select {
			case <-s.stop:
				close(req.reply)
			case s.getCheckResults <- req:
			}
		}()
		return req.reply
	}
}

func provideSubscribeForRegisteredChecks(s *service) SubscribeForRegisteredChecks {
	return func() RegisteredCheckSubscription {
		closedChan := func() RegisteredCheckSubscription {
			ch := make(chan RegisteredCheck)
			close(ch)
			return RegisteredCheckSubscription{ch}
		}

		reply := make(chan chan RegisteredCheck, 1) // a chan buf size 1 decouples the producer from the consumer

		select {
		case <-s.stop:
			return closedChan()
		case s.subscribeForRegisteredChecks <- subscribeForRegisteredChecksRequest{reply}:
			select {
			case <-s.stop:
				return closedChan()
			case ch, ok := <-reply:
				if ok {
					return RegisteredCheckSubscription{ch}
				}
				return closedChan()
			}
		}
	}
}

func provideSubscribeForCheckResults(s *service) SubscribeForCheckResults {
	return func(filter func(result Result) bool) CheckResultsSubscription {
		closedChan := func() CheckResultsSubscription {
			ch := make(chan Result)
			close(ch)
			return CheckResultsSubscription{ch}
		}

		reply := make(chan chan Result, 1) // a chan buf size 1 decouples the producer from the consumer

		select {
		case <-s.stop:
			return closedChan()
		case s.subscribeForCheckResults <- subscribeForCheckResults{reply, filter}:
			select {
			case <-s.stop:
				return closedChan()
			case ch, ok := <-reply:
				if ok {
					return CheckResultsSubscription{ch}
				}
				return closedChan()
			}
		}
	}
}

func checkHealthOnStart(lc fx.Lifecycle, checks RegisteredChecks, checkResults CheckResults) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			registeredCheckChan := checks()
			select {
			case <-ctx.Done():
				return ErrContextTimout
			case registeredChecks, ok := <-registeredCheckChan:
				if !ok {
					return errors.New("failed to get registered health checks because the channel is closed")
				}
				if len(registeredChecks) == 0 {
					return nil
				}

				// Health checks are run immediately in the background after they are registered. Thus, get all of the cached
				// green results. If there is no cached green result, then run the health check now. If any health check fails,
				// then return the error, which will cause the app start up to fail.
				select {
				case <-ctx.Done():
					return ErrContextTimout
				case results, ok := <-checkResults(func(result Result) bool { return result.Status == Green }):
					if !ok {
						return errors.New("failed to get health check results because the channel is closed")
					}
					// if any health checks are not green, then run them now. If any fail, i.e., not green then app start up will fail
				RegisteredChecks:
					for _, registeredCheck := range registeredChecks {
						for _, result := range results {
							if result.ID == registeredCheck.ID {
								continue RegisteredChecks
							}
						}
						if result := registeredCheck.Checker(); result.Status != Green {
							return result.Err
						}
					}
				}
			}

			return nil
		},
	})
}

func provideOverallHealth(s *service) OverallHealth {
	return func() Status {
		reply := make(chan Status, 1)
		select {
		case <-s.stop:
			return Red
		case s.getOverallHealth <- reply:
			select {
			case <-s.stop:
				return Red
			case status := <-reply:
				return status
			}
		}
	}
}
