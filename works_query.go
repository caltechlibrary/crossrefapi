package crossrefapi

import (
	"encoding"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

// WorksQueryFields represents the fields that can be queried in the CrossRef API
type WorksQueryFields struct {
	Affiliation          string `url:"query.affiliation,omitempty"`
	Author               string `url:"query.author,omitempty"`
	Bibliographic        string `url:"query.bibliographic,omitempty"`
	Chair                string `url:"query.chair,omitempty"`
	ContainerTitle       string `url:"query.container-title,omitempty"`
	Contributor          string `url:"query.contributor,omitempty"`
	Degree               string `url:"query.degree,omitempty"`
	Description          string `url:"query.description,omitempty"`
	Editor               string `url:"query.editor,omitempty"`
	EventAcronym         string `url:"query.event-acronym,omitempty"`
	EventLocation        string `url:"query.event-location,omitempty"`
	EventName            string `url:"query.event-name,omitempty"`
	EventTheme           string `url:"query.event-theme,omitempty"`
	FunderName           string `url:"query.funder-name,omitempty"`
	PublisherLocation    string `url:"query.publisher-location,omitempty"`
	PublisherName        string `url:"query.publisher-name,omitempty"`
	StandardsBodyAcronym string `url:"query.standards-body-acronym,omitempty"`
	StandardsBodyName    string `url:"query.standards-body-name,omitempty"`
	Title                string `url:"query.title,omitempty"`
	Translator           string `url:"query.translator,omitempty"`
}

type SortOrder string

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

type SortKey string

const (
	Created             SortKey = "created"
	Deposited           SortKey = "deposited"
	Indexed             SortKey = "indexed"
	IsReferencedByCount SortKey = "is-referenced-by-count"
	Issued              SortKey = "issued"
	Published           SortKey = "published"
	PublishedOnline     SortKey = "published-online"
	PublishedPrint      SortKey = "published-print"
	ReferencesCount     SortKey = "references-count"
	Relevance           SortKey = "relevance"
	Score               SortKey = "score"
	LastUpdate          SortKey = "updated" // renamed from `Updated` to avoid name collision
)

type QuerySortOptions struct {
	Key   SortKey   `url:"sort,omitempty"`
	Order SortOrder `url:"order,omitempty"` // default is "desc"
}

// License represents license-specific filter parameters
type LicenseFilter struct {
	URL     string `yaml:"url,omitempty"`
	Version string `yaml:"version,omitempty"`
	Delay   *int   `yaml:"delay,omitempty"`
}

// Relation represents relation-specific filter parameters
type RelationFilter struct {
	Type       string `yaml:"type,omitempty"`
	ObjectType string `yaml:"object-type,omitempty"`
	Object     string `yaml:"object,omitempty"`
}

type FullTextFilter struct {
	Type        string `yaml:"type,omitempty"`
	Application string `yaml:"application,omitempty"`
	Version     string `yaml:"version,omitempty"`
}

type AwardFilter struct {
	Funder string `yaml:"funder,omitempty"`
	Number *int   `yaml:"number,omitempty"`
}

type DateParameter struct {
	Year  int32
	Month int32
	Day   int32
}

func (d DateParameter) String() string {
	s := fmt.Sprintf("%d", d.Year)
	if d.Month <= 0 {
		return s
	}
	s = fmt.Sprintf("%s-%d", s, d.Month)
	if d.Day <= 0 {
		return s
	}
	s = fmt.Sprintf("%s-%d", s, d.Day)
	return s
}

func (d *DateParameter) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// BoolParameter overrides boolean yaml marshalling to comply with CrossRef API spec
type BoolParameter bool

func (b BoolParameter) MarshalText() ([]byte, error) {
	if b {
		return []byte("1"), nil
	} else {
		return []byte("0"), nil
	}
}

// WorksFilter represents the available filter parameters for the /works endpoint
type WorksFilter struct {
	AlternativeID  string `yaml:"alternative-id,omitempty"`
	Archive        string `yaml:"archive,omitempty"`
	ArticleNumber  string `yaml:"article-number,omitempty"`
	Assertion      string `yaml:"assertion,omitempty"`
	AssertionGroup string `yaml:"assertion-group,omitempty"`

	// Award related fields
	Award *AwardFilter `yaml:"award,omitempty"`

	CategoryName        string `yaml:"category-name,omitempty"`
	CitationID          string `yaml:"citation-id,omitempty"`
	ClinicalTrialNumber string `yaml:"clinical-trial-number,omitempty"`
	ContainerTitle      string `yaml:"container-title,omitempty"`
	ContentDomain       string `yaml:"content-domain,omitempty"`
	DOI                 string `yaml:"doi,omitempty"`

	// From date fields
	FromAcceptedDate   *DateParameter `yaml:"from-accepted-date,omitempty"`
	FromApprovedDate   *DateParameter `yaml:"from-approved-date,omitempty"`
	FromAwardedDate    *DateParameter `yaml:"from-awarded-date,omitempty"`
	FromCreatedDate    *DateParameter `yaml:"from-created-date,omitempty"`
	FromDepositDate    *DateParameter `yaml:"from-deposit-date,omitempty"`
	FromEventEndDate   *DateParameter `yaml:"from-event-end-date,omitempty"`
	FromEventStartDate *DateParameter `yaml:"from-event-start-date,omitempty"`
	FromIndexDate      *DateParameter `yaml:"from-index-date,omitempty"`
	FromIssuedDate     *DateParameter `yaml:"from-issued-date,omitempty"`
	FromOnlinePubDate  *DateParameter `yaml:"from-online-pub-date,omitempty"`
	FromPostedDate     *DateParameter `yaml:"from-posted-date,omitempty"`
	FromPrintPubDate   *DateParameter `yaml:"from-print-pub-date,omitempty"`
	FromPubDate        *DateParameter `yaml:"from-pub-date,omitempty"`
	FromUpdateDate     *DateParameter `yaml:"from-update-date,omitempty"`

	// Full text related fields
	FullText *FullTextFilter `yaml:"full-text,omitempty"`

	// Other fields
	Funder              string `yaml:"funder,omitempty"`
	FunderDoiAssertedBy string `yaml:"funder-doi-asserted-by,omitempty"`
	GroupTitle          string `yaml:"group-title,omitempty"`

	// Boolean flags
	HasAbstract            *BoolParameter `yaml:"has-abstract"`
	HasAffiliation         *BoolParameter `yaml:"has-affiliation"`
	HasArchive             *BoolParameter `yaml:"has-archive"`
	HasAssertion           *BoolParameter `yaml:"has-assertion"`
	HasAuthenticatedOrcid  *BoolParameter `yaml:"has-authenticated-orcid"`
	HasAward               *BoolParameter `yaml:"has-award"`
	HasClinicalTrialNumber *BoolParameter `yaml:"has-clinical-trial-number"`
	HasContentDomain       *BoolParameter `yaml:"has-content-domain"`
	HasDescription         *BoolParameter `yaml:"has-description"`
	HasDomainRestriction   *BoolParameter `yaml:"has-domain-restriction"`
	HasEvent               *BoolParameter `yaml:"has-event"`
	HasFullText            *BoolParameter `yaml:"has-full-text"`
	HasFunder              *BoolParameter `yaml:"has-funder"`
	HasFunderDoi           *BoolParameter `yaml:"has-funder-doi"`
	HasLicense             *BoolParameter `yaml:"has-license"`
	HasOrcid               *BoolParameter `yaml:"has-orcid"`
	HasReferences          *BoolParameter `yaml:"has-references"`
	HasRelation            *BoolParameter `yaml:"has-relation"`
	HasRorID               *BoolParameter `yaml:"has-ror-id"`
	HasUpdate              *BoolParameter `yaml:"has-update"`
	HasUpdatePolicy        *BoolParameter `yaml:"has-update-policy"`
	IsUpdate               *BoolParameter `yaml:"is-update"`

	// ISBN/ISSN
	ISBN string `yaml:"isbn,omitempty"`
	ISSN string `yaml:"issn,omitempty"`

	// License fields
	License *LicenseFilter `yaml:"license,omitempty"`

	// Award amount
	GteAwardAmount int `yaml:"gte-award-amount,omitempty"`
	LteAwardAmount int `yaml:"lte-award-amount,omitempty"`

	// Member and identifiers
	Member string `yaml:"member,omitempty"`
	ORCID  string `yaml:"orcid,omitempty"`
	Prefix string `yaml:"prefix,omitempty"`

	// Relation fields
	Relation *RelationFilter `yaml:"relation,omitempty"`

	// Type fields
	RorID    string `yaml:"ror-id,omitempty"`
	Type     string `yaml:"type,omitempty"`
	TypeName string `yaml:"type-name,omitempty"`

	// Until date fields
	UntilAcceptedDate   *DateParameter `yaml:"until-accepted-date,omitempty"`
	UntilApprovedDate   *DateParameter `yaml:"until-approved-date,omitempty"`
	UntilAwardedDate    *DateParameter `yaml:"until-awarded-date,omitempty"`
	UntilCreatedDate    *DateParameter `yaml:"until-created-date,omitempty"`
	UntilDepositDate    *DateParameter `yaml:"until-deposit-date,omitempty"`
	UntilEventEndDate   *DateParameter `yaml:"until-event-end-date,omitempty"`
	UntilEventStartDate *DateParameter `yaml:"until-event-start-date,omitempty"`
	UntilIndexDate      *DateParameter `yaml:"until-index-date,omitempty"`
	UntilIssuedDate     *DateParameter `yaml:"until-issued-date,omitempty"`
	UntilOnlinePubDate  *DateParameter `yaml:"until-online-pub-date,omitempty"`
	UntilPostedDate     *DateParameter `yaml:"until-posted-date,omitempty"`
	UntilPrintPubDate   *DateParameter `yaml:"until-print-pub-date,omitempty"`
	UntilPubDate        *DateParameter `yaml:"until-pub-date,omitempty"`
	UntilUpdateDate     *DateParameter `yaml:"until-update-date,omitempty"`

	// Update fields
	UpdateType string `yaml:"update-type,omitempty"`
	Updates    string `yaml:"updates,omitempty"`
}

func stringifyValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case *string:
		if t == nil {
			return ""
		}
		return *t
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", t)
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		if t == nil {
			return ""
		}
		return fmt.Sprintf("%d", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func marshalStruct(v interface{}, prefix string) []string {
	val := reflect.ValueOf(v)
	typ := val.Type()

	// If it's a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		typ = val.Type()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var result []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Get the yaml tag
		tag := field.Tag.Get("yaml")
		if tag == "" {
			tag = field.Name
		}
		tagParts := strings.Split(tag, ",")
		name := tagParts[0]
		omitempty := len(tagParts) > 1 && tagParts[1] == "omitempty"

		if field.Type.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}

		// Skip if field is zero value and omitempty is set
		if omitempty && (fieldVal.Kind() != reflect.Ptr && fieldVal.IsZero()) {
			continue
		}

		// Handle text marshaller interface
		if marshaller, ok := fieldVal.Interface().(encoding.TextMarshaler); ok && marshaller != nil {
			text, err := marshaller.MarshalText()
			if err == nil && len(text) > 0 {
				key := name
				if prefix != "" {
					key = prefix + "." + name
				}
				result = append(result, fmt.Sprintf("%s:%s", key, string(text)))
			}
			continue
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() && fieldVal.Elem().Kind() == reflect.Struct {
			newPrefix := name
			if prefix != "" {
				newPrefix = prefix + "." + name
			}
			result = append(result, marshalStruct(fieldVal.Interface(), newPrefix)...)
			continue
		}
		if fieldVal.Kind() == reflect.Struct {
			newPrefix := name
			if prefix != "" {
				newPrefix = prefix + "." + name
			}
			result = append(result, marshalStruct(fieldVal.Interface(), newPrefix)...)
			continue
		}

		// Handle regular fields
		if !fieldVal.IsZero() {
			key := name
			if prefix != "" {
				key = prefix + "." + name
			}
			result = append(result, fmt.Sprintf("%s:%s", key, stringifyValue(fieldVal.Interface())))
		}
	}
	return result
}

func (f WorksFilter) MarshalText() ([]byte, error) {
	parts := marshalStruct(f, "")
	return []byte(strings.Join(parts, ",")), nil
}

func (f WorksFilter) Encode() (string, error) {
	b, err := f.MarshalText()
	if err != nil {
		return "", err
	}
	// Not very efficient but should not even be noticeable
	// s := strings.ReplaceAll(string(b), "\n", ",")
	// s = strings.ReplaceAll(s, ": ", ":")
	return string(b), nil
}

type Pagination struct {
	//The number of items returned in a single response (default is 20, and maximum is 1,000).
	Rows int64 `url:"rows,omitempty"`
	// offset parameter can be used to retrieve items starting from a specific index of the result list
	Offset int64 `url:"offset,omitempty"`
	// see "Deep-paging" section of CrossRef works API doc
	Cursor string `url:"cursor,omitempty"`
}

// WorksQuery represents a query for works in the CrossRef API.
// See https://api.crossref.org/swagger-ui/index.html#/Works/get_works
type WorksQuery struct {
	FreeFormQuery string
	Fields        *WorksQueryFields
	Pagination    *Pagination
	Filters       *WorksFilter
}

func (q WorksQuery) Encode() (values url.Values, err error) {

	// Named query parameters
	values, err = query.Values(q.Fields)
	if err != nil {
		return values, err
	}

	// Free form query
	if strings.TrimSpace(q.FreeFormQuery) != "" {
		values.Add("query", q.FreeFormQuery)
	}

	// Pagination
	if q.Pagination != nil {
		pag, err := query.Values(q.Pagination)
		if err != nil {
			return values, err
		}
		mergeQueries(&values, pag)
	}

	// Filters
	if q.Filters != nil {
		filters, err := q.Filters.Encode()
		if err != nil {
			return values, err
		}
		values.Add("filter", filters)
	}
	return values, nil
}
