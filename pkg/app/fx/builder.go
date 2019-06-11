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
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"go.uber.org/fx"
	"io"
)

type AppBuilder struct {
	desc      *app.Desc
	timeouts  *app.Timeouts
	comps     []*comp.Comp
	opts      []fx.Option
	logWriter io.Writer
}

func NewAppBuilder() *AppBuilder {
	return &AppBuilder{}
}

func (b *AppBuilder) AppDesc(desc app.Desc) *AppBuilder {
	b.desc = &desc
	return b
}

func (b *AppBuilder) Options(opts ...fx.Option) *AppBuilder {
	if len(opts) > 0 {
		b.opts = append(b.opts, opts...)
	}
	return b
}

func (b *AppBuilder) Comps(comps ...*comp.Comp) *AppBuilder {
	if len(comps) > 0 {
		b.comps = append(b.comps, comps...)
	}
	return b
}

func (b *AppBuilder) Build() (*App, error) {
	if len(b.comps) == 0 && len(b.opts) == 0 {
		return nil, OptionsRequiredErr.New()
	}

	if b.desc == nil {
		desc, e := app.LoadDesc()
		if e != nil {
			return nil, e
		}
		b.desc = &desc
	}

	if b.timeouts == nil {
		timeouts, e := app.LoadTimeouts()
		if e != nil {
			return nil, e
		}
		b.timeouts = &timeouts
	}

	for _, c := range b.comps {
		b.opts = append(b.opts, c.FxOptions())
	}

	return NewApp(*b.desc, *b.timeouts, b.logWriter, b.opts...)
}
