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

package option

import "github.com/oysterpack/partire-k8s/pkg/app/err"

// Option related error descriptors
var (
	UnassignableBindingErrClass = err.NewDesc("01DCVHM65XN2E0QF5W8N0M3RKB", "UnassignableBindingErr", "option type is not assignable to type defined by Desc.FuncType")
)

// Option related errors
var (
	UnassignableBindingErr = err.New(UnassignableBindingErrClass, "01DCVHPTRVDNWNVDW2KQ71ZZXR")
)
