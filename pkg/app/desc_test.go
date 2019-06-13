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

package app_test

import (
	"crypto/rand"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"strings"
	"testing"
	"time"
)

func TestDescConstruction(t *testing.T) {
	t.Parallel()
	v := app.Version(*semver.MustParse("0.0.1"))
	desc := &app.Desc{
		ID:        app.ID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)),
		Name:      app.Name("foo"),
		Version:   &v,
		ReleaseID: app.ReleaseID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)),
	}
	t.Logf("%s", desc)
}

func TestLoadDesc(t *testing.T) {
	t.Run("all required env vars are set and are valid", testLoadDesc_ValidScenario)

	t.Run("required fields are missing", testLoadDesc_RequiredFieldsMissing)

	t.Run("using invalid ULIDs", testLoadDesc_UsingInvalidULIDs)

	t.Run("using invalid Version", testLoadDesc_UsingInvalidVersion)
}

func testLoadDesc_UsingInvalidVersion(t *testing.T) {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	version := semver.MustParse("0.0.1")
	name := app.Name("foobar")

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.Name, string(name))
	apptest.Setenv(apptest.ReleaseID, releaseID.String())
	apptest.Setenv(apptest.Version, version.String())

	// Given we are starting from a valid config state
	_, err := app.LoadDesc()
	if err != nil {
		t.Fatal(err)
	}

	// Given Version is invalid
	apptest.Setenv(apptest.Version, "---")
	// When the Desc is loaded from the env
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("*** app.Desc should have failed to load because Version should be invalid")
	} else {
		t.Log(err)
	}
}

func testLoadDesc_UsingInvalidULIDs(t *testing.T) {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	version := semver.MustParse("0.0.1")
	name := app.Name("foobar")

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.Name, string(name))
	apptest.Setenv(apptest.ReleaseID, releaseID.String())
	apptest.Setenv(apptest.Version, version.String())

	// Given we are starting from a valid config state
	_, err := app.LoadDesc()
	if err != nil {
		t.Fatal(err)
	}

	// Given ID is invalid
	apptest.Setenv(apptest.ID, "---")
	// When the Desc is loaded from the env
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because ID should be invalid")
	} else {
		t.Log(err)
	}
	// reset ID
	apptest.Setenv(apptest.ID, id.String())

	// Given ReleaseID is invalid
	apptest.Setenv(apptest.ReleaseID, "---")
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because releaseID should be invalid")
	} else {
		t.Log(err)
	}
}

func testLoadDesc_RequiredFieldsMissing(t *testing.T) {
	apptest.ClearAppEnvSettings()

	// When the Desc is loaded from the env
	_, err := app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because required env vars were not defined")
	} else {
		t.Log(err)
	}

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	apptest.Setenv(apptest.ID, id.String())
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because required env vars were not defined")
	} else {
		t.Log(err)
	}

	name := app.Name("foobar")
	apptest.Setenv(apptest.Name, string(name))
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because required env vars were not defined")
	} else {
		t.Log(err)
	}

	version := semver.MustParse("0.0.1")
	apptest.Setenv(apptest.Version, version.String())
	_, err = app.LoadDesc()
	if err == nil {
		t.Error("app.Desc should have failed to load because required env vars were not defined")
	} else {
		t.Log(err)
	}
}

func testLoadDesc_ValidScenario(t *testing.T) {
	apptest.ClearAppEnvSettings()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	version := semver.MustParse("0.0.1")
	name := app.Name("foobar")

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.Name, string(name))
	apptest.Setenv(apptest.ReleaseID, releaseID.String())
	apptest.Setenv(apptest.Version, version.String())

	// When the Desc is loaded from the env
	desc, err := app.LoadDesc()

	// Then it is loaded successfully
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", &desc)

	// And its properties match what was specified in the env
	if desc.ID != app.ID(id) {
		t.Errorf("ID did not match: %s != %s", desc.ID, id)
	}
	if desc.Name != name {
		t.Errorf("Name did not match: %s != %s", desc.Name, name)
	}
	if !(*semver.Version)(desc.Version).Equal(version) {
		t.Errorf("Version did not match: %s != %s", (*semver.Version)(desc.Version), version)
	}
	if desc.ReleaseID != app.ReleaseID(releaseID) {
		t.Errorf("ReleaseID did not match: %s != %s", desc.ReleaseID, releaseID)
	}
}

func TestDesc_Validate(t *testing.T) {
	var desc app.Desc

	e := desc.Validate()
	if e == nil {
		t.Fatal("*** desc should not be valid because it is the zero value")
	}
	t.Log(e)

	desc.Name = "foo"
	e = desc.Validate()
	if e == nil {
		t.Fatal("*** desc should not be valid")
	}
	t.Log(e)

	version := semver.MustParse("0.1.0")
	appVersion := app.Version(*version)
	desc.Version = &appVersion
	e = desc.Validate()
	if e == nil {
		t.Fatal("*** desc should not be valid")
	}
	t.Log(e)

	desc.ID = app.ID(ulidgen.MustNew())
	e = desc.Validate()
	if e == nil {
		t.Fatal("*** desc should not be valid")
	}
	t.Log(e)

	desc.ReleaseID = app.ReleaseID(ulidgen.MustNew())
	if e := desc.Validate(); e != nil {
		t.Fatalf("*** desc should be valid: %v", e)
	}
	t.Log(desc)

	t.Run("0.0.0 version should not be valid", func(t *testing.T) {
		testDescZeroVersionNotValid(t, desc)
	})

	t.Run("invalid names", func(t *testing.T) {
		testDescInvalidNames(t, desc)
	})

}

func testDescZeroVersionNotValid(t *testing.T, desc app.Desc) {
	v := semver.Version{}
	zeroVersion := app.Version(v)
	desc.Version = &zeroVersion

	e := desc.Validate()
	if e == nil {
		t.Fatalf("*** desc should not be valid: %v", desc)
	}
	t.Log(e)
}

func testDescInvalidNames(t *testing.T, desc app.Desc) {
	desc.Name = "f" // too short
	e := desc.Validate()
	if e == nil {
		t.Errorf("*** desc.Name should not be valid: %q", desc.Name)
	} else {
		t.Log(e)
	}

	desc.Name = "ff" // too short
	e = desc.Validate()
	if e == nil {
		t.Errorf("*** desc.Name should not be valid: %q", desc.Name)
	} else {
		t.Log(e)
	}

	desc.Name = "fff"
	e = desc.Validate()
	if e != nil {
		t.Errorf("*** desc.Name should be valid: %q : %v", desc.Name, e)
	}

	desc.Name = "fff_ggg-kkk"
	e = desc.Validate()
	if e != nil {
		t.Errorf("*** desc.Name should be valid: %q : %v", desc.Name, e)
	}

	desc.Name = app.Name(strings.Repeat("f", 51)) //too long
	e = desc.Validate()
	if e == nil {
		t.Errorf("*** desc.Name should not be valid: %q", desc.Name)
	} else {
		t.Log(e)
	}

	desc.Name = "1abc" // begins with number
	e = desc.Validate()
	if e == nil {
		t.Errorf("*** desc.Name should not be valid: %q", desc.Name)
	} else {
		t.Log(e)
	}

	desc.Name = "abc@1" // contains invalid char
	e = desc.Validate()
	if e == nil {
		t.Errorf("*** desc.Name should not be valid: %q", desc.Name)
	} else {
		t.Log(e)
	}
}
