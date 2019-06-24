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
	AppInitializedEventID          = "01DE4STZ0S24RG7R08PAY1RQX3"
	AppInitializationFailedEventID = "01DE4SWMZXD1ZB40QRT7RGQVPN"

	AppStartingEventID    = "01DE4SXMG8W3KSPZ9FNZ8Z17F8"
	AppStartFailedEventID = "01DE4SY6RYCD0356KYJV7G7THW"

	AppStartedEventID = "01DE4X10QCV1M8TKRNXDK6AK7C"

	AppStoppingEventID   = "01DE4SZ1KY60JQTF7XP4DQ8WGC"
	AppStopFailedEventID = "01DE4T0W35RPD6QMDS42WQXR48"

	AppStoppedEventID = "01DE4T1V9N50BB67V424S6MG5C"
)

// AppInitialized indicates the application has successfully initialized
type AppInitialized struct {
	App
}

func (event AppInitialized) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("start_timeout", event.StartTimeout())
	e.Dur("stop_timeout", event.StopTimeout())

	e.Strs("provides", typeNames(event.App.ConstructorTypes()))
	e.Strs("invokes", typeNames(event.App.FuncTypes()))
}

func typeNames(types []reflect.Type) []string {
	var names []string
	for _, t := range types {
		names = append(names, t.String())
	}
	return names
}

// AppInitializationFailed indicates the application failed to be built and initialized
type AppInitializationFailed struct {
	Err error
}

// AppStartFailed indicates the app failed to start.
type AppStartFailed struct {
	Err error
}

// AppStarted indicates the app has successfully been started.
type AppStarted struct {
	time.Duration
}

func (event AppStarted) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

// AppStopping indicates the app has been triggered to shutdown.
type AppStopping struct {
	os.Signal
}

// AppStopFailed indicates that the app did not shutdown cleanly, i.e., an error was encountered while shutting down the app.
type AppStopFailed struct {
	Err error
}

// AppStopped indicates that the app has been stopped.
// This will always be logged, regardless whether the app failed to shutdown cleanly or not, i.e., if an error occurs
// while shutting down the app, then both the AppStopFailed and AppStopped events will be logged.
type AppStopped struct {
	time.Duration
}

func (event AppStopped) MarshalZerologObject(e *zerolog.Event) {
	e.Dur("duration", event.Duration)
}

func logAppInitialized(logger *zerolog.Logger, app App) {
	logEvent := NewLogEventFunc(logger, zerolog.NoLevel, AppInitializedEventID)
	logEvent(AppInitialized{App: app}, "app initialized")
}

func logAppStarting(logger *zerolog.Logger) {
	logEvent := NewLogEventFunc(logger, zerolog.NoLevel, AppStartingEventID)
	logEvent(nil, "app starting")
}

func logAppStarted(logger *zerolog.Logger, startupTime time.Duration) {
	logEvent := NewLogEventFunc(logger, zerolog.NoLevel, AppStartedEventID)
	logEvent(AppStarted{startupTime}, "app started")
}

func logAppStopping(logger *zerolog.Logger) {
	logEvent := NewLogEventFunc(logger, zerolog.NoLevel, AppStoppingEventID)
	logEvent(nil, "app stopping")
}

func logAppStopped(logger *zerolog.Logger, shutdownDuration time.Duration) {
	logEvent := NewLogEventFunc(logger, zerolog.NoLevel, AppStoppedEventID)
	logEvent(AppStopped{shutdownDuration}, "app stopped")
}
