package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	ansiEscLtGreen = "\033[0;92m"
	ansiEscNoColor = "\033[0m"
	resultPageSize = 5

	sourceEdamamAPI      = "edamam"
	sourceSpoonacularAPI = "spoonacular"
	sourceCSV            = "csv"
)

// recipe is a type for holding app-level recipe data.
type recipe struct {
	name           string
	url            string
	location       string
	imageURL       string
	time           string
	numIngredients int
	yield          int
}

// recipeSource is a source of recipe results based on the
// passed-in space separated search query.
type recipeSource interface {
	fetchRecipes(searchQ string) ([]recipe, error)
}

func newRecipeSource(cfg config) (recipeSource, error) {
	log.Printf("input source is '%s'", cfg.source)

	switch cfg.source {
	case sourceEdamamAPI:
		return newEdamamAPI(cfg.edamam.appID, cfg.edamam.appKey)
	case sourceSpoonacularAPI:
		return newSpoonacularAPI(cfg.spoonacular.appKey)
	case sourceCSV:
		return newRecipeCSV(cfg.csv.location)
	}

	return nil, fmt.Errorf("'%s' is not a valid choice of source", cfg.source)
}

func (r recipe) String() string {
	var rs string

	rs += fmt.Sprintf("%s%s%s\n", ansiEscLtGreen, r.name, ansiEscNoColor)

	rs += msgOnCond(r.imageURL != "", fmt.Sprintln("picture →", r.imageURL))
	rs += msgOnCond(r.url != "", fmt.Sprintln("url →", r.url))
	rs += msgOnCond(r.time != "", fmt.Sprintln("time →", r.time))
	rs += msgOnCond(r.numIngredients > 0, fmt.Sprintln("# ingredients →", r.numIngredients))
	rs += msgOnCond(r.yield > 0, fmt.Sprintln("# yield →", r.yield))

	return rs
}

func msgOnCond(cond bool, s string) string {
	if !cond {
		return ""
	}
	return s
}

func (r recipe) longString() error {
	var err error

	fmt.Println("")

	fmt.Printf("%s%s%s\n", ansiEscLtGreen, r.name, ansiEscNoColor)

	if r.url != "" {
		fmt.Println(r.url)
	}
	if r.location != "" {
		fmt.Println(r.location)
	}

	fmt.Println("")

	if strings.HasSuffix(r.imageURL, ".jpg") {
		cmd := exec.Command("jp2a", "--width=40", r.imageURL)
		var out bytes.Buffer
		cmd.Stdout = &out

		if err = cmd.Run(); err != nil {
			log.Printf("error running jp2a command, %s", err)
			return err
		}

		fmt.Print(out.String())
	}

	fmt.Println("")

	return nil
}
