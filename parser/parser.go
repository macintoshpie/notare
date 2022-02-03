package parser

import (
	"bufio"
	"bytes"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/antchfx/htmlquery"
	"github.com/gomarkdown/markdown"
	"golang.org/x/net/html"
)

type Row struct {
	Doc       string
	DocSpan   int
	DocHTML   template.HTML
	Code      string
	CodeHTML  template.HTML
	CodeEmpty bool
	FirstCode bool
}

type Example struct {
	Id              string
	Name            string
	Rows            []*Row
	FullCode        template.JS
	PreviousExample *Example
	NextExample     *Example
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ParseExample(examplesDir, exampleFileName string) *Example {
	file, err := os.Open(path.Join(examplesDir, exampleFileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var rows []*Row
	var row *Row = &Row{}
	inSpan := false
	var spanningRow *Row = nil
	fullCode := ""
	foundFirstCode := false
	for scanner.Scan() {
		line := scanner.Text()
		commentPrefix := regexp.MustCompile(`^\s*#\s*`)
		if commentPrefix.MatchString(line) {
			// Process comment line (ie documentation)
			line = commentPrefix.ReplaceAllString(line, "")

			switch strings.TrimSpace(line) {
			case "::span-comment":
				inSpan = true
				spanningRow = row
			case "::end-span":
				inSpan = false
				row = &Row{}
			case "::newline":
				row.Doc += "  \n\n"
			default:
				// update the row's documentation
				if len(row.Doc) > 0 {
					line = " " + line
				}
				row.Doc += line
			}
		} else {
			// Process code line

			// skip completely empty lines
			if strings.TrimSpace(line) == "" && row.Doc == "" {
				continue
			}
			if strings.TrimSpace(line) != "" {
				fullCode += line + "\n"
			}

			row.Code = line
			if !foundFirstCode && line != "" {
				foundFirstCode = true
				row.FirstCode = true
			}

			if inSpan {
				spanningRow.DocSpan += 1
				rows = append(rows, row)
				// since we're still spanning a doc, the next row should have no doc (ie no span)
				row = &Row{}
				row.DocSpan = -1
			} else {
				row.DocSpan = 1
				rows = append(rows, row)
				row = &Row{}
			}
		}
	}

	err = scanner.Err()
	if err != nil {
		log.Fatal(err)
	}

	b := new(bytes.Buffer)
	style := styles.Get("autumn")
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	lexer := lexers.Get("yaml")
	iterator, err := lexer.Tokenise(nil, fullCode)
	if err != nil {
		log.Fatal(err)
	}
	err = formatter.Format(b, style, iterator)
	if err != nil {
		log.Fatal(err)
	}
	codeDoc, err := htmlquery.Parse(strings.NewReader(b.String()))
	codeRows := htmlquery.Find(codeDoc, "//span[@class=\"line\"]")

	codeRowIdx := -1
	for _, row := range rows {
		// change docs markdown to html
		row.DocHTML = template.HTML(markdown.ToHTML([]byte(row.Doc), nil, nil))

		// if code is empty
		row.CodeEmpty = strings.TrimSpace(row.Code) == ""

		if row.FirstCode {
			codeRowIdx = 0
		}

		if codeRowIdx >= 0 {
			b := new(bytes.Buffer)
			err := html.Render(b, codeRows[codeRowIdx])
			if err != nil {
				log.Fatal(err)
			}
			row.CodeHTML = template.HTML(b.String())
			codeRowIdx += 1
		}
	}

	// make the code safe for inserting in template string
	fullCode = strings.Replace(fullCode, "`", "\\`", -1)
	fullCode = strings.Replace(fullCode, "$", "\\$", -1)

	return &Example{
		Id:              strings.TrimSuffix(exampleFileName, filepath.Ext(exampleFileName)),
		Name:            strings.TrimSuffix(exampleFileName, filepath.Ext(exampleFileName)),
		Rows:            rows,
		FullCode:        template.JS(fullCode),
		PreviousExample: nil,
		NextExample:     nil,
	}
}
