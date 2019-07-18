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
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"sync/atomic"
	"time"
)

// New initializes a new fx App with the following augmentations:
//	- app life cycle events are logged
func New(options ...fx.Option) *fx.App {
	appOptions := make([]fx.Option, 0, len(options)+2)

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
