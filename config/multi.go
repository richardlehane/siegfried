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

package config

// Multi defines how identifiers treat multiple results.
type Multi int

const (
	Single        Multi = iota // Return a single result. If there is more than one result with the highest score, return UNKNOWN and a warning
	Conclusive                 // Default. Return only the results with the highest score.
	Positive                   // Return any result with a strong score (or if only weak results, return all). This means a byte match, container match or XML match. Text/MIME/extension-only matches are considered weak.
	Comprehensive              // Same as positive but also turn off the priority rules during byte matching.
	Exhaustive                 // Turn off priority rules during byte matching and return all weak as well as strong results.
)

func (m Multi) String() string {
	switch m {
	case Single:
		return "single (0)"
	case Conclusive:
		return "conclusive (1)"
	case Positive:
		return "positive (2)"
	case Comprehensive:
		return "comprehensive (3)"
	case Exhaustive:
		return "exhaustive (4)"
	}
	return ""
}
