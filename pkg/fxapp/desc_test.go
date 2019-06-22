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
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"os"
	"strings"
	"testing"
)

func TestDesc_Build(t *testing.T) {
	id := ulidgen.MustNew()
	releaseID := ulidgen.MustNew()
	desc, err := fxapp.NewDescBuilder().
		SetID(id).
		SetName("foo").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(releaseID).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** desc build error: %v", err)
	default:
		t.Log(desc.String())

		if desc.Name() != "foo" {
			t.Errorf("*** name did not match: %v", desc.Name())
		}
		if desc.ID() != id {
			t.Error("*** ID did not match")
		}
		if desc.ReleaseID() != releaseID {
			t.Error("*** ReleaseID did not match")
		}
		if !desc.Version().Equal(semver.MustParse("0.1.0")) {
			t.Errorf("*** version did not match: %v", desc.Version())
		}
	}
}

func TestDesc_ZeroID(t *testing.T) {
	_, err := fxapp.NewDescBuilder().
		SetID(ulid.ULID{}). // zero value is invalid
		SetName("foo").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulidgen.MustNew()).
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build because the ID is the zero value")
	default:
		t.Log(err)
	}
}

func TestDesc_ZeroReleaseID(t *testing.T) {
	_, err := fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName("foo").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulid.ULID{}). // zero value is invalid
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build because the ReleaseID is the zero value")
	default:
		t.Log(err)
	}
}

func TestDesc_NilVersion(t *testing.T) {
	_, err := fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName("foo").
		SetVersion(nil).
		SetReleaseID(ulidgen.MustNew()). // zero value is invalid
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build because the Version is nil")
	default:
		t.Log(err)
	}
}

func TestDesc_InvalidName(t *testing.T) {
	_, err := fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName("1foo").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulidgen.MustNew()). // zero value is invalid
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build name must start with an alpha char")
	default:
		t.Log(err)
	}

	_, err = fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName("foo:2323").
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulidgen.MustNew()). // zero value is invalid
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build name contains an invalid char")
	default:
		t.Log(err)
	}

	_, err = fxapp.NewDescBuilder().
		SetID(ulidgen.MustNew()).
		SetName(strings.Repeat("a", 51)).
		SetVersion(semver.MustParse("0.1.0")).
		SetReleaseID(ulidgen.MustNew()). // zero value is invalid
		Build()

	switch {
	case err == nil:
		t.Error("*** desc should have failed to build name max length = 50")
	default:
		t.Log(err)
	}
}

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

func TestLoadDescFromEnv(t *testing.T) {
	keys := []string{"ID", "NAME", "VERSION", "RELEASE_ID"}
	unsetenv(keys...)
	defer unsetenv(keys...)

	desc, e := fxapp.LoadDescFromEnv()
	switch {
	case e == nil:
		t.Error("*** desc should have failed to load")
	default:
		t.Log(e)
	}

	id := ulidgen.MustNew()
	releaseID := ulidgen.MustNew()
	setenv("ID", id.String())
	setenv("NAME", "foo")
	setenv("VERSION", "0.1.0")
	setenv("RELEASE_ID", releaseID.String())

	desc, e = fxapp.LoadDescFromEnv()

	switch {
	case e != nil:
		t.Errorf("*** desc failed to load from env: %v", e)
	default:
		t.Log(desc)
		if desc.Name() != "foo" {
			t.Error("*** name did not match")
		}
		if !desc.Version().Equal(semver.MustParse("0.1.0")) {
			t.Error("*** version did not match")
		}
		if desc.ID() != id {
			t.Error("*** ID did not match")
		}
		if desc.ReleaseID() != releaseID {
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
	_, e = fxapp.LoadDescFromEnv()
	checkLoadDescFromEnvFailed(e)

	setenv("ID", id.String())
	setenv("VERSION", "INVALID")
	_, e = fxapp.LoadDescFromEnv()
	checkLoadDescFromEnvFailed(e)

	setenv("VERSION", "0.1.0")
	setenv("RELEASE_ID", "INVALID")
	_, e = fxapp.LoadDescFromEnv()
	checkLoadDescFromEnvFailed(e)
}
