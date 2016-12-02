package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"

	"github.com/giagiannis/data-profiler/core"
)

type simcomparisonParams struct {
	output       *string // the output file
	logfile      *string // the logfile
	similarities []*core.DatasetSimilarities
	modules      map[string]bool // modules to enable
}

func simcomparisonParseParams() *simcomparisonParams {
	params := new(simcomparisonParams)
	params.output =
		flag.String("o", "", "where to store similarities file")
	params.logfile =
		flag.String("l", "", "logfile (default: stderr)")
	input :=
		flag.String("i", "", "comma separated dataset similarities files")
	modulesString :=
		flag.String("m", "aprx", "comma separated comparisons to execute (list to view which are they)")
	flag.Parse()
	setLogger(*params.logfile)

	// list modules
	if *modulesString == "list" {
		fmt.Println("aprx - compares the SMs based on their number of fully calculated nodes")
		os.Exit(0)
	} else {
		params.modules = make(map[string]bool)
		for _, s := range strings.Split(*modulesString, ",") {
			params.modules[s] = true
		}
	}
	if *input == "" ||
		*params.output == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	// parse similarities
	simSlice := strings.Split(*input, ",")
	params.similarities = make([]*core.DatasetSimilarities, len(simSlice))
	for i, sim := range simSlice {
		params.similarities[i] = core.NewDatasetSimilarities(nil)
		f, err := os.Open(sim)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		params.similarities[i].Deserialize(buf)
	}

	return params
}

func simcomparisonRun() {
	params := simcomparisonParseParams()
	if !sanityCheck(params.similarities) {
		fmt.Fprintln(os.Stderr, "Similiarity files do not have the same datasets")
		os.Exit(1)
	}
	if val, ok := params.modules["aprx"]; ok && val {
		runAppxBasedSimilarity(params)
	}

}

// Returns the frobenius distance between two similarity matrices
func frobenius(a, b *core.DatasetSimilarities) float64 {
	datasets := a.Datasets()
	sum := 0.0
	for i, d1 := range datasets {
		for j, d2 := range datasets {
			if i > j {
				sum += math.Pow(a.Get(d1.Path(), d2.Path())-b.Get(d1.Path(), d2.Path()), 2)
			}
		}
	}
	return math.Sqrt(sum)

}

// Checks whether the provided similarity files refer to the same datasets
func sanityCheck(similarities []*core.DatasetSimilarities) bool {
	datasets := similarities[0].Datasets()
	for _, s := range similarities {
		for i, d := range s.Datasets() {
			if d.Path() != datasets[i].Path() {
				return false
			}
		}
	}
	return true
}

func runAppxBasedSimilarity(params *simcomparisonParams) {
	outfile, er := os.OpenFile(*params.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if er != nil {
		fmt.Fprintln(os.Stderr, er)
		os.Exit(1)
	}
	defer outfile.Close()

	maxFullNodes, maxFullNodesIdx := 0, 0
	for i, s := range params.similarities {
		current := s.FullyCalculatedNodes()
		if current > maxFullNodes {
			maxFullNodes = current
			maxFullNodesIdx = i
		}
	}
	fmt.Fprintf(outfile, "\"nodes\"\t\"frobenius\"\n")
	for _, s := range params.similarities {
		val := frobenius(s, params.similarities[maxFullNodesIdx])
		fmt.Fprintf(outfile, "%d\t%.5f\n",
			s.FullyCalculatedNodes(),
			val)
	}

}