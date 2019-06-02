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

package error_test

import (
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/error"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"testing"
)

type empty struct{}

var (
	pkg = app.GetPackage(empty{})
)

// error descriptors are defined in the code
var (
	InvalidRequest = &error.Desc{
		ID:      ulid.MustParse("01DC9HDP0X3R60GWDZZY18CVB8"),
		Name:    "InvalidRequest",
		Message: "Invalid request",
	}

	// SrcID is used to identify the error source, i.e. where in the source did the error originate from.
	// - by assigning a ULID, it will make it easy to locate where the error was created in the code, without a stack trace.
	// - by seeing how many SrcID(s) are defined, tells us how many locations in the code could potentially trigger errors
	TestErrorInvalidRequestErr = error.New(InvalidRequest, "01DC9JRXD98HS9BEXJ1MBXWWM8")

	DGraphQueryTimeout = &error.Desc{
		ID:           ulid.MustParse("01DCC447HWNM5MP7D4Z0DKK0SQ"),
		Name:         "DatabaseTimeout",
		Message:      "query timeout",
		Tags:         []string{DGraphTag.String(), DatabaseTag.String()},
		IncludeStack: true,
	}

	TestErrorDGraphQueryTimeoutErr = error.New(DGraphQueryTimeout, "01DCC4JF4AAK63F6XYFFN8EJE1")
)

const (
	DatabaseTag error.Tag = "db"
	DGraphTag   error.Tag = "dgraph"
)

func TestError_New(t *testing.T) {
	// When a new Error is created
	err := TestErrorInvalidRequestErr.New()
	t.Logf("err: %+v", err)
	// Then the error.Desc is referenced by the Error
	if err.Desc == nil {
		t.Error("Desc is required")
	}
	if err.Desc.ID != InvalidRequest.ID {
		t.Error("Desc.ID did not match")
	}
	// And the Error is assigned a unique InstanceID
	zeroULID := ulid.ULID{}
	if err.InstanceID == zeroULID {
		t.Error("InstanceID is required")
	}
}

func TestError_Log(t *testing.T) {

	t.Run("no tags - with stacktrace", func(t *testing.T) {
		// Given an Error
		err := TestErrorInvalidRequestErr.New()
		// When the Error is logged
		logger := apptest.NewTestLogger(pkg)
		err.Log(logger.Logger).Msg("")
		logEventMsg := logger.Buf.String()
		t.Log(logEventMsg)

		var logEvent apptest.LogEvent
		if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
			t.Fatalf("Invalid JSON log event: %v", err)
		}
		t.Logf("JSON log event: %#v", logEvent)
		checkInvalidRequestErrLogEvent(t, &logEvent, err)
	})

	t.Run("with tags - with no stacktrace", func(t *testing.T) {
		// Given an Error
		err := TestErrorDGraphQueryTimeoutErr.New()

		// When the Error is logged
		logger := apptest.NewTestLogger(pkg)
		err.Log(logger.Logger).Msg("")
		logEventMsg := logger.Buf.String()
		t.Log(logEventMsg)

		var logEvent apptest.LogEvent
		if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
			t.Fatalf("Invalid JSON log event: %v", err)
		}
		checkDGraphQueryTimeErrLogEvent(t, &logEvent, err)
	})
}

func checkDGraphQueryTimeErrLogEvent(t *testing.T, logEvent *apptest.LogEvent, err *error.ErrInstance) {
	if len(logEvent.Error.Tags) != len(DGraphQueryTimeout.Tags) {
		t.Error("the number of logged tags does not match what is expected")
	} else {
		// Then error tags should be logged
		for i := 0; i < len(DGraphQueryTimeout.Tags); i++ {
			if logEvent.Error.Tags[i] != DGraphQueryTimeout.Tags[i] {
				t.Errorf("tag did not match: %v != %v", logEvent.Tags[i], DGraphQueryTimeout.Tags[i])
			}
		}

		// And there should be no stacktrace logged
		if len(logEvent.Stack) == 0 {
			t.Error("stacktrace should have been logged")
		}
	}
}

func checkInvalidRequestErrLogEvent(t *testing.T, logEvent *apptest.LogEvent, err *error.ErrInstance) {
	// Then the log level will be ErrorLevel
	if logEvent.Level != zerolog.ErrorLevel.String() {
		t.Error("log level did not match")
	}

	// And the error message will match
	if logEvent.ErrorMessage != InvalidRequest.Message {
		t.Error("error message did not match")
	}

	if logEvent.Error == nil {
		t.Error("error details were not logged")
	} else {
		if logEvent.Error.ID != err.Desc.ID.String() {
			t.Error("error ID did not match")
		}

		if logEvent.Error.InstanceID != err.InstanceID.String() {
			t.Errorf("error InstanceID did not match: %v != %v", logEvent.Error.InstanceID, err.InstanceID.String())
		}

		if logEvent.Error.Name != err.Desc.Name {
			t.Error("error Name did not match")
		}

		if logEvent.Error.SrcID != err.SrcID.String() {
			t.Error("error source ID did not match")
		}

		if len(logEvent.Tags) > 0 {
			t.Error("log event should have not tags")
		}

		if len(logEvent.Stack) > 0 {
			t.Error("stacktrace should not have been logged")
		}
	}
}

func TestErr_CausedBy(t *testing.T) {
	cause := errors.New("err root cause")
	err := TestErrorInvalidRequestErr.CausedBy(cause)
	errCause := err.Cause
	if errCause.Error() != cause.Error() {
		t.Fatal("error message did not match")
	}

	// When the Error is logged
	logger := apptest.NewTestLogger(pkg)
	err.Log(logger.Logger).Msg("")
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent apptest.LogEvent
	if err := json.Unmarshal([]byte(logEventMsg), &logEvent); err != nil {
		t.Fatalf("Invalid JSON log event: %v", err)
	}

	if logEvent.ErrorMessage != fmt.Sprintf("%s : %s", TestErrorInvalidRequestErr.Message, cause.Error()) {
		t.Error("error message did not match")
	}
}
