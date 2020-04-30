package main

import (
	"bufio"
	"fmt"
	"github.com/docopt/docopt-go"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

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

func displayCollector(w http.ResponseWriter, r *http.Request) {
	indexHTML := `<!DOCTYPE html>
<html>
<head>
  <link rel="stylesheet" href="https://cdn.rawgit.com/Chalarangelo/mini.css/v3.0.1/dist/mini-default.min.css">
</head>
<body>

<h2>Feedback Form</h2>

<p>Please provide up to three negative and three positive comments about the course. <br>
For example, you can mention only three negative but no positive things, two positive and one negative things, nothing at all, etc. <br>
</p>

<p>
In the beginning of tomorrow's session you will see all the comments from everybody else and you will be able to indicate to which of these you agree. <br>
That is, this evaluation form is not meant for sending private/personal feedback about the course.
</p>
<form action="/reply">
  <h3>Negative Comments</h3>
  <label for="fname">First negative:</label><br>
  <input type="text" id="neg_a" name="neg_a" value="" size="100"><br>
  <label for="fname">Second negative:</label><br>
  <input type="text" id="neg_b" name="neg_b" value="" size="100"><br>
  <label for="fname">Third negative:</label><br>
  <input type="text" id="neg_c" name="neg_c" value="" size="100"><br>

  <hr>

  <h3>Positive Comments</h3>
  <label for="fname">First positive:</label><br>
  <input type="text" id="pos_a" name="pos_a" value="" size="100"><br>
  <label for="fname">Second positive:</label><br>
  <input type="text" id="pos_b" name="pos_b" value="" size="100"><br>
  <label for="fname">Third positive:</label><br>
  <input type="text" id="pos_c" name="pos_c" value="" size="100"><br>

  <br>
  <input type="submit" value="Submit">
</form>
</body>
</html>
`

	fmt.Fprintf(w, indexHTML)
}

// Inspired by https://www.veracode.com/blog/2013/12/golangs-context-aware-html-templates
func collectResponses(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form values!", http.StatusInternalServerError)
		return
	}
	// // get the form values
	for key, value := range r.Form {
		// TODO: Think about inserting an IP address based filter for already
		// collected responses
		commentStr := value[0]
		commentStr = strings.Replace(commentStr, "\"", "\"\"", -1)
		commentStr = fmt.Sprintf("\"%s\"", commentStr)
		data := []string{r.RemoteAddr, key, commentStr}
		appendTo("responses.txt", data)
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

func appendTo(file string, data []string) {

	fileHandle, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	timeStamp := time.Now().Format(time.RFC3339)
	dataStr := strings.Join(data, ",")
	csvLine := fmt.Sprintf("%s,%s", timeStamp, dataStr)
	fmt.Fprintln(writer, csvLine)
	writer.Flush()
}

func main() {
	usage := `Delphi Evaluation Phase One Server.

Start me for example like:
$ ./phase_one_server --addr=0.0.0.0 --port=8888 >> log/phase_one.log 2>&1 &

Usage:
  phase_one_server [--addr=<addr>] [--port=<port>]
  phase_one_server -h | --help
  phase_one_server --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --addr=<addr>  Address for serving [default: 127.0.0.1].
  --port=<port>  Port for serving [default: 8080].
`
	arguments, _ := docopt.Parse(usage, nil, true, "Delphi Evaluation Phase One Server 1.0", false)
	port := arguments["--port"].(string)
	addr := arguments["--addr"].(string)

	log.Print(fmt.Sprintf("Serving on http://%s:%s", addr, port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", addr, port), startServing()))
}
