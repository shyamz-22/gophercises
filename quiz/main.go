package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"fmt"
	"os"
	"encoding/csv"
	"time"
)

var (
	quizFile      = kingpin.Flag("quiz-file", "csv file containing quiz questions").Default("fixtures/problems.csv").Short('q').String()
	totalDuration = kingpin.Flag("timeout", "Total duration of Quiz").Default("10s").Short('t').Duration()
)

type Problem struct {
	Question       string
	ExpectedAnswer string
}

func main() {
	kingpin.Parse()

	f, err := os.Open(*quizFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		panic(err)
	}

	problems := parseLines(lines)

	correctAnswers := 0
	timer := time.NewTimer(*totalDuration)

problemLoop:
	for _, p := range problems {
		fmt.Println(fmt.Sprintf("What is  %s?", p.Question))
		answerCh := make(chan string)
		go func() {
			var answer string
			fmt.Scanln(&answer)
			answerCh <- answer
		}()

		select {
		case <-timer.C:
			fmt.Println("Your Time is up :/")
			break problemLoop
		case actualAnswer := <-answerCh:
			if p.ExpectedAnswer == actualAnswer {
				correctAnswers++
			}
		}
	}

	fmt.Println(fmt.Sprintf("you answered %d out of %d correctly", correctAnswers, len(problems)))
}
func parseLines(lines [][]string) []Problem {
	problems := make([]Problem, len(lines))

	for i, p := range lines {
		problems[i] = Problem{
			Question:       p[0],
			ExpectedAnswer: p[1],
		}
	}

	return problems
}
