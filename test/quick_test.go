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
	"io"
	"log"
	"testing"
)

type user struct {
	Name string `json:"n"`
	Age  uint
}

type users map[string]user

func (self users) Get(id string) (u user, ok bool) {
	u, ok = self[id]
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

	io.EOF = errors.New("EOF!!!")
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
