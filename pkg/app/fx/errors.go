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

import "github.com/oysterpack/partire-k8s/pkg/app/err"

// App related errors
var (
	// InvokeErr indicates that a function that was invoked by the app failed.
	// This is only used to wrap non-standard errors, i.e., whose type is not *err.Instance
	//
	// - the error stack is included to help track down where the error came from
	InvokeErr = err.NewDesc("01DCFB3H7DDT7PG5WD5MHVSZ25", "InvokeErr", "invoking app function failed", err.IncludeStack)

	// AppStartErr indicates the app failed to start
	AppStartErr = err.NewDesc("01DCFMV6VJ6QS9B22Z7Q38EC8V", "AppStartErr", "app failed to start", err.ExcludeStack)

	// AppStopErr indicates that the app failed to stop cleanly
	AppStopErr = err.NewDesc("01DCFPF53Z0YF0QDM6YW7818JE", "AppStartErr", "app failed to stop cleanly", err.ExcludeStack)
)
