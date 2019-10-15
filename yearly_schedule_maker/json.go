package main

import (
	"birnenlabs.com/lib/conf"
	"fmt"
	"time"
)

type Config struct {
	// Array of people to generate the schedule for
	People []string
	// Array of holidays
	Holidays []Holiday
	// Array of localized day names to use, starts with Sunday, use short name (e.g. Su, Mo..)
	Days []string
	// Array of localized month names, use long name (e.g. January, February...)
	Months []string
	// Text added when the schedule is free for everyone to use
	Free string
	// Url to be displayed on the generated calendar
	Url string
	// Event title used in ical file, should contain exactly one %d, where person number will be added
	Event string
	// Timezone used in ical file. This is only used in the ical file and does not affect schedule calculations.
	Timezone string
}

type Holiday struct {
	Year  int
	Month int
	Day   int
	Name  string
}

func loadConfig(configName string, year int) (*Config, error) {
	var c Config
	err := conf.LoadConfigFromJson(configName, &c)
	if err != nil {
		return nil, err
	}

	if len(c.People) < 2 {
		return nil, fmt.Errorf("unexpected size of people array: %d", len(c.People))
	}
	if len(c.Days) != 7 {
		return nil, fmt.Errorf("unexpected size of days array: %d", len(c.Days))
	}
	if len(c.Months) != 12 {
		return nil, fmt.Errorf("unexpected size of months array: %d", len(c.Months))
	}
	if len(c.Holidays) < 5 {
		return nil, fmt.Errorf("less than 5 holidays defined: %d", len(c.Holidays))
	}

	for _, h := range c.Holidays {
		if h.Day == 0 {
			return nil, fmt.Errorf("unexpected holiday Day value: %+v", h)
		}
		if h.Month == 0 {
			return nil, fmt.Errorf("unexpected holiday Month value: %+v", h)
		}
		if h.Year != 0 && h.Year != year {
			return nil, fmt.Errorf("unexpected holiday Year value: %+v", h)
		}
		if len(h.Name) == 0 {
			return nil, fmt.Errorf("unexpected holiday Name value: %+v", h)
		}
	}
	return &c, nil

}

func (c *Config) Day(t time.Time) string {
	return c.Days[t.Weekday()]
}

func (c *Config) Month(month int) string {
	return c.Months[month-1]
}

func (c *Config) Holiday(t time.Time) string {
	for _, h := range c.Holidays {
		if h.Matches(t) {
			return h.Name
		}
	}
	return ""
}

func (c *Config) Person(person int) string {
	if person <= 0 || person > len(c.People) {
		return ""
	}
	return c.People[person-1]
}

func (h *Holiday) Matches(t time.Time) bool {
	// Ignoring year, as the program works on one year intervals
	return t.Day() == h.Day && t.Month() == time.Month(h.Month)
}
