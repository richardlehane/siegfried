// Copyright 2016 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package writers

import (
	"fmt"
	"strings"
)

func header(fields []string) string {
	headings := make([]string, len(fields))
	var max int
	for _, v := range fields {
		if v != "namespace" && len(v) > max {
			max = len(v)
		}
	}
	pad := fmt.Sprintf("%%-%ds", max)
	for i, v := range fields {
		if v == "namespace" {
			v = "ns"
		}
		headings[i] = fmt.Sprintf(pad, v)
	}
	return "  - " + strings.Join(headings, " : %v\n    ") + " : %v\n"
}

func yamlizer(fields []string) func([]string) string {
	hdr := header(fields)
	vals := make([]interface{}, len(fields))
	return func(values []string) string {
		for i, v := range values {
			if v == "" {
				vals[i] = ""
				continue
			}
			vals[i] = "'" + v + "'"
		}
		return fmt.Sprintf(hdr, vals...)
	}
}
