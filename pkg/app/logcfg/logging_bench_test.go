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

package logcfg_test

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"log"
	"testing"
	"time"
)

// Benchmark Summary
// ================
// Benchmarks show that zerolog performance is on par with std log, i.e., the logging to stderr is IO bound.
// However, zerolog is more efficient with memory allocations - zerolog has zero allocations :)

// {"l":"info","a":{"i":"01DC2PY850NHGE3ABSDXBB0H5K","r":"01DC2PY8508607HG8TWS11AM3Y","n":"foobar","v":"0.0.1","x":"01DC2PY850WXPFKP1QM48XKJT3"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","t":1559164299}
// 100000             22576 ns/op               0 B/op          0 allocs/op
func BenchmarkLoggingWithNoMessage(b *testing.B) {
	desc := apptest.InitEnv()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := logging.PackageLogger(logcfg.NewLogger(instanceID, desc), PACKAGE)
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("")
	}
}

// {"l":"info","a":{"i":"01DC2PZQVFBXJWD1JWSQX0FVJC","r":"01DC2PZQVFMTZ0CYBDC2E6QV45","n":"foobar","v":"0.0.1","x":"01DC2PZQVFCS82MVK35R73NMT2"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","t":1559164348,"m":"message"}
// 100000             22843 ns/op               0 B/op          0 allocs/op
func BenchmarkLoggingWithMessage(b *testing.B) {
	desc := apptest.InitEnv()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := logging.PackageLogger(logcfg.NewLogger(instanceID, desc), PACKAGE)
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("message")
	}
}

// {"l":"warn","a":{"i":"01DC2Q1A7KYQB7VZT1EPYAQ4BG","r":"01DC2Q1A7K5V738F3RR2TWJ2RN","n":"foobar","v":"0.0.1","x":"01DC2Q1A7KXEFAMBTVA9RFTAR2"},"p":"github.com/oysterpack/partire-k8s/pkg/app_test","n":"foo","t":1559164398,"m":"message"}
// 50000             22037 ns/op               0 B/op          0 allocs/op
func BenchmarkLogEvent_Log(b *testing.B) {
	desc := apptest.InitEnv()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := logging.PackageLogger(logcfg.NewLogger(instanceID, desc), PACKAGE)
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	fooEvent := logging.Event{
		Name:  "foo",
		Level: zerolog.WarnLevel,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fooEvent.Log(logger).Msg("message")
	}
}

// 2019/05/27 21:18:11 message
// 100000             20960 ns/op               8 B/op          1 allocs/op
func BenchmarkStdLog(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Println("message")
	}
}
