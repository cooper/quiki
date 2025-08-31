# quiki test runner
.PHONY: test test-verbose test-markdown test-basic clean

# build and run all tests
test: build
	cd test && go run main.go

# run tests with verbose output
test-verbose: build
	cd test && go run main.go -v

# run specific test suite
test-markdown: build
	cd test && go run main.go -suite suites/markdown.json

test-basic: build
	cd test && go run main.go -suite suites/basic.json

test-colors: build
	cd test && go run main.go -suite suites/colors.json

test-escaping: build
	cd test && go run main.go -suite suites/escaping.json

# run tests matching a filter
test-filter: build
	cd test && go run main.go -filter "$(FILTER)"

# build quiki binary
build:
	go build -o quiki

# clean up
clean:
	rm -f quiki
