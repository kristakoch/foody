package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// https://developer.edamam.com/edamam-docs-recipe-api

type edamamAPI struct {
	apiSearchBase string
	appID         string
	appKey        string
}

func newEdamamAPI(
	appID string,
	appKey string,
) (*edamamAPI, error) {
	var e edamamAPI

	if appID == "" {
		return nil, errors.New("edamam: missing app id")
	}
	if appKey == "" {
		return nil, errors.New("edamam: missing app key")
	}

	e.appID = appID
	e.appKey = appKey

	e.apiSearchBase = "https://api.edamam.com/search"

	return &e, nil
}

type edamamResponse struct {
	Hits []struct {
		Recipe struct {
			Label       string     `json:"label"`
			Image       string     `json:"image"`
			URL         string     `json:"url"`
			Ingredients []struct{} `json:"ingredients"`
			Yield       float64    `json:"yield"`
		} `json:"recipe"`
	} `json:"hits"`
}

func (e edamamAPI) fetchRecipes(searchQ string) ([]recipe, error) {
	var err error

	apiReqURL := e.buildRequestURL(searchQ)

	log.Printf("making request to edamam API with url: %s", apiReqURL)

	var resp *http.Response
	if resp, err = http.Get(apiReqURL); err != nil {
		return nil, err
	}

	var bs []byte
	if bs, err = ioutil.ReadAll(resp.Body); nil != err {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-ok response %d, %s", resp.StatusCode, string(bs))
	}

	var er edamamResponse
	if err = json.Unmarshal(bs, &er); nil != err {
		return nil, err
	}

	log.Printf("found %d total reslts for query %s", len(er.Hits), searchQ)

	recipes := e.mapResponse(er)

	return recipes, nil
}

func (e edamamAPI) buildRequestURL(searchQ string) string {
	vals := url.Values{}

	vals.Set("app_key", e.appKey)
	vals.Set("app_id", e.appID)
	vals.Set("from", "0")
	vals.Set("to", "100")

	vals.Set("q", url.QueryEscape(searchQ))

	return fmt.Sprintf("%s?%s", e.apiSearchBase, vals.Encode())
}

func (e edamamAPI) mapResponse(
	er edamamResponse,
) []recipe {
	var recipes []recipe

	for _, hit := range er.Hits {
		r := recipe{
			name:           hit.Recipe.Label,
			url:            hit.Recipe.URL,
			imageURL:       hit.Recipe.Image,
			numIngredients: len(hit.Recipe.Ingredients),
			yield:          int(hit.Recipe.Yield),
		}
		recipes = append(recipes, r)
	}

	return recipes
}
