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

package err_test

import (
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	// error tags
	const (
		DatabaseTag err.Tag = "db"
		DGraphTag   err.Tag = "dgraph"
	)

	// error descriptors
	var (
		InvalidRequestErr = err.NewDesc("01DC9HDP0X3R60GWDZZY18CVB8", "InvalidRequest", "Invalid request")

		DGraphQueryTimeoutErr = err.NewDesc(
			"01DCC447HWNM5MP7D4Z0DKK0SQ",
			"DatabaseTimeout",
			"query timeout",
			DGraphTag, DatabaseTag,
		).WithStacktrace()
	)

	// errors
	var (
		InvalidRequestErr1 = err.New(InvalidRequestErr, "01DC9JRXD98HS9BEXJ1MBXWWM8")
		InvalidRequestErr2 = err.New(InvalidRequestErr, "01DCGXN8ZE1WT0NBDNVYRN2695")

		DGraphQueryTimeoutErr1 = err.New(DGraphQueryTimeoutErr, "01DCC4JF4AAK63F6XYFFN8EJE1")
	)

	registry := err.NewRegistry()

	if e := registry.Register(
		InvalidRequestErr1,
		InvalidRequestErr2,
		DGraphQueryTimeoutErr1,
	); e != nil {
		t.Error(e)
	}

	t.Run("register conflicting Err", func(t *testing.T) {
		// Given an Err is already registered

		// When we try to register a new Err that reuses the same Err.SrcID that is already registered but with a different Desc.ID
		e := registry.Register(err.New(DGraphQueryTimeoutErr1.Desc, InvalidRequestErr2.SrcID.String()))
		// Then an the Err registration fails
		if e == nil {
			t.Error("Err registration should have failed because InvalidRequestErr2 is already registered, but with a different Desc.ID")
		} else {
			switch e := e.(type) {
			case *err.Instance:
				t.Logf("%v", e)
			default:
				t.Errorf("unexpected error type: %T: %[1]1v", e)
			}
		}
	})

	t.Run("register the same error again", func(t *testing.T) {
		InvalidRequestErr3 := err.New(InvalidRequestErr2.Desc, ulidgen.MustNew().String())

		registeredErrCount := registry.Size()

		// When the same error is registered
		e := registry.Register(InvalidRequestErr1,
			InvalidRequestErr2,
			InvalidRequestErr3,
			DGraphQueryTimeoutErr1)
		// Then it succeeds as a noop
		if e != nil {
			t.Error(e)
		} else {
			expectedCount := registeredErrCount + 1
			t.Log(registry.Errs())
			if len(registry.Errs()) != expectedCount {
				t.Errorf("registered error count (%v) should be %d", registry.Size(), expectedCount)
			}
			if !registry.Registered(InvalidRequestErr3.SrcID) {
				t.Errorf("InvalidRequestErr3 is not registered - registered Errs = %v", registry.Errs())
			}
		}
	})
}

func TestRegistry_Read(t *testing.T) {
	t.Parallel()

	// error tags
	const (
		DatabaseTag err.Tag = "db"
		DGraphTag   err.Tag = "dgraph"
	)

	// error descriptors
	var (
		InvalidRequestErr = err.NewDesc("01DC9HDP0X3R60GWDZZY18CVB8", "InvalidRequest", "Invalid request")

		DGraphQueryTimeoutErr = err.NewDesc(
			"01DCC447HWNM5MP7D4Z0DKK0SQ",
			"DatabaseTimeout",
			"query timeout",
			DGraphTag, DatabaseTag,
		).WithStacktrace()
	)

	// errors
	var (
		InvalidRequestErr1 = err.New(InvalidRequestErr, "01DC9JRXD98HS9BEXJ1MBXWWM8")
		InvalidRequestErr2 = err.New(InvalidRequestErr, "01DCGXN8ZE1WT0NBDNVYRN2695")

		DGraphQueryTimeoutErr1 = err.New(DGraphQueryTimeoutErr, "01DCC4JF4AAK63F6XYFFN8EJE1")
	)

	registry := err.NewRegistry()

	if e := registry.Register(
		InvalidRequestErr1,
		InvalidRequestErr2,
		DGraphQueryTimeoutErr1,
	); e != nil {
		t.Error(e)
	}

	t.Run("get all Descs", func(t *testing.T) {
		descs := registry.Descs()
		t.Log(descs)
		// err.RegistryConflictErrClass is automatically registered
		if len(descs) != 3 {
			t.Errorf("expected 2 Descs, but got back: %v", len(descs))
		}
		if descs[err.RegistryConflictErrClass.ID] != err.RegistryConflictErrClass {
			t.Error("err.RegistryConflictErrClass should be registered")
		}
		if descs[InvalidRequestErr1.ID] != InvalidRequestErr {
			t.Error("InvalidRequestErr should be registered")
		}
		if descs[DGraphQueryTimeoutErr1.ID] != DGraphQueryTimeoutErr {
			t.Error("DGraphQueryTimeoutErr should be registered")
		}
	})

	t.Run("filter errors by Desc.ID", func(t *testing.T) {
		findByDescID := func(id ulid.ULID) func(*err.Err) bool {
			return func(err *err.Err) bool {
				return err.ID == id
			}
		}

		errs := registry.Filter(findByDescID(InvalidRequestErr.ID))
		if len(errs) != 2 {
			t.Errorf("2 InvalidRequestErr errors should be registered: %v", errs)
		}

		errs = registry.Filter(findByDescID(DGraphQueryTimeoutErr.ID))
		if len(errs) != 1 {
			t.Errorf("1 DGraphQueryTimeoutErr errors should be registered: %v", errs)
		}

		// When trying to search for errors using an unregistered ID
		errs = registry.Filter(findByDescID(ulidgen.MustNew()))
		// Then none should be returned
		if len(errs) != 0 {
			t.Errorf("none should have been found: %v", errs)
		}

	})
}
