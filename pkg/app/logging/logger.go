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

package logging

import (
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/rs/zerolog"
)

// PackageLogger add the specified package as a field to the logger
func PackageLogger(logger *zerolog.Logger, p app.Package) *zerolog.Logger {
	pkgLogger := logger.With().
		Str(string(Package), string(p)).
		Logger()
	return &pkgLogger
}
