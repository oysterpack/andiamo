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

// apptest package is used to support testing
package apptest

import (
	"fmt"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"os"
	"strings"
)

type Key string

const (
	ID            = Key("ID")
	NAME          = Key("NAME")
	VERSION       = Key("VERSION")
	RELEASE_ID    = Key("RELEASE_ID")
	START_TIMEOUT = Key("START_TIMEOUT")
	STOP_TIMEOUT  = Key("STOP_TIMEOUT")
)

// Setenv prefixes the key with "APPX12" and then sets the value of the environment variable named by the prefixed key.
func Setenv(key Key, value string) {
	if err := os.Setenv(prefix(key), value); err != nil {
		panic(err)
	}
}

// Unsetenv prefixes the key with "APPX12" and then tries to unset the env var
func Unsetenv(key Key) {
	if err := os.Unsetenv(prefix(key)); err != nil {
		panic(err)
	}
}

func ClearAppEnvSettings() {
	Unsetenv(ID)
	Unsetenv(VERSION)
	Unsetenv(NAME)
	Unsetenv(RELEASE_ID)
	Unsetenv(START_TIMEOUT)
	Unsetenv(STOP_TIMEOUT)
}

// Getenv prefixes the key with "APP12X" and then retrieves the value of the environment variable named by the prefixed key.
// If the variable is present in the environment the value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will be false.
func LookupEnv(key Key) (string, bool) {
	return os.LookupEnv(prefix(key))
}

// prefixes the specified key
func prefix(key Key) string {
	return fmt.Sprintf("%s_%s", app.ENV_PREFIX, strings.ToUpper(string(key)))
}
