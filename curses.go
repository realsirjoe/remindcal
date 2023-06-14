package main

/*
#cgo LDFLAGS: -lncursesw
#cgo LDFLAGS: -L/usr/local/opt/ncurses/lib
#include <stdlib.h>
#include <stdio.h>
#include <locale.h>
#include <ncurses.h>

int go_mvprintw(int y, int x, char *name) {
	return mvprintw(y, x, name);
}
int go_mvwprintw(WINDOW *win, int y, int x, char *name) {
	return mvwprintw(win, y, x, name);
}

// normally a macro but to make it go ready we use pointers instead
void go_getmaxyx(WINDOW *win, int *y, int *x) {
	getmaxyx(win, *y, *x);
}
// usually a macro
int go_COLOR_PAIR(int p) { return COLOR_PAIR(p); }
*/
import "C"

import (
	"unsafe"
	"errors"
)

// Go Curses Lib 

type Window struct {
	win *C.WINDOW
}
func Initscr() (stdscr *Window, err error) {
	stdscr = &Window{C.initscr()}
	if unsafe.Pointer(stdscr.win) == nil {
		err = errors.New("An error occurred initializing ncurses")
	}
	return
}
func Raw() {
	C.raw()
}
func Echo() {
	C.echo()
}
func Noecho() {
	C.noecho()
}
func Keypad(w *Window, bf bool) {
	C.keypad(w.win, C.bool(bf));
}
func Curs_set(visibility int) {
	C.curs_set(C.int(visibility))
}
func Halfdelay(tenth int) {
	C.halfdelay(C.int(tenth));
}

func Mvprintw(y int, x int, text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.go_mvprintw(C.int(y), C.int(x), cs)
}
func Mvwprintw(w *Window, y int, x int, text string) {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.go_mvwprintw(w.win, C.int(y), C.int(x), cs)
}
func Refresh() {
	C.refresh()
}
func Endwin() {
	C.endwin()
}
func Getch() int {
	return int(C.getch())
}
const KEY_RESIZE        = C.KEY_RESIZE

///////////////// WINDOW ////////////////////
func Newwin(h int, w int, y int, x int) (window *Window, err error) {
	window = &Window{C.newwin(C.int(h), C.int(w), C.int(y), C.int(x))}
	if window.win == nil {
		err = errors.New("Failed to create a new window")
	}
	return 
}
func (w *Window) Getmaxyx() (y int, x int) {
	var cy, cx C.int
	C.go_getmaxyx(w.win, &cy, &cx)
	return int(cy), int(cx)
}
func (w *Window) Mv(y int, x int) {
	C.mvwin(w.win, C.int(y), C.int(x));
}
func (w *Window) Resize(lines int, cols int) {
	C.wresize(w.win, C.int(lines), C.int(cols));
}
func (w *Window) Refresh() {
	C.wrefresh(w.win)
}
func (w *Window) Erase() {
	C.werase(w.win)
}

///////////////// BOX ///////////////////////
const ACS_ULCORNER      = C.A_ALTCHARSET + 'l'
const ACS_LLCORNER      = C.A_ALTCHARSET + 'm'
const ACS_URCORNER      = C.A_ALTCHARSET + 'k'
const ACS_LRCORNER      = C.A_ALTCHARSET + 'j'
const ACS_LTEE          = C.A_ALTCHARSET + 't' 
const ACS_RTEE          = C.A_ALTCHARSET + 'u' 
const ACS_BTEE          = C.A_ALTCHARSET + 'v'
const ACS_TTEE          = C.A_ALTCHARSET + 'w'
const ACS_HLINE         = C.A_ALTCHARSET + 'q'
const ACS_VLINE         = C.A_ALTCHARSET + 'x'

func Mvhline(y int, x int, ch int, n int) {
	C.mvhline(C.int(y), C.int(x), C.uint(ch), C.int(n))
}
func Mvvline(y int, x int, ch int, n int) {
	C.mvvline(C.int(y), C.int(x), C.uint(ch), C.int(n))
}
func Mvwhline(w *Window, y int, x int, ch int, n int) {
	C.mvwhline(w.win, C.int(y), C.int(x), C.uint(ch), C.int(n))
}
func Mvwvline(w *Window, y int, x int, ch int, n int) {
	C.mvwvline(w.win, C.int(y), C.int(x), C.uint(ch), C.int(n))
}

func (w *Window) Box(vch, hch int) error {
	if C.box(w.win, C.chtype(vch), C.chtype(hch)) == C.ERR {
		return errors.New("Failed to draw box around window")
	}
	return nil
}

///////////////// COLOR ///////////////////
const A_BOLD   = int(C.A_BOLD)

const COLOR_BLACK   = 0
const COLOR_RED     = 1
const COLOR_GREEN   = 2
const COLOR_YELLOW  = 3
const COLOR_BLUE    = 4
const COLOR_MAGENTA = 5
const COLOR_CYAN    = 6
const COLOR_WHITE   = 7

func Start_color() {
	C.start_color()
}
func Use_default_colors() {
	C.use_default_colors()
}
func Init_pair(pair int, f int, b int) {
	C.init_pair(C.short(pair), C.short(f), C.short(b))
}
func COLOR_PAIR(n int) int {
	return int(C.go_COLOR_PAIR(C.int(n)))
}
func Attron(attrs int) {
	C.attron(C.int(attrs))
}
func Attroff(attrs int) {
	C.attroff(C.int(attrs))
}
func Wattron(w *Window, attrs int) {
	C.wattron(w.win, C.int(attrs));
}
func Wattroff(w *Window, attrs int) {
	C.wattroff(w.win, C.int(attrs));
}

//////////////////// LOCALE ///////////////////////
////////////// (Unicode support) //////////////////
const LC_ALL = int(C.LC_ALL)

func Setlocale(category int, locale string) {
	cs := C.CString(locale)
	C.setlocale(C.int(category), cs)
	C.free(unsafe.Pointer(cs))
}
