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

func TestScheduler_Start(t *testing.T) {
	t.Parallel()

	registry := health.NewRegistry()

	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
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

		<-scheduler.Done()

		if scheduler.HealthCheckCount() != 0 {
			t.Errorf("*** scheduler health check count should be 0: %v", scheduler.HealthCheckCount())
		}

		select {
		case _, ok := <-scheduler.Subscribe(nil):
			if ok {
				t.Error("*** subscription chan should be closed")
			}
		default:
			t.Error("*** since the scheduler is shutdown, the subscription chan should always return false")
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

func TestScheduler_HealthCheckResults(t *testing.T) {
	t.Parallel()

	registry := health.NewRegistry()

	DatabaseHealthCheckDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Executes database query").
		YellowImpact("Slow query").
		RedImpact("Query times out or fails").
		MustBuild()

	UserDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
		Description("Queries the USERS DB").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			return nil
		}).
		RunInterval(time.Second).
		MustBuild()

	SessionsDBHealthCheck := health.NewBuilder(DatabaseHealthCheckDesc, ulidgen.MustNew()).
		Description("Queries the SESSIONS DB").
		RedImpact("Users will not be able to access the app").
		Checker(func(ctx context.Context) health.Failure {
			return health.YellowFailure(errors.New("query took 1.1 sec"))
		}).
		RunInterval(time.Second).
		MustBuild()

	err := registry.Register(UserDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}
	err = registry.Register(SessionsDBHealthCheck)
	if err != nil {
		t.Errorf("*** failed to register health check: %v", err)
		return
	}

	scheduler := health.StartScheduler(registry)

	for {
		results := <-scheduler.Results(nil)
		if len(results) == 2 {
			break
		}
		time.Sleep(time.Microsecond)
	}

	results := <-scheduler.Results(func(result health.Result) bool {
		return result.Status() != health.Green
	})
	if len(results) != 1 {
		t.Errorf("*** only 1 result should have been returned: %v", results)
		return
	}
	if results[0].Status() == health.Green {
		t.Errorf("*** result status should not be Green: %v", results[0])
	}

	scheduler.StopAsync()
	<-scheduler.Done()
	select {
	case _, ok := <-scheduler.Results(nil):
		if ok {
			t.Error("*** results chan should be closed because the scheduler is shutdown")
		}
	default:
		t.Error("*** results chan should be closed because the scheduler is shutdown")
	}

}
