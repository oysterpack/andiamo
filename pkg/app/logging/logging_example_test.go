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

package logging_test

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"log"
	"time"
)

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
	// {"a":{"i":"01DC2PKQA8AVT00HAJRX7NBNP1","r":"01DC2PKQA8VWCKXAQY92356R5H","n":"foobar","v":"0.0.1","x":"01DC2PKQA877R22Q59C6J0RHR1"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","n":"start","t":1559163952}
	// {"a":{"i":"01DC2PKQA8AVT00HAJRX7NBNP1","r":"01DC2PKQA8VWCKXAQY92356R5H","n":"foobar","v":"0.0.1","x":"01DC2PKQA877R22Q59C6J0RHR1"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","n":"stop","t":1559163952}
	// {"l":"warn","a":{"i":"01DC2PKQA8AVT00HAJRX7NBNP1","r":"01DC2PKQA8VWCKXAQY92356R5H","n":"foobar","v":"0.0.1","x":"01DC2PKQA877R22Q59C6J0RHR1"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","n":"timeout","Duration":{"ms":2000},"t":1559163952}

	// Output:
	//
}

// Application log events
var (
	Timeout = logging.Event{
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

// Start logs an event when the app is started
func (l *AppLog) Started() {
	logging.Start.Log(l.Logger).Msg("")
}

// Start logs an event when the app is stopped
func (l *AppLog) Stopped() {
	logging.Stop.Log(l.Logger).Msg("")
}

// LogTimeoutEvent is used to log timeout events
type LogTimeoutEvent func(timeout time.Duration)

func NewLogTimeoutEvent(logger *zerolog.Logger) LogTimeoutEvent {
	return func(timeout time.Duration) {
		Timeout.Log(logger).Dict("Duration", zerolog.Dict().Dur("ms", timeout)).Msg("")
	}
}

func newLogger() *zerolog.Logger {
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	logger := logcfg.NewLogger(instanceID, desc)
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		log.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	return logging.PackageLogger(logger, PACKAGE)
}
