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

package health

import "errors"

// Failure represents a health check Failure
type Failure struct {
	error
	status Status
}

// YellowFailure constructs a new Failure with a Yellow status
func YellowFailure(err error) Failure {
	return Failure{err, Yellow}
}

// RedFailure constructs a new Failure with a Red status
func RedFailure(err error) Failure {
	return Failure{err, Red}
}

// ErrTimeout indicates a health check timed out.
// Healthcheck timeout errors are flagged as Red.
var ErrTimeout = errors.New("health check timed out")
