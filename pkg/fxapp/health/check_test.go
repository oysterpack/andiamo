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
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheckID := ulidgen.MustNew()
	UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
		Description("Queries the USERS DB").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			time.Sleep(time.Millisecond)
			return health.RedFailure(errors.New("failed to connect to the database"))
		}).
		MustBuild()

	beforeRunningHealthCheck := time.Now()
	result := UserDBHealthCheck.Run()
	t.Log(result)
	if result.Status() != health.Red {
		t.Error("*** health check result should be Red")
	}
	if result.Duration() < time.Millisecond {
		t.Error("*** health check should have taken at least 1 msec to run")
	}
	if result.Time().Before(beforeRunningHealthCheck) {
		t.Error("*** healthcheck run time is not possible")
	}

	t.Run("description cannot be blank", func(t *testing.T) {
		_, err := (health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			})).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because description was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("   ").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			}).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because description is blank")
		}
	})

	t.Run("red impact cannot be blank", func(t *testing.T) {
		_, err := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Description").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			}).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because red impact was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Description").
			RedImpact("   ").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			}).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because red impact is blank")
		}
	})

	t.Run("check function is required", func(t *testing.T) {
		_, err := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Description").
			RedImpact("impact").
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because check func was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Description").
			RedImpact("impact").
			Checker(nil).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because check func was nil")
		}
	})

	t.Run("run green health check", func(t *testing.T) {
		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return nil
			}).
			MustBuild()

		result := UserDBHealthCheck.Run()
		t.Log(result)
		if result.Status() != health.Green {
			t.Errorf("*** status should be green")
		}
		if result.Error() != nil {
			t.Error("*** error should be nil")
		}
	})

	t.Run("health check times out", func(t *testing.T) {
		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return nil
			}).
			Timeout(time.Microsecond).
			MustBuild()

		result := UserDBHealthCheck.Run()
		t.Log(result)
		if result.Status() != health.Red {
			t.Errorf("*** status should be Red")
		}
		if result.Error() == nil {
			t.Error("*** health check should have timed out")
		}
	})
}
