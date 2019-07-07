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

package fxapp

import (
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"time"
)

// app lifecycle event IDs
const (
	// 	type Data struct {
	//		StartTimeout uint `json:"start_timeout"`
	//		StopTimeout  uint `json:"stop_timeout"`
	//		Provides     []string
	//		Invokes      []string
	//	}
	InitializedEventID Event = "01DE4STZ0S24RG7R08PAY1RQX3"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	InitFailedEventID Event = "01DE4SWMZXD1ZB40QRT7RGQVPN"

	StartingEventID Event = "01DE4SXMG8W3KSPZ9FNZ8Z17F8"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	StartFailedEventID Event = "01DE4SY6RYCD0356KYJV7G7THW"

	// 	type Data struct {
	//		Duration uint
	//	}
	StartedEventID Event = "01DE4X10QCV1M8TKRNXDK6AK7C"

	ReadyEventID Event = "01DEJ5RA8XRZVECJDJFAA2PWJF"

	StoppingEventID Event = "01DE4SZ1KY60JQTF7XP4DQ8WGC"
	// 	type Data struct {
	//		Err string `json:"e"`
	//	}
	StopFailedEventID Event = "01DE4T0W35RPD6QMDS42WQXR48"

	// 	type Data struct {
	//		Duration uint
	//	}
	StoppedEventID Event = "01DE4T1V9N50BB67V424S6MG5C"
)

// AppInitialized indicates the application has successfully initialized
type AppInitialized struct {
	App
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppInitialized) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("start_timeout", event.StartTimeout())
	e.Dur("stop_timeout", event.StopTimeout())

	typeNames := func(types []reflect.Type) []string {
		var names []string
		for _, t := range types {
			names = append(names, t.String())
		}
		return names
	}

	e.Strs("provides", typeNames(event.App.ConstructorTypes()))
	e.Strs("invokes", typeNames(event.App.FuncTypes()))
}

// AppStarted indicates the app has successfully been started.
type AppStarted struct {
	time.Duration
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppStarted) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

// AppStopping indicates the app has been triggered to shutdown.
type AppStopping struct {
	os.Signal
}

// AppStopped indicates that the app has been stopped.
// This will always be logged, regardless whether the app failed to shutdown cleanly or not, i.e., if an error occurs
// while shutting down the app, then both the AppStopFailed and AppStopped events will be logged.
type AppStopped struct {
	time.Duration
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppStopped) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

// AppFailed indicates the application failed to be built and initialized
type AppFailed struct {
	Err error
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler interface
func (event AppFailed) MarshalZerologObject(e *zerolog.Event) {
	e.Err(event.Err)
}

// health check related events
const (
	// HealthCheck is used to generate the event data, e.g.,
	//
	// "01DF3FV60A2J1WKX5NQHP47H61": {
	//    "id": "01DF3MNDKPB69AJR7ZGDNB3KA1",
	//    "desc_id": "01DF3MNDKP8DS3B04E2TKFHXD9",
	//    "description": [
	//      "Foo",
	//      "FooBar"
	//    ],
	//    "red_impact": [
	//      "app is unavailable",
	//      "fatal"
	//    ],
	//    "yellow_impact": [
	//      "app response times are slow"
	//    ],
	//    "timeout": 5000,
	//    "run_interval": 15000
	//  }
	//
	// - description, red_impact, yellow_impact are combined from health.Desc and health.Check
	HealthCheckRegisteredEventID Event = "01DF3FV60A2J1WKX5NQHP47H61"

	HealthCheckResultEventID Event = "01DF3X60Z7XFYVVXGE9TFFQ7Z1"

	HealthCheckGaugeRegistrationErrorEventID Event = "01DF6M0T7K3DNSFMFQ26TM7XX4"
)
