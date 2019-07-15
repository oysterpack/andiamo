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

package test

import (
	"context"
	"go.uber.org/fx"
	"testing"
)

func TestTypeAlias(t *testing.T) {
	var shutdowner fx.Shutdowner
	app := fx.New(
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					t.Log("starting")
					return nil
				},
				OnStop: func(context.Context) error {
					t.Log("stopping")
					return nil
				},
			})
		}),
		fx.Populate(&shutdowner),
	)
	app.Start(context.Background())
	if err := app.Stop(context.Background()); err != nil {
		t.Error(err)
	}
	if err := app.Start(context.Background()); err != nil {
		t.Error(err)
	}
	if err := app.Stop(context.Background()); err != nil {
		t.Error(err)
	}
}
