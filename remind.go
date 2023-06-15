package main

import (
	"fmt"
	"bytes"
	"strings"
	"os"
	"os/exec"
	"strconv"
	"time"
	"encoding/json"
	"path/filepath"
)

// Gets reminders for today and puts the output into array of lines ([]string) 
// If lines are longer than maxLineWidth they are cut off and the remainder(s) added as new line(s)
func getToday(filename string, year int, month int, day int, maxLineWidth int) []string {
	var outb, errb bytes.Buffer
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	cmd := exec.Command("remind", filename, dateStr)	
    cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil { 
		if errb.Len() > 0 { err = fmt.Errorf(errb.String()) }
		panic(err) 
	}
	if outb.Len() <= 0 {
		panic(fmt.Errorf("remind did not return any output"))
	}

	// entire outb is transformed into lines wrapped if too long
	lines := []string{}
	for i, line := range strings.Split(outb.String(), "\n") {
		if i == 0 {
			// patching first line because it is so long
			lines = append(lines, "Todays Reminders:")
			continue
		}
		for {
			if len(line) <= maxLineWidth {
				lines = append(lines, line)
				break
			} else {
				lines = append(lines, line[:maxLineWidth])
				line = line[maxLineWidth:]
			}
		}
	}

	return lines
}

// Calls remind -pppn -g filename date and parses returned reminders into []Event
// All returned dates are valid
func getEvents(filename string, year int, month int, nrOfMonth int) (eventsArr []Event) {
	var outb, errb bytes.Buffer
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, 1)

	cmd := exec.Command("remind", "-ppp" + strconv.Itoa(nrOfMonth), "-g", filename, dateStr)	
    cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil { 
		if errb.Len() > 0 { err = fmt.Errorf(errb.String()) }
		panic(err) 
	}
	if outb.Len() <= 0 {
		panic(fmt.Errorf("remind did not return any output"))
	}

	eventsArr, err = parseRemindEventsJSON(outb.String())
	if err != nil { panic(err) }
	return eventsArr
}

// Parses pure JSON output ( remind -ppp )
// This is more useful than simple format since it also includes
// Event information like Filename and Lineno
func parseRemindEventsJSON(str string) ([]Event, error) {
	eventsArr := []Event{}

	type Entry struct {
		Date string
		Filename string
		Lineno int
		Body string
	}
	type MonthDescriptor struct { Entries []Entry }
	
	mds := []MonthDescriptor{}
	err := json.Unmarshal([]byte(str), &mds)
	if err != nil { return eventsArr, err }
	
	for _, md := range mds {
	for _, entry := range md.Entries {
		t, err := time.Parse("2006-01-02", entry.Date)
		if err != nil { 
			return eventsArr, fmt.Errorf("Could not parse REM Event date %s: %w", entry.Date, err) 
		}
		event, err := NewEvent(t.Year(), int(t.Month()), t.Day(), entry.Body)
		if err != nil { return eventsArr, err }

		event.Filename = entry.Filename
		event.Lineno = entry.Lineno

		eventsArr = append(eventsArr, event)
	}
	}

	return eventsArr, nil
}

// Parses line of simple format specified according to remind source code ( remind -s )
// General structure: date passhtru tags duration time body 
// all fields are seperated by 1 space char ' '
// date: YEAR/MONTH/DAY of format %04d/%02d/%02d 
// passthru: Out of band reminders when SPECIAL keyword used 
//           possible values are MOON, SHADE, COLOR, ... 
//           it does not seem possible to use multiple keywords at once
// tags: A tag can consist of any char except ',' and ' '. Tags are seperated by ','
// duration: single number, in minutes
// time: single number, in minutes e.g. 8:30am = 8*60+30
// body: event message
//
// Currently just using date and body
func parseRemindEventSimpleFormat(str string) (Event, error) {
	dateStr, rest, found := strings.Cut(str, " ") 
	if !found { return Event{}, fmt.Errorf("Invalid REM Event: %s", str) }

	t, err := time.Parse("2006/01/02", dateStr)
	if err != nil { 
		return Event{}, fmt.Errorf("Could not parse REM Event date %s: %w", str[:10], err) 
	}

	// passthru
	_, rest, found = strings.Cut(rest, " ")
	if !found { return Event{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// tags
	_, rest, found = strings.Cut(rest, " ")
	if !found { return Event{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// duration
	_, rest, found = strings.Cut(rest, " ")
	if !found { return Event{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// time
	_, rest, found = strings.Cut(rest, " ")
	if !found { return Event{}, fmt.Errorf("Invalid REM Event: %s", str) }

	return NewEvent(t.Year(), int(t.Month()), t.Day(), rest)
}

///////////////// OTHER ////////////////////////////////
func openEditor(filename string, lineno int) {
	editor := os.Getenv("EDITOR")
	if editor == "" { editor = "vim" }

	// for certain editors jump to lineno as well
	var cmd *exec.Cmd
	switch filepath.Base(editor) {
	case "vi", "vim", "nano", "emacs":
		cmd = exec.Command(editor, "+" + strconv.Itoa(lineno), filename)
	default:
		cmd = exec.Command(editor, filename)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil { panic(err) }
}
