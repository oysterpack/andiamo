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
)

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
			health.Options(),
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

	t.Run("register invalid health check", func(t *testing.T) {
		var shutdowner fx.Shutdowner
		app := fx.New(
			health.Options(),
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
			health.Options(),
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
			health.Options(),
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
