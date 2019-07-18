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

package app

import (
	"context"
	"github.com/oysterpack/andiamo/pkg/eventlog"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"sync/atomic"
	"time"
)

// New initializes a new fx App with the following augmentations:
//	- app life cycle events are logged:
//	  - InitializedEvent
//	  - StartingEvent
//	  - StartedEvent
//	  - StoppingEvent
//	  - StoppedEvent
//	- app lifecycle error events:
//	  - InitFailedEvent
//	  - StartFailedEvent
//    - StopFailedEvent
func New(opts Opts, options ...fx.Option) *fx.App {
	appOptions := make([]fx.Option, 0, len(options)+3)

	var startUnixTime, startNanosecond int64
	appOptions = append(appOptions,
		fx.Invoke(func(lc fx.Lifecycle, log Logger) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					now := time.Now()
					atomic.StoreInt64(&startUnixTime, now.Unix())
					atomic.StoreInt64(&startNanosecond, int64(now.Nanosecond()))
					log(StartingEvent, zerolog.NoLevel)(nil, "app is starting")
					return nil
				},
				OnStop: func(context.Context) error {
					dur := duration(time.Since(time.Unix(atomic.LoadInt64(&startUnixTime), atomic.LoadInt64(&startNanosecond))))
					log(StoppedEvent, zerolog.NoLevel)(dur, "app is stopped")
					return nil
				},
			})
		}),
	)

	appOptions = append(appOptions, Module(opts))
	appOptions = append(appOptions, options...)
	appOptions = append(appOptions,
		fx.Invoke(
			func(lc fx.Lifecycle, log Logger) {
				lc.Append(fx.Hook{
					OnStart: func(context context.Context) error {
						dur := duration(time.Since(time.Unix(atomic.LoadInt64(&startUnixTime), atomic.LoadInt64(&startNanosecond))))
						log(StartedEvent, zerolog.NoLevel)(dur, "app is started")
						return nil
					},
					OnStop: func(context.Context) error {
						now := time.Now()
						atomic.StoreInt64(&startUnixTime, now.Unix())
						atomic.StoreInt64(&startNanosecond, int64(now.Nanosecond()))
						log(StoppingEvent, zerolog.NoLevel)(nil, "app is stopping")
						return nil
					},
				})
			},
			func(dotgraph fx.DotGraph, log Logger) {
				log(InitializedEvent, zerolog.NoLevel)(appInfo{dotgraph}, "app is initialized")
			},
		),
	)

	return fx.New(appOptions...)
}

// Go runs the app on a background goroutine.
// It returns a shutdowner, which can be used to trigger application shutdown.
// Once shutdown is triggered, the done channel can be used to wait until the app shutdown is complete.
// The done channel will return an error if any error occurs during app initialization, startup, or shutdown.
// An error is returned if the app initialization failed.
func Go(opts Opts, options ...fx.Option) (shutdowner fx.Shutdowner, done chan error, err error) {
	done = make(chan error, 1)

	starting := make(chan struct{})
	publishAppStartingEvent := func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				close(starting)
				return nil
			},
		})
	}
	appOptions := make([]fx.Option, 0, len(options)+2)
	var logger Logger
	appOptions = append(appOptions,
		fx.Invoke(publishAppStartingEvent),
		fx.Populate(&shutdowner, &logger),
	)
	app := New(opts, append(appOptions, options...)...)

	if err := app.Err(); err != nil {
		logger(InitFailedEvent, zerolog.ErrorLevel)(eventlog.NewError(err), "app failed to initialize")
		done <- err
		close(done)
		return nil, done, err
	}

	go func() {
		defer close(done)

		// start the app
		{
			ctx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
			defer cancel()
			err := app.Start(ctx)
			if err != nil {
				logger(StartFailedEvent, zerolog.ErrorLevel)(eventlog.NewError(err), "app failed to start")
				done <- err
				return
			}
		}

		{
			<-app.Done() // wait for stop signal
			// stop the app
			ctx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()
			err := app.Stop(ctx)
			if err != nil {
				logger(StopFailedEvent, zerolog.ErrorLevel)(eventlog.NewError(err), "app stopped with an error")
				done <- err
				return
			}
		}
	}()

	<-starting
	return shutdowner, done, nil
}
