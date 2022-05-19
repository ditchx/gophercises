package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	var filename string
	var limit int
	var shuffle bool

	flag.StringVar(&filename, "filename", "problems.csv", "CSV file containing the quiz questions")
	flag.IntVar(&limit, "limit", 30, "Time limit in seconds. Default is 30.")
	flag.BoolVar(&shuffle, "shuffle", false, "Shuffle questions")
	flag.Parse()

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Unable to open file %s, %s", filename, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	questions, err := r.ReadAll()

	if err != nil {
		log.Fatalf("Failed to read CSV file %s, %s", filename, err)
	}

	if shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	}

	fmt.Println("Press any key to start...")
	fmt.Scanln()

	score := ask(questions, limit)

	fmt.Printf("Your score is %d/%d.\n", score, len(questions))
}

func ask(q [][]string, limit int) int {

	var score int
	timesOut := time.After(time.Duration(limit) * time.Second)
	points := make(chan int)
	done := true
	for i, row := range q {
		var ans string
		fmt.Printf("Problem #%d: %s = ", i+1, row[0])

		go func() {
			fmt.Scanln(&ans)

			if strings.ToLower(strings.Trim(ans, " ")) == strings.ToLower(strings.Trim(row[1], " ")) {
				points <- 1
			} else {
				points <- 0
			}

		}()

		select {
		case p := <-points:
			score += p
		case <-timesOut:
			done = false
			fmt.Printf("\n\nTime's up!\n")
			return score

		}

	}

	if done {
		fmt.Printf("\nAll done!\n")
	}

	return score
}
