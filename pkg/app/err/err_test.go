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

package err_test

import (
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
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
	InvalidRequestErr = err.NewDesc("01DC9HDP0X3R60GWDZZY18CVB8", "InvalidRequest", "Invalid request")

	InvalidRequestErr1 = err.New(InvalidRequestErr, "01DC9JRXD98HS9BEXJ1MBXWWM8")
	InvalidRequestErr2 = err.New(InvalidRequestErr, "01DCGXN8ZE1WT0NBDNVYRN2695")

	DGraphQueryTimeoutErr = err.NewDesc(
		"01DCC447HWNM5MP7D4Z0DKK0SQ",
		"DatabaseTimeout",
		"query timeout",
		DGraphTag, DatabaseTag,
	).WithStacktrace()

	DGraphQueryTimeoutErr1 = err.New(DGraphQueryTimeoutErr, "01DCC4JF4AAK63F6XYFFN8EJE1")
)

const (
	DatabaseTag err.Tag = "db"
	DGraphTag   err.Tag = "dgraph"
)

func TestError_New(t *testing.T) {
	// When a new Error is created
	e := InvalidRequestErr1.New()
	t.Logf("e: %+v", e)
	// Then the error.Desc is referenced by the Error
	if e.Desc == nil {
		t.Error("Desc is required")
	}
	if e.Desc.ID != InvalidRequestErr.ID {
		t.Error("Desc.ID did not match")
	}
	// And the Error is assigned a unique InstanceID
	zeroULID := ulid.ULID{}
	if e.InstanceID == zeroULID {
		t.Error("InstanceID is required")
	}
}

func TestError_Log(t *testing.T) {

	t.Run("no tags - with stacktrace", func(t *testing.T) {
		// Given an Error
		e := InvalidRequestErr2.New()
		// When the Error is logged
		logger := apptest.NewTestLogger(pkg)
		e.Log(logger.Logger).Msg("")
		logEventMsg := logger.Buf.String()
		t.Log(logEventMsg)

		var logEvent apptest.LogEvent
		if e := json.Unmarshal([]byte(logEventMsg), &logEvent); e != nil {
			t.Fatalf("Invalid JSON log event: %v", e)
		}
		t.Logf("JSON log event: %#v", logEvent)
		checkInvalidRequestErrLogEvent(t, &logEvent, e)
	})

	t.Run("with tags - with no stacktrace", func(t *testing.T) {
		// Given an Error
		e := DGraphQueryTimeoutErr1.New()

		// When the Error is logged
		logger := apptest.NewTestLogger(pkg)
		e.Log(logger.Logger).Msg("")
		logEventMsg := logger.Buf.String()
		t.Log(logEventMsg)

		var logEvent apptest.LogEvent
		if e := json.Unmarshal([]byte(logEventMsg), &logEvent); e != nil {
			t.Fatalf("Invalid JSON log event: %v", e)
		}
		checkDGraphQueryTimeErrLogEvent(t, &logEvent, e)
	})
}

func checkDGraphQueryTimeErrLogEvent(t *testing.T, logEvent *apptest.LogEvent, _ *err.Instance) {
	if len(DGraphQueryTimeoutErr.Tags) != 2 {
		t.Fatalf("tags were not added when constructing the ErrorDesc: %+v", DGraphQueryTimeoutErr)
	}
	if len(logEvent.Error.Tags) != len(DGraphQueryTimeoutErr.Tags) {
		t.Error("the number of logged tags does not match what is expected")
	} else {
		// Then error tags should be logged
		for i := 0; i < len(DGraphQueryTimeoutErr.Tags); i++ {
			if logEvent.Error.Tags[i] != DGraphQueryTimeoutErr.Tags[i] {
				t.Errorf("tag did not match: %v != %v", logEvent.Tags[i], DGraphQueryTimeoutErr.Tags[i])
			}
		}

		// And there should be no stacktrace logged
		if len(logEvent.Stack) == 0 {
			t.Error("stacktrace should have been logged")
		}
	}
}

func checkInvalidRequestErrLogEvent(t *testing.T, logEvent *apptest.LogEvent, err *err.Instance) {
	// Then the log level will be ErrorLevel
	if logEvent.Level != zerolog.ErrorLevel.String() {
		t.Error("log level did not match")
	}

	// And the error message will match
	if logEvent.ErrorMessage != InvalidRequestErr.Message {
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
	cause := errors.New("e root cause")
	e := err.New(InvalidRequestErr, "01DCGXQJ70CBPCJCYMP9N15ZB1").CausedBy(cause)
	errCause := e.Cause
	if errCause.Error() != cause.Error() {
		t.Fatal("error message did not match")
	}

	// When the Error is logged
	logger := apptest.NewTestLogger(pkg)
	e.Log(logger.Logger).Msg("")
	logEventMsg := logger.Buf.String()
	t.Log(logEventMsg)

	var logEvent apptest.LogEvent
	if e := json.Unmarshal([]byte(logEventMsg), &logEvent); e != nil {
		t.Fatalf("Invalid JSON log event: %v", e)
	}

	if logEvent.ErrorMessage != fmt.Sprintf("%s : %s", InvalidRequestErr.Message, cause.Error()) {
		t.Error("error message did not match")
	}
	if logEvent.Error.SrcID != e.SrcID.String() {
		t.Error("error source ID did not match")
	}
}
