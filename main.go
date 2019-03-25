/*
	SURVIVOR SCHEDULER
	Josh Spicer <https://joshspicer.com/>
	2019 March 23
*/

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type DailyAvailability struct {
	Halfhours [16]bool // From 4:30pm-12:00pm in half hour increments
}

type Availability struct {
	Days [7]DailyAvailability // 0 == Sunday, 6 == Saturday
}

/*
Defines a .survive file datatype.
*/
type SurviveFile struct {
	Week         int
	Player       string
	Availability Availability
	CreatedAt    time.Time
}

// ========= ENVIRONMENT VARIABLES ==========
const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"

//func echoTest() {
//	fmt.Print("TEST")
//}

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
	avail, err := ss.Availability.availabilityToString()
	if err != nil {
		return err
	}

	stringToWrite := fmt.Sprintf("%s%s", avail, "\n")

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

func (aa Availability) availabilityToString() (string, error) {
	var sb strings.Builder

	for idx, day := range aa.Days {
		// For each day
		str, err := day.dailyAvailabilityToString()
		if err != nil {
			log.Print("Error converting day availability to string in availabilityToString.")
			return "", err
		}
		// No error, write to master string.
		sb.WriteString(str)
		// Add a colon if not last day of week.
		if idx != 6 {
			sb.WriteString(":")
		}
	}

	return sb.String(), nil
}

func (dd DailyAvailability) dailyAvailabilityToString() (string, error) {
	// Input: a [16]bool
	// Output: 4-character hex string. e.g: 4fcb

	const BASE = 16
	const DIFF = BASE - 1
	var count int64 = 0
	for idx, halfhour := range dd.Halfhours {
		//0000 0000 0000 0000

		if halfhour {
			count += int64(math.Pow(float64(2), float64(DIFF-idx)))
		}
	}

	// Convert the count (a base-10 number) into base-16
	base16 := strconv.FormatInt(count, 16)
	return base16, nil
}

// Convert the hex-encoded availability string into an Availability
func stringToAvailability(availStr string) (*Availability, error) {

	// Input eg: 0:c000:0:13:0888:4560:15a0
	// where a segment: 0000 <-> FFFF (hex-encoded 16-bit number)

	// [1] Split into array with all 7 pieces
	arr := strings.Split(availStr, ":")
	if len(arr) != 7 {
		log.Print("Expect 7 parts of an availability string, got ", len(arr))
		return &Availability{}, errors.New("expected 7 parts")
	}

	AA := Availability{}

	// [2] For each chunk (Sunday => Saturday)
	for idx, hexNumStr := range arr {

		day := &AA.Days[idx]

		// [3] Parse hex number into integer
		num, err := strconv.ParseInt(hexNumStr, 16, 32)
		if err != nil {
			log.Print("Unable to parse out the hexNumStr")
			return &Availability{}, err
		}

		// [5] Iterate through string and flip bool of DailyAvail.
		for i := 0; i < 16; i++ {
			// ANDs the hex value and the bit in question.
			if num&(1<<uint(16-i-1)) > 0 {
				day.Halfhours[i] = true
			}
		}
	}

	return &AA, nil
}

func initFile(week int, player string) (*SurviveFile, error) {

	// Init player with zero'd out (free) availability
	tmp := &SurviveFile{Week: week, Player: player, Availability: Availability{}, CreatedAt: time.Now()}
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
		//log.Fatal("Error loading page!!")
		return nil, err
	}

	// Grab the latest file update
	split := strings.Split(string(body), "\n")
	// Per convention: every entry end in a newline, so go two lines up to get last entry.
	avail := split[len(split)-2]

	availability, err := stringToAvailability(avail)
	if err != nil {
		return nil, err
	}

	return &SurviveFile{Week: week, Player: player, Availability: *availability}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/view/"):]
	weekAndName := strings.Split(path, "/")
	if len(weekAndName) != 2 {
		log.Print("Could not parse view input correctly.")
		http.NotFound(w, r)
		return
	}

	num, _ := strconv.ParseInt(weekAndName[0], 10, 32)
	p, err := loadFile(int(num), weekAndName[1])

	if err != nil {
		log.Print("Could not load page view handler")
		http.NotFound(w, r)
		return
	}

	t, _ := template.ParseFiles("templates/view.html")

	//funcMap := template.FuncMap {
	//	"echoTest": echoTest,
	//}
	//
	//t.Funcs(funcMap)

	err = t.Execute(w, p)

	if err != nil {
		log.Fatal(err)
	}
}

/*
Main function. Entry point of program.
*/
func main() {

	//initFile(1, "Joe")
	//initFile(2, "Mike")
	//initFile(1, "Tim")

	s, _ := loadFile(1, "Joe")

	fmt.Print(s)

	http.HandleFunc("/view/", viewHandler) //  .../view/{week}/{player_name} || .../view/{week}
	//http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
