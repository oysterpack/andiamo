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

package app_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/oysterpack/andiamo/pkg/fx/app"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"testing"
)

func TestNewApp(t *testing.T) {
	buf := new(bytes.Buffer)
	a := app.New(
		app.Module(app.Opts{
			ID:        ulids.MustNew(),
			ReleaseID: ulids.MustNew(),
			LogWriter: buf,
		}),
		fx.Invoke(
			func(logger app.Logger) {
				event := logger("TestNewApp", zerolog.InfoLevel)
				event(nil, "CIAO MUNDO!!!")
			},
		),
	)

	assert.NoError(t, a.Err())
	assert.NoError(t, a.Start(context.Background()))
	assert.NoError(t, a.Stop(context.Background()))

	type Data struct {
		Duration uint64
	}

	// {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
	type LogEvent struct {
		Level   string `json:"l"`
		Name    string `json:"n"`
		Message string `json:"m"`

		AppID        string `json:"a"`
		AppReleaseID string `json:"r"`
		InstanceID   string `json:"x"`

		Data `json:"d"`
	}

	expectedEvents := map[string]struct{}{
		app.InitializedEvent: struct{}{},
		app.StartingEvent:    struct{}{},
		app.StartedEvent:     struct{}{},
		app.StoppingEvent:    struct{}{},
		app.StoppedEvent:     struct{}{},
	}
	r := bufio.NewReader(buf)
	var logEvent LogEvent
	for {
		line, err := r.ReadString('\n')
		t.Log(line)
		if err != nil {
			break
		}
		assert.NoError(t, json.Unmarshal([]byte(line), &logEvent), "failed to parse line: %s", line)

		delete(expectedEvents, logEvent.Name)
	}

	assert.Empty(t, expectedEvents)

}
