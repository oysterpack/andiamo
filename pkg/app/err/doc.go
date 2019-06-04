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
Package err standardizes how errors are defined, created, and logged.

Desc is used to define the types of errors.
Err is used to define the the number of source code locations that can produce an error as defined by a Desc.
Instance represents and actual error instance and is assigned a unique instance ID. The Instance knows how to log itself.

*/
package err
