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
	"fmt"
	"github.com/oklog/ulid"
	"sync"
)

// Registry related errors
var (
	RegistryConflictErrClass = MustNewDesc("01DCMMN0342H9FBMVHZ4MGWS2J", "RegistryConflict", "an Err with the same SrcID but different Desc.ID is already registered")
	RegistryConflictErr      = New(RegistryConflictErrClass, "01DCMMT8M89SE1JX3SGNXZZMST")
)

// Registry is used to register application errors. It is good to know what types of application errors can occur upfront.
//
// Registry is safe to use concurrently.
type Registry struct {
	m sync.RWMutex
	// Err.SrcID -> *Err
	errs []*Err
}

// NewRegistry is the registry constructor. It automatically registers RegistryConflictErr.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register registers the specified errors, if not already registered. Err.SrcID is used as the registry key.
//
// Errors
//
// RegistryConflictErr - if an Err is already registered with the same Err.SrcID, but with a different Desc.ID
func (r *Registry) Register(errs ...*Err) error {
	r.m.Lock()
	defer r.m.Unlock()
	for _, e := range errs {
		if registeredErr := r.findBySrcID(e.SrcID); registeredErr != nil {
			// check that the registered error references the same error descriptor
			if registeredErr.ID != e.ID {
				return RegistryConflictErr.CausedBy(fmt.Errorf(""))
			}
		} else {
			r.errs = append(r.errs, e)
		}
	}
	return nil
}

// Registered returns true if an Err is registered for the specified Err.SrcID
func (r *Registry) Registered(srcID ulid.ULID) bool {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.registered(srcID)
}

// Registered returns true if an Err is registered for the specified Err.SrcID
func (r *Registry) registered(srcID ulid.ULID) bool {
	for _, e := range r.errs {
		if e.SrcID == srcID {
			return true
		}
	}
	return false
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

// Count returns the number of registered Err(s)
func (r *Registry) Count() int {
	r.m.RLock()
	defer r.m.RUnlock()
	return len(r.errs)
}

// Descs returns all registered error Desc(s)
func (r *Registry) Descs() map[ulid.ULID]*Desc {
	r.m.RLock()
	defer r.m.RUnlock()
	descs := make(map[ulid.ULID]*Desc)
	for _, e := range r.errs {
		descs[e.ID] = e.Desc
	}
	return descs
}

// Filter returns all errors that match the filter
func (r *Registry) Filter(filter func(*Err) bool) []*Err {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.filter(filter)
}

func (r *Registry) filter(filter func(*Err) bool) []*Err {
	var errs []*Err
	for _, e := range r.errs {
		if filter(e) {
			errs = append(errs, e)
		}
	}
	return errs
}

func (r *Registry) findBySrcID(srcID ulid.ULID) *Err {
	for _, e := range r.errs {
		if e.SrcID == srcID {
			return e
		}
	}
	return nil
}
