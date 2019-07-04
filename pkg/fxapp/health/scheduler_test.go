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
	"testing"
	"time"
)

func TestScheduler_Start(t *testing.T) {
	t.Parallel()

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
			return nil
		}).
		RunInterval(time.Second).
		MustBuild()

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	scheduler := health.StartScheduler(registry)
	defer func() {
		scheduler.StopAsync() // triggers shutdown
		// stopping the scheduler again should be ok
		scheduler.StopAsync()
		if !scheduler.Stopping() {
			t.Error("*** scheduler is reporting that it is not yet stopping")
		}
		timeout := time.After(time.Second)
		select {
		case <-scheduler.Done():
			t.Log("scheduler shutdown is complete")
		case <-timeout:
			t.Error("*** scheduler shutdown timed out")
		}

	}()

	for {
		if scheduler.HealthCheckCount() == 0 {
			t.Log("waiting for health check to get scheduled ...")
			time.Sleep(time.Millisecond)
		} else {
			break
		}
	}
	if count := scheduler.HealthCheckCount(); count != 1 {
		t.Errorf("*** there should be 1 health check scheduled: %d", count)
	}

	if scheduler.Stopping() {
		t.Error("*** scheduler should not be stopping")
	}

}

func TestScheduler_Subscribe(t *testing.T) {
	// shorten the run interval to run the test fast
	minRunInterval := health.MinRunInterval()
	health.SetMinRunInterval(time.Nanosecond)
	defer func() {
		// reset
		health.SetMinRunInterval(minRunInterval)
	}()

	t.Parallel()

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
			return nil
		}).
		RunInterval(1 * time.Microsecond).
		MustBuild()

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	scheduler := health.StartScheduler(registry)
	defer func() {
		scheduler.StopAsync() // triggers shutdown
		if !scheduler.Stopping() {
			t.Error("*** scheduler is reporting that it is not yet stopping")
		}
		timeout := time.After(1 * time.Second)
		select {
		case <-scheduler.Done():
		case <-timeout:
			t.Error("*** scheduler shutdown timed out")
		}
	}()

	healthCheckResults := scheduler.Subscribe(func(check health.Check) bool {
		return true
	})
	result := <-healthCheckResults
	t.Log(result)
	if result.Status() != health.Green {
		t.Errorf("*** health check result status should be Green: %v", result)
	}

}

func TestReceivingOnNilChan(t *testing.T) {
	var ch chan struct{}
	select {
	case <-ch:
	default:
		t.Log("no message")
	}
}
