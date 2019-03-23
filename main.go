/*
	SURVIVOR SCHEDULER
	Josh Spicer <https://joshspicer.com/>
	2019 March 23
*/

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
Defines a .survive file datatype.
*/
type SurviveFile struct {
	Category  int // Will divide between player data, global data, etc..
	week      int
	Title     string
	Body      string /////TODO: use the "bits per half hour" idea sketched out. convert to hex for coolness?
	CreatedAt time.Time
}

// ========= ENVIRONMENT VARIABLES ==========
const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"

func (ss *SurviveFile) save() error {
	// Define the filename with our filesystem naming convention based on struct fields.
	filename := fmt.Sprintf("%s/c%d-w%d-%s.survive", ENV_ROOT, ss.Category, ss.week, ss.Title)

	// Open the file it is exists, or make a new one.
	// Either way, mark file as APPENDABLE
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)

	// We will use colons as internal delimiters. Scrub from input.
	cleanedBody := strings.Replace(ss.Body, ":", " ", -1)

	if err != nil {
		log.Fatal("Error opening file.", err)
		return err
	}

	// Format data.
	stringToWrite := fmt.Sprintf("%s: %s%s", ss.CreatedAt.Format(time.RFC3339), cleanedBody, "\n")

	// Write string to open file.
	_, err = f.WriteString(stringToWrite)
	if err != nil {
		log.Fatal("Error writing to file.", err)
		return err
	}

	// Close the file
	err = f.Close()
	if err != nil {
		log.Fatal("Error closing open file in save()")
		return err
	}

	return nil
}

func loadPage(category int, week int, title string) (*SurviveFile, error) {

	// Compute file name based off convention
	filename := fmt.Sprintf("%s/c%d-w%d-%s.survive", ENV_ROOT, category, week, title)

	// Read file from file system.
	body, err := ioutil.ReadFile(filename)

	// Catch error reading file.
	if err != nil {
		log.Fatal("Error loading page!")
		return nil, err
	}

	// TODO: only return last line of file?
	return &SurviveFile{Category: category, week: week, Title: title, Body: string(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(4, 55, title)

	if err != nil {
		log.Fatal("Could not load page view handler")
		//TODO: redirect back home?
	}
	t, _ := template.ParseFiles("templates/home.html")
	err = t.Execute(w, p)
	if err != nil {
		log.Fatal(err)
	}
}

/*
Main function. Entry point of program.
*/
func main() {

	t1 := SurviveFile{Category: 4, week: 55, Title: "hello", Body: "This is a sample page.", CreatedAt: time.Now()}
	err := t1.save()
	fmt.Print(err)

	http.HandleFunc("/view/", viewHandler)
	//http.HandleFunc("/edit/", editHandler)
	//http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/**

0000000000000000  Sunday 4:30pm-12pm Half hour increments (16 bits) === FFFF in hex for biggest

ffff:ffff:ffff:ffff:ffff:ffff:ffff  (busy every time)
0000:0000:0000:0000:0000:0000:0000  (free all the time)

hex -> binary to determine availability every half hour

**/
