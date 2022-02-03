package templatizer

import (
	_ "embed"
	"io"
	"io/ioutil"
	"log"
	"text/template"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/macintoshpie/notare/parser"
)

//go:embed templates/example.tmpl
var defaultExampleTemplate string

//go:embed templates/index.tmpl
var defaultIndexTemplate string

func TemplatizeExample(example *parser.Example, templatePath string, w io.Writer) {
	exampleTmpl := template.New("example")

	var templateString string
	if len(templatePath) > 0 {
		bytes, err := ioutil.ReadFile(templatePath)
		if err != nil {
			log.Fatal(err)
		}
		templateString = string(bytes)
	} else {
		templateString = defaultExampleTemplate
	}

	_, err := exampleTmpl.Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	exampleTmpl.Execute(w, example)
}

func TemplatizeIndex(examples []*parser.Example, templatePath string, w io.Writer) {
	indexTmpl := template.New("index")

	var templateString string
	if len(templatePath) > 0 {
		bytes, err := ioutil.ReadFile(templatePath)
		if err != nil {
			log.Fatal(err)
		}
		templateString = string(bytes)
	} else {
		templateString = defaultIndexTemplate
	}

	_, err := indexTmpl.Parse(templateString)
	if err != nil {
		log.Fatal(err)
	}

	indexTmpl.Execute(w, examples)
}

func GenerateStyles(w io.Writer) {
	style := styles.Get("autumn")
	formatter := chromahtml.New(chromahtml.WithClasses(true))

	err := formatter.WriteCSS(w, style)
	if err != nil {
		log.Fatal(err)
	}
}
