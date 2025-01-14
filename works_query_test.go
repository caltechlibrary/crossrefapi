package crossrefapi

import (
	"net/url"
	"strings"
	"testing"
)

func TestWorksFilter_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		filter  WorksFilter
		want    string
		wantErr bool
	}{
		{
			name:    "Empty filter",
			filter:  WorksFilter{},
			want:    "",
			wantErr: false,
		},
		{
			name: "Single field",
			filter: WorksFilter{
				DOI: []string{"10.1234/test"},
			},
			want:    "doi:10.1234/test",
			wantErr: false,
		},
		{
			name: "Multiple fields",
			filter: WorksFilter{
				DOI:  []string{"10.1234/test"},
				ISSN: []string{"1234-5678"},
				Type: []string{"journal-article"},
			},
			want:    "doi:10.1234/test,issn:1234-5678,type:journal-article",
			wantErr: false,
		},
		{
			name: "With date parameter",
			filter: WorksFilter{
				FromIssuedDate: []DateParameter{{
					Year:  2020,
					Month: 6,
				}},
			},
			want:    "from-issued-date:2020-6",
			wantErr: false,
		},
		{
			name: "With bool parameter",
			filter: WorksFilter{
				HasAbstract: (*BoolParameter)(boolPtr(true)),
				HasLicense:  (*BoolParameter)(boolPtr(false)),
			},
			want:    "has-abstract:1,has-license:0",
			wantErr: false,
		},
		{
			name: "With nested filter",
			filter: WorksFilter{
				License: &LicenseFilter{
					URL:     []string{"http://example.com"},
					Version: []string{"1.0"},
				},
			},
			want:    "license.url:http://example.com,license.version:1.0",
			wantErr: false,
		},
		{
			name: "Multiple values for same field",
			filter: WorksFilter{
				DOI:  []string{"10.1234/test1", "10.1234/test2"},
				Type: []string{"journal-article", "book-chapter"},
			},
			want:    "doi:10.1234/test1,doi:10.1234/test2,type:journal-article,type:book-chapter",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Encode()
			if string(got) != tt.want {
				t.Errorf("WorksFilter.MarshalText() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestWorksQuery_Encode(t *testing.T) {
	tests := []struct {
		name    string
		query   WorksQuery
		want    url.Values
		wantErr bool
	}{
		{
			name:    "Empty query",
			query:   WorksQuery{},
			want:    url.Values{},
			wantErr: false,
		},
		{
			name: "Free form query",
			query: WorksQuery{
				FreeFormQuery: "test query",
			},
			want: url.Values{
				"query": []string{"test query"},
			},
			wantErr: false,
		},
		{
			name: "Fields query",
			query: WorksQuery{
				Fields: &WorksQueryFields{
					Author: "John Doe",
				},
			},
			want: url.Values{
				"query.author": []string{"John Doe"},
			},
			wantErr: false,
		},
		{
			name: "Pagination query",
			query: WorksQuery{
				Pagination: &Pagination{
					Rows: 10,
				},
			},
			want: url.Values{
				"rows": []string{"10"},
			},
			wantErr: false,
		},
		{
			name: "Filters query",
			query: WorksQuery{
				Filters: &WorksFilter{
					DOI: []string{"10.1234/example.doi"},
				},
			},
			want: url.Values{
				"filter": []string{"doi:10.1234/example.doi"},
			},
			wantErr: false,
		},
		{
			name: "Multiple query parameters and filters",
			query: WorksQuery{
				FreeFormQuery: "test query",
				Fields: &WorksQueryFields{
					Author: "John Doe",
				},
				Pagination: &Pagination{
					Rows: 10,
				},
				Filters: &WorksFilter{
					DOI:    []string{"10.1234/example.doi"},
					ISSN:   []string{"1234-5678"},
					ORCID:  []string{"0000-0001-2345-6789"},
					Prefix: []string{"10.1234"},
				},
			},
			want: url.Values{
				"query":        []string{"test query"},
				"query.author": []string{"John Doe"},
				"rows":         []string{"10"},
				"filter":       []string{"doi:10.1234/example.doi,issn:1234-5678,orcid:0000-0001-2345-6789,prefix:10.1234"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.query.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("WorksQuery.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equalValues(got, tt.want) {
				t.Errorf("WorksQuery.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateParameter_String(t *testing.T) {
	tests := []struct {
		name string
		date DateParameter
		want string
	}{
		{
			name: "Year only",
			date: DateParameter{Year: 2023},
			want: "2023",
		},
		{
			name: "Year and month",
			date: DateParameter{Year: 2023, Month: 10},
			want: "2023-10",
		},
		{
			name: "Full date",
			date: DateParameter{Year: 2023, Month: 10, Day: 15},
			want: "2023-10-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.String(); got != tt.want {
				t.Errorf("DateParameter.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateParameter_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		date    DateParameter
		want    string
		wantErr bool
	}{
		{
			name:    "Year only",
			date:    DateParameter{Year: 2023},
			want:    "2023",
			wantErr: false,
		},
		{
			name:    "Year and month",
			date:    DateParameter{Year: 2023, Month: 10},
			want:    "2023-10",
			wantErr: false,
		},
		{
			name:    "Full date",
			date:    DateParameter{Year: 2023, Month: 10, Day: 15},
			want:    "2023-10-15",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.date.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("DateParameter.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("DateParameter.MarshalText() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestBoolParameter_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		param   BoolParameter
		want    string
		wantErr bool
	}{
		{
			name:    "True value",
			param:   BoolParameter(true),
			want:    "1",
			wantErr: false,
		},
		{
			name:    "False value",
			param:   BoolParameter(false),
			want:    "0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.param.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("BoolParameter.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("BoolParameter.MarshalText() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestWorksQuery_RealAPICall(t *testing.T) {
	client, err := NewCrossRefClient("crossrefapi_test.go", "ls.duchemin@gmail.com")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	query := WorksQuery{
		Fields: &WorksQueryFields{
			Author: "Einstein",
			Title:  "relativity",
		},
		Pagination: &Pagination{
			Rows: 5,
		},
		Filters: &WorksFilter{
			HasAbstract: (*BoolParameter)(boolPtr(true)),
			FromIssuedDate: []DateParameter{{
				Year: 1905,
			}},
			UntilIssuedDate: []DateParameter{{
				Year: 1955,
			}},
		},
	}

	works, err := client.QueryWorks(query)
	if err != nil || works == nil {
		t.Fatalf("WorksQuery real API call failed: %v", err)
	}

	if len(works.Message.Items) == 0 {
		t.Error("Expected to get some works but got none")
	}

	if len(works.Message.Items) > 5 {
		t.Errorf("Expected max 5 results but got %d", len(works.Message.Items))
	}

	// Verify at least one work matches our criteria
	found := false
	for _, work := range works.Message.Items {
		for _, author := range work.Author {
			if strings.Contains(strings.ToLower(author.Given)+" "+strings.ToLower(author.Family), "einstein") {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Error("Expected to find at least one work by Einstein")
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func equalValues(a, b url.Values) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if len(v) != len(b[k]) {
			return false
		}
		for i, vv := range v {
			if vv != b[k][i] {
				return false
			}
		}
	}
	return true
}
