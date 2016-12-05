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

package writer

import "strings"

func jsonizer(fields []string) func([]string) string {
	for i, v := range fields {
		if v == "namespace" {
			fields[i] = "\"ns\":\""
			continue
		}
		fields[i] = "\"" + v + "\":\""
	}
	vals := make([]string, len(fields))
	return func(values []string) string {
		for i, v := range values {
			vals[i] = fields[i] + v
		}
		return "{" + strings.Join(vals, "\",") + "\"}"
	}
}

/*
// JSON representation of a Siegfried struct.
// This is the provenace block at the beginning of sf results and includes descriptions for each identifier.
func (s *Siegfried) JSON() string {
	version := config.Version()
	str := fmt.Sprintf(
		"{\"siegfried\":\"%d.%d.%d\",\"scandate\":\"%v\",\"signature\":\"%s\",\"created\":\"%v\",\"identifiers\":[",
		version[0], version[1], version[2],
		time.Now().Format(time.RFC3339),
		s.path,
		s.C.Format(time.RFC3339))
	for i, id := range s.ids {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("{\"name\":\"%s\",\"details\":\"%s\"}", id.Name(), id.Details())
	}
	str += "],"
	return str
}
*/
