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

package comp

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
)

// Comp represents an application component.
type Comp struct {
	ID   ulid.ULID
	Name string
	*semver.Version
	app.Package
}

// MustNew constructs a new FindByID instance.
func MustNew(id, name, version string, p app.Package) *Comp {
	return &Comp{
		ID:      ulid.MustParse(id),
		Name:    name,
		Version: semver.MustParse(version),
		Package: p,
	}
}

func (c *Comp) String() string {
	return fmt.Sprintf("FindByID{ID=%s, Name=%s, Version=%s, Package=%s}", c.ID, c.Name, c.Version, c.Package)
}

// Logger adds the comp's package and name to the specified logger
//
// NOTE: if the logger already has the package or component fields, then they will be duplicated.
func (c *Comp) Logger(l *zerolog.Logger) *zerolog.Logger {
	return logging.ComponentLogger(logging.PackageLogger(l, c.Package), c.Name)
}
