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
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"sort"
	"sync"
)

// Registry related error classes
var (
	IDAlreadyRegisteredErrClass   = err.NewDesc("01DCQ59QEKTE9ATFJM5SGF168C", "IDAlreadyregisteredErr", "a component with the same ID is already registered")
	NameAlreadyRegisteredErrClass = err.NewDesc("01DCQF3MT1EPJZQJRC9DCVBK4M", "NameAlreadyregisteredErr", "a component with the same name is already registered")
)

// Registry related errors
var (
	IDAlreadyRegisteredErr   = err.New(IDAlreadyRegisteredErrClass, "01DCQ5CX2SCKAZZD354GAF30T5")
	NameAlreadyRegisteredErr = err.New(NameAlreadyRegisteredErrClass, "01DCQF4NP9J9MKM6Y2JD4SFP0G")
)

// Registry is used as an application component registry.
//
// Business Rules
//
//	1. Comp.ID is unique
//  2. Comp.Name is unique within the application scope.
//  3. Comp.Name max len = 30
type Registry struct {
	m     sync.RWMutex
	comps []*Comp
}

// NewRegistry constructs a new application component registry
func NewRegistry() *Registry {
	return &Registry{
		comps: make([]*Comp, 0, 10),
	}
}

// Register tries to register the component.
//
// It may fail because of the following errors:
// - IDAlreadyRegisteredErr
// - NameAlreadyRegisteredErr
func (r *Registry) Register(c *Comp) error {
	r.m.Lock()
	defer r.m.Unlock()
	for _, registeredComp := range r.comps {
		if c.ID == registeredComp.ID {
			return IDAlreadyRegisteredErr.New()
		}
		if c.Name == registeredComp.Name {
			return NameAlreadyRegisteredErr.New()
		}
	}
	r.comps = append(r.comps, c)
	sort.Slice(r.comps, func(i, j int) bool {
		return r.comps[i].Name < r.comps[j].Name
	})
	return nil
}

// FindByID looks up the Comp by ID
func (r *Registry) FindByID(id ulid.ULID) *Comp {
	r.m.RLock()
	defer r.m.RUnlock()
	for _, registeredComp := range r.comps {
		if registeredComp.ID == id {
			return registeredComp
		}
	}
	return nil
}

// FindByName looksup the Comp by ID
func (r *Registry) FindByName(name string) *Comp {
	r.m.RLock()
	defer r.m.RUnlock()
	for _, registeredComp := range r.comps {
		if registeredComp.Name == name {
			return registeredComp
		}
	}
	return nil
}

// Comps returns all registered components
func (r *Registry) Comps() []*Comp {
	r.m.RLock()
	defer r.m.RUnlock()
	comps := make([]*Comp, len(r.comps))
	copy(comps, r.comps)
	return comps
}
