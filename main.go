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
	"strconv"
	"strings"
	"time"
)

/*
Defines a .survive file datatype.
*/
type SurviveFile struct {
	Week         int
	Player       string
	Availability string
	CreatedAt    time.Time
}

// ========= ENVIRONMENT VARIABLES ==========
const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"

func (ss *SurviveFile) save() error {
	// Define the filename with our filesystem naming convention based on struct fields.
	filename := fmt.Sprintf("%s/%d-%s.survive", ENV_ROOT, ss.Week, ss.Player)

	// Open the file it is exists, or make a new one.
	// Either way, mark file as APPENDABLE
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)

	if err != nil {
		log.Fatal("Error opening file.", err)
		return err
	}

	// Format data.
	stringToWrite := fmt.Sprintf("%s: %s%s", ss.CreatedAt.Format(time.RFC3339), ss.Availability, "\n")

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

func initFile(category int, week int, player string) (*SurviveFile, error) {

	// Init player with zero'd out (free) availability
	tmp := &SurviveFile{Week: week, Player: player, Availability: "0000:0000:0000:0000:0000:0000:0000", CreatedAt: time.Now()}
	err := tmp.save()
	if err != nil {
		log.Fatal("Error initializing file", err)
		return nil, err
	}

	return tmp, nil

}

func loadFile(week int, player string) (*SurviveFile, error) {

	// Compute file name based off convention
	filename := fmt.Sprintf("%s/%d-%s.survive", ENV_ROOT, week, player)

	// Read file from file system.
	body, err := ioutil.ReadFile(filename)

	// Catch error reading file.
	if err != nil {
		log.Fatal("Error loading page!")
		return nil, err
	}

	// TODO: only return last line of file?
	return &SurviveFile{Week: week, Player: player, Availability: string(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):]
	weekAndName := strings.Split(path, "/")
	if len(weekAndName) != 2 {
		log.Fatal("Could not parse view input correctly.")
		// TODO: redirect back home or to 404.
	}
	num, _ := strconv.ParseInt(weekAndName[0], 10, 32)
	p, err := loadFile(int(num), weekAndName[1])

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

	//initFile(4, 1, "Joe")
	//initFile(4, 2, "Mike")
	//initFile(4, 1, "Tim")

	http.HandleFunc("/view/", viewHandler) //  .../view/week/player_name
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
