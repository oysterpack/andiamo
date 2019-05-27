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
	"github.com/oysterpack/partire-k8s/pkg/app/apptest"
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

func TestLoadDescFromEnv(t *testing.T) {
	// Given all of the required environment variables are set
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	version := semver.MustParse("0.0.1")
	name := app.Name("foobar")

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.NAME, string(name))
	apptest.Setenv(apptest.RELEASE_ID, releaseID.String())
	apptest.Setenv(apptest.VERSION, version.String())

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
