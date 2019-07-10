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
	"github.com/oysterpack/partire-k8s/pkg/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheckID := ulidgen.MustNew()
	UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
		Description("Queries the USERS DB").
		YellowImpact("Users will experience longer response times").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			time.Sleep(time.Millisecond)
			return health.RedFailure(errors.New("failed to connect to the database"))
		}).
		MustBuild()

	if UserDBHealthCheck.Description() != "Queries the USERS DB" {
		t.Errorf("*** description did not match: %v", UserDBHealthCheck.Description())
	}
	if UserDBHealthCheck.YellowImpact() != "Users will experience longer response times" {
		t.Errorf("*** yellow impact did not match: %v", UserDBHealthCheck.YellowImpact())
	}
	if UserDBHealthCheck.RedImpact() != "Users will not be able to access the app" {
		t.Errorf("*** red impact did not match: %v", UserDBHealthCheck.RedImpact())
	}
	if UserDBHealthCheck.Desc().ID() != DatabaseHealthCheckDesc.ID() {
		t.Error("*** desc did not match")
	}
	if UserDBHealthCheck.Timeout() != time.Second*5 {
		t.Errorf("*** default timeout should be 10 secs: %v", UserDBHealthCheck.Timeout())
	}
	if UserDBHealthCheck.RunInterval() != time.Second*15 {
		t.Errorf("*** default run interval should be every 15 secs: %v", UserDBHealthCheck.RunInterval())
	}

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

	t.Run("run green health check", func(t *testing.T) {
		t.Parallel()

		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			MustBuild()

		result := UserDBHealthCheck.Run()
		t.Log(result)
		if result.HealthCheckID() != UserDBHealthCheck.ID() {
			t.Errorf("*** ID did not match: %v", result.HealthCheckID())
		}
		if result.Status() != health.Green {
			t.Errorf("*** status should be green")
		}
		if result.Error() != nil {
			t.Error("*** error should be nil")
		}
	})

	t.Run("health check times out", func(t *testing.T) {
		t.Parallel()

		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, UserDBHealthCheckID).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(10 * time.Millisecond)
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

func TestCheck_Run(t *testing.T) {
	t.Parallel()

	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	t.Run("run green health check", func(t *testing.T) {
		t.Parallel()

		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
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

	t.Run("run green health check", func(t *testing.T) {
		t.Parallel()

		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.YellowFailure(errors.New("YELLOW"))
			}).
			MustBuild()

		result := UserDBHealthCheck.Run()
		t.Log(result)
		if result.Status() != health.Yellow {
			t.Errorf("*** status should be green")
		}
		if result.Error() == nil {
			t.Error("*** error should not be nil")
		}
	})

	t.Run("health check times out", func(t *testing.T) {
		t.Parallel()

		UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Queries the USERS DB").
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(10 * time.Millisecond)
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

func TestCheck_Validation(t *testing.T) {
	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	t.Run("description cannot be blank", func(t *testing.T) {
		t.Parallel()

		_, err := (health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			RedImpact("Users will not be able to access the app").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			})).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because description was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
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
		t.Parallel()

		_, err := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			Checker(func(ctx context.Context) health.Failure {
				time.Sleep(time.Millisecond)
				return health.RedFailure(errors.New("failed to connect to the database"))
			}).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because red impact was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
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
		t.Parallel()

		_, err := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because check func was not specified")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(nil).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because check func was nil")
		}
	})

	t.Run("timeout cannot be zero", func(t *testing.T) {
		t.Parallel()

		_, err := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			Timeout(time.Duration(0)).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because the timeout was set to 0")
			return
		}
		t.Log(err)
	})

	t.Run("timeout cannot be greater than 10 secs", func(t *testing.T) {
		t.Parallel()

		_, err := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			Timeout(10 * time.Second).
			Build()

		if err != nil {
			t.Error("*** health check should have built because 10 secs is the max timeout")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			Timeout(time.Millisecond * 10001).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because the timeout was set to > 10 secs")
			return
		}
		t.Log(err)
	})

	t.Run("timeout cannot be greater than 10 secs", func(t *testing.T) {
		t.Parallel()

		_, err := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			RunInterval(time.Second).
			Build()

		if err != nil {
			t.Error("*** health check should have built because 1 sec is the min interval")
		}

		_, err = health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
			Description("Description").
			RedImpact("impact").
			Checker(func(ctx context.Context) health.Failure {
				return nil
			}).
			RunInterval(time.Millisecond * 999).
			Build()

		if err == nil {
			t.Error("*** health check should have failed to build because the min interval is 1 sec")
			return
		}
		t.Log(err)
	})
}

func TestBuilder_MustBuild_Panics(t *testing.T) {
	t.Parallel()

	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	defer func() {
		err := recover()
		if err == nil {
			t.Error("*** Builder.MustBuild() should have panicked because the health check is not valid")
		}
	}()

	health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).MustBuild()
}
