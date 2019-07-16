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
	"context"
	"errors"
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"testing"
	"time"
)

// By default, as the part of the app start up process, all health checks are run. If any health checks fail, i.e., are
// not Green, then the app will fail to start.
func TestModuleConfiguredToRunChecksOnStartUp(t *testing.T) {

	t.Run("health check fails", func(t *testing.T) {
		opts := health.DefaultOpts()
		opts.FailFastOnStartup = true
		app := fx.New(
			health.Module(opts),
			fx.Invoke(
				func(lc fx.Lifecycle, register health.Register) error {
					return register(health.Check{
						ID:          ulids.MustNew().String(),
						Description: "Foo",
						RedImpact:   "RED",
					}, health.CheckerOpts{}, func() (status health.Status, e error) {
						return health.Red, errors.New("BOOM")
					})
				},
			),
		)

		assert.NoError(t, app.Err(), "app failed to initialize")

		if err := app.Start(context.Background()); err == nil {
			assert.Fail(t, "app should have failed to start")
		} else {
			t.Log(err)
			assert.Contains(t, err.Error(), "health check failed")
		}
	})

	t.Run("app start up times out before health checks complete", func(t *testing.T) {
		opts := health.DefaultOpts()
		opts.FailFastOnStartup = true
		app := fx.New(
			health.Module(opts),
			fx.Invoke(
				func(lc fx.Lifecycle, register health.Register) error {
					return register(health.Check{
						ID:          ulids.MustNew().String(),
						Description: "Foo",
						RedImpact:   "RED",
					}, health.CheckerOpts{}, func() (status health.Status, e error) {
						time.Sleep(2 * time.Second)
						return health.Red, errors.New("BOOM")
					})
				},
			),
		)

		assert.NoError(t, app.Err(), "app failed to initialize")

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		if err := app.Start(ctx); err == nil {
			assert.Fail(t, "app should have failed to start")
		} else {
			t.Log(err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		}
	})

	t.Run("app has no health checks", func(t *testing.T) {
		opts := health.DefaultOpts()
		opts.FailFastOnStartup = true
		app := fx.New(
			health.Module(opts),
		)

		assert.NoError(t, app.Err(), "app failed to initialize")

		if err := app.Start(context.Background()); err != nil {
			assert.Error(t, err, "app failed to start")
		}
		if err := app.Stop(context.Background()); err != nil {
			assert.Error(t, err, "app failed to stop")
		}
	})

}
