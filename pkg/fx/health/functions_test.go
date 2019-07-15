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
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
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
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.Nil(t, app.Err(), "app initialization failed : %v", app.Err())
		runApp(t, app, shutdowner)
	})

	t.Run("register invalid health check - no fields set", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					InvalidHealthCheck := health.Check{}
					return register(InvalidHealthCheck, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.NotNil(t, app.Err(), "app initialization should have failed")
		t.Log(app.Err())
	})

	t.Run("register invalid health check - tag is not valid ULID", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					InvalidHealthCheck := health.Check{
						ID:           "01DFGJ4A2GBTSQR11YYMV0N086",
						Description:  "Foo",
						RedImpact:    "App is unusable",
						YellowImpact: "App performance degradation",
						Tags:         []string{Database, "INVALID"},
					}
					return register(InvalidHealthCheck, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.NotNil(t, app.Err(), "app initialization should have failed")
		t.Log(app.Err())
	})

	t.Run("register invalid health check - nil checker", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{}, nil)
				},
			),
			fx.Populate(&shutdowner),
		)

		require.NotNil(t, app.Err(), "app initialization should have failed")
		t.Log(app.Err())
	})

	t.Run("register invalid health check - invalid checker opts", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{Timeout: time.Minute, RunInterval: time.Millisecond}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.NotNil(t, app.Err(), "app initialization should have failed")
		t.Log(app.Err())
	})

	t.Run("register duplicate health check", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
				func(register health.Register) error {
					return register(Foo, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.NotNil(t, app.Err(), "app initialization should have failed")
		t.Log(app.Err())
	})

	t.Run("register valid health check with fields whitespace padding", func(t *testing.T) {
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

		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
			fx.Invoke(
				func(register health.Register) error {
					// And CheckerOpts were not specified
					return register(Foo, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					})
				},
				func(getRegisteredChecks health.RegisteredChecks) error {
					checks := <-getRegisteredChecks()
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

		require.Nil(t, app.Err(), "*** app initialization failed : %v", app.Err())
		runApp(t, app, shutdowner)
	})

}

func TestRegisteredChecks(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	t.Run("register 10 health checks", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Module(health.DefaultOpts()),
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
						if err := register(check, health.CheckerOpts{}, func() (health.Status, error) {
							return health.Green, nil
						}); err != nil {
							return err
						}
					}

					return nil
				},
				func(getRegisteredChecks health.RegisteredChecks) error {
					checks := <-getRegisteredChecks()
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

		require.Nil(t, app.Err(), "app initialization failed : %v", app.Err())
		runApp(t, app, shutdowner)
	})

}

func TestSubscribeForRegisteredChecks(t *testing.T) {
	t.Parallel()

	const (
		Database = "01DFGP2MJB9B8BMWA6Q2H4JD9Z"
		MongoDB  = "01DFGP3TS31D016DHS9415JFBB"
	)

	var shutdowner fx.Shutdowner
	var registeredChecks health.RegisteredCheckSubscription
	app := fx.New(
		health.Module(health.DefaultOpts()),
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
					if err := register(check, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
					}); err != nil {
						return err
					}
				}

				return nil
			},
		),
		fx.Populate(&shutdowner),
	)

	require.Nil(t, app.Err(), "app initialization failed : %v", app.Err())

	var registeredCheckCount int
	for check := range registeredChecks.Chan() {
		t.Log(registeredCheckCount, check)
		registeredCheckCount++
		if registeredCheckCount == 10 {
			break
		}
	}

	runApp(t, app, shutdowner)
}

func TestCheckResults(t *testing.T) {
	var shutdowner fx.Shutdowner
	app := fx.New(
		health.Module(health.DefaultOpts()),
		fx.Invoke(
			func(register health.Register) error {
				for i := 0; i < 10; i++ {
					check := health.Check{
						ID:           ulids.MustNew().String(),
						Description:  fmt.Sprintf("Desc %d", i),
						RedImpact:    fmt.Sprintf("Red %d", i),
						YellowImpact: fmt.Sprintf("Yellow %d", i),
					}
					if err := register(check, health.CheckerOpts{}, func() (health.Status, error) {
						return health.Green, nil
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
				checks := <-registeredChecks()
				if len(checks) != 10 {
					return fmt.Errorf("failed to retrieve all registered health checks: %v", len(checks))
				}

			CHECK_LOOP:
				for _, check := range checks {
					results := <-checkResults(func(result health.Result) bool {
						return result.ID == check.ID
					})
					for _, result := range results {
						if result.ID == check.ID {
							continue CHECK_LOOP
						}
					}
					t.Errorf("*** health check result was not returned for: %v", check)
				}
				return nil
			},
		),
		fx.Populate(&shutdowner),
	)

	require.Nil(t, app.Err(), "%v", app.Err())
	runApp(t, app, shutdowner)
}

func TestSubscribeForCheckResults(t *testing.T) {
	var shutdowner fx.Shutdowner
	var subscription health.CheckResultsSubscription
	app := fx.New(
		health.Module(health.DefaultOpts()),
		fx.Invoke(
			func(subscribe health.SubscribeForCheckResults) {
				subscription = subscribe(func(result health.Result) bool {
					return result.Status != health.Green
				})
			},
			func(register health.Register) error {
				t.Log("register RED health check")
				return register(
					health.Check{
						ID:          ulids.MustNew().String(),
						Description: "desc",
						RedImpact:   "FATAL",
					},
					health.CheckerOpts{},
					func() (status health.Status, e error) {
						return health.Red, errors.New("BOOM #1")
					},
				)
			},
			func(register health.Register) error {
				t.Log("register GREEN health check")
				return register(
					health.Check{
						ID:          ulids.MustNew().String(),
						Description: "desc",
						RedImpact:   "FATAL",
					},
					health.CheckerOpts{},
					func() (status health.Status, e error) {
						return health.Green, nil
					},
				)
			},
			func(register health.Register) error {
				t.Log("register YELLOW health check")
				return register(
					health.Check{
						ID:          ulids.MustNew().String(),
						Description: "desc",
						RedImpact:   "FATAL",
					},
					health.CheckerOpts{},
					func() (status health.Status, e error) {
						return health.Yellow, errors.New("BOOM #2")
					},
				)
			},
		),
		fx.Populate(&shutdowner),
	)

	results := []health.Result{<-subscription.Chan(), <-subscription.Chan()}
	t.Log(results)
	statusCounts := make(map[health.Status]int)
	for _, result := range results {
		statusCounts[result.Status] += 1
	}
	assert.Equal(t, 0, statusCounts[health.Green], "there should be no Green results")
	assert.Equal(t, 1, statusCounts[health.Yellow], "there should be 1 Yellow result")
	assert.Equal(t, 1, statusCounts[health.Red], "there should be 1 Red result")

	select {
	case result := <-subscription.Chan():
		t.Errorf("*** no more check results should have been received: %v", result)
	default:
	}

	runApp(t, app, shutdowner)
}

func TestRunningScheduledHealthChecks(t *testing.T) {
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

	t.Run("health check runs on schedule", func(t *testing.T) {
		t.Parallel()

		opts := health.DefaultOpts()
		opts.MinRunInterval = time.Nanosecond

		var shutdowner fx.Shutdowner
		var resultsSubscription health.CheckResultsSubscription
		app := fx.New(
			health.Module(opts),
			fx.Invoke(
				func(subscribe health.SubscribeForCheckResults) {
					resultsSubscription = subscribe(nil)
				},
				// register a health check that is scheduled to run every microsecond
				func(register health.Register) error {
					checkerOpts := health.CheckerOpts{
						Timeout:     time.Millisecond,
						RunInterval: time.Microsecond,
					}
					return register(Foo, checkerOpts, func() (health.Status, error) {
						return health.Yellow, errors.New("error")
					})
				},
			),
			fx.Populate(&shutdowner),
		)

		require.Nil(t, app.Err(), "%v", app.Err())

		runApp(t, app, shutdowner, func() {
			// wait for health check results to be reported
			count := 1
			for {
				result := <-resultsSubscription.Chan()
				t.Logf("[%d] %s", count, result)
				if count == 5 {
					// after we have received at least 5 results, then we are confident that the health checks are being run and reported properly
					return
				}
				count++
			}
		})
	})

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
					resultsSubscription = subscribe(nil)
				},
				func(register health.Register) error {
					checkerOpts := health.CheckerOpts{
						Timeout: time.Nanosecond,
					}
					return register(Foo, checkerOpts, func() (health.Status, error) {
						time.Sleep(time.Microsecond)
						return health.Green, nil
					})
				},
				// verify that the health check timeout is 1 ns
				func(registeredChecks health.RegisteredChecks) {
					registeredCheck := <-registeredChecks()
					t.Log(registeredCheck)
					if registeredCheck[0].Timeout != time.Nanosecond {
						t.Errorf("*** timeout should be 1 ns: %v", registeredCheck)
					}
				},
			),
			fx.Populate(&shutdowner),
		)

		require.Nil(t, app.Err(), "%v", app.Err())

		runApp(t, app, shutdowner, func() {
			result := <-resultsSubscription.Chan()
			t.Log(result)
			assert.Equal(t, health.Red, result.Status, "health check should have timed out, which is considered a Red failure")
			assert.Contains(t, result.Err.Error(), health.ErrTimeout.Error(), "error should have been timeout : %v", result.Err)
		})
	})
}

func TestInvokingFunctionsAfterServiceIsShutDown(t *testing.T) {
	var Foo = health.Check{
		ID:          "01DFGJ4A2GBTSQR11YYMV0N086",
		Description: "Foo",
		RedImpact:   "App is unusable",
	}

	var shutdowner fx.Shutdowner
	var register health.Register
	var registeredChecks health.RegisteredChecks
	var checkResults health.CheckResults
	var subscribeForRegisteredChecks health.SubscribeForRegisteredChecks
	var subscribeForCheckResults health.SubscribeForCheckResults
	app := fx.New(
		health.Module(health.DefaultOpts()),
		fx.Invoke(
			func(register health.Register) error {
				return register(Foo, health.CheckerOpts{}, func() (health.Status, error) {
					return health.Green, nil
				})
			},
		),
		fx.Populate(
			&shutdowner,
			&register,
			&registeredChecks,
			&checkResults,
			&subscribeForRegisteredChecks,
			&subscribeForCheckResults,
		),
	)

	require.Nil(t, app.Err(), "app initialization failed : %v", app.Err())
	runApp(t, app, shutdowner)

	assert.Error(t, register(
		health.Check{
			ID:          ulids.MustNew().String(),
			Description: "Foo",
			RedImpact:   "App is unusable",
		},
		health.CheckerOpts{},
		func() (health.Status, error) {
			return health.Green, nil
		},
	))

	_, ok := <-registeredChecks()
	assert.False(t, ok, "channel should be closed")

	_, ok = <-checkResults(nil)
	assert.False(t, ok, "channel should be closed")

	_, ok = <-subscribeForRegisteredChecks().Chan()
	assert.False(t, ok, "channel should be closed")

	_, ok = <-subscribeForCheckResults(nil).Chan()
	assert.False(t, ok, "channel should be closed")
}
