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

/*
Package `app` standardizes how to compose applications using reusable modules leveraging ["go.uber.org/fx"](https://godoc.org/go.uber.org/fx)
Applications follow [12-Factor App](https://12factor.net/) design guidelines.

Features
========
1. App start and stop timeouts are specified via env vars:
- APPX12_START_TIMEOUT (default = 15s)
- APPX12_STOP_TIMEOUT (default = 15s)

2. `app.Desc` is loaded from the env and provided within the fx.App context:
- APPX12_ID (required)
  - app identifier - specified as a [ULID](https://github.com/ulid/spec)
- APPX12_NAME (required)
  - app name within the given context. Within k8s, the name must be unique within a namespace context.
- APPX12_VERSION (required)
  - follows semver convention
- APPX12_RELEASE_ID (required)
  - app release ID - specified as a [ULID](https://github.com/ulid/spec)

3. Each fx.App instance is assigned a unique `app.InstanceID` ULID, which is provided within the fx.App context

// TODO App Features:
- zap logger is provided within the fx.App context
- Application life cycle events are logged
   - app.new
   - app.starting
   - app.started
   - app.stopping
   - app.stopped
   - app.error
- Application dependency graph is logged in [DOT](https://graphviz.gitlab.io/_pages/doc/info/lang.html) format
- metrics
   - app_start_duration - how long did it take for the app to start
   - app_stop_duration - how long did it take for the app to stop
- events
   - app life cycle events
   - error

*/
package app
