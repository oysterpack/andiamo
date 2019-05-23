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

package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

type user struct {
	Name string `json:"n"`
	Age  uint
}

type users map[string]user

func (us users) Get(id string) (u user, ok bool) {
	u, ok = us[id]
	return
}

func TestItWorks(t *testing.T) {

	type IntSet []uint64

	type Celsius float64

	log.Println("It works!!!")

	var a []int
	log.Println(a)
	for i := 0; i < 10; i++ {
		a = append(a, i)
	}
	log.Println(a)

	db := users(make(map[string]user, 10))
	db["a"] = user{"Alfio", 46}
	db["b"] = user{"Bella", 16}

	if u, ok := db.Get("a"); ok {
		u.Age++
		log.Println(u)
	}
	if u, err := json.Marshal(db["a"]); err == nil {
		log.Printf("json(%v) => %s\n", db["a"], u)
	}

	io.EOF = errors.New("boom: EOF")
	log.Println(io.EOF)

	bits := IntSet{0, 1, 2}
	log.Println(bits)

	var temp Celsius = 101
	log.Println(temp)

	compute := func() int {
		return 5
	}

	switch a := compute(); {
	case a <= 10:
		log.Printf("%v <= 10\n", a)
	case a <= 20:
		log.Printf("%v <= 20\n", a)
	default:
		log.Printf("%v", a)
	}

}

func TestNewFoo(t *testing.T) {
	foo := NewFoo("bar")
	log.Printf("foo: %[1]v: %[1]s", &foo)
}

func TestPassingPointerOverChannels(t *testing.T) {
	counter := 0
	c := make(chan *int)
	go func() {
		c <- &counter
	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter := <-c
		*counter++
	}()
	wg.Wait()
	log.Printf("counter = %d", counter)
	if counter != 1 {
		log.Fatalf("expected counter to have been incremented because a pointer was passed via the channel")
	}
}

func TestPassingValueOverChannels(t *testing.T) {
	counter := 0
	c := make(chan int)
	go func() {
		c <- counter
	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter := <-c
		counter++
	}()
	wg.Wait()
	log.Printf("counter = %d", counter)
	if counter != 0 {
		log.Fatalf("expected counter to have not been incremented by above go routine because a copy of the counter was sent")
	}
}

// based on the bench results, increasing the channel buffer size increases the message throughput
//
// BenchmarkUnbufferedChanMessaging-8       5000000               358 ns/op
// BenchmarkBuffered1ChanMessaging-8        5000000               305 ns/op
// BenchmarkBuffered2ChanMessaging-8        5000000               261 ns/op
// BenchmarkBuffered8ChanMessaging-8       10000000               162 ns/op
// BenchmarkBuffered16ChanMessaging-8      10000000               143 ns/op
// BenchmarkBuffered32ChanMessaging-8      10000000               132 ns/op
// BenchmarkBuffered64ChanMessaging-8      20000000               114 ns/op
func benchBufferedChanMessaging(b *testing.B, chanSize int) {
	c := make(chan struct{}, chanSize)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// drain the channel
		for range c {
		}
		wg.Done()
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c <- struct{}{}
	}
	close(c)
	wg.Wait()
}

func BenchmarkUnbufferedChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 0)
}

func BenchmarkBuffered1ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 1)
}

func BenchmarkBuffered2ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 2)
}

func BenchmarkBuffered8ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 8)
}

func BenchmarkBuffered16ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 16)
}

func BenchmarkBuffered32ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 32)
}

func BenchmarkBuffered64ChanMessaging(b *testing.B) {
	benchBufferedChanMessaging(b, 64)
}

func TestFoo_Name(t *testing.T) {
	bar := NewFoo("bar")
	log.Printf("bar.name = %q", bar.Name())
}

func TestGenericMap(t *testing.T) {
	type foo int

	type bar int

	const (
		FOO_1 foo = iota
		FOO_2
		FOO_3
	)

	const (
		BAR_1 bar = iota
		BAR_2
		BAR_3
	)

	data := make(map[interface{}]interface{})

	data[FOO_1] = "foo 1"
	data[BAR_1] = "bar 1"

	log.Println(data)

	if data[FOO_1] != "foo 1" {
		log.Fatalf("invalid entry for FOO_1: %v", data[FOO_1])
	}

	if data[BAR_1] != "bar 1" {
		log.Fatalf("invalid entry for BAR_1: %v", data[BAR_1])
	}

}

type UID func() string

func uid() string {
	return fmt.Sprintf("%v", time.Now())
}

type UIDFactory struct{}

func (u UIDFactory) uid() string {
	return fmt.Sprintf("%v", time.Now())
}

func TestFuncType(t *testing.T) {
	u := UIDFactory{}
	var f UID = u.uid
	log.Printf("uid = %v", f())
}

type Time func() time.Time

func now() time.Time {
	return time.Now()
}

type F func()

func foo() {}

type Bar struct {
}

func (b *Bar) foo() {
}

type FooFunc interface {
	foo()
}

func BenchmarkFuncDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now()
	}
}

func BenchmarkFuncPtr(b *testing.B) {
	var f Time = now
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

func BenchmarkFuncDirect2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		foo()
	}
}

// invoking a function via a pointer adds ~1.6 ns
func BenchmarkFuncPtr2(b *testing.B) {
	var f F = foo
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

// invoking a function via a pointer adds ~1.6 ns
func BenchmarkMethodFuncPtr(b *testing.B) {
	bar := &Bar{}
	var f F = bar.foo
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

// invoking a function via a pointer adds ~1.6 ns
// invoking the function pointer through the interface add an additional ~1.1 ns
func BenchmarkInterfaceFuncPtr(b *testing.B) {
	var bar FooFunc = &Bar{}
	var f F = bar.foo
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

// invoking a function via a pointer adds ~1.6 ns
func BenchmarkStructFunc(b *testing.B) {
	var bar = Bar{}
	var f F = bar.foo
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f()
	}
}

func TestSwitch(t *testing.T) {
	switch a := 10; {
	case a <= 3:
		log.Println("MATCHED 3 !!!")
	case a == 10:
		log.Println("MATCHED 10 !!!")
	}

	for i, value := range []int{10, 20, 30} {
		log.Println(i, value)
	}

	data := []int{}
	log.Printf("%v : len(data) = %d, cap(data) = %d", data, len(data), cap(data))
	data = append(data, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
	log.Printf("%v : len(data) = %d, cap(data) = %d", data, len(data), cap(data))
	log.Printf("%v", map[string]int{
		"a": 1,
		"b": 2,
	})

	type User struct {
		fname string
		lname string
	}

	user := new(User)
	log.Printf("%#v", user)
	log.Printf("%#v", [...]int{1, 2, 3})
}

func TestGoRoutines(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Microsecond)
			log.Println(index)
		}()
	}

	wg.Wait()
}

func TestSelectChan(t *testing.T) {
	c1 := make(chan struct{}, 1)
	c2 := make(chan struct{}, 1)

	const loopCount = 10
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < loopCount; i++ {
			t.Logf("i = %d\n", i)
			c1 <- struct{}{}
			c2 <- struct{}{}
			if i == 0 {
				wg.Done()
			}
		}

		t.Log("closing channels")
		close(c1)
		close(c2)
		t.Log("closed channels")
	}()

	wg.Wait()
LOOP:
	for i := 1; ; i++ {
		select {
		case _, ok := <-c1:
			if !ok {
				t.Log("NO MESSAGES")
				break LOOP
			}
			t.Logf("received msg from c1 #%d\n", i)
		case _, ok := <-c2:
			if !ok {
				t.Log("NO MESSAGES")
				break LOOP
			}
			t.Logf("received msg from c2 #%d\n", i)
		}
	}

}

func TestRaceCondition(t *testing.T) {
	t.Run("data race", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				log.Printf("i = %d", i)
			}(i)
		}
		wg.Wait()
	})
}
