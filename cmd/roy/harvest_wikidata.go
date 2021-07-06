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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/ross-spencer/wikiprov/pkg/spargo"
	"github.com/ross-spencer/wikiprov/pkg/wikiprov"
)

// harvestWikidata will connect to the configured Wikidata query service
// and save the results of the configured query to disk.
func harvestWikidata() error {

	log.Printf(
		"Roy (Wikidata): Harvesting Wikidata definitions: lang '%s'",
		config.WikidataLang(),
	)
	err := os.MkdirAll(config.WikidataHome(), os.ModePerm)
	if err != nil {
		return fmt.Errorf(
			"Roy (Wikidata): Error harvesting Wikidata definitions: '%s'",
			err,
		)
	}
	log.Printf(
		"Roy (Wikidata): Harvesting definitions from: '%s'",
		config.WikidataEndpoint(),
	)

	wikiprov.SetWikibasePermalinkBaseURL(config.WikidataWikibaseURL())

	res, err := spargo.SPARQLWithProv(
		config.WikidataEndpoint(),
		config.WikidataSPARQL(),
		config.WikidataSPARQLRevisionParam(),
		config.GetWikidataRevisionHistoryLen(),
		config.GetWikidataRevisionHistoryThreads(),
	)

	if err != nil {
		return fmt.Errorf("Roy (Wikidata): Error trying to retrieve SPARQL with revision history: %s", err)
	}

	path := config.WikidataDefinitionsPath()
	err = ioutil.WriteFile(path, []byte(fmt.Sprintf("%s", res)), config.WikidataFileMode())
	if err != nil {
		return fmt.Errorf(
			"Roy (Wikidata): Error harvesting Wikidata: '%s'",
			err,
		)
	}
	log.Printf(
		"Roy (Wikidata): Harvesting Wikidata definitions '%s' complete",
		path,
	)
	return nil
}
