#
# Simple Makefile for conviently testing, building and deploying experiment.
#
PROJECT = crossrefapi

VERSION = $(shell grep '"version":' codemeta.json | cut -d \" -f 4)

BRANCH = $(shell git branch | grep '* ' | cut -d\  -f 2)

MAN_PAGES = $(shell ls -1 *.1.md | sed -E 's/\.1.md/.1/g')

PANDOC = $(shell which pandoc)

PROJECT_LIST = crossrefapi

OS = $(shell uname)

EXT =
ifeq ($(OS), Windows)
	EXT = .exe
endif

build: version.go man $(PROJECT_LIST) CITATION.cff about.md

version.go: .FORCE
	@echo	"package	$(PROJECT)"	>version.go
	@echo	''	>>version.go
	@echo	'const	('	>>version.go
	@echo	'	Version	=	"$(VERSION)"'	>>version.go
	@echo	''	>>version.go
	@echo	'	LicenseText	=	`'	>>version.go
	@cat	LICENSE	>>version.go
	@echo	'`'	>>version.go
	@echo	')'	>>version.go
	@echo	''	>>version.go
	@git	add	version.go

crossrefapi$(EXT): bin/crossrefapi$(EXT) CITATION.cff

bin/crossrefapi$(EXT): crossrefapi.go works.go cmd/crossrefapi/crossrefapi.go
	go build -o bin/crossrefapi$(EXT) cmd/crossrefapi/crossrefapi.go
	./bin/crossrefapi -help>crossrefapi.1.md

man: $(MAN_PAGES)

$(MAN_PAGES): .FORCE
	mkdir -p man/man1
	$(PANDOC) $@.md --from markdown --to man -s >man/man1/$@

install:
	env GOBIN=$(GOPATH)/bin go install cmd/crossrefapi/crossrefapi.go

website: page.tmpl codemeta.json README.md INSTALL.md LICENSE css/site.css about.md
	make -f website.mak


CITATION.cff: codemeta.json .FORCE
	@cat	codemeta.json	|	sed	-E	's/"@context"/"at__context"/g;s/"@type"/"at__type"/g;s/"@id"/"at__id"/g'	>_codemeta.json
	@if	[	-f	$(PANDOC)	];	then	echo	""	|	$(PANDOC)	--metadata	title="Cite	$(PROJECT)"	--metadata-file=_codemeta.json	--template=codemeta-cff.tmpl	>CITATION.cff;	fi

about.md: codemeta.json
	@cat	codemeta.json	|	sed	-E	's/"@context"/"at__context"/g;s/"@type"/"at__type"/g;s/"@id"/"at__id"/g'	>_codemeta.json
	@if	[	-f	$(PANDOC)	];	then	echo	""	|	pandoc	--metadata-file=_codemeta.json	--template	codemeta-md.tmpl	>about.md	2>/dev/null;	fi
	@if	[	-f	_codemeta.json	];	then	rm	_codemeta.json;	fi


test: clean bin/crossrefapi$(EXT)
	go test -mailto="jane.doe@example.edu"

format:
	gofmt -w crossrefapi.go
	gofmt -w crossrefapi_test.go
	gofmt -w cmd/crossrefapi/crossrefapi.go

lint:
	golint crossrefapi.go
	golint crossrefapi_test.go
	golint cmd/crossrefapi/crossrefapi.go

clean:
	if [ -f index.html ]; then rm *.html; fi
	if [ -d bin ]; then rm -fR bin; fi
	if [ -d dist ]; then rm -fR dist; fi
	if [ -d testout ]; then rm -fR testout; fi


dist/linux-amd64:
	mkdir -p dist/bin
	env  GOOS=linux GOARCH=amd64 go build -o dist/bin/crossrefapi cmd/crossrefapi/crossrefapi.go
	cd dist && zip -r $(PROJECT)-$(VERSION)-linux-amd64.zip README.md LICENSE INSTALL.md bin/* man/*
	rm -fR dist/bin

dist/windows-amd64:
	mkdir -p dist/bin
	env  GOOS=windows GOARCH=amd64 go build -o dist/bin/crossrefapi.exe cmd/crossrefapi/crossrefapi.go
	cd dist && zip -r $(PROJECT)-$(VERSION)-windows-amd64.zip README.md LICENSE INSTALL.md bin/* man/*
	rm -fR dist/bin

dist/macos-amd64:
	mkdir -p dist/bin
	env  GOOS=darwin GOARCH=amd64 go build -o dist/bin/crossrefapi cmd/crossrefapi/crossrefapi.go
	cd dist && zip -r $(PROJECT)-$(VERSION)-macos-amd64.zip README.md LICENSE INSTALL.md bin/* man/*
	rm -fR dist/bin

dist/macos-arm64:
	mkdir -p dist/bin
	env  GOOS=darwin GOARCH=arm64 go build -o dist/bin/crossrefapi cmd/crossrefapi/crossrefapi.go
	cd dist && zip -r $(PROJECT)-$(VERSION)-macos-arm64.zip README.md LICENSE INSTALL.md bin/* man/*
	rm -fR dist/bin

dist/raspbian-arm7:
	mkdir -p dist/bin
	env  GOOS=linux GOARCH=arm GOARM=7 go build -o dist/bin/crossrefapi cmd/crossrefapi/crossrefapi.go
	cd dist && zip -r $(PROJECT)-$(VERSION)-raspbian-arm7.zip README.md LICENSE INSTALL.md bin/* man/*
	rm -fR dist/bin

distribute_docs:
	if [ -d dist ]; then rm -fR dist; fi
	mkdir -p dist
	cp -v README.md dist/
	cp -v LICENSE dist/
	cp -v INSTALL.md dist/
	cp -vR man dist/

release: clean crossrefapi.go distribute_docs dist/linux-amd64 dist/windows-amd64 dist/macos-amd64 dist/macos-arm64 dist/raspbian-arm7

status:
	git status

save:
	if [ "$(msg)" != "" ]; then git commit -am "$(msg)"; else git commit -am "Quick Save"; fi
	git push origin $(BRANCH)

publish: website
	bash publish.bash


.FORCE:
