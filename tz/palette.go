package main

import "github.com/charmbracelet/lipgloss"

var ClockEmoji = map[int]string{
	0:  "🕛",
	12: "🕛",
	24: "🕛",

	1:  "🕐",
	13: "🕐",

	2:  "🕑",
	14: "🕑",

	3:  "🕒",
	15: "🕒",

	4:  "🕓",
	16: "🕓",

	5:  "🕔",
	17: "🕔",

	6:  "🕕",
	18: "🕕",

	7:  "🕖",
	19: "🕖",

	8:  "🕗",
	20: "🕗",

	9:  "🕘",
	21: "🕘",

	10: "🕙",
	22: "🕙",

	11: "🕚",
	23: "🕚",
}

type Palette struct {
	Night     string
	Dawn      string
	Morning   string
	Noon      string // work hours
	AfterNoon string
	Dusk      string
	Evening   string

	FgNight     string
	FgDawn      string
	FgMorning   string
	FgNoon      string // work hours
	FgAfterNoon string
	FgDusk      string
	FgEvening   string

	Highlight   string
	FgHighlight string

	Member   string
	FgMember string

	LipglossPalette struct {
		Night     lipgloss.Style
		Dawn      lipgloss.Style
		Morning   lipgloss.Style
		Noon      lipgloss.Style // work hours
		AfterNoon lipgloss.Style
		Dusk      lipgloss.Style
		Evening   lipgloss.Style
		Highlight lipgloss.Style

		Member lipgloss.Style
	}
}

func NewPalette(theme string) *Palette {
	p := &BlueYellow

	p.LipglossPalette.Night = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgNight)).Background(lipgloss.Color(p.Night))
	p.LipglossPalette.Dawn = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgDawn)).Background(lipgloss.Color(p.Dawn))
	p.LipglossPalette.Morning = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgMorning)).Background(lipgloss.Color(p.Morning))
	p.LipglossPalette.Noon = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgNoon)).Background(lipgloss.Color(p.Noon))
	p.LipglossPalette.AfterNoon = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgAfterNoon)).Background(lipgloss.Color(p.AfterNoon))
	p.LipglossPalette.Dusk = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgDusk)).Background(lipgloss.Color(p.Dusk))
	p.LipglossPalette.Evening = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgEvening)).Background(lipgloss.Color(p.Evening))
	p.LipglossPalette.Highlight = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgHighlight)).Background(lipgloss.Color(p.Highlight))
	p.LipglossPalette.Member = lipgloss.NewStyle().Foreground(lipgloss.Color(p.FgMember)).Background(lipgloss.Color(p.Member))

	return p
}

func (p *Palette) Style(timeOfDay int) lipgloss.Style {
	switch timeOfDay {
	case 22, 23, 24, 0, 1, 2, 3, 4:
		return p.LipglossPalette.Night
	case 5, 6:
		return p.LipglossPalette.Dawn
	case 7, 8:
		return p.LipglossPalette.Morning
	case 9, 10, 11, 12, 13, 14:
		return p.LipglossPalette.Noon
	case 15, 16:
		return p.LipglossPalette.AfterNoon
	case 17, 18, 19:
		return p.LipglossPalette.Dusk
	case 20, 21:
		return p.LipglossPalette.Evening
	default:
		return p.LipglossPalette.Night
	}
}

var BlueYellow = Palette{
	Night: "#003E7F",

	Dawn:    "#406C95",
	Evening: "#406C95",

	Morning: "#7F99AB",
	Dusk:    "#7F99AB",

	Noon: "#FEF4D7",

	AfterNoon: "#B9C49F",

	FgNight:     "#FFFFFF",
	FgDawn:      "#FFFFFF",
	FgMorning:   "#FFFFFF",
	FgNoon:      "#000000",
	FgAfterNoon: "#000000",
	FgDusk:      "#FFFFFF",
	FgEvening:   "#FFFFFF",

	Highlight:   "#fafa00",
	FgHighlight: "#000000",

	// Member: "#5495e1",
	Member: "#A3D8BD",
	// Member:   "#F6F6F6",
	// Member:   "#FEF4D7",
	FgMember: "#000000",
}

var PurpleYellow = Palette{
	Night:     "#3d0066",
	Dawn:      "#5c0099",
	Morning:   "#c86bfa",
	Noon:      "#fdc500",
	AfterNoon: "#ffd500",
	Dusk:      "#ffee32",
	Evening:   "#03071e",
}

var BlueGreen = Palette{
	Night:     "#003e7f",
	Dawn:      "#0068af",
	Morning:   "#5495e1",
	Noon:      "#dddddd",
	AfterNoon: "#44a75e",
	Dusk:      "#008239",
	Evening:   "#00540e",
}
