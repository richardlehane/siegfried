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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/ross-spencer/wikiprov/pkg/spargo"
	"github.com/ross-spencer/wikiprov/pkg/wikiprov"
)

// configureCustomWikibase captures all the functions needed to harvest
// signature information from a custom Wikibase instance. It only
// impacts 'harvest'. Roy's 'build' stage needs to be managed differently.
func configureCustomWikibase() error {
	if err := config.SetCustomWikibaseEndpoint(
		*harvestWikidataEndpoint,
		*harvestWikidataWikibaseURL); err != nil {
		return err
	}
	if err := config.SetCustomWikibaseQuery(); err != nil {
		return err
	}
	return nil
}

// jsonEscape can be used to escape a string for adding to a JSON structure
// without fear of letting unescaped special characters slip by.
func jsonEscape(str string) string {
	jsonStr, err := json.Marshal(str)
	if err != nil {
		panic(err)
	}
	str = string(jsonStr)
	return str[1 : len(str)-1]
}

// addEndpoint is designed to augment the harvest data from Wikidata
// with the source endpoint used. This information provides greater
// context for the caller.
//
// In the fullness of time, this might also be added to Wikiprov, and
// if it is, it will make the process more reliable, and this function
// redundant.
func addEndpoint(repl string, endpoint string) string {
	replacement := fmt.Sprintf(
		"{\n  \"endpoint\": \"%s\",",
		jsonEscape(endpoint),
	)
	return strings.Replace(repl, "{", replacement, 1)
}

// harvestWikidata will connect to the configured Wikidata query service
// and save the results of the configured query to disk.
func harvestWikidata() error {

	log.Printf(
		"harvesting Wikidata definitions: lang '%s'",
		config.WikidataLang(),
	)
	err := os.MkdirAll(config.WikidataHome(), os.ModePerm)
	if err != nil {
		return fmt.Errorf(
			"error harvesting Wikidata definitions: '%s'",
			err,
		)
	}
	log.Printf(
		"harvesting definitions from: '%s'",
		config.WikidataEndpoint(),
	)

	// Set the Wikibase server URL for wikiprov to construct index.php
	// and api.php links for permalinks and revision history.
	wikiprov.SetWikibaseURLs(config.WikidataWikibaseURL())

	log.Printf(
		"harvesting revision history from: '%s'",
		config.WikidataWikibaseURL(),
	)

	res, err := spargo.SPARQLWithProv(
		config.WikidataEndpoint(),
		config.WikidataSPARQL(),
		config.WikidataSPARQLRevisionParam(),
		config.GetWikidataRevisionHistoryLen(),
		config.GetWikidataRevisionHistoryThreads(),
	)

	if err != nil {
		return fmt.Errorf(
			"Error trying to retrieve SPARQL with revision history: %s",
			err,
		)
	}

	// Create a modified JSON output containing the endpoint the query
	// was run against. In future this could be added to Wikiprov.
	modifiedJSON := addEndpoint(
		fmt.Sprintf("%s", res), config.WikidataEndpoint(),
	)

	path := config.WikidataDefinitionsPath()
	err = os.WriteFile(
		path,
		[]byte(fmt.Sprintf("%s", modifiedJSON)),
		config.WikidataFileMode(),
	)
	if err != nil {
		return fmt.Errorf(
			"Error harvesting Wikidata: '%s'",
			err,
		)
	}
	log.Printf(
		"harvesting Wikidata definitions '%s' complete",
		path,
	)
	return nil
}
