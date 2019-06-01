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

package error

import (
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var (
	newULID = ulidgen.MonotonicULIDGenerator()
)

type Err struct {
	*Desc
	SrcID ulid.ULID
}

// NewError constructs a new Error instance
func New(desc *Desc, srcULID string) *Err {
	return &Err{
		Desc:  desc,
		SrcID: ulid.MustParse(srcULID),
	}
}

func (e *Err) New() *ErrInstance {
	return &ErrInstance{
		Err:        e,
		InstanceID: newULID(),
	}
}

func (e *Err) CausedBy(cause error) *ErrInstance {
	return &ErrInstance{
		Err:        e,
		InstanceID: newULID(),
		Cause:      cause,
	}
}

type Desc struct {
	ID      ulid.ULID
	Name    string
	Message string
}

type ErrInstance struct {
	*Err
	InstanceID ulid.ULID
	Cause      error
}

// Error implements the Error interface
func (e *ErrInstance) Error() string {
	if e.Cause == nil {
		return e.Err.Message
	}
	return fmt.Sprintf("%s : %s", e.Err.Message, e.Cause.Error())
}

func Log(logger *zerolog.Logger, e *ErrInstance) *zerolog.Event {
	return logger.Error().
		Stack().
		Err(errors.WithStack(e)).
		Dict(string(logging.ERR), zerolog.Dict().
			Str(string(logging.ERR_ID), e.ID.String()).
			Str(string(logging.ERR_NAME), e.Name).
			Str(string(logging.ERR_SRC_ID), e.SrcID.String()).
			Str(string(logging.ERR_INSTANCE_ID), e.InstanceID.String()),
		)
}
