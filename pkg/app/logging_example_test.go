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

package app_test

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/apptest"
	"github.com/rs/zerolog"
	"log"
	"time"
)

// Application log events
var (
	AppStarted = app.LogEvent{
		Name:  "app_started",
		Level: zerolog.InfoLevel,
	}
	AppStopped = app.LogEvent{
		Name:  "app_stopped",
		Level: zerolog.InfoLevel,
	}
	Timeout = app.LogEvent{
		Name:  "timeout",
		Level: zerolog.WarnLevel,
	}
)

// AppLog groups app related events
type AppLog struct {
	*zerolog.Logger
}

// NewAppLog constructs a new AppLog
func NewAppLog(logger *zerolog.Logger) AppLog {
	return AppLog{logger}
}

// Started logs an event when the app is started
func (l *AppLog) Started() {
	AppStarted.Log(l.Logger).Msg("")
}

// Started logs an event when the app is stopped
func (l *AppLog) Stopped() {
	AppStopped.Log(l.Logger).Msg("")
}

// LogTimeoutEvent is used to log timeout events
type LogTimeoutEvent func(timeout time.Duration)

func NewLogTimeoutEvent(logger *zerolog.Logger) LogTimeoutEvent {
	return func(timeout time.Duration) {
		Timeout.Log(logger).Dict("Duration", zerolog.Dict().Dur("ms", timeout)).Msg("")
	}
}

// This example demonstrates a couple of approaches to log events in a consistent and standardized manner.
// Log events should be carefully vetted and logged via an expressive and type-safe library.
func ExampleLogEvent() {
	logger := newLogger()

	appLog := NewAppLog(logger)
	appLog.Started()
	appLog.Stopped()

	timedout := NewLogTimeoutEvent(logger)
	timedout(2 * time.Second)

	// Sample log output
	// {"l":"info","a":{"i":"01DC0G120KG2HP5V2GXMYE8VHP","r":"01DC0G120KREEXZS2RH78T5NZQ","n":"foobar","v":"0.0.1","x":"01DC0G120KPFKF5SYWVWRKBJ2R"},"n":"app_started","t":1559089940}
	// {"l":"info","a":{"i":"01DC0G120KG2HP5V2GXMYE8VHP","r":"01DC0G120KREEXZS2RH78T5NZQ","n":"foobar","v":"0.0.1","x":"01DC0G120KPFKF5SYWVWRKBJ2R"},"n":"app_stopped","t":1559089940}
	// {"l":"warn","a":{"i":"01DC0G120KG2HP5V2GXMYE8VHP","r":"01DC0G120KREEXZS2RH78T5NZQ","n":"foobar","v":"0.0.1","x":"01DC0G120KPFKF5SYWVWRKBJ2R"},"n":"timeout","Duration":{"ms":2000},"t":1559089940}

	// Output:
	//
}

func newLogger() *zerolog.Logger {
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	logger := app.NewLogger(instanceID, desc)
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		log.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	return logger
}
