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
Package fx standardizes how to compose applications using reusable modules leveraging ["go.uber.org/fx"](https://godoc.org/go.uber.org/fx)
Applications follow [12-Factor App](https://12factor.net/) design guidelines.

Features

Basic App Config

	1. `app.Timeouts` is loaded from the env and made available within the fx.App context:
	   - APPX12_START_TIMEOUT (default = 15s)
	   - APPX12_STOP_TIMEOUT (default = 15s)

	2. `app.Desc` is loaded from the env and made available within the fx.App context:
	   - APPX12_ID (required)
		 - app identifier - specified as a [ULID](https://github.com/ulid/spec)
	   - APPX12_NAME (required)
		 - app name within the given context. Within k8s, the name must be unique within a namespace context.
	   - APPX12_VERSION (required)
		 - follows semver convention
	   - APPX12_RELEASE_ID (required)
		 - app release ID - specified as a [ULID](https://github.com/ulid/spec)

	3. Logging Config (zerolog)
		- APPX12_LOG_GLOBAL_LEVEL
		   - configures the global log level
		   - default = info
		 - APPX12_LOG_DISABLE_SAMPLING
		   - if true, then log sampling is disabled

App Logging

[zerolog](https://github.com/rs/zerolog) is used as the logging framework.

	1. *zerolog.Logger is made available within the fx.App context
	2. zerolog settings:
	   - writes to stderr
	   - event timestamp is in Unix time
	   - event context contains app info - see `app.NewLogger()` for details
	3. zerolog is used as the fx.App logger
	   - fx.App log events use debug level
	4. zerolog is used as the go std logger
	5. app start and stop events are logged
	   - they are logged with no log level to ensure they are always logged, regardless of the global log level setting

Components

Applications are composed of components.

App Context Injections

	1. app.InstanceID
	2. app.Desc
	3. *zerolog.Logger
	4. *err.Registry
	5. *logging.EventRegistry
	6. *comp.Registry

// TODO App Features:
- Application dependency graph is logged in [DOT](https://graphviz.gitlab.io/_pages/doc/info/lang.html) format
- metrics
   - app_start_duration - how long did it take for the app to start
   - app_stop_duration - how long did it take for the app to stop
*/
package fx
