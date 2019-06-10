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
	"go.uber.org/fx"
)

// CompRegistry is used as the fx value group name for components.
// Components will tag themselves with this group in order to auto-registered with the comp.Registry.
const CompRegistry = "comp.Registry"

// Comp represents an application component. Components are constructed using component descriptors.
//
// see - Desc.MustNewComp()
type Comp struct {
	*Desc
	Options []option.Option
}

func (c *Comp) String() string {
	return fmt.Sprintf("Comp{ID=%s, Name=%s, Version=%s, Package=%s, Options=%v}", c.ID, c.Name, c.Version, c.Package, c.Options)
}

// FxOptions returns component's application options
func (c *Comp) FxOptions() fx.Option {
	options := make([]fx.Option, len(c.Options), len(c.Options)+1)
	for i, opt := range c.Options {
		options[i] = opt.Option
	}
	// provide itself, which will register the component
	options = append(options, fx.Provide(fx.Annotated{
		Group:  CompRegistry,
		Target: func() *Comp { return c },
	}))
	return fx.Options(options...)
}

// MustNewComp builds a new component using the specified options
//
// Panics if the options don't match the options defined by the component descriptor. The order of the options doesn't matter.
// The options must match on the option types declared by the descriptor. They will be sorted according to the order they
// are listed in the descriptor
func MustNewComp(desc *Desc, options ...option.Option) *Comp {
	return desc.MustNewComp(options...)
}
