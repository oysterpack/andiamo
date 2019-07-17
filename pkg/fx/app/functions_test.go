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

package app_test

import (
	"github.com/oklog/ulid"
	"github.com/oysterpack/andiamo/pkg/fx/app"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"os"
	"testing"
)

func TestAppIDs(t *testing.T) {
	t.Parallel()

	t.Run("IDs are explicity set", func(t *testing.T) {
		t.Parallel()
		opts := app.Opts{ID: ulids.MustNew(), ReleaseID: ulids.MustNew()}
		a := fx.New(
			app.Module(opts),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), opts.ID)
					assert.Equal(t, releaseID(), opts.ReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")
	})

	t.Run("load from env", func(t *testing.T) {
		prefix := ulids.MustNew().String()
		appID := ulids.MustNew()
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		a := fx.New(
			app.Module(app.Opts{EnvPrefix: prefix}),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), appID)
					assert.Equal(t, releaseID(), appReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")
	})

	t.Run("load from env", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), appID)
					assert.Equal(t, releaseID(), appReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")

	})

	t.Run("load from env - missing ID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())

	})

	t.Run("load from env - missing RELEASE_ID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(
				func(app.ReleaseID) {},
			),
		)

		assert.Error(t, a.Err(), a.Err().Error())
	})

	t.Run("load from env - invalid ID ULID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		os.Setenv(prefix+"_RELEASE_ID", ulids.MustNew().String())
		os.Setenv(prefix+"_ID", "INVALID")

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())

	})

	t.Run("load from env - invalid RELEASE_ID ULID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", "INVALID")

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ReleaseID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())

	})

}
