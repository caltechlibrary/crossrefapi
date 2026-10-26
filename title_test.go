package crossrefapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// fixtureMessage loads a captured CrossRef response into the package's Message.
func fixtureMessage(t *testing.T, doiFile string) *Message {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "crossref-"+doiFile+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Message Message `json:"message"`
	}
	if err := json.Unmarshal(src, &r); err != nil {
		t.Fatal(err)
	}
	return &r.Message
}

// TestFullTitle_Fixtures checks real captured records, expected strings written
// from decisions 1, C and D of crossrefapi/DR-0001 (title, ": ", subtitle; the
// subtitle skipped when empty or already in the title; the first title only).
func TestFullTitle_Fixtures(t *testing.T) {
	cases := []struct {
		file string
		want string
		warn []string // Fragment of each multiple-titles warning
	}{
		{"10.1051_0004-6361_202660273", "Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood: Parameter inference with a non-Gaussian likelihood in Fourier and configuration space", nil},
		{"10.1051_0004-6361_202659007", "A Virgo Environmental Survey Tracing Ionised Gas Emission (VESTIGE): XX. Star formation in the tidal tail of NGC 4254", nil},
		{"10.1007_978-3-319-20325-6", "Data Assimilation: A Mathematical Introduction", nil},
		{"10.1007_978-3-319-93031-2_4", "An Efficient Relaxed Projection Method for Constrained Non-negative Matrix Factorization with Application to the Phase-Mapping Problem in Materials Science", nil},
		{"10.1103_physrevb.85.144303", "Line of continuous phase transitions in a three-dimensional U(1) loop model with 1/r² current-current interactions", nil},
		// two titles: the first is used, the other is reported
		{"10.1007_s00717-026-00622-0", "X-chromosomale juvenile Retinoschisis", []string{"X-linked juvenile retinoschisis"}},
		// two titles and a subtitle that is the empty string
		{"10.24272_j.issn.2095-8137.2025.121", "巴马香猪与野猪肾脏发育的动态单核转录图谱", []string{"Dynamic single-nucleus transcriptomic profiling of kidney development in Bamaxiang pigs and wild boar"}},
		// the second title differs only in capitalization: nothing distinct to report
		{"10.67686_didaktika.v10i2.2338", "Pengaruh STEM-PJBL Berbasis Flipped Classroom Dalam Meningkatkan Kreativitas Dan Pemecahan Masalah", nil},
		// a journal issue has no title at all
		{"10.1111_jdv.v40.6", "", nil},
	}
	for _, c := range cases {
		got, ws := fixtureMessage(t, c.file).FullTitle()
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.file, got, c.want)
		}
		var frag []string
		for _, w := range ws {
			if w.Code != WarnMultipleTitles {
				t.Errorf("%s: unexpected warning %v", c.file, w)
				continue
			}
			frag = append(frag, w.Fragment)
		}
		if len(frag) != len(c.warn) || strings.Join(frag, "|") != strings.Join(c.warn, "|") {
			t.Errorf("%s: multiple-titles warnings %q, want %q", c.file, frag, c.warn)
		}
	}
}

// TestFullTitle_ContainedSubtitle checks the real book review, whose title
// already holds its subtitle (with a comma the subtitle has and the title lacks),
// is not doubled.
func TestFullTitle_ContainedSubtitle(t *testing.T) {
	msg := fixtureMessage(t, "10.1080_00253359.2024.2340221")
	got, ws := msg.FullTitle()
	want, _ := PlainText(msg.Title[0])
	if got != want || len(ws) != 0 {
		t.Errorf("got %q (warnings %v), want the title alone %q", got, ws, want)
	}
	if strings.Count(got, "by T. Michals") != 1 {
		t.Errorf("the subtitle appears %d times", strings.Count(got, "by T. Michals"))
	}
}

// TestFullTitle_Synthetic covers cases no fixture has.
func TestFullTitle_Synthetic(t *testing.T) {
	cases := []struct {
		name string
		msg  *Message
		want string
	}{
		{"nil message", nil, ""},
		{"empty message", &Message{}, ""},
		{"title only", &Message{Title: []string{"T"}}, "T"},
		{"title and subtitle", &Message{Title: []string{"T"}, SubTitle: []string{"S"}}, "T: S"},
		{"subtitle without a title is not a title", &Message{SubTitle: []string{"S"}}, ""},
		{"blank title leaves the subtitle", &Message{Title: []string{" "}, SubTitle: []string{"S"}}, "S"},
		{"empty subtitle", &Message{Title: []string{"T"}, SubTitle: []string{""}}, "T"},
		{"blank subtitle", &Message{Title: []string{"T"}, SubTitle: []string{" \n "}}, "T"},
		{"title ending in a question mark", &Message{Title: []string{"Reine Sprache?"}, SubTitle: []string{"Privatbriefe"}}, "Reine Sprache? Privatbriefe"},
		{"title ending in an exclamation mark", &Message{Title: []string{"Wow!"}, SubTitle: []string{"S"}}, "Wow! S"},
		{"title ending in a period", &Message{Title: []string{"Euclid preparation."}, SubTitle: []string{"CII"}}, "Euclid preparation. CII"},
		{"title ending in a colon", &Message{Title: []string{"T:"}, SubTitle: []string{"S"}}, "T: S"},
		{"title ending in a bracket", &Message{Title: []string{"T (A)"}, SubTitle: []string{"S"}}, "T (A): S"},
		{"subtitle equal to the title", &Message{Title: []string{"T and S"}, SubTitle: []string{"T and S"}}, "T and S"},
		{"subtitle inside the title, other case and punctuation", &Message{Title: []string{"Lame Admirals: index ISBN 12"}, SubTitle: []string{"ADMIRALS index, ISBN 12"}}, "Lame Admirals: index ISBN 12"},
		{"subtitle that is only part of a word is kept", &Message{Title: []string{"Party planning"}, SubTitle: []string{"art"}}, "Party planning: art"},
		{"numbers count as words when comparing", &Message{Title: []string{"Part 2 of 3"}, SubTitle: []string{"2 of 4"}}, "Part 2 of 3: 2 of 4"},
		{"a subtitle with the same numbers is contained", &Message{Title: []string{"Part 2 of 3"}, SubTitle: []string{"2 of 3"}}, "Part 2 of 3"},
		{"a subtitle of only punctuation is kept", &Message{Title: []string{"T"}, SubTitle: []string{"..."}}, "T: ..."},
		{"subtitle with markup", &Message{Title: []string{"T"}, SubTitle: []string{"<i>S</i> two"}}, "T: S two"},
		{"title with markup", &Message{Title: []string{"H<sub>2</sub>O"}, SubTitle: []string{"S"}}, "H₂O: S"},
		{"only the first subtitle is used", &Message{Title: []string{"T"}, SubTitle: []string{"S1", "S2"}}, "T: S1"},
	}
	for _, c := range cases {
		got, _ := c.msg.FullTitle()
		if got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
	// a warning from the title's own conversion comes through
	_, ws := (&Message{Title: []string{"x<sub>q</sub>"}, SubTitle: []string{"S"}}).FullTitle()
	if len(ws) != 1 || ws[0].Code != WarnScriptUnmapped {
		t.Errorf("title warnings: %+v", ws)
	}
	// and so does one from the subtitle's
	_, ws = (&Message{Title: []string{"T"}, SubTitle: []string{"x<sub>q</sub>"}}).FullTitle()
	if len(ws) != 1 || ws[0].Code != WarnScriptUnmapped {
		t.Errorf("subtitle warnings: %+v", ws)
	}
	// the multiple-titles warning joins every distinct alternate
	_, ws = (&Message{Title: []string{"A", "B", "C"}}).FullTitle()
	if len(ws) != 1 || ws[0].Code != WarnMultipleTitles || ws[0].Fragment != "B | C" {
		t.Errorf("multiple titles: %+v", ws)
	}
}

// TestAlternateTitles_Fixtures checks real records.
func TestAlternateTitles_Fixtures(t *testing.T) {
	cases := []struct {
		file string
		want []string
	}{
		{"10.24272_j.issn.2095-8137.2025.121", []string{"Dynamic single-nucleus transcriptomic profiling of kidney development in Bamaxiang pigs and wild boar"}},
		{"10.1007_s00717-026-00622-0", []string{"X-linked juvenile retinoschisis"}},
		{"10.67686_didaktika.v10i2.2338", nil},
		{"10.1007_978-3-319-20325-6", nil},
		{"10.1111_jdv.v40.6", nil},
	}
	for _, c := range cases {
		got, ws := fixtureMessage(t, c.file).AlternateTitles()
		if len(got) != len(c.want) || strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("%s: got %q, want %q", c.file, got, c.want)
		}
		if len(ws) != 0 {
			t.Errorf("%s: unexpected warnings %v", c.file, ws)
		}
	}
}

// TestAlternateTitles_Synthetic covers duplicates, markup and blanks.
func TestAlternateTitles_Synthetic(t *testing.T) {
	cases := []struct {
		name   string
		titles []string
		want   []string
	}{
		{"none", nil, nil},
		{"one title", []string{"A"}, nil},
		{"case-only duplicates of the first are dropped", []string{"A", "a", "A"}, nil},
		{"each distinct title once, in order", []string{"A", "B", "b", "C", "B"}, []string{"B", "C"}},
		{"markup is removed before comparing", []string{"A", "<i>B</i>", "B"}, []string{"B"}},
		{"blank titles are skipped", []string{"A", "", "  \n "}, nil},
		{"the first title may carry markup", []string{"<i>A</i>", "A", "B"}, []string{"B"}},
	}
	for _, c := range cases {
		got, ws := (&Message{Title: c.titles}).AlternateTitles()
		if len(got) != len(c.want) || strings.Join(got, "|") != strings.Join(c.want, "|") || len(ws) != 0 {
			t.Errorf("%s: got %q (warnings %v), want %q", c.name, got, ws, c.want)
		}
	}
	var nilMsg *Message
	if got, ws := nilMsg.AlternateTitles(); got != nil || ws != nil {
		t.Errorf("nil message: %q, %v", got, ws)
	}
	// a warning from an alternate's own conversion comes through
	got, ws := (&Message{Title: []string{"A", "x<sub>q</sub>"}}).AlternateTitles()
	if len(got) != 1 || len(ws) != 1 || ws[0].Code != WarnScriptUnmapped {
		t.Errorf("alternate warnings: %q, %+v", got, ws)
	}
}

// TestMessage_HTMLAbstract checks the method is the converter applied to the
// message's abstract, and is safe on nothing.
func TestMessage_HTMLAbstract(t *testing.T) {
	msg := fixtureMessage(t, "10.1002_2016GL070081")
	got, ws := msg.HTMLAbstract()
	want, wantWs := HTMLAbstract(msg.Abstract)
	if got != want || len(ws) != len(wantWs) || got == "" {
		t.Errorf("got %q, want %q", got, want)
	}
	var nilMsg *Message
	if got, ws := nilMsg.HTMLAbstract(); got != "" || ws != nil {
		t.Errorf("nil message: %q, %v", got, ws)
	}
	if got, ws := (&Message{}).HTMLAbstract(); got != "" || ws != nil {
		t.Errorf("empty message: %q, %v", got, ws)
	}
}

// writeSymbols writes an override file for a test.
func writeSymbols(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "symbols.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestNormalizer checks a Normalizer converts with its own table, leaves the
// package defaults alone, and works as a zero value.
func TestNormalizer(t *testing.T) {
	def, err := NewNormalizer("")
	if err != nil {
		t.Fatal(err)
	}
	over, err := NewNormalizer(writeSymbols(t, `{"entities": {"eacute": "E"}, "dropped_headings": []}`))
	if err != nil {
		t.Fatal(err)
	}
	msg := &Message{Title: []string{"caf&eacute;"}, SubTitle: []string{"S"}, Abstract: "<jats:title>Abstract</jats:title><jats:p>y</jats:p>"}

	// the override changes the title and keeps the Abstract heading it no longer drops
	if got, _ := over.FullTitle(msg); got != "cafE: S" {
		t.Errorf("override FullTitle = %q", got)
	}
	if got, _ := over.PlainText("caf&eacute;"); got != "cafE" {
		t.Errorf("override PlainText = %q", got)
	}
	if got, _ := over.AbstractHTML(msg); got != "<p><b>Abstract</b></p><p>y</p>" {
		t.Errorf("override AbstractHTML = %q", got)
	}
	if got, _ := over.HTMLAbstract("<jats:title>Abstract</jats:title><jats:p>y</jats:p>"); got != "<p><b>Abstract</b></p><p>y</p>" {
		t.Errorf("override HTMLAbstract = %q", got)
	}
	// the default normalizer, the zero value, a nil one and the package functions agree
	for name, n := range map[string]*Normalizer{"NewNormalizer(\"\")": def, "zero value": {}, "nil": nil} {
		if got, _ := n.FullTitle(msg); got != "café: S" {
			t.Errorf("%s FullTitle = %q", name, got)
		}
		if got, _ := n.AbstractHTML(msg); got != "<p>y</p>" {
			t.Errorf("%s AbstractHTML = %q", name, got)
		}
		if got, _ := n.HTMLAbstract("<jats:title>Abstract</jats:title><jats:p>y</jats:p>"); got != "<p>y</p>" {
			t.Errorf("%s HTMLAbstract = %q", name, got)
		}
		if got, _ := n.PlainText("caf&eacute;"); got != "café" {
			t.Errorf("%s PlainText = %q", name, got)
		}
	}
	if got, _ := msg.FullTitle(); got != "café: S" {
		t.Errorf("(*Message).FullTitle = %q after an override normalizer was built", got)
	}
	if got, _ := PlainText("caf&eacute;"); got != "café" {
		t.Errorf("PlainText = %q after an override normalizer was built", got)
	}
	if got, _ := over.AlternateTitles(&Message{Title: []string{"A", "caf&eacute;"}}); len(got) != 1 || got[0] != "cafE" {
		t.Errorf("override AlternateTitles = %q", got)
	}
}

// TestNormalizer_Errors checks a bad override gives an error and no Normalizer.
func TestNormalizer_Errors(t *testing.T) {
	n, err := NewNormalizer(filepath.Join(t.TempDir(), "absent.json"))
	if err == nil || n != nil {
		t.Errorf("missing file: %v, %v", n, err)
	}
	n, err = NewNormalizer(writeSymbols(t, `{"superscript": {}}`))
	if err == nil || n != nil {
		t.Errorf("unknown section: %v, %v", n, err)
	}
}

// TestNormalizer_Concurrent checks one Normalizer can be shared: run it with
// -race to catch any change to the table.
func TestNormalizer_Concurrent(t *testing.T) {
	n, err := NewNormalizer("")
	if err != nil {
		t.Fatal(err)
	}
	msg := fixtureMessage(t, "10.1007_s00717-026-00622-0")
	want, _ := n.FullTitle(msg)
	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if got, _ := n.FullTitle(msg); got != want {
					errs <- fmt.Sprintf("FullTitle = %q", got)
				}
				n.AbstractHTML(fixtureMessage(t, "10.1002_2016GL070081"))
				n.PlainText("H<sub>2</sub>O")
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
