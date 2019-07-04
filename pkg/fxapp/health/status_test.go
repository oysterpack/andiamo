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
	"testing"
)

func TestStatus_String(t *testing.T) {
	if health.Green.String() != "Green" {
		t.Errorf("*** health.Green.String() should be 'Green': %s", health.Green.String())
	}

	if health.Yellow.String() != "Yellow" {
		t.Errorf("*** health.Yellow.String() should be 'Yellow': %s", health.Yellow.String())
	}

	if health.Red.String() != "Red" {
		t.Errorf("*** health.Red.String() should be 'Red': %s", health.Red.String())
	}
}
