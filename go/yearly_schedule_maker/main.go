package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"os"
	"time"
)

var year = flag.Int("year", defaultYear(), "Year to generate configuration for, defaults to current year when current month is May or earlier, and to next year if current month is June or later.")
var startPerson = flag.Int("start", 0, "REQUIRED: Person to start the schedule with, should be greater than 0 and less or equal to 'people'.")
var dir = flag.String("dir", "", "REQUIRED: directory to put the files in.")

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()

	glog.Infof("Generating schedule for year %d.", *year)

	conf, err := loadConfig(fmt.Sprintf("yearly-schedule-maker-%d", *year), *year)
	if err != nil {
		glog.Exitf("Could not load config: %v.", err)
	}

	if *startPerson <= 0 || *startPerson > len(conf.People) || len(*dir) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	directory, err := os.Open(*dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(*dir, os.ModePerm)
		if err != nil {
			glog.Exitf("Could not create %q.", *dir)
		}
		directory, err = os.Open(*dir)
		if err != nil {
			glog.Exitf("Directory %q was created but cannot open it.", *dir)
		}
	} else if err != nil {
		glog.Exitf("Directory %q exists but cannot open it.", *dir)
	}
	defer directory.Close()

	stat, err := directory.Stat()
	if err != nil {
		glog.Exitf("Directory %q exists but cannot stat it.", *dir)
	}

	if !stat.IsDir() {
		glog.Exitf("%v exists and is not a directory.", *dir)
	}

	Generate(conf, *year, *startPerson, *dir)
}

func defaultYear() int {
	currentTime := time.Now()
	year := currentTime.Year()
	if currentTime.Month() > 6 {
		year++
	}
	return year
}
