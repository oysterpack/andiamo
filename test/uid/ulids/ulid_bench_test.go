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

/*
BenchmarkULIDChan-8                      2000000               748 ns/op              16 B/op          1 allocs/op
BenchmarkNewULIDCrypto-8                 2000000               731 ns/op              16 B/op          1 allocs/op
BenchmarkNewULIDMontonicCrypto-8        10000000               135 ns/op              16 B/op          1 allocs/op
*/
package ulids

import (
	"context"
	"crypto/rand"
	"github.com/oklog/ulid"
	"runtime"
	"sync"
	"testing"
)

func BenchmarkULIDChan(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithCancel(context.Background())
	bufSize := runtime.NumCPU() / 2
	ulids := make(chan ulid.ULID, bufSize)
	defer cancel()
	for i := 0; i < bufSize; i++ {
		go func() {
			entropy := ulid.Monotonic(rand.Reader, 0)
			for {
				select {
				case <-ctx.Done():
					return
				case ulids <- ulid.MustNew(ulid.Now(), entropy):
				}
			}
		}()
	}

	<-ulids
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ulids
	}
}

func BenchmarkNewULIDCrypto(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.MustNew(ulid.Now(), rand.Reader)
	}
}

func BenchmarkNewULIDMontonicCrypto(b *testing.B) {
	b.ReportAllocs()
	entropy := ulid.Monotonic(rand.Reader, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.MustNew(ulid.Now(), entropy)
	}
}

var lock sync.Mutex

func BenchmarkNewULIDMontonicCryptoMutexProtected(b *testing.B) {
	b.ReportAllocs()
	entropy := ulid.Monotonic(rand.Reader, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lock.Lock()
		ulid.MustNew(ulid.Now(), entropy)
		lock.Unlock()
	}
}
