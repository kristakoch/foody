package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type recipeCSV struct {
	location string
	fieldMap map[int]string
}

func newRecipeCSV(
	location string,
) (*recipeCSV, error) {
	var r recipeCSV

	if location == "" {
		return nil, errors.New("recipe csv: missing file location")
	}
	if !strings.HasSuffix(location, ".csv") {
		return nil, errors.New("recipe csv must have .csv extension")
	}
	if _, err := os.Stat(location); nil != err {
		return nil, err
	}

	r.location = location

	r.fieldMap = map[int]string{
		0: "name",
		1: "url",
		2: "time",
		3: "num_ingredients",
		4: "ingredients",
		5: "directions",
	}

	return &r, nil
}

type csvRow struct {
	name           string
	url            string
	time           string
	numIngredients string
	ingredients    string
	directions     string
	location       string
}

func (r recipeCSV) fetchRecipes(searchQ string) ([]recipe, error) {
	var err error

	var csvf *os.File
	if csvf, err = os.Open(r.location); nil != err {
		return nil, err
	}
	defer csvf.Close()

	csvr := csv.NewReader(csvf)

	var row []string
	if row, err = csvr.Read(); nil != err {
		return nil, err
	}

	if err = r.validateFirstRow(row); nil != err {
		return nil, err
	}

	searchWords := strings.Split(searchQ, " ")

	var hits []csvRow
	rowNumber := 1
	for {
		if row, err = csvr.Read(); nil != err {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if len(row) != len(r.fieldMap) {
			// Skip rows with incorrect number of fields.
			rowNumber++
			continue
		}

		csvRcp := csvRow{
			name:           row[0],
			url:            row[1],
			time:           row[2],
			numIngredients: row[3],
			ingredients:    row[4],
			directions:     row[5],
			location:       fmt.Sprintf("%s: row %d", r.location, rowNumber),
		}

		title := strings.ToLower(csvRcp.name)
		for _, w := range searchWords {
			if strings.Contains(title, w) {
				hits = append(hits, csvRcp)
			}
		}

		rowNumber++
	}

	recipes := r.mapRows(hits)

	return recipes, nil
}

func (r recipeCSV) validateFirstRow(row []string) error {
	if len(row) < len(r.fieldMap) {
		return fmt.Errorf("csv contains %d cols in the first row, expected %d", len(row), len(r.fieldMap))
	}

	for idx, hdr := range r.fieldMap {
		if row[idx] != hdr {
			return fmt.Errorf("expected header %s for col %d, got header %s", hdr, idx, row[idx])
		}
	}

	return nil
}

func (r recipeCSV) mapRows(
	rows []csvRow,
) []recipe {
	var recipes []recipe

	for _, r := range rows {
		numIngts, err := strconv.Atoi(r.numIngredients)
		if nil != err {
			numIngts = 0
		}

		r := recipe{
			name:           r.name,
			url:            r.url,
			numIngredients: numIngts,
			time:           r.time,
			location:       r.location,
		}
		recipes = append(recipes, r)
	}

	return recipes
}
