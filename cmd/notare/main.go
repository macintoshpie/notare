package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/macintoshpie/notare/parser"
	"github.com/macintoshpie/notare/templatizer"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Starting...")
	file, err := os.Open("examples.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	err = os.MkdirAll("public", 0777)
	if err != nil {
		log.Fatal(err)
	}

	examplesDir := "./examples"
	allExamples := []*parser.Example{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		exampleName := scanner.Text()
		example := parser.ParseExample(examplesDir, exampleName)
		allExamples = append(allExamples, example)
	}

	for idx, example := range allExamples {
		if idx > 0 {
			example.PreviousExample = allExamples[idx-1]
		}
		if idx < len(allExamples)-1 {
			example.NextExample = allExamples[idx+1]
		}

		exampleOutput, err := os.Create(fmt.Sprintf("./public/%s.html", example.Name))
		if err != nil {
			log.Fatal(err)
		}
		defer exampleOutput.Close()
		templatizer.TemplatizeExample(example, "", exampleOutput)
	}

	indexOutput, err := os.Create("./public/index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer indexOutput.Close()
	templatizer.TemplatizeIndex(allExamples, "", indexOutput)

	fmt.Println("Generating CSS styles...")
	cssOutput, err := os.Create("./public/highlight.css")
	if err != nil {
		log.Fatal(err)
	}
	defer cssOutput.Close()
	templatizer.GenerateStyles(cssOutput)

	fmt.Println("Finished Successfully")
}
