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

/*
Package fxapp builds upon https://godoc.org/go.uber.org/fx to provide a standardized functional driven application container.

DevOps Application Aspects

  - all app deployments must have an identity
    - each app is assigned a unique ID
      - application names may change, but the app ID is immutable
    - each app deployment is assigned a release ID, which maps to related app information, e.g.,
	  - release notes
      - who were the persons involved - developers, testers, product managers, etc
      - discussions
      - test reports
        - unit test reports
        - acceptance test reports
        - performance test reports
          - performance profiles
      - etc
  - all running application deployment instances must be identified via an instance ID
    - used for troubleshooting, e.g., querying for application instance logs, metrics, etc
  - application logging is structured
    - zerolog is used to provided structured JSON logging
    - log events are strongly typed, i.e., domain specific
  - metrics
  - health checks
*/
package fxapp
