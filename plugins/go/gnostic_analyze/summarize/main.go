// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// summarize is a tool for summarizing the results of gnostic_analyze runs.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/googleapis/gnostic/plugins/go/gnostic_analyze/statistics"
)

// Results are collected in this global slice.
var stats []statistics.DocumentStatistics

// walker is called for each summary file found.
func walker(p string, info os.FileInfo, err error) error {
	basename := path.Base(p)
	if basename != "summary.json" {
		return nil
	}
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	var s statistics.DocumentStatistics
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	stats = append(stats, s)
	return nil
}

func printFrequencies(m map[string]int) {
	for _, pair := range rankByCount(m) {
		fmt.Printf("%6d %s\n", pair.Value, pair.Key)
	}
}

func rankByCount(frequencies map[string]int) PairList {
	pl := make(PairList, len(frequencies))
	i := 0
	for k, v := range frequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	// Collect all statistics in the current directory and its subdirectories.
	stats = make([]statistics.DocumentStatistics, 0)
	filepath.Walk(".", walker)

	// Compute some interesting properties.
	apis_with_anonymous_ops := 0
	op_frequencies := make(map[string]int, 0)
	parameter_type_frequencies := make(map[string]int, 0)
	result_type_frequencies := make(map[string]int, 0)
	definition_field_type_frequencies := make(map[string]int, 0)
	definition_array_type_frequencies := make(map[string]int, 0)

	for _, api := range stats {
		if api.Operations["anonymous"] != 0 {
			apis_with_anonymous_ops += 1
			fmt.Printf("%s has anonymous operations %d/%d\n",
				api.Name,
				api.Operations["anonymous"],
				api.Operations["total"])
		}
		for k, v := range api.Operations {
			op_frequencies[k] += v
		}
		for k, v := range api.ParameterTypes {
			parameter_type_frequencies[k] += v
		}
		for k, v := range api.ResultTypes {
			result_type_frequencies[k] += v
		}
		for k, v := range api.DefinitionFieldTypes {
			definition_field_type_frequencies[k] += v
		}
		for k, v := range api.DefinitionArrayTypes {
			definition_array_type_frequencies[k] += v
		}
	}

	// Report the results.
	fmt.Printf("\n")
	fmt.Printf("Collected information on %d APIs.\n\n", len(stats))
	fmt.Printf("APIs with anonymous operations: %d\n", apis_with_anonymous_ops)
	fmt.Printf("\nOperation frequencies:\n")
	printFrequencies(op_frequencies)
	fmt.Printf("\nParameter type frequencies:\n")
	printFrequencies(parameter_type_frequencies)
	fmt.Printf("\nResult type frequencies:\n")
	printFrequencies(result_type_frequencies)
	fmt.Printf("\nDefinition object field type frequencies:\n")
	printFrequencies(definition_field_type_frequencies)
	fmt.Printf("\nDefinition array type frequencies:\n")
	printFrequencies(definition_array_type_frequencies)
}