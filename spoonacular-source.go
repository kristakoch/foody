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

// https://spoonacular.com/food-api/docs

type spoonacularAPI struct {
	appKey        string
	apiSearchBase string
}

func newSpoonacularAPI(
	appKey string,
) (*spoonacularAPI, error) {
	var s spoonacularAPI

	if appKey == "" {
		return nil, errors.New("spoonacular: missing app key")
	}

	s.appKey = appKey

	s.apiSearchBase = "https://api.spoonacular.com/recipes/complexSearch"

	return &s, nil
}

type spoonacularResponse struct {
	Results []struct {
		Title    string `json:"title"`
		Image    string `json:"image"`
		URL      string `json:"sourceUrl"`
		Servings int    `json:"servings"`
	} `json:"results"`
}

func (s spoonacularAPI) fetchRecipes(searchQ string) ([]recipe, error) {
	var err error

	apiReqURL := s.buildRequestURL(searchQ)

	log.Printf("making request to spoonacular API with url: %s", apiReqURL)

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

	var sr spoonacularResponse
	if err = json.Unmarshal(bs, &sr); nil != err {
		return nil, err
	}

	if len(sr.Results) == 0 {
		return nil, fmt.Errorf("no results found for search query '%s'", searchQ)
	}

	log.Printf("found %d total reslts for query %s", len(sr.Results), searchQ)

	recipes := s.mapResponse(sr)

	return recipes, nil
}

func (s spoonacularAPI) buildRequestURL(searchQ string) string {
	vals := url.Values{}

	vals.Set("apiKey", s.appKey)
	vals.Set("addRecipeInformation", "true")
	vals.Set("number", "100")

	vals.Set("query", url.QueryEscape(searchQ))

	return fmt.Sprintf("%s?%s", s.apiSearchBase, vals.Encode())
}

func (s spoonacularAPI) mapResponse(
	sr spoonacularResponse,
) []recipe {
	var recipes []recipe

	for _, result := range sr.Results {
		r := recipe{
			name:     result.Title,
			imageURL: result.Image,
			url:      result.URL,
			yield:    result.Servings,
		}
		recipes = append(recipes, r)
	}

	return recipes
}
