package crossrefapi

import (
	"encoding/json"
	"os"
	"path"
	"testing"
)

func TestIdentifier(t *testing.T) {
	src := []byte(`[
  {
    "type": "print",
    "value": "0004-637X"
  },
  {
    "type": "electronic",
    "value": "1538-4357"
  }
]`)
	idList := []*Identifier{}
	if err := json.Unmarshal(src, &idList); err != nil {
		t.Error(err)
		t.FailNow()
	}
	expectedList := []*Identifier{
		&Identifier{
			Type:  "print",
			Value: "0004-637X",
		},
		&Identifier{
			Type:  "electronic",
			Value: "1538-4357",
		},
	}
	if len(expectedList) != len(idList) {
		t.Errorf("expected %+v, got %+v", expectedList, idList)
		t.FailNow()
	}
	for i, expected := range expectedList {
		if expected.Type != idList[i].Type {
			t.Errorf("expected (%d) type %q, got %q", i, expected.Type, idList[i].Type)
		}
		if len(expected.Value) != len(idList[i].Value) {
			t.Errorf("expected (%d) value %q, got %q", i, expected.Value, idList[i].Value)
		} else {
			if expected.Value != idList[i].Value {
				t.Errorf("expected (%d) value %q, got %q", i, expected.Value, idList[i].Value)
			}
		}
	}
}

func TestReferences(t *testing.T) {
	fName := path.Join("testdata", "acaf5-references.json")
	src, err := os.ReadFile(fName)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	references := []*Reference{}
	if err := json.Unmarshal(src, &references); err != nil {
		t.Error(err)
		t.FailNow()
	}
	expectedReferences := []*Reference{
		&Reference{
			DOI:           "10.1046/j.1365-8711.1999.02978.x",
			Author:        "Arnouts",
			DoiAssertedBy: "publisher",
			FirstPage:     "540",
			JournalTitle:  "MNRAS",
			Key:           "apjacacf5bib1",
			Volume:        "310",
			Year:          "1999",
		},
		&Reference{
			Author:       "Bolzonella",
			FirstPage:    "476",
			JournalTitle: "A&A",
			Key:          "apjacacf5bib2",
			Volume:       "363",
			Year:         "2000",
		},
		&Reference{
			DOI:           "10.1023/A:1010933404324",
			Author:        "Breiman",
			DoiAssertedBy: "publisher",
			FirstPage:     "5",
			JournalTitle:  "Mach. Learn.",
			Key:           "apjacacf5bib3",
			Volume:        "45",
			Year:          "2001",
		},
		&Reference{
			DOI:           "10.1046/j.1365-8711.2003.06897.x",
			Author:        "Bruzual",
			DoiAssertedBy: "publisher",
			FirstPage:     "1000",
			JournalTitle:  "MNRAS",
			Key:           "apjacacf5bib4",
			Volume:        "344",
			Year:          "2003",
		},
		&Reference{
			DOI:           "10.1086/308692",
			Author:        "Calzetti",
			DoiAssertedBy: "publisher",
			FirstPage:     "682",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib5",
			Volume:        "533",
			Year:          "2000",
		},
		&Reference{
			DOI:           "10.1093/mnras/stt2456",
			Author:        "Carrasco Kind",
			DoiAssertedBy: "publisher",
			FirstPage:     "3409",
			JournalTitle:  "MNRAS",
			Key:           "apjacacf5bib6",
			Volume:        "438",
			Year:          "2014",
		},
		&Reference{
			DOI:           "10.1086/376392",
			Author:        "Chabrier",
			DoiAssertedBy: "publisher",
			FirstPage:     "763",
			JournalTitle:  "PASP",
			Key:           "apjacacf5bib7",
			Volume:        "115",
			Year:          "2003",
		},
		&Reference{
			DOI:           "10.1093/mnras/stz2486",
			Author:        "Davidzon",
			DoiAssertedBy: "publisher",
			FirstPage:     "4817",
			JournalTitle:  "MNRAS",
			Key:           "apjacacf5bib8",
			Volume:        "489",
			Year:          "2019",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/201936695",
			Author:        "Dobbels",
			DoiAssertedBy: "publisher",
			FirstPage:     "A57",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib9",
			Volume:        "634",
			Year:          "2020",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/202039403",
			Author:        "Desprez",
			DoiAssertedBy: "publisher",
			FirstPage:     "A31",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib10",
			Volume:        "644",
			Year:          "2020",
		},
		&Reference{
			DOI:           "10.3847/2041-8213/ab3418",
			Author:        "Hemmati",
			DoiAssertedBy: "publisher",
			FirstPage:     "L14",
			JournalTitle:  "ApJL",
			Key:           "apjacacf5bib11",
			Volume:        "881",
			Year:          "2019",
		},
		&Reference{
			Author: "Hoaglin",
			Key:    "apjacacf5bib12",
			Year:   "1983",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/201425176",
			Author:        "Ilbert",
			DoiAssertedBy: "publisher",
			FirstPage:     "A2",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib13",
			Volume:        "579",
			Year:          "2015",
		},
		&Reference{
			DOI:           "10.1051/0004-6361:20065138",
			Author:        "Ilbert",
			DoiAssertedBy: "publisher",
			FirstPage:     "841",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib14",
			Volume:        "457",
			Year:          "2006",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/201321100",
			Author:        "Ilbert",
			DoiAssertedBy: "publisher",
			FirstPage:     "A55",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib15",
			Volume:        "556",
			Year:          "2013",
		},
		&Reference{
			DOI:           "10.1007/BF00337288",
			Author:        "Kohonen",
			DoiAssertedBy: "publisher",
			FirstPage:     "59",
			JournalTitle:  "Biol. Cybern.",
			Key:           "apjacacf5bib16",
			Volume:        "43",
			Year:          "1982",
		},
		&Reference{
			DOI:           "10.1103/PhysRevE.69.066138",
			Author:        "Kraskov",
			DoiAssertedBy: "publisher",
			JournalTitle:  "PhRvE",
			Key:           "apjacacf5bib17",
			Volume:        "69",
			Year:          "2004",
		},
		&Reference{
			DOI:           "10.3847/0067-0049/224/2/24",
			Author:        "Laigle",
			DoiAssertedBy: "publisher",
			FirstPage:     "24",
			JournalTitle:  "ApJS",
			Key:           "apjacacf5bib18",
			Volume:        "224",
			Year:          "2016",
		},
		&Reference{
			Author: "Laureijs",
			Key:    "apjacacf5bib19",
			Year:   "2011",
		},
		&Reference{
			DOI:           "10.1088/0004-637X/813/1/53",
			Author:        "Masters",
			DoiAssertedBy: "publisher",
			FirstPage:     "53",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib20",
			Volume:        "813",
			Year:          "2015",
		},
		&Reference{
			DOI:           "10.3847/1538-4357/aa6f08",
			Author:        "Masters",
			DoiAssertedBy: "publisher",
			FirstPage:     "111",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib21",
			Volume:        "841",
			Year:          "2017",
		},
		&Reference{
			DOI:           "10.3847/1538-4357/ab184d",
			Author:        "Masters",
			DoiAssertedBy: "publisher",
			FirstPage:     "81",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib22",
			Volume:        "877",
			Year:          "2019",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/201219507",
			Author:        "McCracken",
			DoiAssertedBy: "publisher",
			FirstPage:     "A156",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib23",
			Volume:        "544",
			Year:          "2012",
		},
		&Reference{
			Author: "McInnes",
			Key:    "apjacacf5bib24",
			Year:   "2018",
		},
		&Reference{
			DOI:           "10.1051/0004-6361/202142361",
			Author:        "Moneti",
			DoiAssertedBy: "publisher",
			FirstPage:     "A126",
			JournalTitle:  "A&A",
			Key:           "apjacacf5bib25",
			Volume:        "658",
			Year:          "2022",
		},
		&Reference{
			DOI:           "10.1093/mnras/stab164",
			Author:        "Mucesh",
			DoiAssertedBy: "publisher",
			FirstPage:     "2770",
			JournalTitle:  "MNRAS",
			Key:           "apjacacf5bib26",
			Volume:        "502",
			Year:          "2021",
		},
		&Reference{
			Author:       "Pedregosa",
			FirstPage:    "2825",
			JournalTitle: "J. Mach. Learn. Res.",
			Key:          "apjacacf5bib27",
			Volume:       "12",
			Year:         "2011",
		},
		&Reference{
			DOI:           "10.1088/0004-637X/737/2/103",
			Author:        "Schlafly",
			DoiAssertedBy: "publisher",
			FirstPage:     "103",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib28",
			Volume:        "737",
			Year:          "2011",
		},
		&Reference{
			DOI:           "10.1086/516585",
			Author:        "Scoville",
			DoiAssertedBy: "publisher",
			FirstPage:     "1",
			JournalTitle:  "ApJS",
			Key:           "apjacacf5bib29",
			Volume:        "172",
			Year:          "2007",
		},
		&Reference{
			DOI:           "10.1002/j.1538-7305.1948.tb01338.x",
			Author:        "Shannon",
			DoiAssertedBy: "publisher",
			FirstPage:     "379",
			JournalTitle:  "BSTJ",
			Key:           "apjacacf5bib30",
			Volume:        "27",
			Year:          "1948",
		},
		&Reference{
			DOI:           "10.3847/1538-4357/abd179",
			Author:        "Simet",
			DoiAssertedBy: "publisher",
			FirstPage:     "47",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib31",
			Volume:        "908",
			Year:          "2021",
		},
		&Reference{
			DOI:           "10.3847/1538-4357/ab76be",
			Author:        "Steinhardt",
			DoiAssertedBy: "publisher",
			FirstPage:     "136",
			JournalTitle:  "ApJ",
			Key:           "apjacacf5bib32",
			Volume:        "891",
			Year:          "2020",
		},
		&Reference{
			Author:       "van der Maaten",
			FirstPage:    "2579",
			JournalTitle: "J. Mach. Learn. Res.",
			Key:          "apjacacf5bib33",
			Volume:       "9",
			Year:         "2008",
		},
		&Reference{
			DOI:           "10.3847/1538-4365/ac3078",
			Author:        "Weaver",
			DoiAssertedBy: "publisher",
			FirstPage:     "11",
			JournalTitle:  "ApJS",
			Key:           "apjacacf5bib34",
			Volume:        "258",
			Year:          "2022",
		},
	}

	if len(expectedReferences) != len(references) {
		t.Errorf("expected references not the same length as references")
		t.FailNow()
	}
	for i, expected := range expectedReferences {
		reference := references[i]
		if !expected.IsSame(reference) {
			t.Errorf("expected (%d) reference\n%+v, got\n%+v", i, expected, reference)
			t.FailNow()
		}
	}
}

func TestWorks(t *testing.T) {
	fName := path.Join("testdata", "acaf5.json")
	src, err := os.ReadFile(fName)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	work := &Works{}
	if err := json.Unmarshal(src, &work); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if work.Message == nil {
		t.Errorf("Missing message in work %q", fName)
		t.FailNow()
	}
	expected := "10.3847/1538-4357/acacf5"
	if work.Message.DOI != expected {
		t.Errorf("expected DOI %q, got %q", expected, work.Message.DOI)
	}
	if len(work.Message.ISSN) != 2 {
		t.Errorf("expected two ISSN in .message.ISSN, got %d", len(work.Message.ISSN))
	} else {
		for i, expected := range []string{"0004-637X", "1538-4357"} {
			if work.Message.ISSN[i] != expected {
				t.Errorf("expected ISSN[%d] %q, got %q", i, expected, work.Message.ISSN[i])
			}
		}
	}
	expected = "http://dx.doi.org/10.3847/1538-4357/acacf5"
	if work.Message.URL != expected {
		t.Errorf("expected URL %q, got %q", expected, work.Message.URL)
	}
	expected = `<jats:title>Abstract</jats:title><jats:p>We present a new method based on information theory to find the optimal number of bands required to measure the physical properties of galaxies with desired accuracy. As a proof of concept, using the recently updated COSMOS catalog (COSMOS2020), we identify the most relevant wave bands for measuring the physical properties of galaxies in a Hawaii Two-0- (H20) and UVISTA-like survey for a sample of <jats:italic>i</jats:italic> &lt; 25 AB mag galaxies. We find that with the available <jats:italic>i</jats:italic>-band fluxes, <jats:italic>r</jats:italic>, <jats:italic>u</jats:italic>, IRAC/<jats:italic>ch</jats:italic>2, and <jats:italic>z</jats:italic> bands provide most of the information regarding the redshift with importance decreasing from <jats:italic>r</jats:italic> band to <jats:italic>z</jats:italic> band. We also find that for the same sample, IRAC/<jats:italic>ch</jats:italic>2, <jats:italic>Y</jats:italic>, <jats:italic>r</jats:italic>, and <jats:italic>u</jats:italic> bands are the most relevant bands in stellar-mass measurements with decreasing order of importance. Investigating the intercorrelation between the bands, we train a model to predict UVISTA observations in near-IR from H20-like observations. We find that magnitudes in the <jats:italic>YJH</jats:italic> bands can be simulated/predicted with an accuracy of 1<jats:italic>σ</jats:italic> mag scatter ≲0.2 for galaxies brighter than 24 AB mag in near-IR bands. One should note that these conclusions depend on the selection criteria of the sample. For any new sample of galaxies with a different selection, these results should be remeasured. Our results suggest that in the presence of a limited number of bands, a machine-learning model trained over the population of observed galaxies with extensive spectral coverage outperforms template fitting. Such a machine-learning model maximally comprises the information acquired over available extensive surveys and breaks degeneracies in the parameter space of template fitting inevitable in the presence of a few bands.</jats:p>`
	if work.Message.Abstract != expected {
		t.Errorf("expected Abstract to match\n%q\n%q", expected, work.Message.Abstract)
	}

	expectedAssertion := []*Assertion{
		&Assertion{
			Label: "Article Title",
			Name:  "article_title",
			Value: "A Machine-learning Approach to Predict Missing Flux Densities in Multiband Galaxy Surveys",
		},
		&Assertion{
			Label: "Journal Title",
			Name:  "journal_title",
			Value: "The Astrophysical Journal",
		},
		&Assertion{
			Label: "Article Type",
			Name:  "article_type",
			Value: "paper",
		},
		&Assertion{
			Label: "Copyright Information",
			Name:  "copyright_information",
			Value: "© 2023. The Author(s). Published by the American Astronomical Society.",
		},
		&Assertion{
			Group: &Group{
				Label: "Publication dates",
				Name:  "publication_dates",
			},
			Label: "Date Received",
			Name:  "date_received",
			Value: "2022-01-27",
		},
		{
			Group: &Group{
				Label: "Publication dates",
				Name:  "publication_dates",
			},
			Label: "Date Accepted",
			Name:  "date_accepted",
			Value: "2022-08-29",
		},
		{
			Group: &Group{
				Label: "Publication dates",
				Name:  "publication_dates",
			},
			Label: "Online publication date",
			Name:  "date_epub",
			Value: "2023-01-17",
		},
	}

	if len(work.Message.Assertion) != 7 {
		t.Errorf("expected 7 assertions, got %d -> %+v", len(work.Message.Assertion), work.Message.Assertion)
	} else {
		for i, expectedA := range expectedAssertion {
			asserted := work.Message.Assertion[i]
			if expectedA.Group != nil {
				if expectedA.Group.Name != asserted.Group.Name {
					t.Errorf("expected group name %q, got %q", expectedA.Group.Name, asserted.Group.Name)
				}
				if expectedA.Group.Label != asserted.Group.Label {
					t.Errorf("expected group label %q, got %q", expectedA.Group.Label, asserted.Group.Label)
				}
			}
		}
	}

	if len(work.Message.Author) != 24 {
		t.Errorf("expected 24 authors, got %d -> %+v", len(work.Message.Author), work.Message.Author)
	} else {
		// FIXME: Just checking the first five authors should check them all
		expectedAuthors := []*Person{
			&Person{
				ORCID:              "http://orcid.org/0000-0003-3691-937X",
				Affiliation:        []*Organization{},
				AuthenticatedOrcid: true,
				Family:             "Chartab",
				Given:              "Nima",
				Sequence:           "first",
			},
			&Person{
				Affiliation: []*Organization{},
				Family:      "Mobasher",
				Given:       "Bahram",
				Sequence:    "additional",
			},
			&Person{
				Affiliation: []*Organization{},
				Family:      "Cooray",
				Given:       "Asantha R.",
				Sequence:    "additional",
			},
			&Person{
				ORCID:              "http://orcid.org/0000-0003-2226-5395",
				Affiliation:        []*Organization{},
				AuthenticatedOrcid: true,
				Family:             "Hemmati",
				Given:              "Shoubaneh",
				Sequence:           "additional",
			},
		}
		for i, expectedA := range expectedAuthors {
			author := work.Message.Author[i]
			if expectedA.ORCID != author.ORCID {
				t.Errorf("expected (%d) ORCID %q, got %q -> %+v", i, expectedA.ORCID, author.ORCID, author)
			} else {
				if expectedA.ORCID != "" && author.AuthenticatedOrcid != true {
					t.Errorf("expected (%d) authenticated orcid for %q", i, author.ORCID)
				}
			}
			if expectedA.Family != author.Family {
				t.Errorf("expected (%d) family %q, got %q -> %+v", i, expectedA.Family, author.Family, author)
			}
			if expectedA.Given != author.Given {
				t.Errorf("expected (%d) given %q, got %q -> %+v", i, expectedA.Given, author.Given, author)
			}
			if expectedA.Sequence != author.Sequence {
				t.Errorf("expected (%d) sequence %q, got %q -> %+v", i, expectedA.Sequence, author.Sequence, author)
			}
		}
	}
	if len(work.Message.ContainerTitle) != 1 {
		t.Errorf("expected one item in container title, got none")
	} else {
		if work.Message.ContainerTitle[0] != "The Astrophysical Journal" {
			t.Errorf(`expected container title 0 to be "The Astrophysical Journal" got, %s`, work.Message.ContainerTitle[0])
		}
	}
	expectedDC := &ContentDomain{
		CrossmarkRestriction: false,
		Domain: []string{
			"iopscience.iop.org",
		},
	}

	if work.Message.ContentDomain == nil {
		t.Errorf("expected a content domain, got none")
	} else {
		if expectedDC.CrossmarkRestriction != work.Message.ContentDomain.CrossmarkRestriction ||
			len(work.Message.ContentDomain.Domain) != 1 ||
			work.Message.ContentDomain.Domain[0] != expectedDC.Domain[0] {
			t.Errorf("expected content domain %+v, got %+v", expectedDC, work.Message.ContentDomain)
		}
	}
	expectedDo := &DateObject{
		DateParts: [][]int{
			[]int{
				2023,
				1,
				18,
			},
		},
		DateTime:  "2023-01-18T12:18:32Z",
		Timestamp: 1674044312000,
	}
	if !expectedDo.IsSame(work.Message.Created) {
		t.Errorf("expected created %+v, got %+v", expectedDo, work.Message.Created)
	}
	expectedDo.DateTime = "2023-01-18T12:18:36Z"
	expectedDo.Timestamp = 1674044316000
	if !expectedDo.IsSame(work.Message.Deposited) {
		t.Errorf("expected deposited %+v, got %+v", expectedDo, work.Message.Created)
	}
	expectedDo.DateParts[0][2] = 19
	expectedDo.DateTime = "2023-01-19T06:15:40Z"
	expectedDo.Timestamp = 1674108940059
	if !expectedDo.IsSame(work.Message.Indexed) {
		t.Errorf("expected indexed %+v, got %+v", expectedDo, work.Message.Created)
	}
}
