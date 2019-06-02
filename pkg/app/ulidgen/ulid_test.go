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

package ulidgen_test

import (
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"runtime"
	"sync"
	"testing"
)

func TestMonotonicULIDGenerator(t *testing.T) {
	ulids := make(map[ulid.ULID]bool)

	newULID := ulidgen.MonotonicULIDGenerator()

	for i := 0; i < 100; i++ {
		uid := newULID()
		if ulids[uid] {
			t.Fatal("duplicate ULID found")
		}
		ulids[uid] = true
	}
}

func TestRandomULIDGenerator(t *testing.T) {
	ulids := make(map[ulid.ULID]bool)

	newULID := ulidgen.RandomULIDGenerator()

	for i := 0; i < 100; i++ {
		uid := newULID()
		if ulids[uid] {
			t.Fatal("duplicate ULID found")
		}
		ulids[uid] = true
	}
}

// Single Threaded
// ---------------
// BenchmarkULIDGenerator/monotonic_baseline-8                     10000000               145 ns/op
// BenchmarkULIDGenerator/monotonic_mutex_protected-8              10000000               159 ns/op
// BenchmarkULIDGenerator/monotonic_chan_based-8                    3000000               477 ns/op
// BenchmarkULIDGenerator/random_mutex_protected-8                  2000000               756 ns/op
//
// Parallel
// --------
// BenchmarkMonotonicULIDGeneratorParallel-8                        3000000               432 ns/op
// BenchmarkMonotonicULIDGeneratorChanParallel-8                    5000000               353 ns/op
// BenchmarkRandomULIDGeneratorParallel-8                           3000000               448 ns/op
//
// summary
// -------
// - under highly concurrent parallel load, the chan based monotonic generator performs best - no locking
// - when minimal thread concurrent generation is expected, the mutex sync version is the best performing
func BenchmarkULIDGenerator(b *testing.B) {
	b.Run("monotonic baseline", func(b *testing.B) {
		entropy := ulid.Monotonic(rand.Reader, 0)
		newULID := func() ulid.ULID {
			return ulid.MustNew(ulid.Now(), entropy)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})

	b.Run("monotonic mutex protected", func(b *testing.B) {
		newULID := ulidgen.MonotonicULIDGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})

	b.Run("monotonic chan based", func(b *testing.B) {
		newULID := MonotonicULIDGeneratorChan()
		<-newULID
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			<-newULID
		}
	})

	b.Run("random mutex protected", func(b *testing.B) {
		newULID := ulidgen.RandomULIDGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})
}

func BenchmarkMonotonicULIDGeneratorParallel(b *testing.B) {
	newULID := ulidgen.MonotonicULIDGenerator()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newULID()
		}
	})
}

func BenchmarkMonotonicULIDGeneratorChanParallel(b *testing.B) {
	newULID := MonotonicULIDGeneratorChan()
	<-newULID
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			<-newULID
		}
	})
}

func BenchmarkRandomULIDGeneratorParallel(b *testing.B) {
	newULID := ulidgen.RandomULIDGenerator()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			newULID()
		}
	})
}

func MonotonicULIDGeneratorChan() <-chan ulid.ULID {
	count := runtime.NumCPU()
	c := make(chan ulid.ULID, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			entropy := ulid.Monotonic(rand.Reader, 0)
			wg.Done()
			for {
				c <- ulid.MustNew(ulid.Now(), entropy)
			}
		}()
	}
	wg.Wait()
	return c
}
