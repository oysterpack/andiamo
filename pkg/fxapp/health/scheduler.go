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

// Scheduler is used to schedule health checks to run.
//
// Design
// - Only 1 health check will be allowed to run at a time to prevent application / system overload.
// - The health check's next run is scheduled when the health check run is complete.
// - As health checks are registered, then they will get scheduled to run.
type Scheduler interface {
	Running() <-chan struct{}
}

type scheduler struct {
	Registry

	running chan struct{}
}

func StartScheduler(registry Registry) Scheduler {
	return &scheduler{
		Registry: registry,
	}
}

func (s *scheduler) Running() <-chan struct{} {
	return s.running
}
