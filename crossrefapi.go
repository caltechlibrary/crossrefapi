package crossrefapi


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/doitools"
)


type CrossRefClient struct {
	AppName           string
	MailTo            string `json:"mailto"`
	API               string `json:"api"`
	RateLimitLimit    int    `json:"limit"`
	RateLimitInterval int    `json:"interval"`
	LimitCount        int    `json:"limit_count"`
	Status            string
	StatusCode        int
	LastRequest       time.Time `json:"last_request"`
}

// Object is the general holder of what get back after unmarshaling json
type Object = map[string]interface{}

// Custom JSON decoder so we can treat numbers easier
func JsonDecode(src []byte, obj interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	err := dec.Decode(&obj)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

// MarshalObject provide a custom json encoder to solve a an issue with
// HTML entities getting converted to UTF-8 code points by json.Marshal()
// in recent versions of go (~= go1.21).
func MarshalObject(obj interface{}, prefix string, indent string) ([]byte, error) {
	buf := []byte{}
	w := bytes.NewBuffer(buf)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), err
}

// NewCrossRefClient creates a client and makes a request
// and returns the JSON source as a []byte or error if their is
// a problem.
func NewCrossRefClient(appName string, mailTo string) (*CrossRefClient, error) {
	if strings.TrimSpace(mailTo) == "" {
		return nil, fmt.Errorf("An mailto value is required for politeness")
	}
	client := new(CrossRefClient)
	client.AppName = appName
	client.API = `https://api.crossref.org`
	client.MailTo = mailTo
	return client, nil
}

func (c *CrossRefClient) calcDelay() time.Duration {
	if c.RateLimitLimit == 0 {
		return time.Duration(0)
	}
	return time.Duration(int64(math.Ceil(float64(c.RateLimitInterval) / float64(c.RateLimitLimit))))
}

// getJSON retrieves the path from the CrossRef API maintaining politeness.
// It returns a []byte of JSON source or an error
func (c *CrossRefClient) getJSON(p string) ([]byte, error) {
	var src []byte

	u, err := url.Parse(c.API)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("mailto", c.MailTo)
	u.RawQuery = q.Encode()
	u.Path = p

	client := http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", fmt.Sprintf("%s, based on crossrefapi/%s (github.com/caltechlibrary/crossrefapi/; mailto: %s), A golang cli based on https://github.com/CrossRef/rest-api-doc", c.AppName, Version, c.MailTo))

	// NOTE: Next request can be made based on last request time plus
	// the duration suggested by X-Rate-Limit-Interval / X-Rate-Limit-Limit
	if delay := c.calcDelay(); delay.Seconds() > 0 {
		time.Sleep(delay)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// Save the response status
	c.Status = res.Status
	c.StatusCode = res.StatusCode
	// Process the body buffer
	if c.StatusCode == 200 {
		src, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}

	// NOTE: we want to track the current values for any limits
	// `X-Rate-Limit-Limit` and `X-Rate-Limit-Interval` as well
	// as LastRequest time
	if s := res.Header.Get("X-Rate-Limit-Limit"); s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			c.RateLimitLimit = i
		}
	} else if c.RateLimitLimit == 0 {
		c.RateLimitLimit = 1
	}
	if s := res.Header.Get("X-Rate-Limit-Interval"); s != "" {
		if i, err := strconv.Atoi(strings.TrimSuffix(s, "s")); err == nil {
			c.RateLimitInterval = i
		}
	} else if c.RateLimitInterval == 0 {
		c.RateLimitInterval = 1
	}
	c.LastRequest = time.Now()
	return src, nil
}

// TypesJSON return a list of types in JSON source
func (c *CrossRefClient) TypesJSON() ([]byte, error) {
	return c.getJSON("types")
}

// Types returns the list of supported types as a Object
func (c *CrossRefClient) Types() (Object, error) {
	src, err := c.TypesJSON()
	if err != nil {
		return nil, err
	}
	object := make(Object)
	err = JsonDecode(src, &object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

// WorksJSON return the work JSON source or error for a client and DOI
func (c *CrossRefClient) WorksJSON(doi string) ([]byte, error) {
	s, err := doitools.NormalizeDOI(doi)
	if err != nil {
		return nil, err
	}
	return c.getJSON(path.Join("works", s))
}

// Works return the Work unmarshaled into a Object (i.e. map[string]interface{})
func (c *CrossRefClient) Works(doi string) (*Works, error) {
	src, err := c.WorksJSON(doi)
	if err != nil {
		return nil, err
	}
	if len(src) > 0 {
		work := &Works{}
		err = JsonDecode(src, &work)
		if err != nil {
			return nil, err
		}
		return work, nil
	}
	return nil, nil
}
