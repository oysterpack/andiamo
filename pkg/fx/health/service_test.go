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

package health_test

import (
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/fx/health"
	"github.com/oysterpack/partire-k8s/pkg/ulids"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"runtime"
	"strings"
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

func TestService_Register(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	var Foo = health.Check{
		ID:           "01DFGJ4A2GBTSQR11YYMV0N086",
		Description:  "Foo",
		RedImpact:    "App is unusable",
		YellowImpact: "App performance degradation",
		Tags:         []string{Database, MongoDB},
	}

	t.Run("register valid health check", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.ModuleWithDefaults(),
			fx.Invoke(
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{}, func() error {
						return nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() != nil {
			t.Errorf("*** app initialization failed : %v", app.Err())
			return
		}

		runApp(app, shutdowner)
	})

	t.Run("register invalid health check - no fields set", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.ModuleWithDefaults(),
			fx.Invoke(
				func(register health.Register) error {
					InvalidHealthCheck := health.Check{}
					return register(InvalidHealthCheck, health.CheckerOpts{}, func() error {
						return nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() == nil {
			t.Error("*** app initialization should have failed")
			return
		}
		t.Log(app.Err())
	})

	t.Run("register invalid health check - tag is not valid ULID", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.ModuleWithDefaults(),
			fx.Invoke(
				func(register health.Register) error {
					InvalidHealthCheck := health.Check{
						ID:           "01DFGJ4A2GBTSQR11YYMV0N086",
						Description:  "Foo",
						RedImpact:    "App is unusable",
						YellowImpact: "App performance degradation",
						Tags:         []string{Database, "INVALID"},
					}
					return register(InvalidHealthCheck, health.CheckerOpts{}, func() error {
						return nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() == nil {
			t.Error("*** app initialization should have failed")
			return
		}
		t.Log(app.Err())

	})

}

func TestService_SendRegisteredChecks(t *testing.T) {
	t.Parallel()

	// Given Check fields are padded with whitespace
	const (
		Database = "  01DFGP2MJB9B8BMWA6Q2H4JD9Z  "
		MongoDB  = "  01DFGP3TS31D016DHS9415JFBB  "
	)

	var Foo = health.Check{
		ID:           "  01DFGJ4A2GBTSQR11YYMV0N086  ",
		Description:  "  Foo  ",
		RedImpact:    "  App is unusable  ",
		YellowImpact: "  App performance degradation  ",
		Tags:         []string{Database, MongoDB},
	}

	t.Run("register valid health check", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.ModuleWithDefaults(),
			fx.Invoke(
				func(register health.Register) error {
					// And CheckerOpts were not specified
					return register(Foo, health.CheckerOpts{}, func() error {
						return nil
					})
				},
				func(getRegisteredChecks health.RegisteredChecks) error {
					checks := <-getRegisteredChecks(nil)
					switch len(checks) {
					case 0:
						return errors.New("*** no registered health checks were returned")
					case 1:
						check := checks[0]
						t.Logf("%#v", check)

						// Then all Check fields whould be trimmed
						if check.ID != strings.TrimSpace(Foo.ID) {
							return errors.New("*** ID was not trimmed")
						}
						if check.Description != strings.TrimSpace(Foo.Description) {
							return errors.New("*** Description was not trimmed")
						}
						if check.RedImpact != strings.TrimSpace(Foo.RedImpact) {
							return errors.New("*** RedImpact was not trimmed")
						}
						if check.YellowImpact != strings.TrimSpace(Foo.YellowImpact) {
							return errors.New("*** YellowImpact was not trimmed")
						}
						for _, tag := range check.Tags {
							ulid.MustParse(tag)
						}

						// And CheckerOpts have default values
						if check.Timeout != health.DefaultTimeout {
							return errors.New("*** timeout did not match default")
						}
						if check.RunInterval != health.DefaultRunInterval {
							return errors.New("*** RunInterval did not match default")
						}

						return nil
					default:
						return errors.New("*** no registered health checks were returned")
					}
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() != nil {
			t.Errorf("*** app initialization failed : %v", app.Err())
			return
		}

		runApp(app, shutdowner)
	})

	t.Run("register 10 health checks", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.ModuleWithDefaults(),
			fx.Invoke(
				func(register health.Register) error {
					for i := 0; i < 10; i++ {
						check := health.Check{
							ID:           ulids.MustNew().String(),
							Description:  fmt.Sprintf("Desc %d", i),
							RedImpact:    fmt.Sprintf("Red %d", i),
							YellowImpact: fmt.Sprintf("Yellow %d", i),
							Tags:         []string{Database, MongoDB},
						}
						register(check, health.CheckerOpts{}, func() error {
							return nil
						})
					}

					return nil
				},
				func(getRegisteredChecks health.RegisteredChecks) error {
					checks := <-getRegisteredChecks(nil)
					switch len(checks) {
					case 0:
						return errors.New("*** no registered health checks were returned")
					case 10:
						for _, check := range checks {
							t.Logf("%#v", check)
						}
						return nil
					default:
						return errors.New("*** no registered health checks were returned")
					}
				},
			),
			fx.Populate(&shutdowner),
		)

		if app.Err() != nil {
			t.Errorf("*** app initialization failed : %v", app.Err())
			return
		}

		runApp(app, shutdowner)
	})
}

func TestService_SubscribeForRegisteredChecks(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	var shutdowner fx.Shutdowner
	var registeredChecks health.RegisteredCheckSubscription
	app := fx.New(
		health.ModuleWithDefaults(),
		fx.Invoke(
			func(subscribe health.SubscribeForRegisteredChecks) {
				registeredChecks = subscribe()
			},
			func(register health.Register) error {
				for i := 0; i < 10; i++ {
					check := health.Check{
						ID:           ulids.MustNew().String(),
						Description:  fmt.Sprintf("Desc %d", i),
						RedImpact:    fmt.Sprintf("Red %d", i),
						YellowImpact: fmt.Sprintf("Yellow %d", i),
						Tags:         []string{Database, MongoDB},
					}
					if err := register(check, health.CheckerOpts{}, func() error {
						return nil
					}); err != nil {
						return err
					}
				}

				return nil
			},
		),
		fx.Populate(&shutdowner),
	)

	if app.Err() != nil {
		t.Errorf("*** app initialization failed : %v", app.Err())
		return
	}

	var registeredCheckCount int
	for check := range registeredChecks.Chan() {
		t.Log(registeredCheckCount, check)
		registeredCheckCount++
		if registeredCheckCount == 10 {
			break
		}
	}

	runApp(app, shutdowner)
}

func TestService_CheckResults(t *testing.T) {
	var shutdowner fx.Shutdowner
	app := fx.New(
		health.ModuleWithDefaults(),
		fx.Invoke(
			func(register health.Register) error {
				for i := 0; i < 10; i++ {
					check := health.Check{
						ID:           ulids.MustNew().String(),
						Description:  fmt.Sprintf("Desc %d", i),
						RedImpact:    fmt.Sprintf("Red %d", i),
						YellowImpact: fmt.Sprintf("Yellow %d", i),
					}
					if err := register(check, health.CheckerOpts{}, func() error {
						return nil
					}); err != nil {
						return err
					}
				}
				return nil
			},
			func(registeredChecks health.RegisteredChecks, checkResults health.CheckResults) error {
				for {
					results := <-checkResults(nil)
					if len(results) == 10 {
						break
					}
					t.Logf("waiting for results: count = %v", len(results))
					runtime.Gosched()
					time.Sleep(time.Millisecond)
				}
				checks := <-registeredChecks(nil)
				if len(checks) != 10 {
					return fmt.Errorf("failed to retrieve all registered health checks: %v", len(checks))
				}
				for _, check := range checks {
					results := <-checkResults(func(result health.Result) bool {
						return result.HealthCheckID == check.ID
					})
					if len(results) != 1 {
						return fmt.Errorf("failed to get health check result: %v", results)
					}
				}
				return nil
			},
		),
		fx.Populate(&shutdowner),
	)

	if app.Err() != nil {
		t.Errorf("*** app initialization failed : %v", app.Err())
		return
	}

	runApp(app, shutdowner)
}

func TestService_RunningScheduledHealthChecks(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	var Foo = health.Check{
		ID:           "01DFGJ4A2GBTSQR11YYMV0N086",
		Description:  "Foo",
		RedImpact:    "App is unusable",
		YellowImpact: "App performance degradation",
		Tags:         []string{Database, MongoDB},
	}

	t.Run("health check times out", func(t *testing.T) {
		t.Parallel()

		opts := health.DefaultOpts()
		opts.MinRunInterval = time.Nanosecond

		var shutdowner fx.Shutdowner
		var resultsSubscription health.CheckResultsSubscription
		app := fx.New(
			health.Module(opts),
			fx.Invoke(
				func(subscribe health.SubscribeForCheckResults) {
					resultsSubscription = subscribe()
				},
				func(register health.Register) error {
					checkerOpts := health.CheckerOpts{
						Timeout: time.Nanosecond,
					}
					return register(Foo, checkerOpts, func() error {
						time.Sleep(time.Microsecond)
						return nil
					})
				},
				// verify that the health check timeout is 1 ns
				func(registeredChecks health.RegisteredChecks) {
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
			if result.Status != health.Red {
				t.Errorf("*** health check should have timed out, which is considered a Red failure")
			}
			if result.Err() != health.ErrTimeout {
				t.Errorf("*** error should have been timeout : %v", result.Err())
			}
		})

	})
}