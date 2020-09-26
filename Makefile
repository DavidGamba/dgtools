.PHONY: test

ROOT_DIR = $(shell pwd)
dirs = $(shell find * -type d)

test:
	@for dir in $(dirs); do \
		cd $${dir}; \
		if [ `find . -maxdepth 1 -type f -name '*_test.go' | wc -l` -gt 0 ]; then \
			GO111MODULE=on go test -cover ./... || exit 1; \
		fi; \
		cd ${ROOT_DIR}; \
	done
