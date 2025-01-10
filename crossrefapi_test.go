package crossrefapi

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
)

var (
	MailTo string
)

func TestClient(t *testing.T) {
	api, err := NewCrossRefClient("crossrefapi_test.go", MailTo)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if api.RateLimitLimit != 0 {
		t.Errorf("expected 0, got %d", api.RateLimitLimit)
	}
	if api.RateLimitInterval != 0 {
		t.Errorf("expected 0, got %d", api.RateLimitInterval)
	}
	if fmt.Sprintf("%s", api.LastRequest) == "0001-01-01 00:00:00 +0000" {
		t.Errorf("expected 0001-01-01 00:00:00 +0000, got %s", api.LastRequest)
	}

	// test low level getJSON
	src, err := api.getJSON("/types", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if api.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d -> %s", api.StatusCode, api.Status)
	}
	if len(src) == 0 {
		t.Errorf("expected a response body from /types")
	}
	if api.RateLimitLimit == 0 {
		t.Errorf("expected greater than zero, got %d", api.RateLimitLimit)
	}
	if api.RateLimitInterval == 0 {
		t.Errorf("expected greater than zero, got %d", api.RateLimitInterval)
	}
	if fmt.Sprintf("%s", api.LastRequest) == "0001-01-01 00:00:00 +0000" {
		t.Errorf("expected not equal to 0001-01-01 00:00:00 +0000, got %s", api.LastRequest)
	}

	// test Types API end point
	src, err = api.TypesJSON()
	if err != nil {
		t.Errorf("expected a JSON response, got %s", err)
		t.FailNow()
	}
	if api.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d -> %q", api.StatusCode, api.Status)
		t.FailNow()
	}
	obj1 := make(Object)
	err = json.Unmarshal(src, &obj1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if obj1 == nil {
		t.Errorf("expected unmarshaled object, got nil")
		t.FailNow()
	}

	obj2, err := api.Types()
	if err != nil {
		t.Errorf("expected an Object from Types(), got error %s", err)
		t.FailNow()
	}
	if obj2 == nil {
		t.Errorf("expected an non-nil Object from Types(), got nil but no error")
		t.FailNow()
	}

	if len(obj1) != len(obj2) {
		t.Errorf("expected equal lengths for obj1, obj2 ->\n%+v, \n%+v", obj1, obj2)
		t.FailNow()
	}

	// Confirm these valid doi can be retrieved and unpacked.
	doiList := []string{
		"10.1037/0003-066x.59.1.29",
		"10.7554/elife.81398",
		"10.1101/564955",
	}
	for i, doi := range doiList {
		src, err = api.WorksJSON(doi)
		if err != nil {
			t.Errorf("expected a JSON response (%d, %q), got %s", i, doi, err)
			continue
		}
		if api.StatusCode != 200 {
			t.Errorf("expected status code 200 (%d, %q), got %d -> %q", i, doi, api.StatusCode, api.Status)
		}
		// Make sure we can unmashal this
		w := new(Works)
		err = json.Unmarshal(src, &w)
		if err != nil {
			t.Errorf("failed to unmarshal (%d, %q) -> %s", i, doi, err)
			t.FailNow()
		}
	}

	// Now test Works
	doi := "10.1037/0003-066x.59.1.29"                        //"10.1000/xyz123"
	doiURL := "https://dx.doi.org/10.1037/0003-066x.59.1.29" // "https://dx.doi.org/10.1000/xyz123"

	src, err = api.WorksJSON(doi)
	if err != nil {
		t.Errorf("expected a JSON response, got %s", err)
		t.FailNow()
	}
	if api.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d -> %q", api.StatusCode, api.Status)
		t.FailNow()
	}

	work1 := new(Works)
	err = json.Unmarshal(src, &work1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if work1 == nil {
		t.Errorf("expected unmarshaled object, got nil")
		t.FailNow()
	}
	work2, err := api.Works(doiURL)
	if work2 == nil {
		t.Errorf("expected an non-nil Object from Types(), got nil but no error")
		t.FailNow()
	}
	if !work1.Message.IsSame(work2.Message) {
		src, _ := work1.Message.DiffAsJSON(work2.Message)
		t.Errorf("expected work.Message 1 & 2 don't match ->\n%s", src)
	}
	if !work1.IsSame(work2) {
		src, _ := work1.DiffAsJSON(work2)
		t.Errorf("expected work 1 & 2 don't match\n%s", src)
		t.FailNow()
	}


	// This DOI has an article number, 032435
	doiURL = "https://doi.org/10.1103/PhysRevA.105.032435"
	works3, err := api.Works(doiURL)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if works3.Message == nil {
		t.Errorf("Expected a message attribute, got %+v\n", works3)
		t.FailNow()
	}
	if works3.Message.ArticleNumber != "032435" {
		t.Errorf("Expected article number 032435, got %q", works3.Message.ArticleNumber)
		src, _ := json.MarshalIndent(works3.Message, "", "    ") // DEBUG
		fmt.Fprintf(os.Stderr, "DEBUG works\n%s\n", src) // DEBUG
		t.FailNow()
	}


}

func TestMain(m *testing.M) {
	flag.StringVar(&MailTo, "mailto", "", "set the mailto for testing")
	flag.Parse()
	log.Printf("mailto: %q", MailTo)
	os.Exit(m.Run())
}


