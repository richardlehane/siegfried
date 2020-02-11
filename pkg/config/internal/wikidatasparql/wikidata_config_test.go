// Copyright 2020 Ross Spencer, Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

// Tests for the Wikidata SPARQL helper functions.

package wikidatasparql

import (
	"fmt"
	"strings"
	"testing"
)

// TestDefaultLang the simplest test to make sure that 'some' default
// language is set.
func TestDefaultLang(t *testing.T) {
	expected := "en"
	if wikidataLang != "en" {
		t.Errorf(
			"Default language is incorrect, expected '%s' got '%s'",
			expected,
			wikidataLang,
		)
	}
}

// TestReturnWDSparql ensures that the language `<<lang>>` template is
// present in the SPARQL query required to harvest Wikidata results. The
// language template is a functional component required to generate
// results in different languages including English.
func TestReturnWDSparql(t *testing.T) {
	template := "<<lang>>"
	if !strings.Contains(sparql, template) {
		t.Errorf(
			"Lang replacement template '%s' is missing from SPARQL request string:\n%s",
			template,
			sparql,
		)
	}
	res := WikidataSPARQL()
	defaultLang := "\"[AUTO_LANGUAGE], en\""
	if !strings.Contains(res, defaultLang) {
		t.Errorf(
			"Default language `en` missing from SPARQL request string:\n%s",
			res,
		)
	}
	// Change the language string and ensure that replacement occurs.
	newLang := "jp"
	SetWikidataLang(newLang)
	newLangReplacement := fmt.Sprintf("\"[AUTO_LANGUAGE], %s\".", newLang)
	res = WikidataSPARQL()
	if !strings.Contains(res, newLangReplacement) {
		t.Errorf(
			"Language replacement '%s' hasn't been done in returned SPARQL request string:\n%s",
			newLangReplacement,
			res,
		)
	}
}
