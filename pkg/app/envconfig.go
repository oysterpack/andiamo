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

package app

// EnvPrefix is used as the environment variable name prefix to load configs from the env.
// - "APPX12" was chosen to represent 12-factor apps.
// - this is used by `LoadDesc()` and `LoadTimeouts()`
// - for more information, see "github.com/kelseyhightower/envconfig"
const EnvPrefix = "APPX12"
