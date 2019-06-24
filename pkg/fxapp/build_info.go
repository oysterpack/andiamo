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

package fxapp

import (
	"errors"
	"github.com/rs/zerolog"
	"runtime/debug"
)

type BuildInfo struct {
	Path string    // The main package Path
	Main Module    // The main module information
	Deps []*Module // Module dependencies
}

func (b *BuildInfo) MarshalZerologObject(e *zerolog.Event) {
	e.Dict("build", zerolog.Dict().
		Str("path", b.Path).
		Dict("main", zerolog.Dict().
			Str("path", b.Main.Path).
			Str("version", b.Main.Version).
			Str("checksum", b.Main.Checksum)).
		Array("deps", b.depArr()),
	)
}

func (b *BuildInfo) depArr() *zerolog.Array {
	arr := zerolog.Arr()
	for _, d := range b.Deps {
		arr.Object(&Module{d.Path, d.Version, d.Checksum})
	}
	return arr
}

// ReadBuildInfo returns the build information embedded in the running binary.
// The information is available only in binaries built with module support.
func ReadBuildInfo() (*BuildInfo, error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("build information is available only in binaries built with module support")
	}
	var deps []*Module
	for _, dep := range buildInfo.Deps {
		deps = append(deps, NewModule(dep))
	}
	return &BuildInfo{
		buildInfo.Path,
		Module{buildInfo.Main.Path, buildInfo.Main.Version, buildInfo.Main.Sum},
		deps,
	}, nil
}

type Module struct {
	Path     string
	Version  string
	Checksum string
}

func NewModule(m *debug.Module) *Module {
	d := m
	if m.Replace != nil {
		d = m.Replace
	}
	return &Module{d.Path, d.Version, d.Sum}
}

func (m *Module) MarshalZerologObject(e *zerolog.Event) {
	e.Str("path", m.Path)
	e.Str("version", m.Version)
	e.Str("checksum", m.Checksum)
}
