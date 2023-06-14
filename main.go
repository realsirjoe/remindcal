package main

import (
	"fmt"
	"time"
	"strconv"
	"math"
	"os"
	"os/signal"
	"syscall"
)

///////// YEAR Structure ///////////
type YearStructure struct {
	// Dec Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec Jan
	daysInMonths [14]int
}

func GenerateYearStructure(year int) (ys YearStructure) {
	ys.daysInMonths[0] = DaysInMonth(year-1, 12)
	for month := 1; month<=12; month++ {
		ys.daysInMonths[month] =  DaysInMonth(year, time.Month(month))
	}
	ys.daysInMonths[13] = DaysInMonth(year+1, 1)
	return ys
}

///////// Date Structure ///////////
type Date struct {
	Year int
	Month int
	Day int
	daysInMonth int
}
func NewDate(year int, month int, day int) (Date, error) {
	var d Date
	if year < 1 {
		return d, fmt.Errorf("Year cannot be smaller than 1")	
	}	
	if month < 1 || month > 12 {
		return d, fmt.Errorf("Month must be between 1 and 12")
	}
	d.daysInMonth = DaysInMonth(year, time.Month(month))
	if day > d.daysInMonth {
		return d, fmt.Errorf("Day for %d %d cannot be %d", year, month, day)
	}

	d.Year = year; d.Month = month; d.Day = day
	return d, nil
}
func (d *Date) AddDay() {
	d.Day++
	if d.Day > d.daysInMonth { 
		d.Day = 1

		d.Month++
		if d.Month > 12 {
			d.Month = 1
			d.Year++
		} 
		d.daysInMonth = DaysInMonth(d.Year, time.Month(d.Month))
	} 
}
func (d *Date) SubtractDay() {
	d.Day--
	if d.Day < 1 {
		d.Month--
		if d.Month < 1 {
			// date cannot be smaller than 1-1-1
			if d.Year <= 1 { d.Day = 1; d.Month = 1; d.Year = 1; return } 
			d.Month = 12
			d.Year--
		} 
		d.daysInMonth = DaysInMonth(d.Year, time.Month(d.Month))
		d.Day = d.daysInMonth
	} 
}
// These are blazing fast anyway so they dont need optimization
func (d *Date) AddWeek() {
	for i:=0; i<7; i++ {
		d.AddDay()	
	}
}
func (d *Date) SubtractWeek() {
	for i:=0; i<7; i++ {
		d.SubtractDay()	
	}
}
// if day > days in next month go to last day of next month
func (d *Date) AddMonth() {
	d.Month++
	if d.Month > 12 {
		d.Month = 1
		d.Year++
	}
	d.daysInMonth = DaysInMonth(d.Year, time.Month(d.Month))
	if d.Day > d.daysInMonth {
		d.Day = d.daysInMonth
	}
}
func (d *Date) SubtractMonth() {
	d.Month--
	if d.Month < 1 {
		d.Month = 12
		d.Year--
		// date cannot be smaller than 1-1-1
		if d.Year < 1 { d.Day = 1; d.Month = 1; d.Year = 1; return } 
	}
	d.daysInMonth = DaysInMonth(d.Year, time.Month(d.Month))
	if d.Day > d.daysInMonth {
		d.Day = d.daysInMonth
	}
}

func (d *Date) NumericString() string {
	return NumericString(d.Year, d.Month, d.Day)
}
func NumericString(year, month, day int) string {
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

// adds one month and increases year if necessary
func AddMonth(year int, month int) (int, int) {
	month++
	if month > 12 { year++; month = 1 }
	return year, month
}
func SubtractMonth(year int, month int) (int, int) {
	month--
	if month < 1 {
		year--; month = 12
	}
	return year, month
}
///////// Event Structure ////////////

type Event struct {
	Date Date
	Message string
}
func NewEvent(year int, month int, day int, message string) (e Event, err error) {
	e.Date, err = NewDate(year, month, day)
	e.Message = message
	return
}
type RawEvent struct {
	Year int
	Month int
	Day int
	Message string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: remindcal filename\n")
		os.Exit(1)
	}
	filename := os.Args[1]
	todayWinEnabled := false
	debug := false
	DrawingLoop(filename, todayWinEnabled, debug)
}
// changes events in place
func evaluateRawEvents(rawEvents []RawEvent, events map[string][]Event) {
	for _, re := range rawEvents {
		e, err := NewEvent(re.Year, re.Month, re.Day, re.Message)
		if err != nil { panic(err) }
		key := e.Date.NumericString()
		if _, ok := events[key]; !ok {
			events[key] = []Event{e}
		} else {
			events[key] = append(events[key], e)
		}
	}
}

const EVENTS_WIN = 0
const CALENDAR_WIN = 1
const TODAY_WIN = 2

func DrawingLoop(filename string, todayWinEnabled bool, debug bool) {
	var rawEvents = []RawEvent{}
	var events = map[string][]Event{}
	var todayMessageLines = []string{}
	var statusMessage = ""

	t := time.Now()
	var d Date
	d, err := NewDate(t.Year(), int(t.Month()), t.Day())
	if err != nil { panic(err) }
	var today = d

	var activeWin = CALENDAR_WIN // default window
	var selectedEvent = -1
	var yOffsetTodayWin = 0

	var updateSize = true
	var updateEvents = true
	var updateToday = true
	
	var ys = YearStructure{}
	var wPadding = 0
	var prevYear = 0
	var prevMonth = 0

	Setlocale(LC_ALL, "") // unicode support
	stdscr, err := Initscr()
	if err != nil { panic(err) }
	defer Endwin()
	rows, cols := stdscr.Getmaxyx()

	// size and pos of windows is set on updateSize
	eventsWin, err := Newwin(0, 0, 0, 0)
	if err != nil { panic(err) }
	calWidgetWin, err := Newwin(0, 0, 0, 0)
	if err != nil { panic(err) }
	todayWin, err := Newwin(0, 0, 0, 0)
	if err != nil { panic(err) }
	statusWin, err := Newwin(0, 0, 0, 0)
	if err != nil { panic(err) }

	Raw()
	Noecho()
	Curs_set(0)
	Halfdelay(4)

	Start_color()
	Use_default_colors()
	Init_pair(1, COLOR_RED, -1);
	Init_pair(2, COLOR_CYAN, -1);
	Init_pair(3, COLOR_YELLOW, -1);
	Init_pair(5, COLOR_WHITE, COLOR_BLUE);

	// Signal Handling for Terminal Resize Detection
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)
	go func(){ for _ = range c { updateSize = true } }()

	for {
		if updateSize {
			Endwin()
			Refresh()
			rows, cols = stdscr.Getmaxyx()

			eventsWin.Resize(rows-2, cols-34-wPadding)
			calWidgetWin.Resize(10, 34)
			calWidgetWin.Mv(0, cols-34)
			todayWin.Resize(rows-10-2, 34)
			todayWin.Mv(10, cols-34)
			statusWin.Resize(2, cols)
			statusWin.Mv(rows-2, 0)

			updateToday = true // to update todayMessageLines ( based on new rows/cols )
			updateSize = false
		}
		if d.Month != prevMonth || d.Year != prevYear {
			if d.Year != prevYear {
				ys = GenerateYearStructure(d.Year)	
				prevYear = d.Year
			}
			if d.Month != prevMonth { prevMonth = d.Month }
			updateEvents = true
		}
		if updateEvents {
			start := time.Now()

			// Clear events and populate via remind command
			// This takes the longest and could freeze ui but generally only takes 0.03s
			events = map[string][]Event{}
			year, month := SubtractMonth(d.Year, d.Month)
			rawEvents = getEvents(filename, year, month, 3)
			evaluateRawEvents(rawEvents, events) 

			if debug { statusMessage = fmt.Sprintf("Remind took %fs", time.Now().Sub(start).Seconds()) }
			updateEvents = false
		}
		if updateToday && todayWinEnabled {
			todayMessageLines = getToday(filename, today.Year, today.Month, today.Day, cols-34-wPadding-2)
			updateToday = false
		}

		if activeWin != EVENTS_WIN { selectedEvent = -1 } else if selectedEvent == -1 { selectedEvent = 0 }

		eventsWin.Erase()
		drawEvents(eventsWin, rows-2, cols-34-wPadding, 0, 0, activeWin == EVENTS_WIN, d, events, 0, selectedEvent)
		eventsWin.Refresh()

		updateCalendar(calWidgetWin, 0, 0, activeWin == CALENDAR_WIN, ys, d, today, events)
		calWidgetWin.Refresh()

		if todayWinEnabled {
			todayWin.Erase()
			drawToday(todayWin, rows-10-2, 34, 0, 0, yOffsetTodayWin, activeWin == TODAY_WIN, todayMessageLines)
			todayWin.Refresh()
		}

		drawStatus(statusWin, cols, statusMessage)
		statusWin.Refresh()

		c := Getch()
		exit := false
		statusMessage = ""
		switch(c) {
			case 'q': 
				exit = true
			case 'l':
				if activeWin == CALENDAR_WIN { d.AddDay() } 
			case 'h':
				if activeWin == CALENDAR_WIN { d.SubtractDay() }
			case 'j':
				if activeWin == CALENDAR_WIN { d.AddWeek() 
				} else if activeWin == EVENTS_WIN { 
					if dayEvents, ok := events[d.NumericString()]; ok {
						selectedEvent++
						if selectedEvent >= len(dayEvents) {
							selectedEvent = 0
							d.AddDay() 
						}
					} else {
						d.AddDay() 
					}
				} else if activeWin == TODAY_WIN {
					yOffsetTodayWin += 1
					todayHeight := rows-10-2
					if yOffsetTodayWin > len(todayMessageLines) - (todayHeight-2) {
						yOffsetTodayWin = len(todayMessageLines) - (todayHeight-2)
						if yOffsetTodayWin < 0 { yOffsetTodayWin = 0 }

					}
				}
			case 'k': 
				if activeWin == CALENDAR_WIN { d.SubtractWeek() 
				} else if activeWin == EVENTS_WIN { 
					if selectedEvent == 0 {
						d.SubtractDay() 
						if dayEvents, ok := events[d.NumericString()]; ok {
							selectedEvent = len(dayEvents) - 1
						}
					} else {
						selectedEvent--
					}
				} else if activeWin == TODAY_WIN {
					yOffsetTodayWin--
					if yOffsetTodayWin < 0 {
						yOffsetTodayWin = 0
					}
				}
			case 'J':
				if activeWin == CALENDAR_WIN { d.AddMonth() }
			case 'K':
				if activeWin == CALENDAR_WIN { d.SubtractMonth() }
			case 9:
				if activeWin == CALENDAR_WIN { activeWin = EVENTS_WIN 
				} else if activeWin == EVENTS_WIN { 
					if todayWinEnabled {
						activeWin = TODAY_WIN 
					} else { activeWin = CALENDAR_WIN }
				} else if activeWin == TODAY_WIN { activeWin = CALENDAR_WIN }
				statusMessage = "Chg Win"
			case 'e':
				//escaping from curses mode temporarily
				Endwin()
				openEditor(filename)
				updateEvents = true
				updateToday = true
			case -1: // skip ERR ( see halfdelay )
			case KEY_RESIZE:
			default: 
				statusMessage = fmt.Sprintf("Unbound key: '%c'", c)
		}
		if exit { break	}
	}
}

func trimMessage(message string, max int) string {
	if len(message) > max {
		if max < 1 {
			message = ""
		} else if max < 3 {
			message = "."
		} else {
			message = message[:max-3] + "..."
		}
	}
	return message
}


//@TODO really bad way to do it in gerneral
func drawEvents(
	win *Window, 
	h int, w int, y int, x int, active bool,
	d Date, events map[string][]Event, 
	daySelection int, eventSelection int,
	) {
	wPadding := 1
	maxMessage := w - 2 - 2*wPadding

	if active { Wattron(win, COLOR_PAIR(1)) }
	drawBox(win, h, w, y, x)
	Wattroff(win, COLOR_PAIR(1))

	row := 1
	yOffset := 0
	if eventSelection > h-5 {
		yOffset = eventSelection-(h-5)
	}
	for count := 0; ; count++ {
		if row >= h-2+yOffset { break }
		dateLabel := fmt.Sprintf("%s %d, %d", time.Month(d.Month).String(), d.Day, d.Year)
		attrs := COLOR_PAIR(1)
		if count == daySelection { attrs |= A_BOLD }
		if row > yOffset { 
			Wattron(win, attrs)
			Mvwprintw(win, y+row-yOffset, w-len(dateLabel)-1-wPadding, dateLabel) 
			Wattroff(win, attrs)
		}
		row++
		if row >= h-2+yOffset { break }

		if dayEvents, ok := events[d.NumericString()]; ok {
			for ei, event := range dayEvents {
				if row > yOffset { 
					if count == daySelection && ei == eventSelection {  
						Wattron(win, attrs)
					}
					
					Mvwprintw(win, y+row-yOffset, 1+wPadding, trimMessage(event.Message, maxMessage)) 
					Wattroff(win, attrs)
				}
				row++
				if row >= h-2+yOffset { break }
			}
			row--
		}
		row++
		if row >= h-2+yOffset { break }

		if row > yOffset { Mvwhline(win, y+1+row-yOffset, x+1, ACS_HLINE, w-2) }

		row+=2
		if row >= h-2+yOffset { break }
		d.AddDay()
	}
}

func updateCalendar(win *Window, y int, x int, active bool, ys YearStructure, d Date, today Date, events map[string][]Event) {
	monthYearLabel := time.Month(d.Month).String() + " " + strconv.Itoa(d.Year)
	selection := 0
	todayIndex := -1

	daysInMonthPrev := ys.daysInMonths[d.Month-1]
	daysInMonth := ys.daysInMonths[d.Month]

	wdStart := Weekday(d.Year, time.Month(d.Month), 1)
	// weekday transform Mo 0 ... Sun 6
	if wdStart == 0 { wdStart = 7 }; wdStart--

	wdEnd := Weekday(d.Year, time.Month(d.Month), daysInMonth)
	if wdEnd == 0 { wdEnd = 7 }; wdEnd--

	dayNr := 0
	for _, daysInMonth := range ys.daysInMonths[1:d.Month] {
		dayNr += daysInMonth
	}
	weekNr := int((dayNr + 1 - 1) / 7) + 1
	dayNr += d.Day

	days := [42]int{}
	eventsIndex := [42]bool{}
	i := 0
	for j:=daysInMonthPrev-wdStart+1; j<=daysInMonthPrev; j++ {
		days[i] = j

		// add events
		year, month := SubtractMonth(d.Year, d.Month)
		if _, ok := events[NumericString(year, month, j)]; ok { eventsIndex[i] = true }

		if today.Day == j && today.Month == month && today.Year == year { todayIndex = i }

		i++
	}
	for j:=1; j<=daysInMonth; j++ {
		days[i] = j
		if j == d.Day { selection = i }

		// add events
		if _, ok := events[NumericString(d.Year, d.Month, j)]; ok { eventsIndex[i] = true }

		if today.Day == j && today.Month == d.Month && today.Year == d.Year { todayIndex = i }

		i++
	}
	for j:=1; j<=6-wdEnd; j++ {
		days[i] = j

		// add events
		year, month := AddMonth(d.Year, d.Month)
		if _, ok := events[NumericString(year, month, j)]; ok { eventsIndex[i] = true }

		if today.Day == j && today.Month == month && today.Year == year { todayIndex = i }

		i++
	}

	weeks := [6]int{}
	for i:=0; i<6; i++ {
		weeks[i] = weekNr + i
	}
	
	drawCalendar(win, y, x, active, monthYearLabel, days, weeks, dayNr, selection, todayIndex, eventsIndex)
}

// Draws fixed length calendar widget height=10, width=34
// does not require erase overwrites old spots
// if any day is 0 entire row is left empty
func drawCalendar(
	win *Window, y int, x int, active bool, 
	monthYearLabel string, days [42]int, weeks[6]int, dayNr int, 
	selection int, todayIndex int, eventsIndex [42]bool,
	) {

	weekdays := "Mon Tue Wed Thu Fri Sat Sun"

	if active { Wattron(win, COLOR_PAIR(1)) } 
	win.Box(0, 0)
	Wattron(win, COLOR_PAIR(1))
	Mvwprintw(win, y, x+27, fmt.Sprintf("(#%3d)", dayNr))
	Mvwprintw(win, y+1,x+5, "                           ")
	Mvwprintw(win, y+1,x+5+(len(weekdays)-len(monthYearLabel))/2, monthYearLabel)
	Mvwprintw(win, y+2,x+5, weekdays)
	Wattroff(win, COLOR_PAIR(1))

	// add days
	count := 0
	for row:=0; row<6; row++{
		emptyRow := false
		days7 := days[row*7:row*7+7]
		for col, d := range days7 {
			// if any day is 0 entire row is skipped
			if d == 0 { emptyRow = true; break } 

			if eventsIndex[count] { Wattron(win, COLOR_PAIR(2)) }
			if todayIndex == count { Wattron(win, COLOR_PAIR(3)) }
			Mvwprintw(win, y+3+row, x+1+4+col*4, fmt.Sprintf(" %2d ", d))
			Wattroff(win, COLOR_PAIR(2))
			Wattroff(win, COLOR_PAIR(3))
			count++
		}
		weekLabel := fmt.Sprintf(" %2d ", weeks[row])
		if emptyRow {
			Mvwprintw(win, y+3+row, x+1+4, "                            ")
			// if row empty also don't print weekLabel
			weekLabel = "    "
		} 
		Wattron(win, COLOR_PAIR(1))
		Mvwprintw(win, y+3+row, x+1, weekLabel)
		Wattroff(win, COLOR_PAIR(1))
	}
	// add selection
	selectedCalRow := int(selection/7)
	selectedCalCol := int(math.Abs(float64(selection%7)))

	Wattron(win, COLOR_PAIR(1) | A_BOLD)
	Mvwprintw(win, y+3+selectedCalRow, x+1+4+4*selectedCalCol, "[")
	Mvwprintw(win, y+3+selectedCalRow, x+1+7+4*selectedCalCol, "]")
	Wattroff(win, COLOR_PAIR(1) | A_BOLD)
}
func drawBox(win *Window, height int, width int, y int, x int) {
	if height < 3 || width < 3 { 
		panic("drawBox minValue height and width is 3") 
	}

	Mvwvline(win, y, x, ACS_ULCORNER, 1)
	Mvwhline(win, y, x+1, ACS_HLINE, width-2)
	Mvwvline(win, y, x+width-1, ACS_URCORNER, 1)

	Mvwvline(win, y+1, x, ACS_VLINE, height-2)
	Mvwhline(win, y+height-1, x, ACS_LLCORNER, 1)

	Mvwvline(win, y+1, x+width-1, ACS_VLINE, height-2)
	Mvwhline(win, y+height-1, x+width-1, ACS_LRCORNER, 1)
	Mvwhline(win, y+height-1, x+1, ACS_HLINE, width-2)
}

func drawToday(win *Window, h int, w int, y int, x int, yOffset int, active bool, lines []string) {
	for row, line := range lines {
		Mvwprintw(win, 1+row-yOffset, 1, line)
	}

	if active { Wattron(win, COLOR_PAIR(1)) }
	drawBox(win, h, w, y, x)
	Wattroff(win, COLOR_PAIR(1))
}

func drawStatus(win *Window, width int, message string) {
	padding := 1
	maxMessage := width-2*padding

	// status message
	Wattron(win, COLOR_PAIR(5))
	Mvwhline(win, 0, 0, ACS_HLINE, width)
	if len(message) > maxMessage { message = message[:maxMessage] }
	Mvwprintw(win, 0, padding, message)
	Wattroff(win, COLOR_PAIR(5))

	// controls
	Mvwprintw(win, 1, padding, "q:Quit TAB:ChgWin  e:Edit  h:Right  j:Down  k:Up  l:Left")
}


func DaysInMonth(year int, month time.Month) int {
	switch(month) {
	case 1: return 31
	case 2:
		if IsLeapYear(year) { return 29
		} else { return 28 }
	case 3: return 31
	case 4: return 30
	case 5: return 31
	case 6: return 30
	case 7: return 31
	case 8: return 31
	case 9: return 30
	case 10: return 31
	case 11: return 30
	case 12: return 31
	default: panic(fmt.Sprintf("Invalid month %d", month) )
	}
}
func IsLeapYear(year int) bool {
	if year % 4 == 0 {
		if year % 100 == 0 && year % 400 != 0 {
			return false 
		}
		return true 
	}
	return false
}

// 0 = Sunday ... 6 = Saturday
func Weekday(year int, month time.Month, day int) int {
	d := day; m := int(month); y := year
	if month<3 {
		d+= y; y--
	} else {
		d+= y-2 
	}
	return (23*m/9+d+4+y/4-y/100+y/400) % 7
}
