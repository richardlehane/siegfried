// Copyright 2017 Richard Lehane. All rights reserved.
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

package chart

import (
	"bytes"
	"fmt"
	"strings"
)

const max = 10

func squares(num int, rel float64, abs bool) string {
	if rel < 1 && !abs {
		nnum := int(float64(num) * rel)
		if nnum == 0 && num > 0 {
			nnum = 1
		}
		num = nnum
	}
	s := make([]string, num)
	for i := 0; i < num; i++ {
		s[i] = "\xE2\x96\xA0"
	}
	return strings.Join(s, " ")
}

func Chart(title string, sections, fields []string, abs map[string]bool, frequencies map[string]map[string]int) string {
	buf := &bytes.Buffer{}
	if len(title) > 0 {
		buf.WriteString(strings.ToUpper(title))
	}
	var pad int // pad to length longest field
	for _, v := range fields {
		if len(v) > pad {
			pad = len(v) + 1
		}
	}
	template := fmt.Sprintf("%%-%ds %%s (%%d)\n", pad)
	var rel float64 = 1
	for _, m := range frequencies {
		for _, num := range m {
			if num > max && float64(max)/float64(num) < rel {
				rel = float64(max) / float64(num)
			}
		}
	}
	for _, k := range sections {
		if m, ok := frequencies[k]; ok {
			fmt.Fprintf(buf, "\n%s\n", k)
			for _, label := range fields {
				if num, ok := m[label]; ok {
					fmt.Fprintf(buf, template, label+":", squares(num, rel, abs[label]), num)
				}
			}
		}
	}
	return buf.String()
}
