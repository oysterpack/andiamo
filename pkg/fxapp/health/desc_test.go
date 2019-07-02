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

package health_test

import (
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"testing"
)

func TestHealthDesc(t *testing.T) {

	id := ulidgen.MustNew()

	t.Run("description cannot be blank", func(t *testing.T) {
		_, err := health.NewDescBuilder(id).
			YellowImpact("Slow query").
			RedImpact("Query times out or fails").
			Build()

		t.Log(err)
		if err == nil {
			t.Error("*** the desc should have failed to build")
		}

		_, err = health.NewDescBuilder(id).
			Description("    ").
			YellowImpact("Slow query").
			RedImpact("Query times out or fails").
			Build()

		t.Log(err)
		if err == nil {
			t.Error("*** the desc should have failed to build")
		}
	})

	t.Run("red impact cannot be blank", func(t *testing.T) {
		_, err := health.NewDescBuilder(id).
			Description("Executes database query").
			YellowImpact("Slow query").
			Build()

		t.Log(err)
		if err == nil {
			t.Error("*** the desc should have failed to build")
		}

		_, err = health.NewDescBuilder(id).
			Description("Executes database query").
			YellowImpact("Slow query").
			RedImpact("   ").
			Build()

		t.Log(err)
		if err == nil {
			t.Error("*** the desc should have failed to build")
		}
	})

	t.Run("all text fields are trimmed", func(t *testing.T) {
		DatabaseHealthCheckDesc := health.NewDescBuilder(id).
			Description("   Executes database query   ").
			YellowImpact("   Slow query   ").
			RedImpact("   Query times out or fails   ").
			MustBuild()

		if DatabaseHealthCheckDesc.ID() != id {
			t.Errorf("*** ID did not match: %v", DatabaseHealthCheckDesc.ID())
		}
		if DatabaseHealthCheckDesc.Description() != "Executes database query" {
			t.Errorf("*** Description did not match: %v", DatabaseHealthCheckDesc.Description())
		}
		if DatabaseHealthCheckDesc.YellowImpact() != "Slow query" {
			t.Errorf("*** YellowImpact did not match: %v", DatabaseHealthCheckDesc.YellowImpact())
		}
		if DatabaseHealthCheckDesc.RedImpact() != "Query times out or fails" {
			t.Errorf("*** RedImpact did not match: %v", DatabaseHealthCheckDesc.RedImpact())
		}
	})

}
