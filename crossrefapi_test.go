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
	src, err := api.getJSON("/types")
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

	// Now test Works
	doi := "10.1037/0003-066x.59.1.29"                        //"10.1000/xyz123"
	doi_url := "https://dx.doi.org/10.1037/0003-066x.59.1.29" // "https://dx.doi.org/10.1000/xyz123"

	src, err = api.WorksJSON(doi)
	if err != nil {
		t.Errorf("expected a JSON response, got %s", err)
		t.FailNow()
	}
	if api.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d -> %q", api.StatusCode, api.Status)
		t.FailNow()
	}
	obj1 = nil
	err = json.Unmarshal(src, &obj1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if obj1 == nil {
		t.Errorf("expected unmarshaled object, got nil")
		t.FailNow()
	}
	obj2, err = api.Works(doi_url)
	if obj2 == nil {
		t.Errorf("expected an non-nil Object from Types(), got nil but no error")
		t.FailNow()
	}
	if len(obj1) != len(obj2) {
		t.Errorf("expected equal lengths for obj1, obj2 ->\n%+v, \n%+v", obj1, obj2)
		t.FailNow()
	}
}

func TestMain(m *testing.M) {
	flag.StringVar(&MailTo, "mailto", "", "set the mailto for testing")
	flag.Parse()
	log.Printf("mailto: %q", MailTo)
	os.Exit(m.Run())
}
