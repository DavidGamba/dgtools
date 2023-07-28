package version

import (
	"fmt"
	"runtime/debug"
	"time"
)

type BuildInfo struct {
	Revision string
	Time     time.Time
	Modified bool
}

func GetBuildInfo() (*BuildInfo, error) {
	bi := &BuildInfo{}
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				bi.Revision = s.Value
			case "vcs.time":
				vcsTime := s.Value
				date, err := time.Parse("2006-01-02T15:04:05Z", vcsTime)
				if err != nil {
					return bi, fmt.Errorf("failed to parse time: %w", err)
				}
				bi.Time = date
			case "vcs.modified":
				if s.Value == "true" {
					bi.Modified = true
				}
			}
		}
	}
	return bi, nil
}
