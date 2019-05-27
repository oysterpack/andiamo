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
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/ulid"
)

// Desc provides information to identify the application deployment
type Desc struct {
	ID        `required:"true"`
	Name      `required:"true"`
	Version   *Version `required:"true"`
	ReleaseID `required:"true" split_words:"true"`
}

func (d *Desc) String() string {
	return fmt.Sprintf("Desc{ID=%s, Name=%s, Version=%s, ReleaseID=%s}", ulid.ULID(d.ID), d.Name, (*semver.Version)(d.Version), ulid.ULID(d.ReleaseID))
}

// LoadDesc loads the app Desc from the system environment. The following env vars are required:
// - APPX12_ID
// - APPX12_NAME
// - APPX12_VERSION
// - APPX12_RELEASE_ID
func LoadDesc() (desc Desc, err error) {
	err = envconfig.Process(ENV_PREFIX, &desc)
	return
}
