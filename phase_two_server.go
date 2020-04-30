package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/docopt/docopt-go"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type Comments struct {
	Negative []string
	Positive []string
}

func readCSVFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func filterResponses(records [][]string) Comments {
	var negComments []string
	var posComments []string

	for _, row := range records {
		if strings.HasPrefix(row[2], "neg_") {
			if row[3] != "" {
				negComments = append(negComments, row[3])
			}
		} else {
			if row[3] != "" {
				posComments = append(posComments, row[3])
			}
		}
	}

	return Comments{
		Negative: negComments,
		Positive: posComments,
	}
}

// App creates a mux and binds the root route for processing
// static files.
func startServing() http.Handler {

	// Create a new mux for this service.
	m := http.NewServeMux()
	// m.Handle("/", http.FileServer(http.Dir("./static")))
	m.HandleFunc("/", displayCollector)
	m.HandleFunc("/reply", collectResponses)

	return m
}

func appendTo(file string, data []string) {

	fileHandle, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	dataStr := strings.Join(data, ",")
	csvLine := fmt.Sprintf("%s", dataStr)
	fmt.Fprintln(writer, csvLine)
	writer.Flush()
}

func displayCollector(w http.ResponseWriter, r *http.Request) {
	indexHTMLTemplate := `<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="https://cdn.rawgit.com/Chalarangelo/mini.css/v3.0.1/dist/mini-default.min.css">
</head>
<body>

<h2>Do you agree on your class-mates' statements?</h2>

<p>Please select (set a checkmark) on each comment that you agree on.</p>

<form action="/reply">
  <h3>Negative Comments</h3>
  <ul>
    {{range $index, $element := .Negative}}
   
	<li><input type="checkbox" name="neg_{{$index}}" value="{{$element}}"> {{$element}}</li>
    {{end}}
  </ul>
  <hr>

  <h3>Positive Comments</h3>
  <ul>
  {{range $index, $element := .Positive}}
    <li><input type="checkbox" name="pos_{{$index}}" value="{{$element}}"> {{$element}}</li>
  {{end}}
  </ul>
  <br>
  <input type="submit" value="Submit">
</form>
</body>
</html>
`

	t, err := template.New("comments").Parse(indexHTMLTemplate)
	if err != nil {
		// TODO: check how to handle this precisely
		panic(err)
	}

	// the global variable comments of type Comments
	t.Execute(w, comments)
}

// Inspired by https://www.veracode.com/blog/2013/12/golangs-context-aware-html-templates
func collectResponses(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form values!", http.StatusInternalServerError)
		return
	}
	// get the form values
	for key, value := range r.Form {
		commentStr := value[0]
		commentStr = strings.Replace(commentStr, "\"", "\"\"", -1)
		commentStr = fmt.Sprintf("\"%s\"", commentStr)
		data := []string{key, commentStr}
		appendTo("rated_responses.txt", data)
	}

	thankYouHTML := `<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="https://cdn.rawgit.com/Chalarangelo/mini.css/v3.0.1/dist/mini-default.min.css">
</head>
<body>

<h1>Thank you for your feedback!</h1>
</body>
</html>
`
	fmt.Fprintf(w, thankYouHTML)
}

var comments Comments

func main() {
	usage := `Delphi Evaluation Phase Two Server.

Start me for example like:
$ ./phase_two_server --addr=0.0.0.0 --port=9999 >> log/phase_two.log 2>&1 &


Usage:
  phase_two_server [--addr=<addr>] [--port=<port>]
  phase_two_server -h | --help
  phase_two_server --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --addr=<addr>  Address for serving [default: 127.0.0.1].
  --port=<port>  Port for serving [default: 8080].

Hint:
  To quickly count the for the most rated comments run:
  $ sort rated_responses.txt | uniq -c | sort -r
`
	arguments, _ := docopt.Parse(usage, nil, true, "Delphi Evaluation Phase Two Server 1.0", false)
	port := arguments["--port"].(string)
	addr := arguments["--addr"].(string)

	// Be aware of that the responses file is read on server startup! That is
	// do not start the server before that file is in a state that you like
	responses := readCSVFile("./responses.txt")
	comments = filterResponses(responses)

	log.Print(fmt.Sprintf("Serving on http://%s:%s", addr, port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), startServing()))
}
