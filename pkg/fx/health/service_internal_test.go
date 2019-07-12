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

package health

import (
	"go.uber.org/fx"
	"runtime"
	"testing"
)

func runApp(app *fx.App, shutdowner fx.Shutdowner, funcs ...func()) {
	done := make(chan struct{})
	defer func() {
	ShutdownLoop:
		for {
			select {
			case <-done:
				break ShutdownLoop
			default:
				shutdowner.Shutdown()
				runtime.Gosched()
			}
		}
	}()

	running := make(chan struct{})
	go func() {
		defer close(done)
		close(running)
		app.Run()
	}()
	<-running
	runtime.Gosched()
	for _, f := range funcs {
		f()
	}

}

func TestService_TriggerShutdown(t *testing.T) {
	t.Parallel()

	t.Run("trigger shutdown is idempotent", func(t *testing.T) {
		t.Parallel()
		s := newService(DefaultOpts())
		go s.run()
		s.TriggerShutdown()
		// calling it again should have no effect
		s.TriggerShutdown()
		<-s.stop
	})
}
