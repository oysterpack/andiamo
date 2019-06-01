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
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
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

func BenchmarkMonotonicULIDGenerator(b *testing.B) {
	newULID := ulidgen.MonotonicULIDGenerator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newULID()
	}
}

func BenchmarkRandomULIDGenerator(b *testing.B) {
	newULID := ulidgen.RandomULIDGenerator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newULID()
	}
}
