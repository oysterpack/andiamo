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
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"testing"
)

func runApp(t *testing.T, app *fx.App, shutdowner fx.Shutdowner, funcs ...func()) {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
		defer cancel()
		if err := app.Stop(ctx); err != nil {
			assert.Error(t, err, "app failed to stop")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()
	if err := app.Start(ctx); err != nil {
		assert.Error(t, err, "app failed to start")
		return
	}
	for _, f := range funcs {
		f()
	}
}
