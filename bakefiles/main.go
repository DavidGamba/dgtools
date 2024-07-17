package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/DavidGamba/dgtools/run"
	"github.com/DavidGamba/go-getoptions"
	"golang.org/x/mod/semver"
)

func getEnv(file string) ([]string, error) {
	env := []string{}
	// Check if build.env exists and load it if it does
	if _, err := os.Stat(file); err == nil {

		commentRegex := regexp.MustCompile(`^\s*#|^\s*//`)

		file, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if commentRegex.MatchString(line) {
				continue
			}
			Logger.Printf("Adding to env: %s\n", line)
			env = append(env, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}
	return env, nil
}

// install - Build and install the current binary
func Install(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}
		Logger.Printf("Running install on %s\n", wd)

		env, err := getEnv("build.env")
		if err != nil {
			return fmt.Errorf("failed to get env: %w", err)
		}

		cmd := run.CMD("go", "install")
		for _, e := range env {
			cmd.Env(e)
		}
		err = cmd.Log().Run()
		if err != nil {
			return err
		}
		return nil
	}
}

// test - run tests
func Test(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}
		Logger.Printf("Running test on %s\n", wd)

		env, err := getEnv("build.env")
		if err != nil {
			return fmt.Errorf("failed to get env: %w", err)
		}

		cmd := run.CMD("go", "test", "./...")
		for _, e := range env {
			cmd.Env(e)
		}
		err = cmd.Log().Run()
		if err != nil {
			return err
		}
		return nil
	}
}

// tag - show tags for the current tool
func Tag(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}
		Logger.Printf("Running tag on %s\n", wd)

		tags, err := getTags(wd)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
		return nil
	}
}

func Homebrew(opt *getoptions.GetOpt) getoptions.CommandFn {
	return func(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}
		Logger.Printf("Running tag on %s\n", wd)
		tool := filepath.Base(wd)

		tags, err := getTags(wd)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		if len(tags) == 0 {
			return fmt.Errorf("no tags found")
		}
		tag := tags[0]
		url := fmt.Sprintf("https://github.com/DavidGamba/dgtools/archive/refs/tags/%s/%s.tar.gz", tool, tag)

		var buf bytes.Buffer
		err = run.CMD("curl", "-sL", url).Log().Run(&buf, os.Stderr)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		h := sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			return err
		}
		sum := fmt.Sprintf("%x", h.Sum(nil))
		fmt.Printf("url: %s\n", url)
		fmt.Printf("sum: %s\n", sum)
		err = run.CMD("grepp", "--no-pager", "url.*", fmt.Sprintf("../HomebrewFormula/%s.rb", tool), "-r", fmt.Sprintf(`url "%s"`, url), "-f").Log().Run()
		if err != nil {
			return fmt.Errorf("failed to run grepp: %w", err)
		}
		err = run.CMD("grepp", "--no-pager", "sha256.*", fmt.Sprintf("../HomebrewFormula/%s.rb", tool), "-r", fmt.Sprintf(`sha256 "%s"`, sum), "-f").Log().Run()
		if err != nil {
			return fmt.Errorf("failed to run grepp: %w", err)
		}

		return nil
	}
}

func getTags(wd string) ([]string, error) {
	versions := []string{}
	base := filepath.Base(wd)
	filter := base
	output, err := run.CMD("git", "tag").DiscardErr().STDOutOutput()
	if err != nil {
		return versions, fmt.Errorf("failed to get tags: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if !strings.Contains(line, filter) {
			continue
		}
		version := strings.TrimPrefix(line, filter+"/")
		versions = append(versions, version)
	}
	if len(versions) == 0 {
		return versions, nil
	}
	sort.Sort(sort.Reverse(semver.ByVersion(versions)))
	return versions, nil
}
