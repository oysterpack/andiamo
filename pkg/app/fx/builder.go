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
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"io"
	"time"
)

type disableLogSamplingFlag struct{}

// AppBuilder is used to build a new App
type AppBuilder struct {
	desc                      []app.Desc
	startTimeout, stopTimeout time.Duration
	comps                     []*comp.Comp
	opts                      []fx.Option
	logWriter                 io.Writer
	globalLogLevel            zerolog.Level
	disableLogSampling        *disableLogSamplingFlag
}

// NewAppBuilder returns a new AppBuilder
func NewAppBuilder() *AppBuilder {
	return &AppBuilder{
		globalLogLevel: zerolog.NoLevel,
	}
}

// DisableLogSampling disables log sampling. Log sampling is enabled by default.
func (b *AppBuilder) DisableLogSampling() *AppBuilder {
	b.disableLogSampling = &disableLogSamplingFlag{}
	return b
}

// GlobalLogLevel sets the global log level
func (b *AppBuilder) GlobalLogLevel(level zerolog.Level) *AppBuilder {
	b.globalLogLevel = level
	return b
}

// LogWriter specifies the writer that will be used for application logging.
func (b *AppBuilder) LogWriter(w io.Writer) *AppBuilder {
	b.logWriter = w
	return b
}

// AppDesc sets the app descriptor. If not specified, then the builder will try to load the app descriptor from env vars.
func (b *AppBuilder) AppDesc(desc app.Desc) *AppBuilder {
	b.desc = []app.Desc{desc}
	return b
}

// StartTimeout sets the app start timeout. If not set, then it will try to load the timeout from the env.
// If not specified in the env, then it defaults to 15 sec.
func (b *AppBuilder) StartTimeout(timeout time.Duration) *AppBuilder {
	b.startTimeout = timeout
	return b
}

// StopTimeout sets the app stop timeout. If not set, then it will try to load the timeout from the env.
// If not specified in the env, then it defaults to 15 sec.
func (b *AppBuilder) StopTimeout(timeout time.Duration) *AppBuilder {
	b.stopTimeout = timeout
	return b
}

// Options is used to specify application options.
// Only `provide` and `invoke` options should be specified.
// `populate` options come in handy when unit testing.
func (b *AppBuilder) Options(opts ...fx.Option) *AppBuilder {
	if len(opts) > 0 {
		b.opts = append(b.opts, opts...)
	}
	return b
}

// Comps is used to specifiy the application components, which will be registered with the component registry.
func (b *AppBuilder) Comps(comps ...*comp.Comp) *AppBuilder {
	if len(comps) > 0 {
		b.comps = append(b.comps, comps...)
	}
	return b
}

// Build tries to build the app.
func (b *AppBuilder) Build() (*App, error) {
	if len(b.comps) == 0 && len(b.opts) == 0 {
		return nil, OptionsRequiredErr.New()
	}

	if len(b.desc) == 0 {
		desc, e := app.LoadDesc()
		if e != nil {
			return nil, e
		}
		b.AppDesc(desc)
	}

	if e := b.desc[0].Validate(); e != nil {
		return nil, InvalidDescErr.CausedBy(e)
	}

	timeouts, e := app.LoadTimeouts()
	if e != nil {
		return nil, e
	}
	if b.startTimeout != 0 {
		timeouts.StartTimeout = b.startTimeout
	}
	if b.stopTimeout != 0 {
		timeouts.StopTimeout = b.stopTimeout
	}

	for _, c := range b.comps {
		b.opts = append(b.opts, c.FxOptions())
	}

	fxapp, e := NewApp(b.desc[0], timeouts, b.logWriter, b.globalLogLevel, b.opts...)
	if b.disableLogSampling != nil {
		zerolog.DisableSampling(true)
	}
	return fxapp, e
}
