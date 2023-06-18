package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
	"unicode"
)

func main() {
	input := flag.String("input", "", "Location of puzzle to read or stdin by default")
	output := flag.String("output", "", "Output location for solution, or stdout by default")
	printPartials := flag.Bool("print-partials", false, "Print each partial puzzle to stdout")
	delay := flag.Int("delay", 0, "Add delay between each iteration (useful with `--print-partials`)")

	flag.Parse()

	// get input
	var inFile []byte
	if *input == "" {
		f, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Could not read stdin: ", err)
			os.Exit(1)
		}
		inFile = f
	} else {
		f, err := os.ReadFile(*input)
		if err != nil {
			fmt.Println("Could not read ", input, ": ", err)
			os.Exit(1)
		}
		inFile = f
	}

	// get output
	var outFile *os.File
	if *output == "" {
		outFile = os.Stdout
	} else {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Println("Could not open output: ", err)
			os.Exit(1)
		}
		outFile = f
	}

	candidate, err := parse(inFile)
	if err != nil {
		fmt.Println("Could not parse input: ", err)
		os.Exit(1)
	}
	root := Node{
		Candidate:  candidate,
		MostRecent: 0,
		Children:   [9]*Node{nil},
	}
	config := Config{Input: inFile, Output: outFile, PrintPartials: *printPartials, Delay: *delay}

	solution := root.Backtrack(&config)
	if solution == nil {
		fmt.Println("Could not solve puzzle")
		os.Exit(1)
	}

	solution.PrintSolution(&config)
}

type Config struct {
	Input         []byte
	Output        *os.File
	PrintPartials bool
	Delay         int
}

type Node struct {
	Candidate  [81]uint8
	MostRecent uint8
	Children   [9]*Node
}

func (n *Node) Backtrack(config *Config) *Node {
	if config.Delay != 0 {
		time.Sleep(time.Second * time.Duration(config.Delay))
	}
	if config.PrintPartials {
		n.PrintSolution(config)
	}

	if n.Reject() {
		return nil
	} else if n.Accept() {
		return n
	}

	toChange := n.First()
	next := n.Children[n.MostRecent]

	for next != nil {
		solution := next.Backtrack(config)
		if solution != nil {
			return solution
		}

		if n.Next(toChange) {
			next = n.Children[n.MostRecent]
		} else {
			break
		}
	}
	return nil
}

func (n *Node) First() uint8 {
	child := copyCandidate(n)

	var toChange uint8
	for i, v := range n.Candidate {
		if v == 0 {
			toChange = uint8(i)
			break
		}
	}

	child.Candidate[toChange] = 1

	n.Children[n.MostRecent] = &child

	return toChange
}

func (n *Node) Next(toChange uint8) bool {
	if n.MostRecent >= 8 {
		return false
	}
	prev := n.Children[n.MostRecent]
	child := copyCandidate(prev)
	child.Candidate[toChange] += 1

	n.MostRecent += 1
	n.Children[n.MostRecent] = &child

	return true
}

func (n *Node) Accept() bool {
	for _, v := range n.Candidate {
		if v == 0 {
			return false
		}
	}
	return true
}

func (n *Node) Reject() bool {
	counter := [10]uint8{0}
	// Check horizontal rows
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			counter[n.Candidate[(i*9+j)]] += 1
		}

		for q, j := range counter[1:] {
			if j > 1 {
				return true
			}

			counter[q+1] = 0
		}
	}

	// check vertical rows
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			counter[n.Candidate[(j*9+i)]] += 1
		}

		for q, j := range counter[1:] {
			if j > 1 {
				return true
			}

			counter[q+1] = 0
		}
	}

	// check squares
	// traverse rows
	for i := 0; i < 3; i++ {
		// traverse columns
		for j := 0; j < 3; j++ {
			offset := (i * 27) + (j * 3)
			// traverse row in square
			for k := 0; k < 3; k++ {
				counter[n.Candidate[offset+k*9]] += 1
				counter[n.Candidate[offset+k*9+1]] += 1
				counter[n.Candidate[offset+k*9+2]] += 1
			}

			for q, k := range counter[1:] {
				if k > 1 {
					return true
				}

				counter[q+1] = 0
			}
		}
	}

	return false
}

func (n *Node) PrintSolution(config *Config) {
	for i := 0; i < 3; i++ {
		offset := i * 27
		io.WriteString(config.Output, "+-------+-------+-------+\n")
		for j := 0; j < 3; j++ {
			io.WriteString(
				config.Output,
				fmt.Sprintf("| %d %d %d | %d %d %d | %d %d %d |\n",
					n.Candidate[offset],
					n.Candidate[offset+1],
					n.Candidate[offset+2],
					n.Candidate[offset+3],
					n.Candidate[offset+4],
					n.Candidate[offset+5],
					n.Candidate[offset+6],
					n.Candidate[offset+7],
					n.Candidate[offset+8]),
			)
			offset += 9
		}
	}
	io.WriteString(config.Output, "+-------+-------+-------+\n")
}

func parse(contents []byte) ([81]uint8, error) {
	i := 0
	puzzle := [81]uint8{0}

	for _, c := range contents {
		if unicode.IsSpace(rune(c)) {
			continue
		}

		if unicode.IsDigit(rune(c)) {
			n := c - '0'
			puzzle[i] = n
		} else {
			puzzle[i] = 0
		}
		i += 1
		if i == 81 {
			break
		}
	}

	if i < 81 {
		return puzzle, errors.New("not enough input")
	} else {
		return puzzle, nil
	}
}

func copyCandidate(n *Node) Node {
	can := n.Candidate
	return Node{
		MostRecent: 0,
		Children:   [9]*Node{nil},
		Candidate:  can,
	}
}
