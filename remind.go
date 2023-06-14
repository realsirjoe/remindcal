package main

import (
	"fmt"
	"bytes"
	"strings"
	"os"
	"os/exec"
	"strconv"
	"time"
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

// Calls remind -sn -g filename date and parses returned reminders into []RawEvent
// All returned dates are valid
func getEvents(filename string, year int, month int, nrOfMonth int) (rawEvents []RawEvent) {
	var outb, errb bytes.Buffer
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, 1)

	cmd := exec.Command("remind", "-s" + strconv.Itoa(nrOfMonth), "-g", filename, dateStr)	
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

	// lines => rawEvents
	lines := strings.Split(outb.String(), "\n")
	for _, line := range lines {
		if line != "" {
			rawEvent, err := parseRemindEvent(line)
			if err != nil { panic(err) }
			rawEvents = append(rawEvents, rawEvent)
		}
	}
	return rawEvents
}

// Parses simple format specified according to remind source code
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
func parseRemindEvent(str string) (RawEvent, error) {
	dateStr, rest, found := strings.Cut(str, " ") 
	if !found { return RawEvent{}, fmt.Errorf("Invalid REM Event: %s", str) }

	t, err := time.Parse("2006/01/02", dateStr)
	if err != nil { 
		return RawEvent{}, fmt.Errorf("Could not parse REM Event date %s: %w", str[:10], err) 
	}

	// passthru
	_, rest, found = strings.Cut(rest, " ")
	if !found { return RawEvent{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// tags
	_, rest, found = strings.Cut(rest, " ")
	if !found { return RawEvent{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// duration
	_, rest, found = strings.Cut(rest, " ")
	if !found { return RawEvent{}, fmt.Errorf("Invalid REM Event: %s", str) }

	// time
	_, rest, found = strings.Cut(rest, " ")
	if !found { return RawEvent{}, fmt.Errorf("Invalid REM Event: %s", str) }

	return RawEvent{t.Year(), int(t.Month()), t.Day(), rest}, nil
}

///////////////// OTHER ////////////////////////////////
func openEditor(editor string, filename string) {
	cmd := exec.Command(editor, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil { panic(err) }
}
