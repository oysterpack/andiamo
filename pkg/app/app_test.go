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
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"testing"
)

// ID implements the Stringer interface
func TestID_String(t *testing.T) {
	t.Parallel()
	appID := app.NewID()
	id := ulid.ULID(appID)

	t.Logf("%s", appID)
	if id.String() != appID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestReleaseID_String(t *testing.T) {
	t.Parallel()
	releaseID := app.NewReleaseID()
	id := ulid.ULID(releaseID)

	t.Logf("%s", releaseID)
	if id.String() != releaseID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestInstanceID_String(t *testing.T) {
	t.Parallel()
	instanceID := app.NewInstanceID()
	id := ulid.ULID(instanceID)

	t.Logf("%s", instanceID)
	if id.String() != instanceID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestVersion_String(t *testing.T) {
	t.Parallel()
	var appVer app.Version
	if e := appVer.Decode("1.2.3"); e != nil {
		t.Errorf("Failed to decode version: %v", e)
	}

	if appVer.String() != "1.2.3" {
		t.Errorf("Version did not match: %s", &appVer)
	}
}

func TestMustParseVersion(t *testing.T) {

	t.Run("valid version", func(t *testing.T) {
		version := app.MustParseVersion("0.1.0")
		if version.String() != "0.1.0" {
			t.Errorf("parsed version did not match: %v", version)
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		defer func() {
			e := recover()
			if e == nil {
				t.Fatal("*** version should have failed to parse")
			}
			t.Logf("error: %v", e)
		}()
		app.MustParseVersion("---")
	})

}
