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
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestNewHealthCheckRegistration(t *testing.T) {
	t.Parallel()

	DatabaseHealthCheck := fxapp.HealthCheckClass{
		ID:           "01DEMFD9QF7FHG0K2G44MXT6CY",
		Description:  "Executes database query",
		YellowImpact: "Slow query",
		RedImpact:    "Query timedout or failed",
	}

	DatabaseConnectivityHealthCheck := fxapp.HealthCheck{
		Class:       &DatabaseHealthCheck,
		ID:          "01DEMG4KTPH5XFM54JGQ7XZT7V",
		Description: "select 1 from session where session_id = ''", // augments the HealthCheckClass description
		RedImpact:   "users will not be able to access the app",    // augments the HealthCheckClass RedImapct
		Checker: func(ctx context.Context) fxapp.HealthCheckError {
			return fxapp.RedHealthCheckError(errors.New("DB conn failed"))
		},
	}

	healthCheckRegistration, err := fxapp.NewHealthCheckRegistration(&DatabaseConnectivityHealthCheck, fxapp.Interval(10*time.Second), fxapp.Timeout(time.Second))
	switch {
	case err != nil:
		t.Errorf("*** NewHealthCheckRegistration failed: %v", err)
	default:
		err = healthCheckRegistration.HealthCheck.Checker(context.Background())
		t.Log(err)
	}
}

func TestAppProvidesHealthCheckRegistry(t *testing.T) {
	t.Parallel()

	DatabaseHealthCheck := fxapp.HealthCheckClass{
		ID:           "01DEMFD9QF7FHG0K2G44MXT6CY",
		Description:  "Executes database query",
		YellowImpact: "Slow query",
		RedImpact:    "Query timedout or failed",
	}

	var registry fxapp.HealthCheckRegistry
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func() {

		},
		).
		Provide(
			func() (fxapp.HealthCheckRegistration, error) {
				healthCheck := fxapp.HealthCheck{
					Class:       &DatabaseHealthCheck,
					ID:          "01DEMG4KTPH5XFM54JGQ7XZT7V",
					Description: "select 1 from session where session_id = ''",
					RedImpact:   "users will not be able to access the app",
					Checker: func(ctx context.Context) fxapp.HealthCheckError {
						return fxapp.RedHealthCheckError(errors.New("DB conn failed"))
					},
				}
				return fxapp.NewHealthCheckRegistration(&healthCheck, fxapp.Interval(10*time.Second), fxapp.Timeout(time.Second))
			},
			func() (fxapp.HealthCheckRegistration, error) {
				healthCheck := fxapp.HealthCheck{
					Class:       &DatabaseHealthCheck,
					ID:          "01DEMT6E1JBMC0NKE9A3H9A5E9",
					Description: "select 1 from users where user_id = ''",
					RedImpact:   "users will not be able to access the app",
					Checker: func(ctx context.Context) fxapp.HealthCheckError {
						return nil
					},
				}
				return fxapp.NewHealthCheckRegistration(&healthCheck, fxapp.Interval(10*time.Second), fxapp.Timeout(time.Second))
			},
		).
		Populate(&registry).
		DisableHTTPServer().
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build error: %v", err)
	default:

	}
}
