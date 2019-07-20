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

// Package eventlog standardizes structured JSON logging using zerolog as the underlying logging framework.
//
// Zerolog is initialized with the following settings:
//  - the following standard logger field names are shortened
//    - Timestamp -> t
//    - Level -> l
//    - Message -> m
//    - Error -> err
//  - Unix time format is used for performance reasons - seconds granularity is sufficient for log events
//  - an error stack marshaller is configured
//  - time.Duration fields are rendered as int instead float because it's more efficient
//  - each log event is tagged with an XID via a field named "x"
package eventlog
