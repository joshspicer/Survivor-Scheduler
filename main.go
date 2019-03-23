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

// ENVIRONMENT VARIABLES
const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"

func (ss *SurviveFile) save() error {
	filename := fmt.Sprintf("%s/c%d-w%d-%s.survive", ENV_ROOT, ss.Category, ss.week, ss.Title)

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)

	cleanedBody := strings.Replace(ss.Body, ":", " ", -1)

	if err != nil {
		log.Fatal("Error opening file.", err)
		return err
	}

	stringToWrite := fmt.Sprintf("%s: %s%s", ss.CreatedAt.Format(time.RFC3339), cleanedBody, "\n")
	_, err = f.WriteString(stringToWrite)
	if err != nil {
		log.Fatal("Error writing to file.", err)
		return err
	}

	return nil
}

func loadPage(category int, week int, title string) SurviveFile {
	filename := fmt.Sprintf("%s/c%d-w%d-%s.survive", ENV_ROOT, category, week, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error loading page!")
		return SurviveFile{}
	}
	return SurviveFile{Category: category, week: week, Title: title, Body: string(body)}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p := loadPage(4, 55, title)
	fmt.Print(p)

	//if err != nil {
	//	log.Fatal("Could not load page view handler")
	//	//TODO: redirect back home?
	//}
	t, _ := template.ParseFiles("templates/home.html")
	err := t.Execute(w, p)
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
