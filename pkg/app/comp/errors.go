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

package comp

import "github.com/oysterpack/partire-k8s/pkg/app/err"

// Desc related error descriptors
var (
	UniqueOptionTypeConstraintErrClass = err.MustNewDesc("01DCW50DM81QZH5E0G101DKY98", "UniqueOptionTypeConstraintErr", "option types must be unique")
	UniqueOptionTypeConstraintErr      = err.New(UniqueOptionTypeConstraintErrClass, "01DCW5472EZFFK5YE9VT7SRG2V")

	DescInvalidIDErr      = err.New(err.InvalidIDErrClass, "01DDKT67J2SD6K1P0GTDWPYQZG")
	DescInvalidVersionErr = err.New(err.InvalidVersionErrClass, "01DDKTD6R6XEWKXM09B7J5TRCF")
)

// Comp related error descriptors
var (
	OptionsRequiredErrClass = err.MustNewDesc("01DCVNVW0QMJH345SWQR8Q81N7", "OptionsRequiredErr", "at least 1 option is required")
	OptionsRequiredErr      = err.New(OptionsRequiredErrClass, "01DCVNVXSTEBWPM1BBQFM0ZKXJ")

	OptionsNotMatchingDescErrClass = err.MustNewDesc("01DCW3NEMQ7ESTR7D9184Y4DBC", "OptionsNotMatchingDescErr", "options not matching descriptor")
	OptionCountDoesNotMatchErr     = err.New(OptionsNotMatchingDescErrClass, "01DCW3QNQ6H8EMW97W8SPGSNYG")
	OptionDescTypeNotMatchedErr    = err.New(OptionsNotMatchingDescErrClass, "01DCW4SXRKMJ491F6DD5HJS3QF")
)
