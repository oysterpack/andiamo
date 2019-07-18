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

package app

import (
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"time"
)

// app lifecycle event IDs
const (
	// 	type Data struct {
	//		DependencyGraph string `json:"dot_graph"` // DOT language visualization of the app dependency graph
	//	}
	InitializedEvent = "01DE4STZ0S24RG7R08PAY1RQX3"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	InitFailedEvent = "01DE4SWMZXD1ZB40QRT7RGQVPN"

	StartingEvent = "01DE4SXMG8W3KSPZ9FNZ8Z17F8"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	StartFailedEvent = "01DE4SY6RYCD0356KYJV7G7THW"

	// 	type Data struct {
	//		Duration uint
	//      DependencyGraph string `json:"dot_graph"` // DOT language visualization of the app dependency graph
	//	}
	StartedEvent = "01DE4X10QCV1M8TKRNXDK6AK7C"

	StoppingEvent = "01DE4SZ1KY60JQTF7XP4DQ8WGC"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	StopFailedEvent = "01DE4T0W35RPD6QMDS42WQXR48"

	// 	type Data struct {
	//		Duration uint
	//	}
	StoppedEvent = "01DE4T1V9N50BB67V424S6MG5C"
)

type appInfo struct {
	fx.DotGraph
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event appInfo) MarshalZerologObject(e *zerolog.Event) {
	e.Str("dot_graph", string(event.DotGraph))
}

type duration time.Duration

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (d duration) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", time.Duration(d))
}
