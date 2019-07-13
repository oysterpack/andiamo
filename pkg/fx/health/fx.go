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
	"go.uber.org/fx"
)

// ModuleWithDefaults provides the fx Module for the health module
func ModuleWithDefaults() fx.Option {
	return Module(DefaultOpts())
}

// Module provides the fx Module for the health module
func Module(opts Opts) fx.Option {
	return fx.Options(
		fx.Provide(
			startService(opts),

			provideRegisterFunc,

			provideRegisteredChecksFunc,
			provideCheckResultsFunc,

			provideSubscribeForRegisteredChecks,
			provideSubscribeForCheckResults,
		),
	)
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
	return func(check Check, opts CheckerOpts, checker Checker) error {
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
	return func() <-chan []Result {
		reply := make(chan []Result, 1) // a chan buf size 1 decouples the producer from the consumer
		go func() {
			select {
			case <-s.stop:
				close(reply)
			case s.getCheckResults <- reply:
			}
		}()
		return reply
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
	return func() CheckResultsSubscription {
		closedChan := func() CheckResultsSubscription {
			ch := make(chan Result)
			close(ch)
			return CheckResultsSubscription{ch}
		}

		reply := make(chan chan Result, 1) // a chan buf size 1 decouples the producer from the consumer

		select {
		case <-s.stop:
			return closedChan()
		case s.subscribeForCheckResults <- subscribeForCheckResults{reply}:
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
