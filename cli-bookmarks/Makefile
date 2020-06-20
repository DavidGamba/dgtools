# This file is part of cli-bookmarks.
#
# Copyright (C) 2018  David Gamba Rios
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

NAME=cli-bookmarks

BUILD_FLAGS=-ldflags="-X github.com/DavidGamba/$(NAME)/semver.BuildMetadata=`git rev-parse HEAD`"

.PHONY: test debug deps

test:
	go test ./...

debug:
	pwd
	echo ${GOPATH}
	ls **

deps:
	go get -u github.com/DavidGamba/go-getoptions
	go get -u github.com/DavidGamba/ffind
	go get -u github.com/nsf/termbox-go
	go get -u github.com/BurntSushi/toml

cover:
	# Works as of Go v1.10
	go test -coverprofile=coverage.out -covermode=atomic ./...

view:
	go tool cover -html=coverage.out

doc:
	asciidoctor README.adoc

man:
	asciidoctor -b manpage $(NAME).adoc

open:
	open README.html

build:
	go build $(BUILD_FLAGS)

install:
	go install $(BUILD_FLAGS) main.go

rpm:
	rpmbuild -bb rpm.spec \
		--define '_rpmdir ./RPMS' \
		--define '_sourcedir ${PWD}' \
		--buildroot ${PWD}/buildroot
