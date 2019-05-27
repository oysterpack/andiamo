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

package fx

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"go.uber.org/fx"
	"time"
)

// New construct a new fx.App with the following options:
// - app start and stop timeout options are configured from the env - see `LoadConfig()`
// - constructor functions for:
//   - Desc - loaded from the env - see `LoadDesc()`
//   - InstanceID
func New(options ...fx.Option) *fx.App {
	config := app.LoadConfig()
	options = append(options, fx.StartTimeout(config.StartTimeout))
	options = append(options, fx.StopTimeout(config.StopTimeout))

	options = append(options, fx.Provide(app.LoadDesc))
	options = append(options, fx.Provide(func() app.InstanceID {
		return app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	}))
	return fx.New(options...)
}
