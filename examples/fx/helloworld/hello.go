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
	"context"
	"go.uber.org/fx"
	"log"
)

func main() {
	app := fx.New(fx.Invoke(hello))
	log.Printf("app start timeout = %v", app.StartTimeout())
	log.Printf("app stop timeout = %v", app.StartTimeout())
	app.Run()
}

func hello(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Println("app is starting")
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Println("app is stopping")
			return nil
		},
	})
	log.Println("hello ... use [Ctrl-C] to stop the app")
}
