#
# Makefile for crossrefapi, a Go package and command line tool.
#
# Everyday targets:   build, test, test-live, install, uninstall, clean
# Release targets:    generate, release, dist (see "Releasing" below)
#
# A plain `make`, `make build`, `make test` and `make install` need only git,
# make, go and pandoc. `cmt` is run only by `make generate` and by
# `make release`, never by the everyday targets, so a fresh clone builds with
# the committed version.go, CITATION.cff and installers.
#
PROJECT = crossrefapi

GIT_GROUP = caltechlibrary

PROGRAMS = $(shell ls -1 cmd)

# version.go is left out: it has a rule (cmt, under generate), and listing it
# here would make every build run cmt through that rule.
PACKAGE = $(shell ls -1 *.go | grep -v 'version.go')

MAN_PAGES = $(shell ls -1 docs/*.1.md 2>/dev/null | sed -E 's|^docs/||; s/\.1.md/.1/g')

VERSION = $(shell grep '"version":' codemeta.json | cut -d\"  -f 4)

RELEASE_HASH = $(shell git log --pretty=format:'%h' -n 1)

BRANCH = $(shell git branch | grep '* ' | cut -d\  -f 2)

OS = $(shell uname)

PREFIX = $(HOME)

ifneq ($(prefix),)
	 PREFIX = $(prefix)
endif

EXT =
ifeq ($(OS), Windows)
	 EXT = .exe
endif

# One entry per release platform: name:GOOS:GOARCH
PLATFORMS = Linux-x86_64:linux:amd64 Linux-aarch64:linux:arm64 \
	 macOS-x86_64:darwin:amd64 macOS-arm64:darwin:arm64 \
	 Windows-x86_64:windows:amd64 Windows-arm64:windows:arm64

build: $(PROGRAMS) man

# Building a program also writes its manual page (the program's own -help
# output) into docs/, which is where the documentation site is built from.
$(PROGRAMS): $(PACKAGE)
	 @mkdir -p bin
	 go build -o "bin/$@$(EXT)" ./cmd/$@
	 @mkdir -p docs
	 @./bin/$@$(EXT) -help >docs/$@.1.md

# man renders each docs/NAME.1.md into man/man1/NAME.1.
man: $(MAN_PAGES)

$(MAN_PAGES): .FORCE
	 @mkdir -p man/man1
	 pandoc docs/$@.md --from markdown --to man -s >man/man1/$@

# test runs offline: vet, the Go tests (the two that call the live CrossRef API
# skip themselves unless -mailto is given) and the release script's tests.
test: .FORCE
	 go vet ./...
	 go test ./...
	 ./release_test.bash

# test-live also calls the live CrossRef API, which asks callers to identify
# themselves. Run: make test-live MAILTO=you@example.edu
test-live: .FORCE
	 @if [ -z "$(MAILTO)" ]; then echo 'usage: make test-live MAILTO=you@example.edu' >&2; exit 2; fi
	 go test . -mailto "$(MAILTO)"

# check vets the code and lists, without changing them, files gofmt would
# reformat. Existing files are not reformatted as a side effect of other work.
check: .FORCE
	 go vet ./...
	 @UNFORMATTED="$$(gofmt -l . )"; if [ -n "$$UNFORMATTED" ]; then echo "not gofmt-formatted (left as is):"; echo "$$UNFORMATTED"; fi

# NOTE: on macOS you must use "mv" instead of "cp" to avoid problems
install: build .FORCE
	 @if [ ! -d $(PREFIX)/bin ]; then mkdir -p $(PREFIX)/bin; fi
	 @echo "Installing programs in $(PREFIX)/bin"
	 @for FNAME in $(PROGRAMS); do if [ -f ./bin/$$FNAME ]; then mv -v ./bin/$$FNAME $(PREFIX)/bin/$$FNAME; fi; done
	 @echo ""
	 @echo "Make sure $(PREFIX)/bin is in your PATH"
	 @echo ""
	 @if [ ! -d $(PREFIX)/man/man1 ]; then mkdir -p $(PREFIX)/man/man1; fi
	 @for MAN_PAGE in $(MAN_PAGES); do cp -v man/man1/$$MAN_PAGE $(PREFIX)/man/man1/;done
	 @echo ""
	 @echo "Make sure $(PREFIX)/man is in your MANPATH"
	 @echo ""

uninstall: .FORCE
	 @echo "Removing programs in $(PREFIX)/bin"
	 -for FNAME in $(PROGRAMS); do if [ -f $(PREFIX)/bin/$$FNAME ]; then rm -v $(PREFIX)/bin/$$FNAME; fi; done
	 -for MAN_PAGE in $(MAN_PAGES); do if [ -f "$(PREFIX)/man/man1/$$MAN_PAGE" ]; then rm "$(PREFIX)/man/man1/$$MAN_PAGE"; fi; done

hash: .FORCE
	 @git log --pretty=format:'%h' -n 1

clean: .FORCE
	 -if [ -d bin ]; then rm -fR bin; fi
	 -if [ -d dist ]; then rm -fR dist; fi
	 -if [ -d man ]; then rm -fR man; fi
	 -if [ -d testout ]; then rm -fR testout; fi

status:
	 git status

save:
	 @git commit -am "$${msg:-Quick Save}"
	 git push origin $(BRANCH)

refresh:
	 git fetch origin
	 git pull origin $(BRANCH)

#
# Releasing
#
# 1. Edit codemeta.json by hand: version, dateModified, releaseNotes.
# 2. make release      regenerates version.go, CITATION.cff and both installers
#                      from codemeta.json with cmt (the only place cmt runs
#                      besides an explicit `make generate`), rebuilds the
#                      program and its manual page from that version.go, runs
#                      the tests, builds one zip per platform in dist/ and
#                      runs release-check, which stops here if anything does
#                      not carry the version and the hash of the commit.
# 3. ./release.bash    refuses if anything beyond those files is uncommitted,
#                      asks for confirmation, commits and pushes the regenerated
#                      files ("prep for vX.Y.Z"), and creates the DRAFT GitHub
#                      release from dist/*.zip. Publish the draft from the
#                      repository's releases page (or with gh release edit).
#

# generate runs cmt over codemeta.json. Run it when the version or the
# metadata change; a plain `make` leaves the committed copies alone.
generate: version.go CITATION.cff installer.sh installer.ps1

version.go: .FORCE
	 cmt codemeta.json version.go

CITATION.cff: .FORCE
	 cmt codemeta.json CITATION.cff

installer.sh: .FORCE
	 cmt codemeta.json installer.sh

installer.ps1: .FORCE
	 cmt codemeta.json installer.ps1

# dist builds one zip per platform in PLATFORMS, starting from an empty dist/.
dist: build .FORCE
	 @rm -fR dist
	 @mkdir -p dist
	 @for ENTRY in $(PLATFORMS); do \
	 	 NAME=$$(echo $$ENTRY | cut -d: -f1); GOOS=$$(echo $$ENTRY | cut -d: -f2); GOARCH=$$(echo $$ENTRY | cut -d: -f3); \
	 	 SUFFIX=""; if [ "$$GOOS" = "windows" ]; then SUFFIX=".exe"; fi; \
	 	 rm -fR dist/stage; mkdir -p dist/stage/bin; \
	 	 for FNAME in $(PROGRAMS); do \
	 	 	 env GOOS=$$GOOS GOARCH=$$GOARCH go build -o "dist/stage/bin/$${FNAME}$$SUFFIX" ./cmd/$$FNAME || exit 1; \
	 	 done; \
	 	 cp LICENSE codemeta.json CITATION.cff README.md INSTALL.md dist/stage/; \
	 	 cp -R docs man dist/stage/; \
	 	 (cd dist/stage && zip -qr ../$(PROJECT)-v$(VERSION)-$$NAME.zip .) || exit 1; \
	 	 echo "dist/$(PROJECT)-v$(VERSION)-$$NAME.zip"; \
	 done
	 @rm -fR dist/stage

release-version: .FORCE
	 @$(MAKE) -B generate

# release-check stops the release unless version.go, the installers,
# CITATION.cff, the manual pages in docs/ and the zip files all carry the
# version in codemeta.json and the hash of the commit being released.
release-check: .FORCE
	 @HASH=$$(git log --pretty=format:%h -n 1); \
	 if ! grep -q 'Version = "$(VERSION)"' version.go; then \
	 	 echo "error: version.go does not carry version $(VERSION)" >&2; exit 1; fi; \
	 if ! grep -q "ReleaseHash = \"$$HASH\"" version.go; then \
	 	 echo "error: version.go does not carry hash $$HASH" >&2; exit 1; fi; \
	 if ! grep -q '^version: $(VERSION)' CITATION.cff; then \
	 	 echo "error: CITATION.cff does not carry version $(VERSION)" >&2; exit 1; fi; \
	 for FNAME in installer.sh installer.ps1; do \
	 	 if ! grep -q 'VERSION.*"$(VERSION)"' $$FNAME; then \
	 	 	 echo "error: $$FNAME does not carry version $(VERSION)" >&2; exit 1; fi; \
	 done; \
	 for FNAME in docs/*.?.md; do \
	 	 if ! head -4 $$FNAME | grep -q "version $(VERSION) $$HASH"; then \
	 	 	 echo "error: $$FNAME does not carry version $(VERSION) $$HASH" >&2; exit 1; fi; \
	 done; \
	 COUNT=0; \
	 for ENTRY in $(PLATFORMS); do \
	 	 NAME=$$(echo $$ENTRY | cut -d: -f1); ZIP=dist/$(PROJECT)-v$(VERSION)-$$NAME.zip; \
	 	 if [ ! -f $$ZIP ]; then echo "error: missing $$ZIP" >&2; exit 1; fi; \
	 	 for PAGE in $$(unzip -Z1 $$ZIP 'docs/*.?.md'); do \
	 	 	 if ! unzip -p $$ZIP $$PAGE | head -4 | grep -q "version $(VERSION) $$HASH"; then \
	 	 	 	 echo "error: $$ZIP $$PAGE does not carry version $(VERSION) $$HASH" >&2; exit 1; fi; \
	 	 done; \
	 	 COUNT=$$((COUNT + 1)); \
	 done; \
	 echo "release-check: version.go, CITATION.cff, installers, docs/ and $$COUNT zip files carry $(VERSION) $$HASH"

# release: see "Releasing" above. It commits nothing; release.bash does that.
release: clean release-version test dist release-check
	@printf "\nrelease-check passed. Now run: ./release.bash   (or ./release.bash --dry-run first)\n\n"

.FORCE:
