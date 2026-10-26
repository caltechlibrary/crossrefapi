#!/bin/bash
#
# Tests for release.bash. Each case builds a throwaway git repository with a
# bare remote, a codemeta.json, a dist/ holding fake zips and a stub `gh` on
# PATH that records its arguments, so nothing here can reach GitHub.
#
# Run: ./release_test.bash        Needs: bash, git, jq.
#
set -u

HERE="$(cd "$(dirname "$0")" && pwd)"
PASS=0
FAIL=0
ROOT="$(mktemp -d)"
trap 'rm -rf "$ROOT"' EXIT

# new_repo makes $ROOT/<name>/work with a clean tree pushed to a bare remote,
# version 1.2.3, one fake zip and a stub gh. It prints the work directory.
new_repo() {
	local name="$1" base="$ROOT/$1"
	mkdir -p "$base/bin"
	git init -q --bare "$base/remote.git"
	git clone -q "$base/remote.git" "$base/work" 2>/dev/null
	(
		cd "$base/work"
		git config user.email test@example.edu
		git config user.name Test
		jq -n '{name:"demo", version:"1.2.3", releaseNotes:"Notes with `backticks` and $dollars"}' >codemeta.json
		printf '/dist/\n' >.gitignore
		cp "$HERE/release.bash" .
		git add . && git commit -q -m init && git push -q -u origin HEAD 2>/dev/null
		mkdir dist
		printf 'zip-a' >dist/demo-v1.2.3-Linux-x86_64.zip
		printf 'zip-b' >dist/demo-v1.2.3-macOS-arm64.zip
	)
	cat >"$base/bin/gh" <<'STUB'
#!/bin/bash
# Stub gh: log the arguments, and the notes file's content for release create.
LOG="$(dirname "$0")/gh.log"
echo "gh $*" >>"$LOG"
if [ "$1 $2" = "release view" ]; then
	[ "${GH_RELEASE_EXISTS:-}" = "1" ] && exit 0
	exit 1
fi
if [ "$1 $2" = "release create" ]; then
	[ "${GH_FAIL:-}" = "1" ] && exit 1
	while [ $# -gt 0 ]; do
		if [ "$1" = "--notes-file" ] || [ "$1" = "-F" ]; then echo "NOTES: $(cat "$2")" >>"$LOG"; fi
		shift
	done
fi
exit 0
STUB
	chmod +x "$base/bin/gh"
	echo "$base/work"
}

# run_release DIR ARGS... runs release.bash in DIR with the stub gh first on
# PATH, answering any prompt with "n", and sets $CODE and $OUT.
run_release() {
	local dir="$1"
	shift
	OUT="$(cd "$dir" && PATH="$(dirname "$dir")/bin:$PATH" bash ./release.bash "$@" </dev/null 2>&1)"
	CODE=$?
}

check() { # check NAME EXPECTED_CODE [PATTERN]
	local name="$1" want="$2" pat="${3:-}"
	if [ "$CODE" != "$want" ]; then
		echo "FAIL $name: exit $CODE, want $want"
		echo "$OUT" | sed 's/^/     | /'
		FAIL=$((FAIL + 1))
	elif [ -n "$pat" ] && ! printf '%s' "$OUT" | grep -q -- "$pat"; then
		echo "FAIL $name: output does not contain '$pat'"
		echo "$OUT" | sed 's/^/     | /'
		FAIL=$((FAIL + 1))
	else
		echo "ok   $name"
		PASS=$((PASS + 1))
	fi
}

gh_log() { cat "$(dirname "$1")/bin/gh.log" 2>/dev/null; }

# --- usage: exit 2 --------------------------------------------------------
W="$(new_repo usage)"
run_release "$W" --bogus;           check "bad flag"        2 "usage"
run_release "$W" --dry-run extra;   check "surplus argument" 2 "usage"
run_release "$W" --help;            check "--help"          0 "EXIT STATUS"

# --- data and input: 65, 66 -----------------------------------------------
W="$(new_repo badtag)"
(cd "$W" && jq '.version="1.2.3; rm -rf /"' codemeta.json >c && mv c codemeta.json && git commit -q -am bad && git push -q)
run_release "$W" --dry-run;         check "bad tag characters" 65 "version"

W="$(new_repo nover)"
(cd "$W" && jq 'del(.version)' codemeta.json >c && mv c codemeta.json && git commit -q -am nover && git push -q)
run_release "$W" --dry-run;         check "no version"      65 "no version"

W="$(new_repo nocodemeta)"
(cd "$W" && git rm -q codemeta.json && git commit -q -m x && git push -q)
run_release "$W" --dry-run;         check "no codemeta.json" 66 "codemeta.json"

W="$(new_repo nodist)"
rm -f "$W"/dist/*.zip
run_release "$W" --dry-run;         check "no dist zips"    66 "dist"

W="$(new_repo nogh)"
rm -f "$(dirname "$W")/bin/gh"
OUT="$(cd "$W" && PATH="/usr/bin:/bin" bash ./release.bash --dry-run </dev/null 2>&1)"; CODE=$?
check "gh missing" 66 "gh"

# --- state forbids: 1 -----------------------------------------------------
W="$(new_repo dirty)"
(cd "$W" && echo '# change' >>.gitignore)
run_release "$W" --dry-run;         check "dirty tree"      1 "uncommitted"

W="$(new_repo unpushed)"
(cd "$W" && echo x >f && git add f && git commit -q -m local)
run_release "$W" --dry-run;         check "unpushed commit" 1 "not pushed"

W="$(new_repo exists)"
GH_RELEASE_EXISTS=1 run_release "$W" --dry-run; check "release already exists" 1 "already exists"

W="$(new_repo tagelsewhere)"
(cd "$W" && git tag v1.2.3 && echo y >g && git add g && git commit -q -m next && git push -q)
run_release "$W" --dry-run;         check "tag points elsewhere" 1 "tag"

# --- dry run: 0, creates nothing ------------------------------------------
W="$(new_repo dry)"
run_release "$W" --dry-run;         check "dry run succeeds" 0 "v1.2.3"
if gh_log "$W" | grep -q "release create"; then echo "FAIL dry run called gh release create"; FAIL=$((FAIL + 1)); else echo "ok   dry run creates no release"; PASS=$((PASS + 1)); fi
if ls "$W"/dist/*checksums* >/dev/null 2>&1; then echo "FAIL dry run wrote a checksums file"; FAIL=$((FAIL + 1)); else echo "ok   dry run writes no checksums"; PASS=$((PASS + 1)); fi

# --- prompt declined: 1, creates nothing ----------------------------------
W="$(new_repo declined)"
run_release "$W";                   check "prompt declined" 1 "not pushed to GitHub\|cancel\|declin"
if gh_log "$W" | grep -q "release create"; then echo "FAIL declined run called gh release create"; FAIL=$((FAIL + 1)); else echo "ok   declined run creates no release"; PASS=$((PASS + 1)); fi

# --- real run with --yes --------------------------------------------------
W="$(new_repo real)"
run_release "$W" --yes;             check "release with --yes" 0 "draft"
LOG="$(gh_log "$W")"
for want in "release create v1.2.3" "--draft" "demo-v1.2.3-Linux-x86_64.zip" "demo-v1.2.3-macOS-arm64.zip" "demo-v1.2.3-checksums.txt"; do
	if printf '%s' "$LOG" | grep -q -- "$want"; then echo "ok   gh create carries $want"; PASS=$((PASS + 1)); else echo "FAIL gh create missing $want"; echo "$LOG" | sed 's/^/     | /'; FAIL=$((FAIL + 1)); fi
done
if printf '%s' "$LOG" | grep -q 'NOTES: Notes with `backticks` and $dollars'; then echo "ok   release notes pass through unmangled"; PASS=$((PASS + 1)); else echo "FAIL release notes mangled"; echo "$LOG" | sed 's/^/     | /'; FAIL=$((FAIL + 1)); fi
CS="$W/dist/demo-v1.2.3-checksums.txt"
if [ -f "$CS" ] && grep -q "demo-v1.2.3-Linux-x86_64.zip" "$CS" && ! grep -q "dist/" "$CS"; then echo "ok   checksums file lists zips by bare name"; PASS=$((PASS + 1)); else echo "FAIL checksums file wrong"; FAIL=$((FAIL + 1)); fi

# --- gh failure: 69 -------------------------------------------------------
W="$(new_repo ghfail)"
GH_FAIL=1 run_release "$W" --yes;   check "gh release create fails" 69 "gh"

echo
echo "$PASS passed, $FAIL failed"
[ "$FAIL" = 0 ]
