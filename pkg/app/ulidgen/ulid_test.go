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
	crand "crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestMonotonicULIDGenerator(t *testing.T) {
	t.Parallel()
	ulids := make(map[ulid.ULID]bool)
	newULID := ulidgen.MonotonicULIDGenerator()
	for i := 0; i < 100; i++ {
		uid := newULID()
		t.Log(uid)
		if ulids[uid] {
			t.Fatal("duplicate ULID found")
		}
		ulids[uid] = true
	}
}

func TestRandomULIDGenerator(t *testing.T) {
	t.Parallel()
	ulids := make(map[ulid.ULID]bool)
	newULID := ulidgen.RandomULIDGenerator()
	for i := 0; i < 100; i++ {
		uid := newULID()
		t.Log(uid)
		if ulids[uid] {
			t.Fatal("duplicate ULID found")
		}
		ulids[uid] = true
	}
}

func TestMustNew(t *testing.T) {
	t.Parallel()
	ulids := make(map[ulid.ULID]bool)
	for i := 0; i < 100; i++ {
		uid := ulidgen.MustNew()
		if ulids[uid] {
			t.Fatal("duplicate ULID found")
		}
		ulids[uid] = true
	}
}

// Intel(R) Core(TM) i7-3770K CPU @ 3.50GHz
//
// Single Threaded
// ---------------
// BenchmarkULIDGenerator/monotonic_baseline_using_rand-8          10000000               136 ns/op
// BenchmarkULIDGenerator/monotonic_baseline_using_crypto/rand-8   10000000               144 ns/op
// BenchmarkULIDGenerator/MonotonicULIDGenerator-8                 10000000               154 ns/op
// BenchmarkULIDGenerator/MonotonicULIDGeneratorChan-8              3000000               464 ns/op
// BenchmarkULIDGenerator/RandomULIDGenerator-8                     2000000               756 ns/op
//
// Parallel
// --------
// BenchmarkMonotonicMathRandULIDGeneratorParallel-8                3000000               413 ns/op
// BenchmarkMonotonicULIDGeneratorParallel-8                        3000000               412 ns/op
// BenchmarkMonotonicULIDGeneratorChanParallel-8                    5000000               357 ns/op
// BenchmarkRandomULIDGeneratorParallel-8                           3000000               449 ns/op
//
// BenchmarkULIDMustParse-8                                        50000000                26.4 ns/op
//
// summary
// -------
// - crypto/rand adds minimal overhead vs math/rand (~8 ns)
// - under highly concurrent parallel load, the chan based monotonic generator performs best - no locking
// - when minimal thread concurrent generation is expected, the mutex sync version is the best performing
// - parsing ULID is fast
func BenchmarkULIDGenerator(b *testing.B) {
	b.Run("monotonic baseline using rand", func(b *testing.B) {
		entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
		newULID := func() ulid.ULID {
			return ulid.MustNew(ulid.Now(), entropy)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})

	b.Run("monotonic baseline using crypto/rand", func(b *testing.B) {
		entropy := ulid.Monotonic(crand.Reader, 0)
		newULID := func() ulid.ULID {
			return ulid.MustNew(ulid.Now(), entropy)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})

	b.Run("MonotonicULIDGenerator", func(b *testing.B) {
		newULID := ulidgen.MonotonicULIDGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})

	b.Run("MonotonicULIDGeneratorChan", func(b *testing.B) {
		newULID := MonotonicULIDGeneratorChan()
		<-newULID
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			<-newULID
		}
	})

	b.Run("RandomULIDGenerator", func(b *testing.B) {
		newULID := ulidgen.RandomULIDGenerator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			newULID()
		}
	})
}

func BenchmarkMonotonicMathRandULIDGeneratorParallel(b *testing.B) {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	var m sync.Mutex
	newULID := func() ulid.ULID {
		m.Lock()
		u := ulid.MustNew(ulid.Now(), entropy)
		m.Unlock()
		return u
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
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
			entropy := ulid.Monotonic(crand.Reader, 0)
			wg.Done()
			for {
				c <- ulid.MustNew(ulid.Now(), entropy)
			}
		}()
	}
	wg.Wait()
	return c
}

func BenchmarkULIDMustParse(b *testing.B) {
	ulidStr := ulidgen.MustNew().String()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ulid.MustParse(ulidStr)
	}
}
