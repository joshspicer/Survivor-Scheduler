/*
	SURVIVOR SCHEDULER
	Josh Spicer <https://joshspicer.com/>
	2019 March 23
*/

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

/*
Defines a .survive file datatype.
*/
type SurviveFile struct {
	category  int // Will divide between player data, global data, etc..
	week      int
	title     string
	Body      string
	createdAt time.Time
}

// ENVIRONMENT VARIABLES
const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"

func (ss *SurviveFile) save() error {
	filename := fmt.Sprintf("%s/c%d-w%d-%s.survive", ENV_ROOT, ss.category, ss.week, ss.title)

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)

	cleanedBody := strings.Replace(ss.Body, ":", " ", -1)

	if err != nil {
		log.Fatal("Error opening file.", err)
		return err
	}

	stringToWrite := fmt.Sprintf("%s: %s%s", ss.createdAt.Format(time.RFC3339), cleanedBody, "\n")
	_, err = f.WriteString(stringToWrite)
	if err != nil {
		log.Fatal("Error writing to file.", err)
		return err
	}

	return nil
}

/*
Main function. Entry point of program.
*/
func main() {

	//templates := template.Must(template.ParseFiles("templates/welcome-template.html"))
	//http.Handle("/static/", //final url can be anything
	//	http.StripPrefix("/static/",
	//		http.FileServer(http.Dir("static"))))

	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//
	//	//Takes the name from the URL query e.g ?name=Martin, will set welcome.Name = Martin.
	//	if name := r.FormValue("name"); name != "" {
	//		welcome.Name = name
	//	}
	//	//If errors show an internal server error message
	//	//I also pass the welcome struct to the welcome-template.html file.
	//	if err := templates.ExecuteTemplate(w, "welcome-template.html", welcome); err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//	}
	//})

	t1 := SurviveFile{category: 4, week: 55, title: "hello", Body: "This is :a sampl: :::e Page.::", createdAt: time.Now()}
	err := t1.save()
	fmt.Print(err)

	//fmt.Println("Listening")
	//fmt.Println(http.ListenAndServe(":25500", nil))

}
