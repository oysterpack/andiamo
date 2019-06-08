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
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

// Comp represents an application component.
type Comp struct {
	Desc
	Options []option.Option
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

// AppOptions returns component's application options
func (c *Comp) AppOptions() []fx.Option {
	options := make([]fx.Option, len(c.Options))
	for i, opt := range c.Options {
		options[i] = opt.Option
	}
	return options
}
