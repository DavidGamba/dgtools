.PHONY: test

ROOT_DIR = $(shell pwd)
dirs = $(shell find * -type d)

test:
	@for dir in $(dirs); do \
		cd $${dir}; \
		if [ `find . -maxdepth 1 -type f -name '*_test.go' | wc -l` -gt 0 ]; then \
			go test -cover ./... || exit 1; \
		fi; \
		cd ${ROOT_DIR}; \
	done

build:
	@for dir in $(shell ffind main.go | xargs -I{} dirname {}); do \
		cd $${dir}; \
		pwd; \
		go build; \
		cd ${ROOT_DIR}; \
	done
