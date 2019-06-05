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

package err

import (
	"github.com/oklog/ulid"
	"sync"
)

// Registry related errors
var (
	ErrRegistryConflict           = NewDesc("01DCMMN0342H9FBMVHZ4MGWS2J", "ErrRegistryConflict", "an Err with the same SrcID but different Desc.ID is already registered")
	ErrRegistryConflictOnRegister = New(ErrRegistryConflict, "01DCMMT8M89SE1JX3SGNXZZMST")
)

// Registry is used to register application errors. It is good to know what types of application errors can occur upfront.
type Registry struct {
	m sync.RWMutex
	// Err.SrcID -> *Err
	errs map[ulid.ULID]*Err
}

// NewRegistry is the registry constructor. It automatically registers ErrRegistryConflictOnRegister.
func NewRegistry() *Registry {
	registry := &Registry{
		errs: make(map[ulid.ULID]*Err),
	}
	registry.Register(ErrRegistryConflictOnRegister)
	return registry
}

// Register registers the specified errors, if not already registered. Err.SrcID is used as the registry key.
//
// Returns an error if an Err is already registered with the same Err.SrcID, but with a different Desc.ID.
func (r *Registry) Register(errs ...*Err) error {
	r.m.Lock()
	defer r.m.Unlock()
	for _, e := range errs {
		if registeredErr := r.errs[e.SrcID]; registeredErr != nil {
			if registeredErr.ID != e.ID {
				return ErrRegistryConflictOnRegister.New()
			}
		} else {
			r.errs[e.SrcID] = e
		}
	}
	return nil
}

// Registered returns true if an Err is registered for the specified Err.SrcID
func (r *Registry) Registered(srcID ulid.ULID) bool {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.errs[srcID] != nil
}

// Errs returns all registered Err(s)
func (r *Registry) Errs() []*Err {
	r.m.RLock()
	defer r.m.RUnlock()
	errs := make([]*Err, 0, len(r.errs))
	for _, e := range r.errs {
		errs = append(errs, e)
	}
	return errs
}

// Size returns the number of registered Err(s)
func (r *Registry) Size() int {
	r.m.RLock()
	defer r.m.RUnlock()
	return len(r.errs)
}

// Descs returns the number of registered error Desc(s)
func (r *Registry) Descs() map[ulid.ULID]*Desc {
	r.m.RLock()
	defer r.m.RUnlock()
	descs := make(map[ulid.ULID]*Desc)
	for _, e := range r.errs {
		descs[e.ID] = e.Desc
	}
	return descs
}
