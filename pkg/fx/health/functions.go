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

// Register is used to register health checks.
type Register func(check Check, opts CheckerOpts, checker func() (Status, error)) error

// RegisteredChecks returns all registered Checks
type RegisteredChecks func() <-chan []RegisteredCheck

// SubscribeForRegisteredChecks is used to subscribe for health check registrations
//
// Use Cases:
//  - logging - log the registered health checks
type SubscribeForRegisteredChecks func() RegisteredCheckSubscription

// CheckResults returns all current health check results that match the specified filter
type CheckResults func(filter func(result Result) bool) <-chan []Result

// SubscribeForCheckResults is used to subscribe to health check results that match the specified filter
type SubscribeForCheckResults func(filter func(result Result) bool) CheckResultsSubscription

// MonitorOverallHealth is used to subscribe to health status changes
// TODO:
type MonitorOverallHealth func() OverallHealthMonitor

// OverallHealth returns the overall health status.
//  - `Green` if all health checks are `Green`
//  - `Yellow` if there is at least 1 `Yellow` and no `Red`
//  - `Red` if at least 1 health check has a `Red` status
type OverallHealth func() Status
