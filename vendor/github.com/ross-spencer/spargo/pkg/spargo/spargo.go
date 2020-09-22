package spargo

/* Golang SPARQL package

Enable the querying of a SPARQL data store using Golang.

	"...Too rich for some people's tastes..."
*/

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// DefaultAgent user-agent determined by Wikidata User-agent policy: https://meta.wikimedia.org/wiki/User-Agent_policy.
const DefaultAgent string = "spargo/0.0.2 (https://github.com/ross-spencer/spargo/; all.along.the.watchtower+github@gmail.com)"

// TODO: Fix lazy error handling...
func errorHandler(err error) {
	if err != nil {
		panic(err)
	}
}

// SPARQLClient ...
type SPARQLClient struct {
	Client  *http.Client
	BaseURL string
	Agent   string
	Accept  string
	Query   string
}

// SPARQLGo takes our SparqlEndpoint structure and packages that as a request
// for our SPARQL endpoint of choice. For the given
func (endpoint *SPARQLClient) SPARQLGo() SPARQLResult {

	if endpoint.Client == nil {
		endpoint.Client = &http.Client{}
	}
	endpoint.SetUserAgent("")
	endpoint.SetAcceptHeader("")

	req, err := http.NewRequest("GET", endpoint.BaseURL, nil)
	errorHandler(err)

	req.Header.Add("User-Agent", endpoint.Agent)
	req.Header.Add("Accept", endpoint.Accept)

	query := req.URL.Query()
	query.Add("query", endpoint.Query)
	req.URL.RawQuery = query.Encode()

	resp, err := endpoint.Client.Do(req)
	defer resp.Body.Close()

	errorHandler(err)
	if resp.StatusCode != 200 {
		log.Printf("Error: unknown response from server: %s", resp.Status)
		return SPARQLResult{}
	}

	body, err := ioutil.ReadAll(resp.Body)
	errorHandler(err)

	var sparqlResponse SPARQLResult
	sparqlResponse.Human = string(body)

	err = json.Unmarshal(body, &sparqlResponse)
	errorHandler(err)

	return sparqlResponse
}

// SetUserAgent agent allows the user to set a custom user agent or use the
// library's default.
func (endpoint *SPARQLClient) SetUserAgent(agent string) {
	if agent == "" {
		agent = DefaultAgent
	}
	endpoint.Agent = agent
}

// SetQuery enables us to set the SPARQL query.
func (endpoint *SPARQLClient) SetQuery(queryString string) {
	if queryString == "" {
		// Shall we perform some error handling here?
	}
	endpoint.Query = queryString
}

// SetAcceptHeader will allow us to request results in other data formats. Our
// default is SPARQL JSON.
func (endpoint *SPARQLClient) SetAcceptHeader(accept string) {
	endpoint.Accept = "application/sparql-results+json, application/json"
}

// SetURL lets us set the URL of the SPARQL endpoint to query.
func (endpoint *SPARQLClient) SetURL(url string) {
	endpoint.BaseURL = url
}

// ClientInit provides us with a helper function to set endpoint URL and
// query string in a single go.
func (endpoint *SPARQLClient) ClientInit(url string, queryString string) {
	endpoint.SetURL(url)
	endpoint.SetQuery(queryString)
}
