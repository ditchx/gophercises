package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	var filename string
	flag.StringVar(&filename, "filename", "problems.csv", "CSV file containing the quiz questions")
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

	score := ask(questions)

	fmt.Printf("Your score is %d/%d\n", score, len(questions))
}

func ask(q [][]string) int {
	score := 0
	for i, row := range q {
		var ans string
		fmt.Printf("Problem #%d: %s = ", i+1, row[0])
		fmt.Scanln(&ans)

		if strings.ToLower(strings.Trim(ans, " ")) == strings.ToLower(strings.Trim(row[1], " ")) {
			score++
		}
	}

	return score

}
