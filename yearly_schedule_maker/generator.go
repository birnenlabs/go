package main

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"path"
	"time"
)

const fileHtml = "schedule.html"
const fileIcal = "schedule_%02d.ical"

func Generate(conf *Config, year int, startPerson int, dir string) {

	var personList bytes.Buffer
	for person := 1; person <= len(conf.People); person++ {
		personList.WriteString(fmt.Sprintf(personTmpl, person, conf.Person(person)))
	}

	person := startPerson
	days := make(map[Date]Day)
	for month := 1; month <= 12; month++ {
		t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

		glog.Infof("#### %s", t.Month())
		for t.Month() == time.Month(month) {
			day := Day{
				holiday:  conf.Holiday(t),
				day:      conf.Day(t),
				saturday: t.Weekday() == 6,
				sunday:   t.Weekday() == 0,
			}
			if len(day.holiday) == 0 && !day.saturday && !day.sunday {
				day.person = person
				if person < len(conf.People) {
					person++
				} else {
					person = 1
				}
			}

			date := Date{
				day:   t.Day(),
				month: int(t.Month()),
			}
			days[date] = day
			glog.Infof("%s %02d %s %s", day.day, date.day, day.holiday, conf.Person(day.person))

			t = t.AddDate(0, 0, 1)
		}
	}

	// Generate html
	var header bytes.Buffer
	for month := 1; month <= 12; month++ {
		header.WriteString(fmt.Sprintf(monthCellTmpl, conf.Month(month)))
	}

	var calendar bytes.Buffer
	calendar.WriteString(fmt.Sprintf(rowTmpl, header.String()))

	for dayNr := 1; dayNr <= 31; dayNr++ {

		var rows bytes.Buffer
		for month := 1; month <= 12; month++ {
			date := Date{
				day:   dayNr,
				month: month,
			}
			day, ok := days[date]

			class := ""
			if day.saturday {
				class = satClass
			}
			if len(day.holiday) > 0 || day.sunday {
				class = sunClass
			}

			if ok {
				desc := ""
				if day.person != 0 {
					desc = fmt.Sprintf("%d", day.person)
				} else if len(day.holiday) == 0 && day.saturday {
					desc = conf.Free
				}
				rows.WriteString(fmt.Sprintf(dayCellTmpl, class, day.day, class, dayNr, class, desc))
			} else {
				rows.WriteString(fmt.Sprintf(dayCellTmpl, "", "", "", "", "", ""))
			}
		}
		calendar.WriteString(fmt.Sprintf(rowTmpl, rows.String()))
	}

	htmlOutput := fmt.Sprintf(htmlTemplateOuter, year, personList.String(), conf.Url, calendar.String())
	err := ioutil.WriteFile(path.Join(dir, fileHtml), []byte(htmlOutput), 0644)
	if err != nil {
		glog.Errorf("Could not write to file %q: %v.", fileHtml, err)
	}

	// generate ical's
	for person := 1; person <= len(conf.People); person++ {
		firstOccurrence := ""
		var rDates bytes.Buffer
		for month := 1; month <= 12; month++ {
			for dayNr := 1; dayNr <= 31; dayNr++ {
				date := Date{
					day:   dayNr,
					month: month,
				}
				day := days[date]
				if day.person == person {
					if len(firstOccurrence) == 0 {
						firstOccurrence = fmt.Sprintf("%04d%02d%02d", year, month, dayNr)
					} else {
						rDates.WriteString(fmt.Sprintf(rDateTmpl, conf.Timezone, year, month, dayNr))
					}
				}
			}
		}
		eventTitle := fmt.Sprintf(conf.Event, person)
		fileName := fmt.Sprintf(fileIcal, person)
		icalOutput := fmt.Sprintf(icalTmpl, conf.Timezone, firstOccurrence, eventTitle, rDates.String())
		err := ioutil.WriteFile(path.Join(dir, fileName), []byte(icalOutput), 0644)
		if err != nil {
			glog.Errorf("Could not write to file %q: %v.", fileName, err)
		}
	}
}

type Date struct {
	day   int
	month int
}

type Day struct {
	person   int
	holiday  string
	day      string
	saturday bool
	sunday   bool
}
