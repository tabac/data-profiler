package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/giagiannis/data-profiler/core"
)

type expAccuracyParams struct {
	mlScript    *string // script used for approximation
	output      *string // output file path
	repetitions *int    // number of times to repeat experiment
	threads     *int    // number of threads to utilize

	coords []core.DatasetCoordinates // coords of datasets
	scores []float64                 // scores of datasets

	samplingRates []float64 // samplings rates to run
}

func expAccuracyParseParams() *expAccuracyParams {
	params := new(expAccuracyParams)
	params.mlScript =
		flag.String("ml", "", "ML script to use for approximation")
	params.output =
		flag.String("o", "", "output path")
	params.repetitions =
		flag.Int("r", 1, "number of repetitions")
	params.threads =
		flag.Int("t", 1, "number of threads")
	loger :=
		flag.String("l", "", "log file")

	coordsFile :=
		flag.String("c", "", "coordinates file")
	scoresFile :=
		flag.String("s", "", "scores file")
	idxFile :=
		flag.String("i", "", "index file")
	srString :=
		flag.String("sr", "", "comma separated sampling rates")

	flag.Parse()
	setLogger(*loger)
	if *params.mlScript == "" || *params.output == "" || *coordsFile == "" ||
		*scoresFile == "" || *idxFile == "" || *srString == "" {
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// sampling rates parsing
	a := strings.Split(*srString, ",")
	params.samplingRates = make([]float64, 0)
	for i := range a {
		v, err := strconv.ParseFloat(a[i], 64)
		if err == nil {
			params.samplingRates = append(params.samplingRates, v)
		}
	}

	// idx file parsing
	f, err := os.Open(*idxFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	idx := make([]string, 0)
	for i, line := range strings.Split(string(buf), "\n") {
		a := strings.Split(line, "\t")
		if len(a) == 2 {
			j, err := strconv.ParseInt(a[0], 10, 32)
			if err != nil || int(j) != i {
				log.Fatalln(err)
			}
			idx = append(idx, a[1])
		}
	}
	f.Close()

	// coordinates file parsing
	f, err = os.Open(*coordsFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	params.coords = make([]core.DatasetCoordinates, 0)
	for i, line := range strings.Split(string(buf), "\n") {
		a := strings.Split(line, " ")
		res := make(core.DatasetCoordinates, 0)
		if i > 0 && len(a) > 0 {
			for _, s := range a {
				if s != "" {
					v, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Fatalln(err)
					}
					res = append(res, v)
				}
			}
			if len(res) > 0 {
				params.coords = append(params.coords, res)
			}
		}
	}
	f.Close()

	// scores
	f, err = os.Open(*scoresFile)
	if err != nil {
		log.Fatalln(err)
	}
	buf, err = ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	scores := core.NewDatasetScores()
	scores.Deserialize(buf)
	params.scores = make([]float64, len(scores.Scores))
	for i, path := range idx {
		params.scores[i] = scores.Scores[path]
	}
	f.Close()

	return params
}

func expAccuracyRun() {
	// inititializing steps
	params := expAccuracyParseParams()
	rand.Seed(int64(time.Now().Nanosecond()))
	output := setOutput(*params.output)
	fmt.Fprintf(output, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		"sr",
		"avg",
		"perc-0",
		"perc-25",
		"perc-50",
		"perc-75",
		"perc-100",
	)

	// create random permutation
	slice := make([]int, len(params.coords))
	for i := 0; i < len(slice); i++ {
		slice[i] = i
	}

	testset := generateSet(slice[0:int(float64(len(slice))*1.0)], params.coords, params.scores)

	executeScript := func(script, trainset, testset string) float64 {
		cmd := exec.Command(script, trainset, testset)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
		}
		val, err := strconv.ParseFloat(string(out), 64)
		if err != nil {
			log.Println(err)
		}
		return val
	}

	// execute
	for _, sr := range params.samplingRates {
		results := make([]float64, 0)
		eval := func(sr float64) float64 {
			perm := rand.Perm(len(params.coords))
			trainsetIndexes := perm[0:int(float64(len(perm))*sr)]
			trainset := generateSet(trainsetIndexes, params.coords, params.scores)
			modelError := executeScript(*params.mlScript, trainset, testset)
			os.Remove(trainset)
			return modelError
		}
		done := make(chan float64)
		slots := make(chan bool, *params.threads)
		for i := 0; i < *params.threads; i++ {
			slots <- true
		}

		for i := 0; i < *params.repetitions; i++ {
			go func(done chan float64, slots chan bool, repetition int) {
				log.Printf("[thread-%d] Starting calculation for SR %.2f\n", repetition, sr)
				<-slots
				done <- eval(sr)
				slots <- true
				log.Printf("[thread-%d] Done calculation for SR %.2f\n", repetition, sr)
			}(done, slots, i)
		}
		for i := 0; i < *params.repetitions; i++ {
			v := <-done
			results = append(results, v)
		}
		fmt.Fprintf(output, "%.5f\t%.5f\t%.5f\t%.5f\t%.5f\t%.5f\t%.5f\n",
			sr,
			getAverage(results),
			getPercentile(results, 0),
			getPercentile(results, 25),
			getPercentile(results, 50),
			getPercentile(results, 75),
			getPercentile(results, 100),
		)
	}
	os.Remove(testset)
}

// UTILS

// executes the script and get model errors