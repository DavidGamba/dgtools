package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/DavidGamba/go-getoptions"
)

const semVersion = "0.1.0"

var Logger = log.New(io.Discard, "", log.LstdFlags)

func main() {
	os.Exit(program(os.Args))
}

// Time format used for printing
// Choose between "15:04" or "03:04 PM"
var HourMinuteFormat = "15:04"

// TODO
var HourFormat = "15"

func program(args []string) int {
	opt := getoptions.New()
	opt.SetUnknownMode(getoptions.Pass)
	opt.Bool("verbose", false, opt.GetEnv("TZ_VERBOSE"), opt.Description("Enable logging"))
	opt.Bool("format-standard", false, opt.Alias("format-12-hour", "format-12h"), opt.Description("Use standard 12 hour AM/PM time format"))
	opt.Bool("short", false, opt.Alias("s"), opt.Description("Don't show timezone bars"))
	opt.String("config", "", opt.Alias("c"), opt.Description("Config file"))
	opt.String("group", "", opt.Description("Group to show"))
	opt.SetCommandFn(Run)

	list := opt.NewCommand("list", "list all timezones")
	list.SetCommandFn(ListRun)

	cities := opt.NewCommand("cities", "filter cities list")
	cities.Bool("all", false, opt.Alias("a"), opt.Description("Show all cities"))
	cities.String("country-code", "", opt.Alias("cc"), opt.Description("Filter by country code"))
	cities.SetCommandFn(CitiesRun)

	version := opt.NewCommand("version", "show version")
	version.SetCommandFn(VersionRun)

	opt.HelpCommand("help", opt.Alias("?"))
	remaining, err := opt.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	if opt.Called("verbose") {
		Logger.SetOutput(os.Stderr)
	}
	if opt.Called("format-standard") {
		HourMinuteFormat = "03:04 PM"
		HourFormat = "03"
	}
	Logger.Println(remaining)

	ctx, cancel, done := getoptions.InterruptContext()
	defer func() { cancel(); <-done }()

	err = opt.Dispatch(ctx, remaining)
	if err != nil {
		if errors.Is(err, getoptions.ErrorHelpCalled) {
			return 1
		}
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return 1
	}
	return 0
}

func Run(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	configFile := opt.Value("config").(string)
	group := opt.Value("group").(string)
	short := opt.Value("short").(bool)

	if configFile == "" {
		configFile = os.Getenv("HOME") + "/.config/tz/config.cue"
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		configFile = os.Getenv("HOME") + "/.tz.cue"
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file not provided and not found in ~/.config/tz/config.cue or ~/.tz.cue")
	}

	c, err := ReadConfig(ctx, configFile)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	Logger.Printf("%+v\n", c)

	if group == "" {
		group = c.DefaultGroup
	}

	am, err := ConfigToMemberMap(c, group)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	Logger.Printf("%+v\n", am)

	p := NewPalette("BlueYellow")
	PrintMembers(am, short, p)

	return nil
}

func ConfigToMemberMap(c *Config, group string) (MemberMap, error) {
	cc := NewCities()
	am := make(MemberMap)
	for _, member := range c.Group[group].Member {
		at := MemberTime{
			Member:  member.Name,
			Display: member.Name,
			Type:    member.Type,
		}
		if member.TimeZone != "" {
			location := member.TimeZone
			loc, err := time.LoadLocation(location)
			if err != nil {
				return am, fmt.Errorf("failed to load '%s': %w", location, err)
			}
			now := time.Now().In(loc)
			_, offset := now.Zone()
			at.Location = location
			at.Time = now
			at.Offset = offset
			at.Abbreviation = now.Format("MST")
			am[offset] = append(am[offset], at)
		} else if member.City != "" {
			Logger.Printf("Searching for city: %s - %s\n", member.City, member.CountryCode)
			cities, err := cc.Get(member.City, member.Admin1, member.CountryCode)
			if err != nil {
				return am, fmt.Errorf("failed search: %w", err)
			}
			Logger.Printf("Found cities: %+v\n", len(cities))
			if len(cities) > 1 {
				PrintCities(cities)
			}
			if len(cities) == 1 {
				location := cities[0].TimeZone
				loc, err := time.LoadLocation(location)
				if err != nil {
					return am, fmt.Errorf("failed to load '%s': %w", location, err)
				}
				now := time.Now().In(loc)
				_, offset := now.Zone()
				at.Location = location
				at.Time = now
				at.Offset = offset
				at.Abbreviation = now.Format("MST")
				am[offset] = append(am[offset], at)
			}
		}

		Logger.Printf("%+v\n", at)
	}

	return am, nil
}

// List of locations can be found in "/usr/share/zoneinfo" in both Linux and macOS
func ListRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	short := opt.Value("short").(bool)

	locations := []string{
		"Australia/Sydney",
		"Asia/Tokyo",
		"Asia/Hong_Kong",
		"Asia/Shanghai",
		"Europe/Berlin",
		"Europe/Paris",
		"Europe/Madrid",
		"Europe/London",
		"UTC",
		"America/Bogota",
		"America/New_York",
		"America/Toronto",
		"America/Chicago",
		"America/Costa_Rica",
		"America/Edmonton",
		"America/Los_Angeles",
	}

	am := make(MemberMap)
	count := 0
	for _, location := range locations {
		loc, err := time.LoadLocation(location)
		if err != nil {
			return fmt.Errorf("failed to load '%s': %w", location, err)
		}
		now := time.Now().In(loc)
		_, offset := now.Zone()
		at := MemberTime{
			Member:   location,
			Location: location,
			Time:     now,
			Offset:   offset,
			Display:  fmt.Sprintf("@%s (%s)", location, now.Format("MST")),
		}
		Logger.Printf("@%s: %s %d", at.Member, at.Time.Format("01/02 15:04 MST"), offset/3600)
		if _, ok := am[offset]; !ok {
			am[offset] = []MemberTime{at}
		} else {
			am[offset] = append(am[offset], at)
		}
		count++
	}
	Logger.Printf("Total: %d\n", count)

	p := NewPalette("BlueYellow")
	PrintMembers(am, short, p)
	return nil
}

func CitiesRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	Logger.Printf("Cities")
	all := opt.Value("all").(bool)
	countryCode := opt.Value("country-code").(string)

	if len(args) == 0 && !all {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", "Missing city name")
		fmt.Fprintf(os.Stderr, "%s", opt.Help(getoptions.HelpSynopsis))
		return getoptions.ErrorHelpCalled
	}

	nameQuery := strings.Join(args, " ")
	cc := NewCities()
	_, err := cc.Search(nameQuery, countryCode)
	if err != nil {
		return fmt.Errorf("failed search: %w", err)
	}

	return nil
}

func VersionRun(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	var revision, timeStr, modified string
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				revision = s.Value
			case "vcs.time":
				vcsTime := s.Value
				date, err := time.Parse("2006-01-02T15:04:05Z", vcsTime)
				if err != nil {
					return fmt.Errorf("failed to parse time: %w", err)
				}
				timeStr = date.Format("20060102_150405")
			case "vcs.modified":
				if s.Value == "true" {
					modified = "modified"
				}
			}
		}
	}
	v := semVersion
	if revision != "" && timeStr != "" {
		v += fmt.Sprintf("+%s.%s", revision, timeStr)
		if modified != "" {
			v += fmt.Sprintf(".%s", modified)
		}
	}
	fmt.Printf("%s\n", v)
	return nil
}
