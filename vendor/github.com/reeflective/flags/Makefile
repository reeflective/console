.PHONY: all test test_v generate lint vet fmt coverage check check-fast prepare race

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
PKGSDIRS=$(shell find -L . -type f -name "*.go")

all: prepare

travis: info vet lint check test_v coverage

coverage:
	@echo "$(OK_COLOR)Generate coverage$(NO_COLOR)"
	@./scripts/cover_multi.sh

# Golangci-lint generates too many errors and some of them
# are plain wrong (code detected as unused while being used)
# prepare: generate fmt vet lint check test race
prepare: generate test race

test_v:
	@echo "$(OK_COLOR)Test packages$(NO_COLOR)"
	@go test -cover -v ./...

test:
	@echo "$(OK_COLOR)Test packages$(NO_COLOR)"
	@go test -cover ./...

# Lint throws a ton of errors. I tend to lint my code a lot though.
lint:
	@echo "$(OK_COLOR)Run lint$(NO_COLOR)"
	@test -z "$$(golint -min_confidence 0.3 ./... | tee /dev/stderr)"

check:
	@echo "$(OK_COLOR)Run golangci-lint$(NO_COLOR)"
	@golangci-lint run --no-config --exclude-use-default=true --max-same-issues=10 --disable=gosimple --disable=golint --enable=megacheck --enable=interfacer  --enable=goconst --enable=misspell --enable=unparam --enable=goimports --disable=errcheck --disable=ineffassign  --disable=gocyclo --disable=gas

# Vet throws a ton of errors
vet:
	@echo "$(OK_COLOR)Run vet$(NO_COLOR)"
	@go vet ./...

race:
	# Don't run race in subdirectories or in gen/flags.
	# There are data races which will always fail, related to cobra/pflag.
	# Maybe we should take them into account. I don't know.
	@echo "$(OK_COLOR)Test for races$(NO_COLOR)"
	@go test -race .

fmt:
	@echo "$(OK_COLOR)Formatting$(NO_COLOR)"
	@echo $(PKGSDIRS) | xargs -I '{p}' -n1 goimports -w {p}

info:
	depscheck -totalonly -tests .
	golocc --no-vendor ./...

generate:
	@echo "$(OK_COLOR)Go generate$(NO_COLOR)"
	@go generate

# Tools don't install some stuff. Dependency checks are done with other workflows.
# tools:
# 	@echo "$(OK_COLOR)Install tools$(NO_COLOR)"
# 	go get -u github.com/warmans/golocc
# 	go get -u github.com/divan/depscheck
# 	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
# 	cd ${GOPATH}/src/github.com/golangci/golangci-lint/cmd/golangci-lint \
#    && go install -ldflags "-X 'main.version=$(git describe --tags)' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date)'"
