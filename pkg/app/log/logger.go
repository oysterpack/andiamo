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

package log

import (
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/rs/zerolog"
	"os"
)

// NewLogger constructs a new timestamped Logger with standardized fields.
//
// Example log message:
//
// {"l":"info","a":{"i":"01DBXQXE6WS76C2EXYBC06MSWB","r":"01DBXQXE6WSGB4EZGW8TGEH0PV","n":"foobar","v":"0.0.1","x":"01DBXQXE6WR0H602E09TA96X4D"},"t":1558997547228,"m":"info msg"}
//
// 	l   = level
//	t   = timestamp in Unix time
//  m   = message
//	a   = app
//	a.i = app ID
//	a.r = release ID
//	a.n = app name
//	a.v = app version
//	a.x = app instance ID
func NewLogger(instanceID app.InstanceID, desc app.Desc) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Dict(string(APP), zerolog.Dict().
			Str(string(APP_ID), desc.ID.String()).
			Str(string(APP_RELEASE_ID), desc.ReleaseID.String()).
			Str(string(APP_NAME), string(desc.Name)).
			Str(string(APP_VERSION), desc.Version.String()).
			Str(string(APP_INSTANCE_ID), instanceID.String())).
		Logger()

	return &logger
}

// PackageLogger add the specified package as a field to the logger
func PackageLogger(logger *zerolog.Logger, p app.Package) *zerolog.Logger {
	pkgLogger := logger.With().
		Str(string(PACKAGE), string(p)).
		Logger()
	return &pkgLogger
}
