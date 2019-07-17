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

// Package health provides support for application health checks.
//
// Health check status can be `Green`, `Yellow`, or `Red`. A yellow status indicates that health check aspect is still
// functional but may be under stress, experiencing degraded performance, close to resource constraints, etc.
//
// When health checks are registered, they are scheduled to run on a periodic basis. The max number of health checks that
// can be run concurrently is configurable as a module option.
//
// The health check is configured with timeout. If the health check times out, then it is considered a `Red` failure.
// Health checks should be designed to run as fast as possible.
//
// The latest health check results are cached.
// Interested parties can subscribe for the following health check events:
//  - health check registrations
//  - health check results
//  - overall health status changes
//
// TODO:
// 	1. health check http API
//	2. health check grpc API
//     - server streaming APIs for health check results and
package health
