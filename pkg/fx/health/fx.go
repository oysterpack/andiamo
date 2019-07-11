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

// Options provides the fx options for the health module
func Options() fx.Option {
	return fx.Options(
		fx.Provide(
			startService,
			provideRegisterFunc,
			provideRegisteredChecksFunc,
		),
	)
}

func startService(lc fx.Lifecycle) *service {
	s := newService()
	go s.run()
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			s.TriggerShutdown()
			return nil
		},
	})
	return s
}

func provideRegisterFunc(s *service) Register {
	return func(check Check, opts CheckerOpts, checker Checker) error {
		reply := make(chan error)
		req := registerRequest{
			check:   check,
			opts:    opts,
			checker: checker,
			reply:   reply,
		}
		select {
		case <-s.stop:
			return ErrServiceNotRunning
		case s.register <- req:
		}

		select {
		case <-s.stop:
			return ErrServiceNotRunning
		case err := <-reply:
			return err
		}
	}
}

func provideRegisteredChecksFunc(s *service) RegisteredChecks {
	return func(filter func(c Check, opts CheckerOpts) bool) <-chan []RegisteredCheck {
		reply := make(chan []RegisteredCheck)

		go func() {
			select {
			case <-s.stop:
				close(reply)
			case s.getRegisteredChecks <- getRegisteredChecksRequest{filter, reply}:
			}
		}()

		return reply
	}
}
