# This file is part of grepp.
#
# Copyright (C) 2012-2017  David Gamba Rios
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

BUILD_FLAGS=-ldflags="-X github.com/DavidGamba/grepp/semver.BuildMetadata=`git rev-parse HEAD`"

test:
	go test ./...

debug:
	pwd; \
	echo ${GOPATH}; \
	ls **;

deps:
	go get github.com/DavidGamba/go-getoptions

cover:
	go test -coverprofile=c.out -covermode=atomic github.com/DavidGamba/grepp

view:
	go tool cover -html=c.out

doc:
	asciidoctor README.adoc

man:
	asciidoctor -b manpage grepp.adoc

open:
	open README.html

build:
	go build $(BUILD_FLAGS)

install:
	go install $(BUILD_FLAGS) grepp.go

rpm:
	rpmbuild -bb rpm.spec \
		--define '_rpmdir ./RPMS' \
		--define '_sourcedir ${PWD}' \
		--buildroot ${PWD}/buildroot
