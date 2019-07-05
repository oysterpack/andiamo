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

package fxapp_test

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"testing"
)

// The app automatically provides health.Registry and health.Scheduler.
func TestAppHealthCheckRegistry(t *testing.T) {
	t.Parallel()

	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()

	var healthCheckRegistry health.Registry
	var healthCheckScheduler health.Scheduler
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func(registry health.Registry) {
			FooHealth := health.NewBuilder(FooHealthDesc, ulidgen.MustNew()).
				Description("Foo").
				RedImpact("fatal").
				Checker(func(ctx context.Context) health.Failure {
					return nil
				}).
				MustBuild()

			registry.Register(FooHealth)
		}).
		Populate(&healthCheckRegistry, &healthCheckScheduler).
		Build()

	if err != nil {
		t.Errorf("*** app failed to build: %v", err)
	}

	// health checks are scheduled to run as they are registered

	go app.Run()
	<-app.Started()

	// When the app is shutdown
	app.Shutdown()
	<-app.Done()

	// Then the health.Scheduler instance is shutdown

}
