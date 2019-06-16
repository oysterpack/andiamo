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

package main

import (
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/rs/zerolog"
	"log"
	"reflect"
)

type empty struct{}

type Hello func()
type HelloProvider func(logger *zerolog.Logger) Hello

type Command func(hello Hello)

// Hello Comp
var (
	HelloProviderOptionDesc = option.NewDesc(option.Provide, reflect.TypeOf(HelloProvider(nil)))
	HelloInvokeOptionDesc   = option.NewDesc(option.Invoke, reflect.TypeOf(Command(nil)))

	HelloCompDesc = comp.MustNewDesc(
		comp.ID("01DDER5DK2KA7AC0S4YYWFD73V"),
		comp.Name("hello"),
		comp.Version("0.1.0"),
		app.GetPackage(empty{}),
		HelloProviderOptionDesc,
		HelloInvokeOptionDesc,
	)

	HelloComp = HelloCompDesc.MustNewComp(
		HelloProviderOptionDesc.NewOption(func(logger *zerolog.Logger) Hello {
			return func() {
				logger.Info().Msg("hello")
			}
		}),
		HelloInvokeOptionDesc.NewOption(func(hello Hello) {
			hello()
		}),
	)
)

func main() {
	fxapp, e := appfx.NewAppBuilder().
		AppDesc(app.Desc{
			ID:        app.ID(ulidgen.MustNew()),
			Name:      "helloworld",
			Version:   app.MustParseVersion("0.1.0"),
			ReleaseID: app.ReleaseID(ulidgen.MustNew()),
		}).
		Comps(HelloComp).
		Build()
	if e != nil {
		log.Panic(e)
	}
	go func() {
		if e := fxapp.Run(); e != nil {
			log.Fatal(e)
		}
	}()
	<-fxapp.Stopped()
}
