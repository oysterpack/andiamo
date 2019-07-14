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
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"log"
)

func Example() {

	MongoDB := "01DFRABQGWX6HB1HFX3K0GR15G"

	DatabaseHealthCheck := health.Check{
		ID: "01DFR9PN8BEMZFGBC49A698WTS",
		Description: `Performs the following:
1. creates a new session 
2. retrieves the session 
3. deletes the session    

Each operation should not take more than 50 msec.
Yellow -> operation took longer than 50 msec but less than 1 sec
Red -> SQL error or operation took longer than 1 sec
`,
		YellowImpact: "degraded performance",
		RedImpact:    "unacceptable user experience",
		Tags:         []string{MongoDB},
	}

	SmokeTest := health.Check{
		ID:          "01DFRB8XAXMF9XJW2XYCSMN4VE",
		Description: "smoke test",
		RedImpact:   "application is non-functional",
	}

	var registeredCheckSubscription health.RegisteredCheckSubscription
	var checkResultsSubscription health.CheckResultsSubscription
	var checkResults health.CheckResults
	app := fx.New(
		// install the health module using default options
		health.ModuleWithDefaults(),
		fx.Invoke(
			// initialize subscribers
			func(subscribe health.SubscribeForRegisteredChecks) {
				registeredCheckSubscription = subscribe()
			},
			func(subscribe health.SubscribeForCheckResults) {
				checkResultsSubscription = subscribe(func(result health.Result) bool {
					return result.Status != health.Green
				})
			},
			// register some health checks
			func(register health.Register) error {
				return register(DatabaseHealthCheck, health.CheckerOpts{}, func() (status health.Status, e error) {
					return health.Yellow, errors.New("creating new session was too slow")
				})
			},
			func(register health.Register) error {
				return register(SmokeTest, health.CheckerOpts{}, func() (status health.Status, e error) {
					return health.Green, nil
				})
			},
			func(registeredChecks health.RegisteredChecks) {
				log.Print(<-registeredChecks())
			},
		),
		fx.Populate(&checkResults),
	)

	// make sure the app initialized with no errors
	if app.Err() != nil {
		log.Panic(app.Err())
	}
	app.Start(context.Background())
	defer app.Stop(context.Background())

	// 2 health checks were registered
	log.Println(<-registeredCheckSubscription.Chan())
	log.Println(<-registeredCheckSubscription.Chan())

	// we subscribed to receive health checks that are not Green
	log.Println(<-checkResultsSubscription.Chan())
	// get all check results
	log.Println(<-checkResults(nil))

	// Output:
}
