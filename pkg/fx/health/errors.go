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
	"fmt"
	"github.com/pkg/errors"
)

// package errors
var (
	ErrServiceNotRunning = errors.New("health service is not running")

	// ErrTimeout indicates a health check timed out.
	ErrTimeout = errors.New("health check timed out")
)

// YellowError is used to indicate that the health check failed with a warning.
type YellowError struct {
	error
}

// health check registration errors validation errors
var (
	ErrIDNotULID        = errors.New("`ID` must be a ULID")
	ErrBlankDescription = errors.New("`Description` must not be blank")
	ErrBlankRedImpact   = errors.New("`RedImpact` must not be blank")
	ErrTagNotULID       = errors.New("`Tags` must be ULIDs")

	ErrNilChecker             = errors.New("`Checker` is required and must not be nil")
	ErrRunTimeoutTooHigh      = fmt.Errorf("health check run timeout is too high - max allowed timeout is %s", MaxTimeout)
	ErrRunIntervalTooFrequent = fmt.Errorf("health check run interval is too frequent - min allowed run interval is %s", MinRunInterval)
)
