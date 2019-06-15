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

package fx

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"runtime/debug"
	"testing"
)

type startLogEvent struct {
	Build struct {
		Path string
		Main struct {
			Path     string
			Version  string
			Checksum string
		}
		Deps []struct {
			Path     string
			Version  string
			Checksum string
		}
	}
}

/*
{
  "l": "info",
  "build": {
    "path": "build_path",
    "main": {
      "path": "main_path",
      "version": "0.1.0",
      "checksum": "main_mod_checksum"
    },
    "deps": [
      {
        "path": "dep_path_1",
        "version": "0.1.0",
        "checksum": "dep_mod_checksum_1"
      },
      {
        "path": "dep_path_2",
        "version": "0.2.0",
        "checksum": "dep_mod_checksum_2"
      }
    ]
  }
}
*/
func TestLogStartEvent(t *testing.T) {
	b := &buildInfo{
		path: "build_path",
		main: module{
			path:     "main_path",
			version:  "0.1.0",
			checksum: "main_mod_checksum",
		},
		deps: []*module{
			{
				path:     "dep_path_1",
				version:  "0.1.0",
				checksum: "dep_mod_checksum_1",
			},
			{
				path:     "dep_path_2",
				version:  "0.2.0",
				checksum: "dep_mod_checksum_2",
			},
		},
	}

	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	logEvent := logger.Info()
	appendBuildInfo(logEvent, b)
	logEvent.Msg("")
	t.Logf("%s", buf)

	// Verify that logged event structure
	var event startLogEvent
	if e := json.Unmarshal(buf.Bytes(), &event); e != nil {
		t.Fatalf("failed to parse log event")
	}
	t.Logf("event: %v", event)
	if event.Build.Path != b.path {
		t.Errorf("path does not match: %s", event.Build.Path)
	}
	// check main build info matches
	if event.Build.Main.Path != b.main.path {
		t.Errorf("main path does not match: %s", event.Build.Main.Path)
	}
	if event.Build.Main.Version != b.main.version {
		t.Errorf("main version does not match: %s", event.Build.Main.Version)
	}
	if event.Build.Main.Checksum != b.main.checksum {
		t.Errorf("main checksum does not match: %s", event.Build.Main.Checksum)
	}
	// check dependency mod info matches
	if len(event.Build.Deps) != len(b.deps) {
		t.Errorf("dep count does not match: %d", len(event.Build.Deps))
	}
	for i, dep := range event.Build.Deps {
		if dep.Path != b.deps[i].path {
			t.Errorf("dep path does not match: %s", dep.Path)
		}
		if dep.Version != b.deps[i].version {
			t.Errorf("dep version does not match: %s", dep.Version)
		}
		if dep.Checksum != b.deps[i].checksum {
			t.Errorf("dep checksum does not match: %s", dep.Checksum)
		}
	}
}

/*
{
  "l": "info",
  "build": {
    "path": "build_path",
    "main": {
      "path": "main_path",
      "version": "0.1.0",
      "checksum": "main_check_sum"
    },
    "deps": [
      {
        "path": "dep_path_1",
        "version": "0.1.0",
        "checksum": "dep_check_sum_1"
      },
      {
        "path": "dep_path_3",
        "version": "0.3.0",
        "checksum": "dep_check_sum_3"
      }
    ]
  }
}
*/
func TestNewBuildInfo(t *testing.T) {
	b := newBuildInfo(&debug.BuildInfo{
		Path: "build_path",
		Main: debug.Module{
			Path:    "main_path",
			Version: "0.1.0",
			Sum:     "main_check_sum",
		},
		Deps: []*debug.Module{
			{
				Path:    "dep_path_1",
				Version: "0.1.0",
				Sum:     "dep_check_sum_1",
			},
			{
				Path:    "dep_path_2",
				Version: "0.2.0",
				Sum:     "dep_check_sum_2",
				Replace: &debug.Module{
					Path:    "dep_path_3",
					Version: "0.3.0",
					Sum:     "dep_check_sum_3",
				},
			},
		},
	})

	buf := new(bytes.Buffer)
	logger := zerolog.New(buf)
	logEvent := logger.Info()
	appendBuildInfo(logEvent, b)
	logEvent.Msg("")
	t.Logf("%s", buf)

	// Verify that logged event structure
	var event startLogEvent
	if e := json.Unmarshal(buf.Bytes(), &event); e != nil {
		t.Fatalf("failed to parse log event")
	}
	t.Logf("event: %v", event)

	if event.Build.Path != b.path {
		t.Errorf("path does not match: %s", event.Build.Path)
	}
	// check main build info matches
	if event.Build.Main.Path != b.main.path {
		t.Errorf("main path does not match: %s", event.Build.Main.Path)
	}
	if event.Build.Main.Version != b.main.version {
		t.Errorf("main version does not match: %s", event.Build.Main.Version)
	}
	if event.Build.Main.Checksum != b.main.checksum {
		t.Errorf("main checksum does not match: %s", event.Build.Main.Checksum)
	}

	// check dependency mod info matches
	if len(event.Build.Deps) != len(b.deps) {
		t.Errorf("dep count does not match: %d", len(event.Build.Deps))
	}
	for i, dep := range event.Build.Deps {
		if dep.Path != b.deps[i].path {
			t.Errorf("dep path does not match: %s", dep.Path)
		}
		if dep.Version != b.deps[i].version {
			t.Errorf("dep version does not match: %s", dep.Version)
		}
		if dep.Checksum != b.deps[i].checksum {
			t.Errorf("dep checksum does not match: %s", dep.Checksum)
		}
	}

	// dep[1] should be replaced
	if event.Build.Deps[1].Path != "dep_path_3" {
		t.Errorf("dep[1] path does not match: %s", event.Build.Deps[1].Path)
	}
	if event.Build.Deps[1].Version != "0.3.0" {
		t.Errorf("dep[1] version does not match: %s", event.Build.Deps[1].Version)
	}
	if event.Build.Deps[1].Checksum != "dep_check_sum_3" {
		t.Errorf("dep[1] checksum does not match: %s", event.Build.Deps[1].Checksum)
	}
}
