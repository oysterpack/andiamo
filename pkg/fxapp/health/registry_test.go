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
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	registry := health.NewRegistry()

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

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed register health check: %v", err)
		return
	}

	err = registry.Register(UserDBHealthCheck)
	if err == nil {
		t.Error("*** should have failed to register duplicate health check")
		return
	}
	t.Logf("%s", err)

	err = registry.Register(health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
		Description("Queries the SESSIONS DB").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			return nil
		}).
		MustBuild())
	if err != nil {
		t.Errorf("*** failed register health check: %v", err)
		return
	}

	checks := registry.HealthChecks(nil)
	t.Log(checks)
	if len(checks) != 2 {
		t.Errorf("*** registered health check count should be 2 : %v", len(checks))
	}

	checks = registry.HealthChecks(func(check health.Check) bool {
		return check.ID() == UserDBHealthCheckID
	})
	t.Log(checks)
	if len(checks) != 1 {
		t.Errorf("*** only 1 health check should have matched : %v", len(checks))
		return
	}
	if checks[0].ID() != UserDBHealthCheckID {
		t.Errorf("*** health check ID did not match: %v != %v", checks[0].ID(), UserDBHealthCheckID)
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	err := health.NewRegistry().Register(nil)
	if err == nil {
		t.Error("*** registering a nil health check should result in an error")
		return
	}
	t.Log(err)
}

func TestRegistry_Subscribe(t *testing.T) {
	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
		Description("Queries the USERS DB").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			time.Sleep(time.Millisecond)
			return health.RedFailure(errors.New("failed to connect to the database"))
		}).
		MustBuild()

	registry := health.NewRegistry()
	checkChan := make(chan health.Check)
	registry.Subscribe(checkChan)
	if err := registry.Register(UserDBHealthCheck); err != nil {
		t.Errorf("*** failed to register health check: %v", err)
	}
	t.Log(<-checkChan)
}
