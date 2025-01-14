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
	URL     []string `key:"url,omitempty"`
	Version []string `key:"version,omitempty"`
	Delay   []int    `key:"delay,omitempty"`
}

// Relation represents relation-specific filter parameters
type RelationFilter struct {
	Type       []string `key:"type,omitempty"`
	ObjectType []string `key:"object-type,omitempty"`
	Object     []string `key:"object,omitempty"`
}

type FullTextFilter struct {
	Type        []string `key:"type,omitempty"`
	Application []string `key:"application,omitempty"`
	Version     []string `key:"version,omitempty"`
}

type AwardFilter struct {
	Funder []string `key:"funder,omitempty"`
	Number []int    `key:"number,omitempty"`
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

func (d DateParameter) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// BoolParameter overrides boolean text marshalling to comply with CrossRef API spec
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
	AlternativeID  []string `key:"alternative-id,omitempty"`
	Archive        []string `key:"archive,omitempty"`
	ArticleNumber  []string `key:"article-number,omitempty"`
	Assertion      []string `key:"assertion,omitempty"`
	AssertionGroup []string `key:"assertion-group,omitempty"`

	// Award related fields
	Award *AwardFilter `key:"award,omitempty"`

	CategoryName        []string `key:"category-name,omitempty"`
	CitationID          []string `key:"citation-id,omitempty"`
	ClinicalTrialNumber []string `key:"clinical-trial-number,omitempty"`
	ContainerTitle      []string `key:"container-title,omitempty"`
	ContentDomain       []string `key:"content-domain,omitempty"`
	DOI                 []string `key:"doi,omitempty"`

	// From date fields
	FromAcceptedDate   []DateParameter `key:"from-accepted-date,omitempty"`
	FromApprovedDate   []DateParameter `key:"from-approved-date,omitempty"`
	FromAwardedDate    []DateParameter `key:"from-awarded-date,omitempty"`
	FromCreatedDate    []DateParameter `key:"from-created-date,omitempty"`
	FromDepositDate    []DateParameter `key:"from-deposit-date,omitempty"`
	FromEventEndDate   []DateParameter `key:"from-event-end-date,omitempty"`
	FromEventStartDate []DateParameter `key:"from-event-start-date,omitempty"`
	FromIndexDate      []DateParameter `key:"from-index-date,omitempty"`
	FromIssuedDate     []DateParameter `key:"from-issued-date,omitempty"`
	FromOnlinePubDate  []DateParameter `key:"from-online-pub-date,omitempty"`
	FromPostedDate     []DateParameter `key:"from-posted-date,omitempty"`
	FromPrintPubDate   []DateParameter `key:"from-print-pub-date,omitempty"`
	FromPubDate        []DateParameter `key:"from-pub-date,omitempty"`
	FromUpdateDate     []DateParameter `key:"from-update-date,omitempty"`

	// Full text related fields
	FullText *FullTextFilter `key:"full-text,omitempty"`

	// Other fields
	Funder              []string `key:"funder,omitempty"`
	FunderDoiAssertedBy []string `key:"funder-doi-asserted-by,omitempty"`
	GroupTitle          []string `key:"group-title,omitempty"`

	// Boolean flags
	// No point supporting multiple filter values as array here
	HasAbstract            *BoolParameter `key:"has-abstract"`
	HasAffiliation         *BoolParameter `key:"has-affiliation"`
	HasArchive             *BoolParameter `key:"has-archive"`
	HasAssertion           *BoolParameter `key:"has-assertion"`
	HasAuthenticatedOrcid  *BoolParameter `key:"has-authenticated-orcid"`
	HasAward               *BoolParameter `key:"has-award"`
	HasClinicalTrialNumber *BoolParameter `key:"has-clinical-trial-number"`
	HasContentDomain       *BoolParameter `key:"has-content-domain"`
	HasDescription         *BoolParameter `key:"has-description"`
	HasDomainRestriction   *BoolParameter `key:"has-domain-restriction"`
	HasEvent               *BoolParameter `key:"has-event"`
	HasFullText            *BoolParameter `key:"has-full-text"`
	HasFunder              *BoolParameter `key:"has-funder"`
	HasFunderDoi           *BoolParameter `key:"has-funder-doi"`
	HasLicense             *BoolParameter `key:"has-license"`
	HasOrcid               *BoolParameter `key:"has-orcid"`
	HasReferences          *BoolParameter `key:"has-references"`
	HasRelation            *BoolParameter `key:"has-relation"`
	HasRorID               *BoolParameter `key:"has-ror-id"`
	HasUpdate              *BoolParameter `key:"has-update"`
	HasUpdatePolicy        *BoolParameter `key:"has-update-policy"`
	IsUpdate               *BoolParameter `key:"is-update"`

	// ISBN/ISSN
	ISBN []string `key:"isbn,omitempty"`
	ISSN []string `key:"issn,omitempty"`

	// License fields
	License *LicenseFilter `key:"license,omitempty"`

	// Award amount
	GteAwardAmount int `key:"gte-award-amount,omitempty"`
	LteAwardAmount int `key:"lte-award-amount,omitempty"`

	// Member and identifiers
	Member []string `key:"member,omitempty"`
	ORCID  []string `key:"orcid,omitempty"`
	Prefix []string `key:"prefix,omitempty"`

	// Relation fields
	Relation *RelationFilter `key:"relation,omitempty"`

	// Type fields
	RorID    []string `key:"ror-id,omitempty"`
	Type     []string `key:"type,omitempty"`
	TypeName []string `key:"type-name,omitempty"`

	// Until date fields
	UntilAcceptedDate   []DateParameter `key:"until-accepted-date,omitempty"`
	UntilApprovedDate   []DateParameter `key:"until-approved-date,omitempty"`
	UntilAwardedDate    []DateParameter `key:"until-awarded-date,omitempty"`
	UntilCreatedDate    []DateParameter `key:"until-created-date,omitempty"`
	UntilDepositDate    []DateParameter `key:"until-deposit-date,omitempty"`
	UntilEventEndDate   []DateParameter `key:"until-event-end-date,omitempty"`
	UntilEventStartDate []DateParameter `key:"until-event-start-date,omitempty"`
	UntilIndexDate      []DateParameter `key:"until-index-date,omitempty"`
	UntilIssuedDate     []DateParameter `key:"until-issued-date,omitempty"`
	UntilOnlinePubDate  []DateParameter `key:"until-online-pub-date,omitempty"`
	UntilPostedDate     []DateParameter `key:"until-posted-date,omitempty"`
	UntilPrintPubDate   []DateParameter `key:"until-print-pub-date,omitempty"`
	UntilPubDate        []DateParameter `key:"until-pub-date,omitempty"`
	UntilUpdateDate     []DateParameter `key:"until-update-date,omitempty"`

	// Update fields
	UpdateType []string `key:"update-type,omitempty"`
	Updates    []string `key:"updates,omitempty"`
}

func marshalStruct(v interface{}, prefix string) []string {

	withName := func(s string) string {
		if prefix != "" {
			return fmt.Sprintf("%s:%s", prefix, s)
		}
		return s
	}

	withPrefix := func(s string) string {
		if prefix != "" {
			return fmt.Sprintf("%s.%s", prefix, s)
		}
		return s
	}

	val := reflect.ValueOf(v)
	typ := val.Type()
	var result []string

	// Handle text marshaller interface
	if marshaller, ok := val.Interface().(encoding.TextMarshaler); ok && marshaller != nil {
		text, err := marshaller.MarshalText()
		if err == nil && len(text) > 0 {
			result = append(result, withName(string(text)))
			return result
		}
	}

	switch t := v.(type) {
	case string:
		return []string{withName(t)}
	case *string:
		if t == nil {
			return nil
		}
		return []string{withName(*t)}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return []string{withName(fmt.Sprintf("%d", t))}
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		if t == nil {
			return nil
		}
		return []string{withName(fmt.Sprintf("%d", t))}
	default:
		// Proceed
	}

	// If it's a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		typ = val.Type()
	}

	// Handle slice
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			result = append(result, marshalStruct(val.Index(i).Interface(), prefix)...)
		}
		fmt.Printf("%v\n", result)
		return result
	}

	// Not scalar nor slice: ensure value is Struct
	if val.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Get the key tag
		tag := field.Tag.Get("key")
		if tag == "" {
			tag = field.Name
		}
		tagParts := strings.Split(tag, ",")
		name := tagParts[0]
		omitempty := len(tagParts) > 1 && tagParts[1] == "omitempty"
		newPrefix := withPrefix(name)

		// Skip nil pointers and slices
		if (field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Slice) && fieldVal.IsNil() {
			continue
		}

		// Skip if field is zero value and omitempty is set
		if omitempty && (fieldVal.Kind() != reflect.Ptr && fieldVal.IsZero()) {
			continue
		}

		// Recursively marshal struct fields
		result = append(result, marshalStruct(fieldVal.Interface(), newPrefix)...)
	}
	return result
}

func (f WorksFilter) Encode() string {
	parts := marshalStruct(f, "")
	b := []byte(strings.Join(parts, ","))
	return string(b)
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
// See https://api.crossref.org/swagger-ui/index.html#/Works/get_works.
type WorksQuery struct {
	// General query string
	// Note: as stated in CrossRef API docs,
	// Fields.Bibliographic SHOULD be preferred
	// whenever the search query is a partial or full reference
	// (e.g. including authors, date, title, ...)
	FreeFormQuery string
	// Structured query on specific metadata
	Fields *WorksQueryFields
	// Results pagination
	Pagination *Pagination
	// Results filtering
	// Filters on different keys are applied with AND semantic.
	// Filters on the same key are applied with OR semantic.
	Filters *WorksFilter
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
		filters := q.Filters.Encode()
		values.Add("filter", filters)
	}
	return values, nil
}
