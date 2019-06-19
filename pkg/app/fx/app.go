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

package fx

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"os"
	"strings"
	"time"
)

const pkg app.Package = "github.com/oysterpack/partire-k8s/pkg/app/fx"

// StartTime is when the app was started
type StartTime time.Time

// StartedTime is when the the app startup lifecycle phase completed
type StartedTime time.Time

// App is a wrapper around fx.App.
//
// The main reason to wrap fx.App is too override fx.App.Run() with App.Run() that returns an error.
// fx.App.Run() would fatally exit the process if a start or stop error occurs. Instead, we want to simply return an
// error, which also makes it easier to test the App. In addition, the following behaviors are added:
// - all start and stop errors are logged using standardized errors
// - a StopSignal event is logged, which logs exactly what type of os.Signal was received
type App struct {
	*fx.App
	logger  *zerolog.Logger
	stopped chan os.Signal
}

// Run starts the application, blocks on the signals channel, and then gracefully shuts the application down.
// It uses DefaultTimeout to set a deadline for application startup and shutdown, unless the user has configured
// different timeouts with the StartTimeout or StopTimeout options. It's designed to make typical applications simple to run.
//
// Application lifecycle events are logged.
func (a *App) Run() error {
	startCtx, cancel := context.WithTimeout(context.Background(), a.StartTimeout())
	defer cancel()
	defer close(a.stopped)

	stopChan := a.App.Done()

	if e := a.Start(startCtx); e != nil {
		appStartErr := AppStartErr.CausedBy(e)
		appStartErr.Log(a.logger).Msg("")
		return appStartErr
	}

	// wait for the app to be signalled to stop
	signal := <-stopChan
	defer func() {
		a.stopped <- signal
	}()
	StopSignal.Log(a.logger).Msg(strings.ToUpper(signal.String()))

	stopCtx, cancel := context.WithTimeout(context.Background(), a.StopTimeout())
	defer cancel()

	if e := a.Stop(stopCtx); e != nil {
		appStopErr := AppStopErr.CausedBy(e)
		appStopErr.Log(a.logger).Msg("")
		return appStopErr
	}

	return nil
}

// Stopped returns a chan that is used to signal when the app has stopped.
//
// NOTE: the stopped channel should only be used when the application is run via App.Run(), i.e., it is signalled when the
//       the App.Run() method completes
func (a *App) Stopped() <-chan os.Signal {
	return a.stopped
}
