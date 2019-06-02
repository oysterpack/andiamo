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
Package logging standardizes application logging - using zerolog as the logging framework.

`Field`
-------
- standardizes the log field names

`Event`
-------
- used to define all application log events in code
- is used to log events to a `zerolog.Logger`

`PackageLogger()`
-----------------
- used to scope log events to a package, i.e., it adds the `Package` field to the log event

*/
package logging
