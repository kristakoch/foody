package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	source string

	spoonacular struct {
		appKey string
	}

	edamam struct {
		appID  string
		appKey string
	}

	csv struct {
		location string
	}
}

func main() {
	var err error

	log := log.New(os.Stdout, "FOODY ", log.LstdFlags|log.Lshortfile)

	var cfg config
	flag.StringVar(&cfg.source, "source", os.Getenv("SOURCE"), "choice of recipe source")
	flag.StringVar(&cfg.spoonacular.appKey, "spoonacularappkey", os.Getenv("SPOONACULAR_APP_KEY"), "app key for spoonacular")
	flag.StringVar(&cfg.edamam.appID, "edamamappid", os.Getenv("EDAMAM_APP_ID"), "app id for edamam")
	flag.StringVar(&cfg.edamam.appKey, "edamamappkey", os.Getenv("EDAMAM_APP_KEY"), "app key for edamam")
	flag.StringVar(&cfg.csv.location, "csvlocation", os.Getenv("CSV_LOCATION"), "file location for recipe csv")
	flag.Parse()

	var source recipeSource
	if source, err = newRecipeSource(cfg); nil != err {
		log.Fatal(err)
	}

	if err = run(source); nil != err {
		log.Fatal(err)
	}
}

func run(source recipeSource) error {
	var err error

	var searchQ string
	for searchQ == "" {
		log.Println("what kind of food are you looking for? enter one or more words separated by spaces")

		in := bufio.NewReader(os.Stdin)

		searchQ, err = in.ReadString('\n')
		if err != nil {
			return err
		}

		searchQ = strings.TrimSpace(searchQ)
	}

	var recipes []recipe
	if recipes, err = source.fetchRecipes(searchQ); nil != err {
		return err
	}

	if len(recipes) == 0 {
		log.Printf("no results found for search query '%s', exiting", searchQ)
		return nil
	}

	log.Printf("found %d total recipe recommendations", len(recipes))

	var idx int
	for {
		var recipePage []recipe
		switch {
		case idx+resultPageSize > len(recipes):
			recipePage = recipes[idx:]
		default:
			recipePage = recipes[idx : idx+resultPageSize]
		}

		for innerIdx, r := range recipePage {
			fmt.Printf("%d. %s", idx+innerIdx+1, r.String())
			fmt.Println("")
		}

		log.Println("hit 'f' to go the next page of results, 'b' to go back, or enter if you've found a recipe you like")

		in := bufio.NewReader(os.Stdin)

		var navChoice string
		if navChoice, err = in.ReadString('\n'); err != nil {
			return err
		}

		navChoice = strings.TrimSpace(navChoice)

		if navChoice == "" {
			break
		}

		switch navChoice {
		case "f":
			if (idx + resultPageSize) > len(recipes) {
				log.Print("cannot go forward, this is the last page")

				// Sleep so the user has some time to notice the message.
				time.Sleep(2 * time.Second)
				break
			}
			idx += resultPageSize
		case "b":
			if (idx - resultPageSize) < 0 {
				log.Print("cannot go back, this is the first page")

				// Sleep so the user has some time to notice the message.
				time.Sleep(2 * time.Second)
				break
			}
			idx -= resultPageSize
		default:
			log.Printf("'%s' is not a valid choice, please enter 'f', 'b', or 'd'", navChoice)
		}
	}

	// Prompt the user for a recipe choice.
	var recipeIdx int
	var valid bool
	for !valid {
		var choice string
		log.Println("find something good? enter the recipe number or 'n' for no")

		fmt.Scanln(&choice)

		if choice == "n" {
			log.Println("that's fair. dying.")
			return nil
		}
		if recipeIdx, err = strconv.Atoi(choice); nil != err {
			continue
		}
		if recipeIdx < 0 || recipeIdx > len(recipes) {
			continue
		}

		recipeIdx = recipeIdx - 1
		valid = true
	}

	recipe := recipes[recipeIdx]

	log.Println("you chose:")

	if err = recipe.longString(); nil != err {
		return err
	}

	return nil
}
