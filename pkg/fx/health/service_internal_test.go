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
	"go.uber.org/fx"
	"runtime"
	"testing"
	"time"
)

func runApp(app *fx.App, shutdowner fx.Shutdowner, funcs ...func()) {
	done := make(chan struct{})
	defer func() {
	ShutdownLoop:
		for {
			select {
			case <-done:
				break ShutdownLoop
			default:
				shutdowner.Shutdown()
				runtime.Gosched()
			}
		}
	}()

	running := make(chan struct{})
	go func() {
		defer close(done)
		close(running)
		app.Run()
	}()
	<-running
	runtime.Gosched()
	for _, f := range funcs {
		f()
	}

}

func TestService_TriggerShutdown(t *testing.T) {
	t.Parallel()

	t.Run("trigger shutdown is idempotent", func(t *testing.T) {
		t.Parallel()
		s := newService(DefaultOpts())
		go s.run()
		s.TriggerShutdown()
		// calling it again should have no effect
		s.TriggerShutdown()
		<-s.stop
	})
}

func TestService_RunningScheduledHealthChecks(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	var Foo = Check{
		ID:           "01DFGJ4A2GBTSQR11YYMV0N086",
		Description:  "Foo",
		RedImpact:    "App is unusable",
		YellowImpact: "App performance degradation",
		Tags:         []string{Database, MongoDB},
	}

	t.Run("health check times out", func(t *testing.T) {
		t.Parallel()

		opts := DefaultOpts()
		opts.MinRunInterval = time.Nanosecond

		var shutdowner fx.Shutdowner
		var resultsSubscription CheckResultsSubscription
		app := fx.New(
			options(opts),
			fx.Invoke(
				func(subscribe SubscribeForCheckResults) {
					resultsSubscription = subscribe()
				},
				func(register Register) error {
					checkerOpts := CheckerOpts{
						Timeout: time.Nanosecond,
					}
					return register(Foo, checkerOpts, func() error {
						time.Sleep(time.Microsecond)
						return nil
					})
				},
				// verify that the health check timeout is 1 ns
				func(registeredChecks RegisteredChecks) {
					registeredCheck := <-registeredChecks(nil)
					t.Log(registeredCheck)
					if registeredCheck[0].Timeout != time.Nanosecond {
						t.Errorf("*** timeout should be 1 ns: %v", registeredCheck)
					}
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() != nil {
			t.Errorf("*** app initialization failed : %v", app.Err())
			return
		}

		runApp(app, shutdowner, func() {
			result := <-resultsSubscription.Chan()
			t.Log(result)
			if result.Status != Red {
				t.Errorf("*** health check should have timed out, which is considered a Red failure")
			}
			if result.error != ErrTimeout {
				t.Errorf("*** error should have been timeout : %v", result.error)
			}
		})

	})
}
