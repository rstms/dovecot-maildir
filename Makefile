# go makefile

program != basename $$(pwd)

latest_release != gh release list --json tagName --jq '.[0].tagName' | tr -d v

version != cat VERSION

gitclean = if git status --porcelain | grep '^.*$$'; then echo git status is dirty; false; else echo git status is clean; true; fi

build: fmt 
	fix go build

fmt: go.sum
	fix go fmt . ./...

go.mod:
	go mod init

go.sum: go.mod
	go mod tidy

install: build
	go install

test: fmt build
	go test -failfast -v . ./...

debug: fmt
	go test -v -failfast -count=1 -run $(test) . ./...

release:
	$(gitclean)
	echo gh release create v$(version) --notes "v$(version)"

clean: 
	rm -f $(program)
	go clean

sterile: clean
	go clean -r || true
	go clean -cache
	go clean -modcache
	rm -f go.mod go.sum
	rm -f README.md
