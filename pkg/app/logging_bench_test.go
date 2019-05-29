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
	"testing"
	"time"
)

// Benchmark Summary
// ================
// Benchmarks show that zerolog performance is on par with std log, i.e., the logging to stderr is IO bound.
// However, zerolog is more efficient with memory allocations - zerolog has zero allocations :)

// {"l":"info","a":{"i":"01DBY05VDS61D3QB34YC3D6HT4","r":"01DBY05VDSTXQJ1H73F648Z2V3","n":"foobar","v":"0.0.1","x":"01DBY05VDS1S7A68Y3PETT9E05"},"t":1559006212}
// 50000             21167 ns/op               0 B/op          0 allocs/op
func BenchmarkLoggingWithNoMessage(b *testing.B) {
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := app.NewLogger(instanceID, desc)
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("")
	}
}

// {"l":"info","a":{"i":"01DBY07S8PCA10QPNVGMAVQ2H2","r":"01DBY07S8PZXHFANS4WJEPTVYT","n":"foobar","v":"0.0.1","x":"01DBY07S8PRYYY6EYVD0JPS3RE"},"t":1559006275,"m":"message"}
// 100000             21026 ns/op               0 B/op          0 allocs/op
func BenchmarkLoggingWithMessage(b *testing.B) {
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := app.NewLogger(instanceID, desc)
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("message")
	}
}

// {"l":"warn","a":{"i":"01DC0B3ZPYG4P5BNB1CY0MFYDE","r":"01DC0B3ZPYKPAY8NS9AG0ET3QM","n":"foobar","v":"0.0.1","x":"01DC0B3ZPYRWXZH7SF8XVNFXZ5"},"n":"foo","t":1559084794,"m":"message"}
// 50000             21845 ns/op               0 B/op          0 allocs/op
func BenchmarkLogEvent_New(b *testing.B) {
	desc := apptest.InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := app.NewLogger(instanceID, desc)
	// And zerolog is configured
	if err := app.ConfigureZerolog(); err != nil {
		b.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}

	fooEvent := app.LogEvent{
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
