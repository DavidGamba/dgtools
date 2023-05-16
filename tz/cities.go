package main

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed cities-tz.tsv
var f embed.FS

type City struct {
	Name        string
	Admin1Name  string // Admin division 1
	CountryCode string
	TimeZone    string
	Population  string
}

// CityMap is a map of city names to City.
type CityMap struct {
	m      map[string][]*City
	loaded bool
}

func NewCities() *CityMap {
	return &CityMap{
		m:      make(map[string][]*City),
		loaded: false,
	}
}

// Returns a list of cities with the given name and optionally a country code.
func (c *CityMap) Get(name, admin1Name, countryCode string) ([]*City, error) {
	if !c.loaded {
		err := c.load()
		if err != nil {
			return []*City{}, fmt.Errorf("failed to load cities table: %w", err)
		}
	}
	cities, ok := c.m[name]
	if !ok {
		return []*City{}, fmt.Errorf("no cities found for '%s'", name)
	}
	if len(cities) <= 1 || (admin1Name == "" && countryCode == "") {
		return cities, nil
	}
	cc := []*City{}
	ccAdmin1 := []*City{}
	if countryCode != "" {
		for _, city := range cities {
			if strings.EqualFold(countryCode, city.CountryCode) {
				cc = append(cc, city)
			}
		}
	} else {
		cc = cities
	}
	if admin1Name != "" {
		for _, city := range cc {
			if strings.EqualFold(admin1Name, city.Admin1Name) {
				ccAdmin1 = append(ccAdmin1, city)
			}
		}
	} else {
		ccAdmin1 = cc
	}
	return ccAdmin1, nil
}

func (c *CityMap) Search(name, countryCode string) ([]City, error) {
	if !c.loaded {
		err := c.load()
		if err != nil {
			return []City{}, fmt.Errorf("failed to load cities table: %w", err)
		}
	}
	count := 0
	cc := []*City{}
	for n, cities := range c.m {
		for _, city := range cities {
			count++
			if strings.Contains(strings.ToLower(n), strings.ToLower(name)) {
				if countryCode == "" || strings.EqualFold(countryCode, city.CountryCode) {
					cc = append(cc, city)
				}
			}
		}
	}
	sort.Slice(cc, func(i, j int) bool {
		return cc[i].Name < cc[j].Name
	})
	PrintCities(cc)
	return []City{}, nil
}

func (c *CityMap) load() error {
	if c.loaded {
		return nil
	}

	tableFilename := "cities-tz.tsv"
	tableFH, err := f.Open(tableFilename)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", tableFilename, err)
	}
	defer tableFH.Close()

	r := csv.NewReader(tableFH)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read table: %w", err)
		}

		// Column widths are known:
		// select max(length(asciiname)) from cities;
		// 49
		// select max(length(asciiname)) from admin1;
		// 40
		// select max(length(countrycode)) from cities;
		// 2
		// select max(length(timezone)) from cities;
		// 30
		// select max(length(population)) from cities;
		// 8 + 2 commas
		// name, admin1Name, countryCode, timeZone, population
		// Logger.Printf("%#v\n", record)
		p := message.NewPrinter(language.English)

		pop, err := strconv.Atoi(record[4])
		if err != nil {
			return fmt.Errorf("failed to parse population: %w", err)
		}
		population := p.Sprintf("%d", pop)
		if c.m[record[0]] == nil {
			c.m[record[0]] = []*City{{
				Name:        record[0],
				Admin1Name:  record[1],
				CountryCode: record[2],
				TimeZone:    record[3],
				Population:  population,
			}}
		} else {
			c.m[record[0]] = append(c.m[record[0]], &City{
				Name:        record[0],
				Admin1Name:  record[1],
				CountryCode: record[2],
				TimeZone:    record[3],
				Population:  population,
			})
		}
	}
	c.loaded = true

	return nil
}

func PrintCities(cities []*City) {
	light := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#eef2f3"))
	dark := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#dce3e6"))

	fmt.Printf("%s%s%s%s%s%s\n",
		dark.
			Width(19).
			PaddingLeft(1).
			PaddingRight(1).
			Render("Time"),
		dark.
			Width(49).
			PaddingLeft(1).
			PaddingRight(1).
			Render("Name"),
		dark.
			Width(42).
			PaddingLeft(1).
			PaddingRight(1).
			Render("Admin1"),
		dark.
			Width(4).
			PaddingLeft(1).
			PaddingRight(1).
			Render("CC"),
		dark.
			Width(32).
			PaddingLeft(1).
			PaddingRight(1).
			Render("TimeZone"),
		dark.
			Width(12).
			PaddingLeft(1).
			PaddingRight(1).
			Render("Population"),
	)

	for i, city := range cities {
		loc, _ := time.LoadLocation(city.TimeZone)
		now := time.Now().In(loc)
		if i%2 == 0 {
			fmt.Printf("%s%s%s%s%s%s\n",
				light.
					Width(19).
					PaddingLeft(1).
					PaddingRight(1).
					Render(now.Format("01/02 15:04 MST")),
				light.
					Width(49).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Name),
				light.
					Width(42).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Admin1Name),
				light.
					Width(4).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.CountryCode),
				light.
					Width(32).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.TimeZone),
				light.
					Width(12).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Population),
			)
		} else {
			fmt.Printf("%s%s%s%s%s%s\n",
				dark.
					Width(19).
					PaddingLeft(1).
					PaddingRight(1).
					Render(now.Format("01/02 15:04 MST")),
				dark.
					Width(49).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Name),
				dark.
					Width(42).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Admin1Name),
				dark.
					Width(4).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.CountryCode),
				dark.
					Width(32).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.TimeZone),
				dark.
					Width(12).
					PaddingLeft(1).
					PaddingRight(1).
					Render(city.Population),
			)
		}
	}
}
