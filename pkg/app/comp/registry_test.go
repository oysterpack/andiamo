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

package comp_test

import (
	"fmt"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"testing"
)

const (
	Package app.Package = "github.com/oysterpack/partire-k8s/pkg/comp_test"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	t.Run("register comp", testRegistry_Register_ValidScenario)

	t.Run("registering comp using ID that is already registered", testRegistry_Register_CompIDAlreadyRegistered)

	t.Run("registering comp using a name that is already registered", testRegistry_Register_CompNameAlreadyRegistered)

}

func testRegistry_Register_CompIDAlreadyRegistered(t *testing.T) {
	registry := comp.NewRegistry()
	c := NewComp(t, "01DCQ2NV56WAQD6SDDECZPGB2T", "foo", "0.0.1")
	if e := registry.Register(c); e != nil {
		t.Fatalf("Failed to register component: %v", e)
	}
	if e := registry.Register(c); e == nil {
		t.Error("registration should have failed")
	} else {
		t.Log(e)
	}
}

func testRegistry_Register_CompNameAlreadyRegistered(t *testing.T) {
	registry := comp.NewRegistry()
	// Given a comp is registered
	c1 := NewComp(t, ulidgen.MustNew().String(), "foo", "0.0.1")
	if e := registry.Register(c1); e != nil {
		t.Fatalf("Failed to register component: %v", e)
	}
	// When a comp is trying to be registered using a name that is already registered
	c2 := NewComp(t, ulidgen.MustNew().String(), c1.Name, "0.0.1")
	if e := registry.Register(c2); e == nil {
		t.Errorf("registration should have failed: %v", registry.FindByID(c1.ID))
	} else {
		// Then registration will fail
		t.Log(e)
	}
	// When a unique comp is registered
	c3 := NewComp(t, ulidgen.MustNew().String(), ulidgen.MustNew().String(), "0.0.1")
	// Then it will successfully register
	if e := registry.Register(c3); e != nil {
		t.Fatalf("Failed to register component: %v", e)
	}
}

func testRegistry_Register_ValidScenario(t *testing.T) {
	registry := comp.NewRegistry()
	// Given a component is registered
	foo := NewComp(t, "01DCQ2NV56WAQD6SDDECZPGB2T", "foo", "0.0.1")
	bar := NewComp(t, "01DCQ6ZSVE7D7G6VRVDAECJJWF", "bar", "0.0.1")
	for _, c := range []*comp.Comp{foo, bar} {
		if e := registry.Register(c); e != nil {
			t.Fatalf("Failed to register component: %v", e)
		}
	}

	for _, c := range []*comp.Comp{foo, bar} {
		// Then it can be retrieved from the registry by ID
		if cc := registry.FindByID(c.ID); cc == nil {
			t.Error("component was not found by the registry")
		} else {
			t.Log(cc)
			if c.ID != cc.ID {
				t.Errorf("The comp returned has a different ID: %v != %v", c.ID, cc.ID)
			}
		}

		// And it can be retrieved from the registry by name
		if cc := registry.FindByName(c.Name); cc == nil {
			t.Error("component was not found by the registry")
		} else {
			t.Log(cc)
			if c.ID != cc.ID && c.Name == cc.Name {
				t.Errorf("The comp returned has a different ID: %v != %v", c.ID, cc.ID)
			}
		}
	}
}

func TestRegistry_FindByID_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("lookup comp using unregistered ID", func(t *testing.T) {
		registry := comp.NewRegistry()
		// When a random ULID is used to retrieve a component
		// Then nil should be returned
		if c := registry.FindByID(ulidgen.MustNew()); c != nil {
			t.Error("component should not have been returned")
		}
	})

	t.Run("lookup comp using unregistered name", func(t *testing.T) {
		registry := comp.NewRegistry()
		// When a random ULID is used to retrieve a component
		// Then nil should be returned
		if c := registry.FindByName(ulidgen.MustNew().String()); c != nil {
			t.Error("component should not have been returned")
		}
	})
}

func TestRegistry_Comps(t *testing.T) {
	t.Parallel()
	registry := comp.NewRegistry()
	comps := registry.Comps()
	if len(comps) > 0 {
		t.Fatalf("registry should be empty")
	}

	const count = 5
	registeredComps := make([]*comp.Comp, 0, count)
	for i := 0; i < count; i++ {
		c := NewComp(t, ulidgen.MustNew().String(), fmt.Sprintf("c%d", i), "0.0.1")
		if e := registry.Register(c); e != nil {
			t.Fatalf("failed to register component: %v : %v : %v", e, c, registry.Comps())
		}
		registeredComps = append(registeredComps, c)
	}

	comps = registry.Comps()
	t.Log(comps)
	if len(comps) != count {
		t.Errorf("The number of registered comps does not match: %d", len(comps))
	}
}
