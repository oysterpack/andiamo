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

// RegisteredCheckSubscription wraps the channel used to notify subscribers
type RegisteredCheckSubscription struct {
	ch chan RegisteredCheck
}

// Chan returns the chan in read-only mode
func (s RegisteredCheckSubscription) Chan() <-chan RegisteredCheck {
	return s.ch
}

//CheckResultsSubscription wraps the channel used to notify subscribers
type CheckResultsSubscription struct {
	ch chan Result
}

// Chan returns the chan in read-only mode
func (s CheckResultsSubscription) Chan() <-chan Result {
	return s.ch
}

// OverallHealthMonitor publishes overall health changes.
// When first created, it immediately sends the current status.
// From that point on, when ever the overall health status changes, it is published.
type OverallHealthMonitor struct {
	ch chan Status
}

// Chan returns the chan in read-only mode
func (m OverallHealthMonitor) Chan() <-chan Status {
	return m.ch
}
