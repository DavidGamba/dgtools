package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type MemberTime struct {
	Member       string
	Time         time.Time
	Location     string
	Offset       int // in seconds
	Display      string
	Abbreviation string
	Type         string // person, city, country
}

type MemberMap map[int][]MemberTime

func PrintMembers(am MemberMap, short bool, p *Palette) {
	offsets := []int{}
	for offset := range am {
		offsets = append(offsets, offset)
	}
	sort.Ints(offsets)
	Logger.Println(len(offsets))
	for _, offset := range offsets {
		PrintMembersLine(p, am[offset])
		if !short {
			PrintHours(p, am[offset][0].Time, am[offsets[0]][0].Time)
			fmt.Println()
		}
	}
}

func PrintMembersLine(p *Palette, att []MemberTime) {
	// map[abbreviation] = display
	displayLineCountries := map[string][]string{}
	displayLineCities := map[string][]string{}
	displayLinePeople := map[string][]string{}
	for _, at := range att {
		switch at.Type {
		case "country":
			if _, ok := displayLineCountries[at.Abbreviation]; !ok {
				displayLineCountries[at.Abbreviation] = []string{at.Display}
			} else {
				displayLineCountries[at.Abbreviation] = append(displayLineCountries[at.Abbreviation], at.Display)
				// TODO: who cares about the repeated sorting right?
				sort.Strings(displayLineCountries[at.Abbreviation])
			}
		case "city":
			if _, ok := displayLineCities[at.Abbreviation]; !ok {
				displayLineCities[at.Abbreviation] = []string{at.Display}
			} else {
				displayLineCities[at.Abbreviation] = append(displayLineCities[at.Abbreviation], at.Display)
				// TODO: who cares about the repeated sorting right?
				sort.Strings(displayLineCities[at.Abbreviation])
			}
		case "person":
			if _, ok := displayLinePeople[at.Abbreviation]; !ok {
				displayLinePeople[at.Abbreviation] = []string{at.Display}
			} else {
				displayLinePeople[at.Abbreviation] = append(displayLinePeople[at.Abbreviation], at.Display)
				// TODO: who cares about the repeated sorting right?
				sort.Strings(displayLinePeople[at.Abbreviation])
			}
		}
	}

	t := att[0].Time

	// Line length is 141

	fmt.Print(
		p.Style(t.Hour()).Render(ClockEmoji[t.Hour()]+" "),
		p.Style(t.Hour()).Render(t.Format("02 "+HourMinuteFormat+" -07")),
		p.Style(t.Hour()).Render("  "),
	)
	str := ""
	for abb, list := range displayLineCountries {
		str += p.Style(t.Hour()).Render(fmt.Sprintf(" %s ", abb))
		// for _, d := range list {
		// 	str += fmt.Sprint(p.LipglossPalette.Member.Render(fmt.Sprintf("%s", d)))
		// 	str += p.Style(t.Hour()).Render(" ")
		// }
		str += p.LipglossPalette.Member.Render(fmt.Sprintf(" %s ", strings.Join(list, " ")))
	}
	for abb, list := range displayLineCities {
		str += p.Style(t.Hour()).Render(fmt.Sprintf(" %s ", abb))
		str += p.LipglossPalette.Member.Render(fmt.Sprintf(" %s ", strings.Join(list, " ")))
	}
	for abb, list := range displayLinePeople {
		str += p.Style(t.Hour()).Render(fmt.Sprintf(" %s ", abb))
		str += p.LipglossPalette.Member.Render(fmt.Sprintf(" %s ", strings.Join(list, " ")))
	}
	// fmt.Println(lipgloss.PlaceHorizontal(132, lipgloss.Left, str, lipgloss.WithWhitespaceBackground(p.LipglossPalette.Noon.GetBackground())))
	fmt.Println(lipgloss.PlaceHorizontal(126, lipgloss.Left, str, lipgloss.WithWhitespaceBackground(p.Style(t.Hour()).GetBackground())))
	// 🕕 06:56   .
	// fmt.Println(str)
}

func PrintHours(p *Palette, t, base time.Time) {
	x := t.Hour()
	h := t.Hour()
	h -= 12

	for i := 0; i < 24; i++ {
		if h >= 24 {
			h = h - 24
		}
		if h < 0 {
			h = 24 + h
		}
		if h == x {
			PrintBlock(t.Format("15:04"), p.Style(h), h == x, p.LipglossPalette.Highlight)
		} else {
			PrintBlock(fmt.Sprintf("%02d", h), p.Style(h), h == x, p.LipglossPalette.Highlight)
		}
		h++
	}
	fmt.Println()

}

func PrintBlock(hour string, style lipgloss.Style, highlight bool, hStyle lipgloss.Style) {
	normal := style.Copy()

	normal.
		PaddingLeft(2).
		PaddingRight(2)

	if highlight {
		hour = hStyle.Render(hour)
	} else {
		hour = normal.Render(hour)
	}
	fmt.Printf("%s", hour)
}
