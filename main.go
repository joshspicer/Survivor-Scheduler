/*
	SURVIVOR SCHEDULER
	Josh Spicer <https://joshspicer.com/>
	2019 March 23
*/

package main

import (
	"bufio"
	"encoding/json"
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
	Player       string
	Availability Availability
	CreatedAt    time.Time
}

type JSONResponse struct {
	Player string
	I1     int
	I2     int
}

// ========= ENVIRONMENT VARIABLES ==========
//const ENV_ROOT = "/Users/joshspicer/go/src/github.com/joshspicer/survivor-scheduler"
const ENV_ROOT = "/Users/jspicer/go/src/survivor-scheduler-golang"

// ==== STATE ====
var PLAYERS []string

func (ss *SurviveFile) save() error {
	// Define the filename with our filesystem naming convention based on struct fields.
	filename := fmt.Sprintf("%s/%s.survive", ENV_ROOT, ss.Player)

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

/**
HELPER:
Given an Availability and a "target" hour on a "target" day, flips the bit.
Mutates the given Availability as it exists in memory
*/
func (aa *Availability) flipAvailabilityBit(dayIdx int, hourIdx int) {
	currValue := aa.Days[dayIdx].Halfhours[hourIdx]
	day := &aa.Days[dayIdx]
	day.Halfhours[hourIdx] = !currValue

}

/**
Appends to (or creates new) .survive file based on given parameters.
Used to update data file on disk with availability changes.
*/
func updateActor(player string, dayIdx int, hourIdx int) (*SurviveFile, error) {

	// Load this player's file based on the information given.
	sFile, err := loadFile(player)

	if err != nil {
		log.Print("Error loading file")
		return nil, err
	}

	// Edit player's availability with the given instructions
	sFile.Availability.flipAvailabilityBit(dayIdx, hourIdx)

	// Save the edits
	err = sFile.save()
	if err != nil {
		log.Print("Error saving file", err)
		return nil, err
	}

	return sFile, nil

}

// Outputs a given Availability as a string.
// Of form:
// 			0000:0000:<...7 hex groups...>:0000
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

func initFile(player string) (*SurviveFile, error) {

	// Init player with zero'd out (free) availability
	tmp := &SurviveFile{Player: player, Availability: Availability{}, CreatedAt: time.Now()}
	err := tmp.save()
	if err != nil {
		log.Fatal("Error initializing file", err)
		return nil, err
	}

	return tmp, nil

}

func loadFile(player string) (*SurviveFile, error) {

	// Compute file name based off convention
	filename := fmt.Sprintf("%s/%s.survive", ENV_ROOT, player)

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

	return &SurviveFile{Player: player, Availability: *availability}, nil
}

func playerEditHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/edit/"):]
	player := strings.Split(path, "/")
	if len(player) != 1 {
		log.Print("Could not parse view input correctly.")
		http.NotFound(w, r)
		return
	}

	p, err := loadFile(player[0])

	if err != nil {
		log.Print("Could not load page player edit handler")
		http.NotFound(w, r)
		return
	}

	// Map functions from golang to be reflected in HTML
	funcMap := template.FuncMap{
		"tableflip":   func() string { return "(╯°□°）╯︵ ┻━┻" },
		"updateActor": updateActor,
	}

	tpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("templates/*.html"))

	err = tpl.ExecuteTemplate(w, "playerEdit.html", p)

	if err != nil {
		log.Fatal(err)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t JSONResponse
	err := decoder.Decode(&t)
	if err != nil {
		log.Print("Error parsing POST: ", err)
		return
	}

	_, err = updateActor(t.Player, t.I1, t.I2)
	if err != nil {
		log.Print("Error in the updateHandler, from updateActor: ", err)
	}
}

func manageHandler(w http.ResponseWriter, r *http.Request) {

	aggregatedWeeklyAvails, err := aggregatedWeeklyAvails(); if err != nil {
		log.Print(err)
	}

	// Map functions from golang to be reflected in HTML
	funcMap := template.FuncMap{
		"tableflip": func() string { return "(╯°□°）╯︵ ┻━┻" },
	}

	tpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("templates/*.html"))

	err = tpl.ExecuteTemplate(w, "manage.html", aggregatedWeeklyAvails)

	if err != nil {
		log.Fatal(err)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	// Map functions from golang to be reflected in HTML
	funcMap := template.FuncMap{
		"tableflip": func() string { return "(╯°□°）╯︵ ┻━┻" },
	}

	tpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob("templates/*.html"))

	err := tpl.ExecuteTemplate(w, "index.html", PLAYERS)

	if err != nil {
		log.Fatal(err)
	}

}

func errorHandler(w http.ResponseWriter, r *http.Request) {

	tpl := template.Must(template.New("main").ParseGlob("templates/*.html"))

	err := tpl.ExecuteTemplate(w, "error.html", nil)

	if err != nil {
		log.Fatal(err)
	}

}



// Utilizes conf.survive to flatten all player's availabilities
// into one "master availablity".
// If any one person is unavailable, time slot is marked as "busy", else kept "free"

func aggregatedWeeklyAvails() (Availability, error) {

	// "Master" availability
	var master Availability

	// For each Player, get their availability. Flip master if conflict
	for _, player := range PLAYERS {
		file, err := loadFile(player); if err != nil {
			return Availability{}, err
		}

		for dayIdx, day := range file.Availability.Days {
			for hrIdx := 0; hrIdx < 16; hrIdx++ {
				if day.Halfhours[hrIdx] {
					master.Days[dayIdx].Halfhours[hrIdx] = true
				}
			}
		}
	}

	return master, nil
}

/**
*/
func bigBang() error {

	newGame, err := os.OpenFile("conf", os.O_RDONLY, 0600)

	// If there is no new init file, lets see if we have an existing game to restore from!!
	if err != nil {
		existingGame, err := os.OpenFile("conf.processed", os.O_RDONLY, 0600); if err != nil {
			log.Print("Lack of either new OR existing conf file...")
			return err
		}
		scanner := bufio.NewScanner(existingGame)
		for scanner.Scan() {
			PLAYERS = append(PLAYERS, scanner.Text())
		}

		// Existing game restored. Now Return from function with no error.
		return nil
	}


	// === If game has NEVER been initialized we continue down here. ====

	scanner := bufio.NewScanner(newGame)
	for scanner.Scan() {

		PLAYERS = append(PLAYERS, scanner.Text())
		_, err := initFile(scanner.Text())

		if err != nil {
			log.Print("Error init of player. Check config format.")
			return err
		}
	}

	err = os.Rename("conf", "conf.processed")

	return err

}

/*
Main function. Entry point of program.
*/
func main() {

	// If conf.survive exists, initialize game.
	err := bigBang()
	if err != nil {
		log.Print("Error in bigbang", err)
		http.HandleFunc("/", errorHandler)

	} else {
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/edit/", playerEditHandler) //  .../view/{week}/{player_name}
		http.HandleFunc("/manage/", manageHandler)   //  .../manage/{week}/
		http.HandleFunc("/update", updateHandler)    //  POST to /update with {week, player,i1,i2}
	}

	fmt.Print(PLAYERS)


	//http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
