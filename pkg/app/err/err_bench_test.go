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

package err_test

import (
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"testing"
)

// BenchmarkInstance_Log/with_no_stack_trace-8              3000000               555 ns/op               0 B/op          0 allocs/op
// BenchmarkInstance_Log/with_stack_trace-8                  100000             23241 ns/op            5529 B/op        108 allocs/op
//
// Logging the stacktrace is very expensive. Thus, collect the stacktrace only when needed.
func BenchmarkInstance_Log(b *testing.B) {

	ErrWithNoStackTrace := err.NewDesc("01DC9HDP0X3R60GWDZZY18CVB8", "Err", "error")
	ErrWithStackTrace := err.NewDesc("01DC9HDP0X3R60GWDZZY18CVB8", "Err", "error").WithStacktrace()

	logger := apptest.NewDiscardLogger(pkg)

	b.Run("with no stack trace", func(b *testing.B) {
		e := err.New(ErrWithNoStackTrace, "01DCGYD9N4CWBT6A55E5XW5TRT")
		errInstance := e.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errInstance.Log(logger).Msg("")
		}
	})

	b.Run("with stack trace", func(b *testing.B) {
		e := err.New(ErrWithStackTrace, "01DCGYD9N4CWBT6A55E5XW5TRT")
		errInstance := e.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errInstance.Log(logger).Msg("")
		}
	})
}
