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
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
)

var (
	appEventTags = []string{"app"}

	// Start signals that the app is starting.
	Start = logging.Event{
		Name:  "start",
		Level: zerolog.NoLevel,
		Tags:  appEventTags,
	}

	// Running signals that something is running.
	Running = logging.Event{
		Name:  "running",
		Level: zerolog.NoLevel,
		Tags:  appEventTags,
	}

	// Stop signals that something is stopping.
	Stop = logging.Event{
		Name:  "stop",
		Level: zerolog.NoLevel,
		Tags:  appEventTags,
	}

	// Stopped signals that something has stopped.
	Stopped = logging.Event{
		Name:  "stopped",
		Level: zerolog.NoLevel,
		Tags:  appEventTags,
	}
)
