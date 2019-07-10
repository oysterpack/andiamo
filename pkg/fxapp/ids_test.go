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

package fxapp_test

import (
	"fmt"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"os"
	"strings"
	"testing"
)

func setenv(key, value string) {
	e := os.Setenv(fmt.Sprintf("%s_%s", fxapp.EnvconfigPrefix, strings.ToUpper(key)), value)
	if e != nil {
		panic(e)
	}
}

func unsetenv(keys ...string) {
	for _, key := range keys {
		e := os.Unsetenv(fmt.Sprintf("%s_%s", fxapp.EnvconfigPrefix, strings.ToUpper(key)))
		if e != nil {
			panic(e)
		}
	}
}

func TestLoadIDsFromEnv(t *testing.T) {
	keys := []string{"ID", "NAME", "VERSION", "RELEASE_ID"}
	unsetenv(keys...)
	defer unsetenv(keys...)

	id, releaseID, e := fxapp.LoadIDsFromEnv()
	switch {
	case e == nil:
		t.Error("*** should have failed to load IDs")
	default:
		t.Log(e)
	}

	ulidID := ulidgen.MustNew()
	ulidReleaseID := ulidgen.MustNew()
	setenv("ID", ulidID.String())
	setenv("NAME", "foo")
	setenv("VERSION", "0.1.0")
	setenv("RELEASE_ID", ulidReleaseID.String())

	id, releaseID, e = fxapp.LoadIDsFromEnv()

	switch {
	case e != nil:
		t.Errorf("*** desc failed to load from env: %v", e)
	default:
		t.Log(id, releaseID)
		if fxapp.ID(ulidID) != id {
			t.Error("*** ID did not match")
		}
		if fxapp.ReleaseID(ulidReleaseID) != releaseID {
			t.Error("*** ReleaseID did not match")
		}
	}

	checkLoadDescFromEnvFailed := func(e error) {
		switch {
		case e == nil:
			t.Error("*** desc should have failed to load")
		default:
			t.Log(e)
		}
	}

	setenv("ID", "INVALID")
	_, _, e = fxapp.LoadIDsFromEnv()
	checkLoadDescFromEnvFailed(e)

	setenv("ID", ulidgen.MustNew().String())
	setenv("RELEASE_ID", "INVALID")
	_, _, e = fxapp.LoadIDsFromEnv()
	checkLoadDescFromEnvFailed(e)
}
