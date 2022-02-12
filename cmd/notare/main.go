package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/macintoshpie/notare/parser"
	"github.com/macintoshpie/notare/templatizer"
	"github.com/urfave/cli/v2"
)

func build(ctx *cli.Context) error {
	fmt.Println("Starting...")
	file, err := os.Open("examples.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	err = os.MkdirAll("public", 0777)
	if err != nil {
		return err
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
			return err
		}
		defer exampleOutput.Close()
		templatizer.TemplatizeExample(example, "", exampleOutput)
	}

	indexOutput, err := os.Create("./public/index.html")
	if err != nil {
		return err
	}
	defer indexOutput.Close()
	templatizer.TemplatizeIndex(allExamples, "", indexOutput)

	fmt.Println("Generating CSS styles...")
	cssOutput, err := os.Create("./public/highlight.css")
	if err != nil {
		return err
	}
	defer cssOutput.Close()
	templatizer.GenerateStyles(cssOutput)

	fmt.Println("Finished Successfully")

	return nil
}

var builtinMimeTypesLower = map[string]string{
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".htm":  "text/html; charset=utf-8",
	".html": "text/html; charset=utf-8",
	".jpg":  "image/jpeg",
	".js":   "application/javascript",
	".wasm": "application/wasm",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".xml":  "text/xml; charset=utf-8",
}

func staticFileGetMimeType(ext string) string {
	if v, ok := builtinMimeTypesLower[ext]; ok {
		return v
	}
	return mime.TypeByExtension(ext)
}

func serve(ctx *cli.Context) error {
	// A hacky and insecure way to serve static files
	// Not using FileServer b/c we were having issues with incorrect content-types
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resourcePath := r.RequestURI
		if strings.HasSuffix(resourcePath, "/") {
			resourcePath = resourcePath + "index.html"
		}

		fileBytes, err := ioutil.ReadFile("public" + resourcePath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		contentType := staticFileGetMimeType(filepath.Ext(resourcePath))

		w.Header().Add("Content-Type", contentType)
		w.Write(fileBytes)
	})
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:   "build",
				Usage:  "Build HTML from examples",
				Action: build,
			},
			{
				Name:   "serve",
				Usage:  "Start server",
				Action: serve,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
