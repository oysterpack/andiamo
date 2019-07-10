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
	"context"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"testing"
	"time"
)

func TestScheduler_Subscribe(t *testing.T) {
	registry := NewRegistry()

	DatabaseHealthCheckDesc := NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck, err := CheckOpts{
		Desc:        DatabaseHealthCheckDesc,
		ID:          ulidgen.MustNew().String(),
		Description: "Desc",
		RedImpact:   "Users will not be able to access the app",
		Checker: func(ctx context.Context) Failure {
			return nil
		},
	}.new(checkConstraints{minRunInterval: time.Nanosecond, maxRunTimeout: 10 * time.Second})

	if err != nil {
		t.Errorf("*** UserDBHealthCheck is not valid: %v", err)
		return
	}

	SessionDBHealthCheck, err := CheckOpts{
		Desc:        DatabaseHealthCheckDesc,
		ID:          ulidgen.MustNew().String(),
		Description: "Queries the SESSIONS DB",
		RedImpact:   "Users will not be able to access the app",
		Checker: func(ctx context.Context) Failure {
			return nil
		},
	}.new(checkConstraints{minRunInterval: time.Nanosecond, maxRunTimeout: 10 * time.Second})

	if err != nil {
		t.Errorf("*** SessionDBHealthCheck is not valid: %v", err)
		return
	}

	err = registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}
	err = registry.Register(SessionDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	scheduler := StartScheduler(registry)
	defer func() {
		scheduler.StopAsync() // triggers shutdown
		if !scheduler.Stopping() {
			t.Error("*** scheduler is reporting that it is not yet stopping")
		}
		timeout := time.After(1 * 10 * time.Second)
		select {
		case <-scheduler.Done():
		case <-timeout:
			t.Error("*** scheduler shutdown timed out")
		}
	}()

	healthCheckResults := scheduler.Subscribe(func(check Check) bool {
		return check.ID() == UserDBHealthCheck.ID()
	})
	result := <-healthCheckResults
	t.Log(result)
	if result.Status() != Green {
		t.Errorf("*** health check result status should be Green: %v", result)
	}

}

func TestScheduler_Subscribe_GetResultsAfterSchedulerClosed(t *testing.T) {
	registry := NewRegistry()

	DatabaseHealthCheckDesc := NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck := CheckOpts{
		Desc:        DatabaseHealthCheckDesc,
		ID:          ulidgen.MustNew().String(),
		Description: "Queries the USERS DB",
		RedImpact:   "Users will not be able to access the app",
		Checker: func(ctx context.Context) Failure {
			return nil
		},
		Timeout: time.Microsecond,
	}.mustNew(checkConstraints{minRunInterval: time.Nanosecond, maxRunTimeout: 10 * time.Second})

	SessionDBHealthCheck := CheckOpts{
		Desc:        DatabaseHealthCheckDesc,
		ID:          ulidgen.MustNew().String(),
		Description: "Queries the SESSIONS DB",
		RedImpact:   "Users will not be able to access the app",
		Checker: func(ctx context.Context) Failure {
			return nil
		},
		Timeout: time.Microsecond,
	}.mustNew(checkConstraints{minRunInterval: time.Nanosecond, maxRunTimeout: 10 * time.Second})

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}
	err = registry.Register(SessionDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	scheduler := StartScheduler(registry)

	healthCheckResults := scheduler.Subscribe(func(check Check) bool {
		return check.ID() == UserDBHealthCheck.ID()
	})

	scheduler.StopAsync()
	select {
	case <-scheduler.Done():
		t.Log("scheduler is shutdown")
	case result := <-healthCheckResults:
		t.Log(result)
		if result.Status() != Green {
			t.Errorf("*** health check result status should be Green: %v", result)
		}
	}
}

func TestSchedulerAutomaticallySchedulesRegisteredHealthCheck(t *testing.T) {
	registry := NewRegistry()

	DatabaseHealthCheckDesc := NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck := CheckOpts{
		Desc:        DatabaseHealthCheckDesc,
		ID:          ulidgen.MustNew().String(),
		Description: "Queries the USERS DB",
		RedImpact:   "Users will not be able to access the app",
		Checker: func(ctx context.Context) Failure {
			return nil
		},
	}.mustNew(checkConstraints{minRunInterval: time.Nanosecond, maxRunTimeout: 10 * time.Second})

	scheduler := StartScheduler(registry)
	defer func() {
		scheduler.StopAsync() // triggers shutdown
		if !scheduler.Stopping() {
			t.Error("*** scheduler is reporting that it is not yet stopping")
		}
		timeout := time.After(1 * 10 * time.Second)
		select {
		case <-scheduler.Done():
		case <-timeout:
			t.Error("*** scheduler shutdown timed out")
		}
	}()

	healthCheckResults := scheduler.Subscribe(nil)

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	result := <-healthCheckResults
	t.Log(result)
	if result.Status() != Green {
		t.Errorf("*** health check result status should be Green: %v", result)
	}
}
