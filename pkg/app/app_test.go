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

package app

import (
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"testing"
)

func TestID_String(t *testing.T) {
	id, err := ulid.Parse("01DC7RT5S3HF8G25TBT44J083Z")
	if err != nil {
		t.Fatal(err)
	}

	appID := ID(id)

	if "01DC7RT5S3HF8G25TBT44J083Z" != appID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestReleaseID_String(t *testing.T) {
	id, err := ulid.Parse("01DC7RT5S3HF8G25TBT44J083Z")
	if err != nil {
		t.Fatal(err)
	}

	appID := ReleaseID(id)

	if "01DC7RT5S3HF8G25TBT44J083Z" != appID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestInstanceID_String(t *testing.T) {
	id, err := ulid.Parse("01DC7RT5S3HF8G25TBT44J083Z")
	if err != nil {
		t.Fatal(err)
	}

	appID := InstanceID(id)

	if "01DC7RT5S3HF8G25TBT44J083Z" != appID.String() {
		t.Errorf("ID did not match ULID: %s", id)
	}
}

func TestVersion_String(t *testing.T) {
	v := semver.MustParse("1.2.3")
	appVer := Version(*v)

	if appVer.String() != "1.2.3" {
		t.Errorf("Version did not match: %s", &appVer)
	}

}
