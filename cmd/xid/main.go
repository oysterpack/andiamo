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

package main

import (
	"flag"
	"fmt"
	"github.com/rs/xid"
	"log"
)

var id = flag.String("p", "", "parse XID")
var verbose = flag.Bool("v", false, "shows XID components (PID, Time, Counter)")
var help = flag.Bool("h", false, "prints help")

// used to generate and parse XID
//
// Command Line Flags
//  -p is used to specify an XID to parse
//  -v shows XID components (PID, Time, Counter)
func main() {
	flag.Parse()
	if *help {
		fmt.Println(`xid of a tool used to generate or parse an XID (https://github.com/rs/xid)

Usage:

   xid [-p XID] [-v]

   when the -p flag is not specified, then it will generate a new XID

Flags:`)
		flag.PrintDefaults()
		return
	}

	if *id != "" {
		parseXID(*id)
		return
	}
	// generate a new XID
	print(xid.New())
}

func parseXID(id string) {
	id2, err := xid.FromString(id)
	if err != nil {
		log.Fatal(err)
	}
	print(id2)
}

func print(id xid.ID) {
	if *verbose {
		fmt.Printf("%v -> PID(%v) Time(%s) Counter(%d)\n", id, id.Pid(), id.Time(), id.Counter())
		return
	}
	fmt.Println(id)
}
