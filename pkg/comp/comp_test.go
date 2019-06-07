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
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/oysterpack/partire-k8s/pkg/comp"
	"testing"
)

func TestComp_Logger(t *testing.T) {
	logger := apptest.NewAppTestLogger()

	c := comp.MustNew(ulidgen.MustNew().String(), "foo", "0.1.0", Package)
	compLogger := c.Logger(logger.Logger)
	compLogger.Info().Msg("")

	var logEvent apptest.LogEvent
	t.Log(logger.Buf.String())
	if e := json.Unmarshal([]byte(logger.Buf.String()), &logEvent); e != nil {
		t.Fatal(e)
	}
	if logEvent.Package != string(Package) {
		t.Error("package field is missing from the log event")
	}
	if logEvent.Component != c.Name {
		t.Error("component field is missing from the log event")
	}
}
