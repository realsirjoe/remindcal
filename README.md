# Remindcal - Terminal Frontend for Remind

Remind is the way to go for all your future events but I thought it needs a frontend

![Demo](https://raw.githubusercontent.com/realsirjoe/remindcal/master/demo.png)

## Requirements

- remind https://dianne.skoll.ca/projects/remind/
- ncurses ( installed on most systems )

## Build from Source

    go build
    install remindcal /usr/local/bin/

## Download

- Pick a release on github
- Download binary for your architecture
- mv binary to /usr/local/bin/

## Workflow
Remind lets you organize your events in files. The event format is very intuitive and human readable, while still allowing high flexibility and quite complex events. For starters choose a location to store your events. It can be a filepath or a path to a directory containing multiple files ending with .rem, for example type 

    remindcal ~/.reminders

or some other filepath to store all your events. 
To make things easier I created 

    alias cal="remindcal ~/.reminders"

Now you can browse through all your events. You can use vim keys or the arrow keys to walk around the calendar. To switch to a different window press TAB
If you want to exit just press 'q'

In case you haven't gotten any events yet type 'e' which opens the events file(s) in your 
editor of choice ( EDITOR env var )
Simply add events by adding lines to the file(s)
For example:

    REM July 4 MSG Independence Day 

Will create an event for July 4th for all past years and all years to come

Going on vacation:

    REM May 7 2023 THROUGH May 21 2023 MSG Europe Trip

Meeting:

    REM August 27 2023 AT 9:30 UNTIL 11:00

You can also specify events in a more conventional way:

    REM 2023-12-23 MSG Christmas Party

More examples are on the [Remind Wiki](https://dianne.skoll.ca/wiki/Remind)

When you are done simply exit your editor and you'll be back in remindcal with your event added
This workflow allows me to store all my events in a maintainable format while sticking to the unix philosophy
