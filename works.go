package crossrefapi

import (
	"encoding/json"
	"sort"
	"strings"
)

// Works is a representation retrieved the CrossRef REST API using
// the Works path and a DOI. This is based on documentaiton at
// https://api.crossref.org/swagger-ui/index.html#/Works/get_works__doi_
// Captured on 2023-03-28, RSD
//
// NOTE: structure in documentation appears wrong, my test records
// indicate that some things listed as array of string are really
// just strings and visa versa.
type Works struct {
	Status         string   `json:"status,omitempty"`
	MessageType    string   `json:"message-type,omitempty"`
	MessageVersion string   `json:"message-version,omitempty"`
	Message        *Message `json:"message,omitempty"`
}

type Message struct {
	// Institutional information
	Institution []*Organization `json:"institution,omitempty"`
	// Indexed described when the work was last indexed
	Indexed *DateObject `json:"indexed,omitempty"`
	// Posted is when the work was posted to the API??
	Posted *DateObject `json:"posted,omitempty"`
	// PublisherLocation, where they are located as a string
	PublisherLocation string `json:"publisher-location,omitempty"`
	// UpdateTo ????
	UpdateTo []*Updated `json:"updated-to,omitempty"`
	// StandardsBody, ???
	StandardsBody       []*Organization      `json:"standards-body,omitempty"`
	EditionNumber       string               `json:"edition-number,omitempty"`
	GroupTitle          string             `json:"group-title,omitempty"`
	Publisher           string               `json:"publisher,omitempty"`
	Issue               string               `json:"issue,omitempty"`
	IsbnType            []*Identifier        `json:"isbn-type,omitempty"`
	License             []*License           `json:"license,omitempty"`
	Funder              []*Funder            `json:"funder,omitempty"`
	ContentDomain       *ContentDomain       `json:"content-domain,omitempty"`
	Chair               []*Person            `json:"chair,omitempty"`
	ShortContainerTitle []string             `json:"short-container-title,omitempty"`
	Accepted            *DateObject          `json:"accepted,omitempty"`
	ContentUpdated      *DateObject          `json:"content-updated,omitempty"`
	PublishedPrint      *DateObject          `json:"published-print,omitempty"`
	Abstract            string               `json:"abstract,omitempty"`
	DOI                 string               `json:"doi,omitempty"`
	Type                string               `json:"type,omitempty"`
	Created             *DateObject          `json:"created,omitempty"`
	Approved            *DateObject          `json:"approved,omitempty"`
	Page                string               `json:"page,omitempty"`
	UpdatePolicy        string               `json:"update-policy,omitempty"`
	Source              string               `json:"source,omitempty"`
	Title               []string             `json:"title,omitempty"`
	Prefix              string               `json:"prefix,omitempty"`
	Volume              string               `json:"volume,omitempty"`
	ClinicalTrailNumber *ClinicalTrailNumber `json:"clinical-trail-number,omitempty"`
	Author              []*Person            `json:"author,omitempty"`
	Member              string               `json:"member,omitempty"`
	ContentCreated      *DateObject          `json:"content-created,omitempty"`
	PublishedOnline     *DateObject          `json:"published-online,omitempty"`
	Reference           []*Reference         `json:"reference,omitempty"`
	ContainerTitle      []string             `json:"container-title,omitempty"`
	Review              *Review              `json:"review,omitempty"`
	OriginalTitle       []string             `json:"original-title,omitempty"`
	Language            string               `json:"language,omitempty"`
	Link                []*Link              `json:"link,omitempty"`
	Deposited           *DateObject          `json:"deposited,omitempty"`
	Score               int                  `json:"score,omitempty"`
	Degree              string               `json:"degree,omitempty"`
	SubTitle            []string             `json:"subtitle,omitempty"`
	Translator          []*Person            `json:"translator,omitempty"`
	FreeToRead          *DateRange           `json:"free-to-read,omitempty"`
	Editor              []*Person            `json:"editor,omitempty"`
	ComponentNumber     string               `json:"component-number,omitempty"`
	ShortTitle          []string             `json:"short-title,omitempty"`
	Issued              *DateObject          `json:"issued,omitempty"`
	ISBN                []string             `json:"isbn,omitempty"`
	ReferenceCount      int                  `json:"reference-count,omitempty"`
	PartNumber          string               `json:"part-number,omitempty"`
	JournalIssue        *JournalIssue        `json:"journal-issue,omitempty"`
	ArticleNumber       string               `json:"article-number,omitempty"`
	AlternativeId       []string             `json:"alternative-id,omitempty"`
	URL                 string               `json:"URL,omitempty"`
	Archive             []string             `json:"archive,omitempty"`
	Relation            map[string][]*Property `json:"relation,omitempty"`
	ISSN                []string             `json:"issn,omitempty"`
	IssnType            []*Identifier        `json:"issn-type,omitempty"`
	Subject             []string             `json:"subject,omitempty"`
	PublishedOther      *DateObject          `json:"published-other,omitempty"`
	Published           *DateObject          `json:"published,omitempty"`
	Assertion           []*Assertion         `json:"assertion,omitempty"`
}

type Identifier struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	IdType string `json:"id-type,omitempty"`
	Id     string `json:"id,omitempty"`
	AssertedBy string `json:"asserted-by,omitempty"`
}

type Person struct {
	ORCID              string          `json:"ORCID,omitempty"`
	Suffix             string          `json:"suffix,omitempty"`
	Given              string          `json:"given,omitempty"`
	Family             string          `json:"family,omitempty"`
	Affiliation        []*Organization `json:"affiliation,omitempty"`
	Name               string          `json:"name,omitempty"`
	AuthenticatedOrcid bool            `json:"authenticated-orcid,omitempty"`
	Prefix             string          `json:"prefix,omitempty"`
	Sequence           string          `json:"sequence,omitempty"`
}

type Organization struct {
	IDs        []*Identifier `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	Place      []string `json:"place,omitempty"`
	Department []string `json:"department,omitempty"`
	Acronym    []string `json:"acronym,omitempty"`
}

type License struct {
	URL            string      `json:"URL,omitempty"`
	Start          *DateObject `json:"start,omitempty"`
	DelayInDays    int         `json:"delay-in-days,omitempty"`
	ContentVersion string      `json:"content-version,omitempty"`
}

type Funder struct {
	Name          string   `json:"name,omitempty"`
	DOI           string   `json:"DOI,omitempty"`
	DoiAssertedBy string   `json:"doi-asserted-by,omitempty"`
	Award         []string `json:"award,omitempty"`
}

type ClinicalTrailNumber struct {
	ClinicalTrailNumber string `json:"clinical-trail-number,omitempty"`
	Registry            string `json:"registry,omitempty"`
	Type                string `json:"type,omitempty"`
}

type ContentDomain struct {
	Domain               []string `json:"domain,omitempty"`
	CrossmarkRestriction bool     `json:"crossmark-restriction,omitempty"`
}

type Review struct {
	Type                       string `json:"type,omitempty"`
	RunningNumber              string `json:"running-number,omitempty"`
	RevisionRound              string `json:"revision-round,omitempty"`
	Stage                      string `json:"stage,omitempty"`
	CompetingInterestStatement string `json:"competing-interest-statement,omitempty"`
	Recommendation             string `json:"recommendation,omitempty"`
	Language                   string `json:"language,omitempty"`
}

type Updated struct {
	Label   string      `json:"label,omitempty"`
	DOI     string      `json:"doi,omitempty"`
	Type    string      `json:"type,omitempty"`
	Updated *DateObject `json:"updated,omitempty"`
}

// DateObject is a date/timestamp/action timestamp of when
// something happened. It is used repeated in the message object
type DateObject struct {
	// DateParts holds a date an an array of Year, Month and Day integer values
	DateParts [][]int `json:"date-parts,omitempty"`
	// DateTime holds a date/time stamp, e.g. 2023-03-28T18:43:06.364Z
	DateTime string `json:"date-time,omitempty"`
	// Olds an integer representation of a timestamp, Unix epoch?
	Timestamp int64 `json:"timestamp,omitempty"`
}

type Reference struct {
	ISSN               string `json:"issn,omitempty"`
	StandardsBody      string `json:"standards-body,omitempty"`
	Issue              string `json:"issue,omitempty"`
	Key                string `json:"key,omitempty"`
	SeriesTitle        string `json:"series-title,omitempty"`
	IsbnType           string `json:"isbn-type,omitempty"`
	DoiAssertedBy      string `json:"doi-asserted-by,omitempty"`
	FirstPage          string `json:"first-page,omitempty"`
	ISBN               string `json:"isbn,omitempty"`
	DOI                string `json:"doi,omitempty"`
	Component          string `json:"component,omitempty"`
	ArticleTitle       string `json:"article-title,omitempty"`
	VolumeTitle        string `json:"volume-title,omitempty"`
	Volume             string `json:"volume,omitempty"`
	Author             string `json:"author,omitempty"`
	StandardDesignator string `json:"standard-designator,omitempty"`
	Year               string `json:"year,omitempty"`
	Unstructured       string `json:"unstructured,omitempty"`
	Edition            string `json:"edition,omitempty"`
	JournalTitle       string `json:"journal-title,omitempty"`
	IssnType           string `json:"issn-type,omitempty"`
}

type Assertion struct {
	Group *Group `json:"group,omitempty"`
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Group struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Link struct {
	URL                 string `json:"URL,omitempty"`
	ContentType         string `json:"content-type,omitempty"`
	ContentVersion      string `json:"content-version,omitempty"`
	IntendedApplication string `json:"intended-application,omitempty"`
}

type DateRange struct {
	StartDate *DateObject `json:"start-date,omitempty"`
	EndDate   *DateObject `json:"end-date,omitempty"`
}

type JournalIssue struct {
	Issue string `json:"issue,omitempty"`
}

type Property struct {
	IdType     string `json:"id-type,omitempty"`
	Id         string `json:"id,omitempty"`
	AssertedBy string `json:"asserted-by,omitempty"`
}

// IsSame checks if the date objects represent the same date.
// NOTE: if both objects are nil, they are considered the same.
func (do *DateObject) IsSame(t *DateObject) bool {
	if do == nil && t == nil {
		return true
	}
	if do == nil || t == nil {
		return false
	}
	if len(do.DateParts) != len(t.DateParts) {
		return false
	}
	for i := range do.DateParts {
		if len(do.DateParts[i]) != len(t.DateParts[i]) {
			return false
		}
		for j, v := range do.DateParts[i] {
			if v != t.DateParts[i][j] {
				return false
			}
		}
	}
	if do.DateTime != t.DateTime {
		return false
	}
	return (do.Timestamp == t.Timestamp)
}

// IsSame checks of the reference object are the same.
// NOTE: if both objects are nil, they are considered the same.
func (ref *Reference) IsSame(t *Reference) bool {
	if ref == nil && t == nil {
		return true
	}
	if ref == nil || t == nil {
		return false
	}
	if ref.ISSN != t.ISSN {
		return false
	}
	if ref.StandardsBody != t.StandardsBody {
		return false
	}
	if ref.Issue != t.Issue {
		return false
	}
	if ref.Key != t.Key {
		return false
	}
	if ref.SeriesTitle != t.SeriesTitle {
		return false
	}
	if ref.IsbnType != t.IsbnType {
		return false
	}
	if ref.DoiAssertedBy != t.DoiAssertedBy {
		return false
	}
	if ref.FirstPage != t.FirstPage {
		return false
	}
	if ref.ISBN != t.ISBN {
		return false
	}
	if ref.DOI != t.DOI {
		return false
	}
	if ref.Component != t.Component {
		return false
	}
	if ref.ArticleTitle != t.ArticleTitle {
		return false
	}
	if ref.VolumeTitle != t.VolumeTitle {
		return false
	}
	if ref.Volume != t.Volume {
		return false
	}
	if ref.Author != t.Author {
		return false
	}
	if ref.StandardDesignator != t.StandardDesignator {
		return false
	}
	if ref.Year != t.Year {
		return false
	}
	if ref.Unstructured != t.Unstructured {
		return false
	}
	if ref.Edition != t.Edition {
		return false
	}
	if ref.JournalTitle != t.JournalTitle {
		return false
	}
	return (ref.IssnType == t.IssnType)
}

func isSameStrings(s1 []string, s2 []string) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	if len(s1) != len(s2) {
		return false
	}
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}

// IsSame checks if two works object are the same.
// NOTE: if both objects are nil, they are considered the same.
func (org *Organization) IsSame(t *Organization) bool {
	if org == nil && t == nil {
		return true
	}
	if org == nil || t == nil {
		return false
	}
	if !isSameStrings(org.Place, t.Place) {
		return false
	}
	if !isSameStrings(org.Department, t.Department) {
		return false
	}
	if !isSameStrings(org.Acronym, t.Acronym) {
		return false
	}
	return (org.Name == t.Name)
}

func (p *Person) IsSame(t *Person) bool {
	if p == nil && t == nil {
		return true
	}
	if p == nil || t == nil {
		return false
	}
	if p.ORCID != t.ORCID {
		return false
	}
	if p.Suffix != t.Suffix {
		return false
	}
	if p.Given != t.Given {
		return false
	}
	if p.Family != t.Family {
		return false
	}
	if !isSameOrganizations(p.Affiliation, t.Affiliation) {
		return false
	}
	if p.Name != t.Name {
		return false
	}
	if p.AuthenticatedOrcid != t.AuthenticatedOrcid {
		return false
	}
	if p.Prefix != t.Prefix {
		return false
	}
	return (p.Sequence == t.Sequence)
}

func (l *Link) IsSame(t *Link) bool {
	if l == nil && t == nil {
		return true
	}
	if l == nil || t == nil {
		return false
	}
	if l.URL != t.URL {
		return false
	}
	if l.ContentType != t.ContentType {
		return false
	}
	if l.ContentVersion != t.ContentVersion {
		return false
	}
	return (l.IntendedApplication == t.IntendedApplication)
}

func (u *Updated) IsSame(t *Updated) bool {
	if u == nil && t == nil {
		return true
	}
	if u == nil || t == nil {
		return false
	}
	if u.Label != t.Label {
		return false
	}
	if u.DOI != t.DOI {
		return false
	}
	if u.Type != t.Type {
		return false
	}
	return u.Updated.IsSame(t.Updated)
}

func (lic *License) IsSame(t *License) bool {
	if lic.URL != t.URL {
		return false
	}
	if !lic.Start.IsSame(t.Start) {
		return false
	}
	if lic.DelayInDays != t.DelayInDays {
		return false
	}
	return (lic.ContentVersion == t.ContentVersion)
}

func isSameUpdatedTo(u1 []*Updated, u2 []*Updated) bool {
	if len(u1) != len(u2) {
		return false
	}
	for i, u := range u1 {
		if !u.IsSame(u2[i]) {
			return false
		}
	}
	return true
}

func isSameOrganizations(o1 []*Organization, o2 []*Organization) bool {
	if len(o1) != len(o2) {
		return false
	}
	for i, org := range o1 {
		if !org.IsSame(o2[i]) {
			return false
		}
	}
	return true
}

func isSameLicenses(l1 []*License, l2 []*License) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i, lic := range l1 {
		if !lic.IsSame(l2[i]) {
			return false
		}
	}
	return true
}

// IsSame checks if two works object are the same.
// NOTE: if both objects are nil, they are considered the same.
func (f *Funder) IsSame(t *Funder) bool {
	if f.Name != t.Name {
		return false
	}
	if f.DOI != t.DOI {
		return false
	}
	if f.DoiAssertedBy != t.DoiAssertedBy {
		return false
	}
	return isSameStrings(f.Award, t.Award)
}

func isSameFunders(f1 []*Funder, f2 []*Funder) bool {
	if len(f1) != len(f2) {
		return false
	}
	for i, f := range f1 {
		if !f.IsSame(f2[i]) {
			return false
		}
	}
	return true
}

func isSamePersons(p1 []*Person, p2 []*Person) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i, p := range p1 {
		if !p.IsSame(p2[i]) {
			return false
		}
	}
	return true
}

func isSameReferences(r1 []*Reference, r2 []*Reference) bool {
	if len(r1) != len(r2) {
		return false
	}
	for i, r := range r1 {
		if !r.IsSame(r2[i]) {
			return false
		}
	}
	return true
}

func (r *Review) IsSame(t *Review) bool {
	if r == nil && t == nil {
		return true
	}
	if r == nil || t == nil {
		return false
	}
	if r.Type != t.Type {
		return false
	}
	if r.RunningNumber != t.RunningNumber {
		return false
	}
	if r.RevisionRound != t.RevisionRound {
		return false
	}
	if r.Stage != t.Stage {
		return false
	}
	if r.CompetingInterestStatement != t.CompetingInterestStatement {
		return false
	}
	if r.Recommendation != t.Recommendation {
		return false
	}
	return (r.Language == t.Language)
}

func isSameLinks(l1 []*Link, l2 []*Link) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i, l := range l1 {
		if !l.IsSame(l2[i]) {
			return false
		}
	}
	return true
}

func (f *DateRange) IsSame(t *DateRange) bool {
	if f == nil && t == nil {
		return true
	}
	if f == nil || t == nil {
		return false
	}
	if !f.StartDate.IsSame(t.StartDate) {
		return false
	}
	if !f.EndDate.IsSame(t.EndDate) {
		return false
	}
	return true
}

func (p *Property) IsSame(t *Property) bool {
	if p == nil && t == nil {
		return true
	}
	if p == nil || t == nil {
		return false
	}
	if p.IdType != t.IdType {
		return false
	}
	if p.Id != t.Id {
		return false
	}
	return (p.AssertedBy == t.AssertedBy)
}

func isSameRelations(r1 map[string][]*Property, r2 map[string][]*Property) bool {
	if len(r1) != len(r2) {
		return false
	}
	r1Keys := []string{}
	for key := range r1 {
		r1Keys = append(r1Keys, key)
	}
	r2Keys := []string{}
	for key := range r2 {
		r2Keys = append(r2Keys, key)
	}
	if len(r1Keys) != len(r2Keys) {
		return false
	}
	sort.Strings(r1Keys)
	
	for _, key := range r1Keys {
		if propList1, ok := r1[key]; ok {
			if propList2, ok := r2[key]; ok {
				if len(propList1) != len(propList2) {
					return false
				}
				for _, prop1 := range propList1 {
					foundIt :=  false
					for _, prop2 := range propList2 {
						if prop1.IsSame(prop2) {
							foundIt = true
							break
						}
					}
					if !foundIt {
						return false
					}
				}
			}
		} else {
			return false
		}
	}
	return true
}

// IsSame checks if two works object are the same.
// NOTE: if both objects are nil, they are considered the same.
func (i *Identifier) IsSame(t *Identifier) bool {
	if i == nil && t == nil {
		return true
	}
	if i == nil || t == nil {
		return false
	}
	if i.Label != t.Label {
		return false
	}
	if i.Name != t.Name {
		return false
	}
	if i.Type != t.Type {
		return false
	}
	return (i.Value == t.Value)
}

func isSameIdentifiers(i1 []*Identifier, i2 []*Identifier) bool {
	if len(i1) != len(i2) {
		return false
	}

	for i, id := range i1 {
		if !id.IsSame(i2[i]) {
			return false
		}
	}
	return true
}

func (g *Group) IsSame(t *Group) bool {
	if g == nil && t == nil {
		return true
	}
	if g == nil || t == nil {
		return false
	}
	if g.Label != t.Label {
		return false
	}
	return (g.Name == t.Name)
}

func (a *Assertion) IsSame(t *Assertion) bool {
	if !a.Group.IsSame(t.Group) {
		return false
	}
	if a.Label != t.Label {
		return false
	}
	if a.Name != t.Name {
		return false
	}
	return (a.Value == t.Value)
}

func (c *ContentDomain) IsSame(t *ContentDomain) bool {
	if c == nil && t == nil {
		return true
	}
	if c == nil || t == nil {
		return false
	}
	if !isSameStrings(c.Domain, t.Domain) {
		return false
	}
	return (c.CrossmarkRestriction == t.CrossmarkRestriction)
}

func isSameAssertions(a1 []*Assertion, a2 []*Assertion) bool {
	if len(a1) != len(a2) {
		return false
	}
	for i, a := range a1 {
		if !a.IsSame(a2[i]) {
			return false
		}
	}
	return true
}

func (ctn *ClinicalTrailNumber) IsSame(t *ClinicalTrailNumber) bool {
	if ctn == nil && t == nil {
		return true
	}
	if ctn == nil || t == nil {
		return false
	}
	if ctn.ClinicalTrailNumber != t.ClinicalTrailNumber {
		return false
	}
	if ctn.Registry != t.Registry {
		return false
	}
	return (ctn.Type == t.Type)
}

func (i *JournalIssue) IsSame(t *JournalIssue) bool {
	if i == nil && t == nil {
		return true
	}
	if i == nil || t == nil {
		return false
	}
	return (i.Issue == t.Issue)
}

// IsSame checks if two works object are the same.
// NOTE: if both objects are nil, they are considered the same.
func (msg *Message) IsSame(t *Message) bool {
	if msg == nil && t == nil {
		return true
	}
	if msg == nil || t == nil {
		return false
	}
	if !isSameOrganizations(msg.Institution, t.Institution) {
		return false
	}
	if !msg.Indexed.IsSame(t.Indexed) {
		return false
	}
	if !msg.Posted.IsSame(t.Posted) {
		return false
	}
	if msg.PublisherLocation != t.PublisherLocation {
		return false
	}
	if !isSameUpdatedTo(msg.UpdateTo, t.UpdateTo) {
		return false
	}
	if !isSameOrganizations(msg.StandardsBody, t.StandardsBody) {
		return false
	}
	if msg.EditionNumber != t.EditionNumber {
		return false
	}
	if strings.Compare(msg.GroupTitle, t.GroupTitle) != 0 {
		return false
	}
	if msg.Publisher != t.Publisher {
		return false
	}
	if msg.Issue != t.Issue {
		return false
	}
	if !isSameIdentifiers(msg.IsbnType, t.IsbnType) {
		return false
	}
	if !isSameLicenses(msg.License, t.License) {
		return false
	}
	if !isSameFunders(msg.Funder, t.Funder) {
		return false
	}
	if !msg.ContentDomain.IsSame(t.ContentDomain) {
		return false
	}
	if !isSamePersons(msg.Chair, t.Chair) {
		return false
	}
	if !isSameStrings(msg.ShortContainerTitle, t.ShortContainerTitle) {
		return false
	}
	if !msg.Accepted.IsSame(t.Accepted) {
		return false
	}
	if !msg.ContentUpdated.IsSame(t.ContentUpdated) {
		return false
	}
	if !msg.PublishedPrint.IsSame(t.PublishedPrint) {
		return false
	}
	if msg.Abstract != t.Abstract {
		return false
	}
	if msg.DOI != t.DOI {
		return false
	}
	if msg.Type != t.Type {
		return false
	}
	if !msg.Created.IsSame(t.Created) {
		return false
	}
	if !msg.Approved.IsSame(t.Approved) {
		return false
	}
	if msg.Page != t.Page {
		return false
	}
	if msg.UpdatePolicy != t.UpdatePolicy {
		return false
	}
	if msg.Source != t.Source {
		return false
	}
	if !isSameStrings(msg.Title, t.Title) {
		return false
	}
	if msg.Prefix != t.Prefix {
		return false
	}
	if msg.Volume != t.Volume {
		return false
	}
	if !msg.ClinicalTrailNumber.IsSame(t.ClinicalTrailNumber) {
		return false
	}
	if !isSamePersons(msg.Author, t.Author) {
		return false
	}
	if msg.Member != t.Member {
		return false
	}
	if !msg.ContentCreated.IsSame(t.ContentCreated) {
		return false
	}
	if !msg.PublishedOnline.IsSame(t.PublishedOnline) {
		return false
	}
	if !isSameReferences(msg.Reference, t.Reference) {
		return false
	}
	if !isSameStrings(msg.ContainerTitle, t.ContainerTitle) {
		return false
	}
	if !msg.Review.IsSame(t.Review) {
		return false
	}
	if !isSameStrings(msg.OriginalTitle, t.OriginalTitle) {
		return false
	}
	if msg.Language != t.Language {
		return false
	}
	if !isSameLinks(msg.Link, t.Link) {
		return false
	}
	if !msg.Deposited.IsSame(t.Deposited) {
		return false
	}
	if msg.Score != t.Score {
		return false
	}
	if msg.Degree != t.Degree {
		return false
	}
	if !isSameStrings(msg.SubTitle, t.SubTitle) {
		return false
	}
	if !isSamePersons(msg.Translator, t.Translator) {
		return false
	}
	if !msg.FreeToRead.IsSame(t.FreeToRead) {
		return false
	}
	if !isSamePersons(msg.Editor, t.Editor) {
		return false
	}
	if msg.ComponentNumber != t.ComponentNumber {
		return false
	}
	if !isSameStrings(msg.ShortTitle, t.ShortTitle) {
		return false
	}
	if !msg.Issued.IsSame(t.Issued) {
		return false
	}
	if !isSameStrings(msg.ISBN, t.ISBN) {
		return false
	}
	if msg.ReferenceCount != t.ReferenceCount {
		return false
	}
	if msg.PartNumber != t.PartNumber {
		return false
	}
	if !msg.JournalIssue.IsSame(t.JournalIssue) {
		return false
	}
	if !isSameStrings(msg.AlternativeId, t.AlternativeId) {
		return false
	}
	if msg.URL != t.URL {
		return false
	}
	if !isSameStrings(msg.Archive, t.Archive) {
		return false
	}
	if !isSameRelations(msg.Relation, t.Relation) {
		return false
	}
	if !isSameStrings(msg.ISSN, t.ISSN) {
		return false
	}
	if !isSameIdentifiers(msg.IssnType, t.IssnType) {
		return false
	}
	if !isSameStrings(msg.Subject, t.Subject) {
		return false
	}
	if !msg.PublishedOther.IsSame(t.PublishedOther) {
		return false
	}
	if !msg.Published.IsSame(t.Published) {
		return false
	}
	return isSameAssertions(msg.Assertion, t.Assertion)
}

// Changes takes the current Message, a new version of the Message
// and returns a Message object with the new Message object containing
// only the new elements.
func (msg *Message) Changes(t *Message) *Message {
	if msg == nil && t == nil {
		return nil
	}
	if msg == nil && t != nil {
		return t
	}
	// Aggregate the changed fields
	nMsg := new(Message)
	if !isSameOrganizations(msg.Institution, t.Institution) {
		nMsg.Institution = t.Institution
	}
	if !msg.Indexed.IsSame(t.Indexed) {
		nMsg.Indexed = t.Indexed
	}
	if !msg.Posted.IsSame(t.Posted) {
		nMsg.Posted = t.Posted
	}
	if msg.PublisherLocation != t.PublisherLocation {
		nMsg.PublisherLocation = t.PublisherLocation
	}
	if !isSameUpdatedTo(msg.UpdateTo, t.UpdateTo) {
		nMsg.UpdateTo = t.UpdateTo
	}
	if !isSameOrganizations(msg.StandardsBody, t.StandardsBody) {
		nMsg.StandardsBody = t.StandardsBody
	}
	if msg.EditionNumber != t.EditionNumber {
		nMsg.EditionNumber = t.EditionNumber
	}
	if strings.Compare(msg.GroupTitle, t.GroupTitle) != 0{
		nMsg.GroupTitle = t.GroupTitle
	}
	if msg.Publisher != t.Publisher {
		nMsg.Publisher = t.Publisher
	}
	if msg.Issue != t.Issue {
		nMsg.Issue = t.Issue
	}
	if !isSameIdentifiers(msg.IsbnType, t.IsbnType) {
		nMsg.IsbnType = t.IsbnType
	}
	if !isSameLicenses(msg.License, t.License) {
		nMsg.License = t.License
	}
	if !isSameFunders(msg.Funder, t.Funder) {
		nMsg.Funder = t.Funder
	}
	if !msg.ContentDomain.IsSame(t.ContentDomain) {
		nMsg.ContentDomain = t.ContentDomain
	}
	if !isSamePersons(msg.Chair, t.Chair) {
		nMsg.Chair = t.Chair
	}
	if !isSameStrings(msg.ShortContainerTitle, t.ShortContainerTitle) {
		nMsg.ShortContainerTitle = t.ShortContainerTitle
	}
	if !msg.Accepted.IsSame(t.Accepted) {
		nMsg.Accepted = t.Accepted
	}
	if !msg.ContentUpdated.IsSame(t.ContentUpdated) {
		nMsg.ContentUpdated = t.ContentUpdated
	}
	if !msg.PublishedPrint.IsSame(t.PublishedPrint) {
		nMsg.PublishedPrint = t.PublishedPrint
	}
	if msg.Abstract != t.Abstract {
		nMsg.Abstract = t.Abstract
	}
	if msg.DOI != t.DOI {
		nMsg.DOI = t.DOI
	}
	if msg.Type != t.Type {
		nMsg.Type = t.Type
	}
	if !msg.Created.IsSame(t.Created) {
		nMsg.Created = t.Created
	}
	if !msg.Approved.IsSame(t.Approved) {
		nMsg.Approved = t.Approved
	}
	if msg.Page != t.Page {
		nMsg.Page = t.Page
	}
	if msg.UpdatePolicy != t.UpdatePolicy {
		nMsg.UpdatePolicy = t.UpdatePolicy
	}
	if msg.Source != t.Source {
		nMsg.Source = t.Source
	}
	if !isSameStrings(msg.Title, t.Title) {
		nMsg.Title = t.Title
	}
	if msg.Prefix != t.Prefix {
		nMsg.Prefix = t.Prefix
	}
	if msg.Volume != t.Volume {
		nMsg.Volume = t.Volume
	}
	if !msg.ClinicalTrailNumber.IsSame(t.ClinicalTrailNumber) {
		nMsg.ClinicalTrailNumber = t.ClinicalTrailNumber
	}
	if !isSamePersons(msg.Author, t.Author) {
		nMsg.Author = t.Author
	}
	if msg.Member != t.Member {
		nMsg.Member = t.Member
	}
	if !msg.ContentCreated.IsSame(t.ContentCreated) {
		nMsg.ContentCreated = t.ContentCreated
	}
	if !msg.PublishedOnline.IsSame(t.PublishedOnline) {
		nMsg.PublishedOnline = t.PublishedOnline
	}
	if !isSameReferences(msg.Reference, t.Reference) {
		nMsg.Reference = t.Reference
	}
	if !isSameStrings(msg.ContainerTitle, t.ContainerTitle) {
		nMsg.ContainerTitle = t.ContainerTitle
	}
	if !msg.Review.IsSame(t.Review) {
		nMsg.Review = t.Review
	}
	if !isSameStrings(msg.OriginalTitle, t.OriginalTitle) {
		nMsg.OriginalTitle = t.OriginalTitle
	}
	if msg.Language != t.Language {
		nMsg.Language = t.Language
	}
	if !isSameLinks(msg.Link, t.Link) {
		nMsg.Link = t.Link
	}
	if !msg.Deposited.IsSame(t.Deposited) {
		nMsg.Deposited = t.Deposited
	}
	if msg.Score != t.Score {
		nMsg.Score = t.Score
	}
	if msg.Degree != t.Degree {
		nMsg.Degree = t.Degree
	}
	if !isSameStrings(msg.SubTitle, t.SubTitle) {
		nMsg.SubTitle = t.SubTitle
	}
	if !isSamePersons(msg.Translator, t.Translator) {
		nMsg.Translator = t.Translator
	}
	if !msg.FreeToRead.IsSame(t.FreeToRead) {
		nMsg.FreeToRead = t.FreeToRead
	}
	if !isSamePersons(msg.Editor, t.Editor) {
		nMsg.Editor = t.Editor
	}
	if msg.ComponentNumber != t.ComponentNumber {
		nMsg.ComponentNumber = t.ComponentNumber
	}
	if !isSameStrings(msg.ShortTitle, t.ShortTitle) {
		nMsg.ShortTitle = t.ShortTitle
	}
	if !msg.Issued.IsSame(t.Issued) {
		nMsg.Issued = t.Issued
	}
	if !isSameStrings(msg.ISBN, t.ISBN) {
		nMsg.ISBN = t.ISBN
	}
	if msg.ReferenceCount != t.ReferenceCount {
		nMsg.ReferenceCount = t.ReferenceCount
	}
	if msg.PartNumber != t.PartNumber {
		nMsg.PartNumber = t.PartNumber
	}
	if !msg.JournalIssue.IsSame(t.JournalIssue) {
		nMsg.JournalIssue = t.JournalIssue
	}
	if !isSameStrings(msg.AlternativeId, t.AlternativeId) {
		nMsg.AlternativeId = t.AlternativeId
	}
	if msg.URL != t.URL {
		nMsg.URL = t.URL
	}
	if !isSameStrings(msg.Archive, t.Archive) {
		nMsg.Archive = t.Archive
	}
	if !isSameRelations(msg.Relation, t.Relation) {
		nMsg.Relation = t.Relation
	}
	if !isSameStrings(msg.ISSN, t.ISSN) {
		nMsg.ISSN = t.ISSN
	}
	if !isSameIdentifiers(msg.IssnType, t.IssnType) {
		nMsg.IssnType = t.IssnType
	}
	if !isSameStrings(msg.Subject, t.Subject) {
		nMsg.Subject = t.Subject
	}
	if !msg.PublishedOther.IsSame(t.PublishedOther) {
		nMsg.PublishedOther = t.PublishedOther
	}
	if !msg.Published.IsSame(t.Published) {
		nMsg.Published = t.Published
	}
	if !isSameAssertions(msg.Assertion, t.Assertion) {
		nMsg.Assertion = t.Assertion
	}
	return nMsg
}

// Diff takes the current Message, a new version of the Message
// and two Message objects one holding the old values and another
// holding the new values.
func (msg *Message) Diff(t *Message) (*Message, *Message) {
	if msg == nil && t == nil {
		return nil, nil
	}
	if msg == nil && t != nil {
		return nil, t
	}
	if msg != nil && t == nil {
		return msg, nil
	}
	// Aggregate the changed fields
	oMsg := new(Message)
	nMsg := new(Message)
	if !isSameOrganizations(msg.Institution,t.Institution) {
		oMsg.Institution = msg.Institution
		nMsg.Institution = t.Institution
	}
	if !msg.Indexed.IsSame(t.Indexed) {
		oMsg.Indexed = msg.Indexed
		nMsg.Indexed = t.Indexed
	}
	if !msg.Posted.IsSame(t.Posted) {
		oMsg.Posted = msg.Posted
		nMsg.Posted = t.Posted
	}
	if msg.PublisherLocation != t.PublisherLocation {
		oMsg.PublisherLocation = msg.PublisherLocation
		nMsg.PublisherLocation = t.PublisherLocation
	}
	if !isSameUpdatedTo(msg.UpdateTo, t.UpdateTo) {
		oMsg.UpdateTo = msg.UpdateTo
		nMsg.UpdateTo = t.UpdateTo
	}
	if !isSameOrganizations(msg.StandardsBody, t.StandardsBody) {
		oMsg.StandardsBody = msg.StandardsBody
		nMsg.StandardsBody = t.StandardsBody
	}
	if msg.EditionNumber != t.EditionNumber {
		oMsg.EditionNumber = msg.EditionNumber
		nMsg.EditionNumber = t.EditionNumber
	}
	if strings.Compare(msg.GroupTitle, t.GroupTitle) != 0{
		oMsg.GroupTitle = msg.GroupTitle
		nMsg.GroupTitle = t.GroupTitle
	}
	if msg.Publisher != t.Publisher {
		oMsg.Publisher = msg.Publisher
		nMsg.Publisher = t.Publisher
	}
	if msg.Issue != t.Issue {
		oMsg.Issue = msg.Issue
		nMsg.Issue = t.Issue
	}
	if !isSameIdentifiers(msg.IsbnType, t.IsbnType) {
		oMsg.IsbnType = msg.IsbnType
		nMsg.IsbnType = t.IsbnType
	}
	if !isSameLicenses(msg.License, t.License) {
		oMsg.License = msg.License
		nMsg.License = t.License
	}
	if !isSameFunders(msg.Funder, t.Funder) {
		oMsg.Funder = msg.Funder
		nMsg.Funder = t.Funder
	}
	if !msg.ContentDomain.IsSame(t.ContentDomain) {
		oMsg.ContentDomain = msg.ContentDomain
		nMsg.ContentDomain = t.ContentDomain
	}
	if !isSamePersons(msg.Chair, t.Chair) {
		oMsg.Chair = msg.Chair
		nMsg.Chair = t.Chair
	}
	if !isSameStrings(msg.ShortContainerTitle, t.ShortContainerTitle) {
		oMsg.ShortContainerTitle = msg.ShortContainerTitle
		nMsg.ShortContainerTitle = t.ShortContainerTitle
	}
	if !msg.Accepted.IsSame(t.Accepted) {
		oMsg.Accepted = msg.Accepted
		nMsg.Accepted = t.Accepted
	}
	if !msg.ContentUpdated.IsSame(t.ContentUpdated) {
		oMsg.ContentUpdated = msg.ContentUpdated
		nMsg.ContentUpdated = t.ContentUpdated
	}
	if !msg.PublishedPrint.IsSame(t.PublishedPrint) {
		oMsg.PublishedPrint = msg.PublishedPrint
		nMsg.PublishedPrint = t.PublishedPrint
	}
	if msg.Abstract != t.Abstract {
		oMsg.Abstract = msg.Abstract
		nMsg.Abstract = t.Abstract
	}
	if msg.DOI != t.DOI {
		oMsg.DOI = msg.DOI
		nMsg.DOI = t.DOI
	}
	if msg.Type != t.Type {
		oMsg.Type = msg.Type
		nMsg.Type = t.Type
	}
	if !msg.Created.IsSame(t.Created) {
		oMsg.Created = msg.Created
		nMsg.Created = t.Created
	}
	if !msg.Approved.IsSame(t.Approved) {
		oMsg.Approved = msg.Approved
		nMsg.Approved = t.Approved
	}
	if msg.Page != t.Page {
		oMsg.Page = msg.Page
		nMsg.Page = t.Page
	}
	if msg.UpdatePolicy != t.UpdatePolicy {
		oMsg.UpdatePolicy = msg.UpdatePolicy
		nMsg.UpdatePolicy = t.UpdatePolicy
	}
	if msg.Source != t.Source {
		oMsg.Source = msg.Source
		nMsg.Source = t.Source
	}
	if !isSameStrings(msg.Title, t.Title) {
		oMsg.Title = msg.Title
		nMsg.Title = t.Title
	}
	if msg.Prefix != t.Prefix {
		oMsg.Prefix = msg.Prefix
		nMsg.Prefix = t.Prefix
	}
	if msg.Volume != t.Volume {
		oMsg.Volume = msg.Volume
		nMsg.Volume = t.Volume
	}
	if !msg.ClinicalTrailNumber.IsSame(t.ClinicalTrailNumber) {
		oMsg.ClinicalTrailNumber = msg.ClinicalTrailNumber
		nMsg.ClinicalTrailNumber = t.ClinicalTrailNumber
	}
	if !isSamePersons(msg.Author, t.Author) {
		oMsg.Author = msg.Author
		nMsg.Author = t.Author
	}
	if msg.Member != t.Member {
		oMsg.Member = msg.Member
		nMsg.Member = t.Member
	}
	if !msg.ContentCreated.IsSame(t.ContentCreated) {
		oMsg.ContentCreated = msg.ContentCreated
		nMsg.ContentCreated = t.ContentCreated
	}
	if !msg.PublishedOnline.IsSame(t.PublishedOnline) {
		oMsg.PublishedOnline = msg.PublishedOnline
		nMsg.PublishedOnline = t.PublishedOnline
	}
	if !isSameReferences(msg.Reference, t.Reference) {
		oMsg.Reference = msg.Reference
		nMsg.Reference = t.Reference
	}
	if !isSameStrings(msg.ContainerTitle, t.ContainerTitle) {
		oMsg.ContainerTitle = msg.ContainerTitle
		nMsg.ContainerTitle = t.ContainerTitle
	}
	if !msg.Review.IsSame(t.Review) {
		oMsg.Review = msg.Review
		nMsg.Review = t.Review
	}
	if !isSameStrings(msg.OriginalTitle, t.OriginalTitle) {
		oMsg.OriginalTitle = msg.OriginalTitle
		nMsg.OriginalTitle = t.OriginalTitle
	}
	if msg.Language != t.Language {
		oMsg.Language = msg.Language
		nMsg.Language = t.Language
	}
	if !isSameLinks(msg.Link, t.Link) {
		oMsg.Link = msg.Link
		nMsg.Link = t.Link
	}
	if !msg.Deposited.IsSame(t.Deposited) {
		oMsg.Deposited = msg.Deposited
		nMsg.Deposited = t.Deposited
	}
	if msg.Score != t.Score {
		oMsg.Score = msg.Score
		nMsg.Score = t.Score
	}
	if msg.Degree != t.Degree {
		oMsg.Degree = msg.Degree
		nMsg.Degree = t.Degree
	}
	if !isSameStrings(msg.SubTitle, t.SubTitle) {
		oMsg.SubTitle = msg.SubTitle
		nMsg.SubTitle = t.SubTitle
	}
	if !isSamePersons(msg.Translator, t.Translator) {
		oMsg.Translator = msg.Translator
		nMsg.Translator = t.Translator
	}
	if !msg.FreeToRead.IsSame(t.FreeToRead) {
		oMsg.FreeToRead = msg.FreeToRead
		nMsg.FreeToRead = t.FreeToRead
	}
	if !isSamePersons(msg.Editor, t.Editor) {
		oMsg.Editor = msg.Editor
		nMsg.Editor = t.Editor
	}
	if msg.ComponentNumber != t.ComponentNumber {
		oMsg.ComponentNumber = msg.ComponentNumber
		nMsg.ComponentNumber = t.ComponentNumber
	}
	if !isSameStrings(msg.ShortTitle, t.ShortTitle) {
		oMsg.ShortTitle = msg.ShortTitle
		nMsg.ShortTitle = t.ShortTitle
	}
	if !msg.Issued.IsSame(t.Issued) {
		oMsg.Issued = msg.Issued
		nMsg.Issued = t.Issued
	}
	if !isSameStrings(msg.ISBN, t.ISBN) {
		oMsg.ISBN = msg.ISBN
		nMsg.ISBN = t.ISBN
	}
	if msg.ReferenceCount != t.ReferenceCount {
		oMsg.ReferenceCount = msg.ReferenceCount
		nMsg.ReferenceCount = t.ReferenceCount
	}
	if msg.PartNumber != t.PartNumber {
		oMsg.PartNumber = msg.PartNumber
		nMsg.PartNumber = t.PartNumber
	}
	if !msg.JournalIssue.IsSame(t.JournalIssue) {
		oMsg.JournalIssue = msg.JournalIssue
		nMsg.JournalIssue = t.JournalIssue
	}
	if !isSameStrings(msg.AlternativeId, t.AlternativeId) {
		oMsg.AlternativeId = msg.AlternativeId
		nMsg.AlternativeId = t.AlternativeId
	}
	if msg.URL != t.URL {
		oMsg.URL = msg.URL
		nMsg.URL = t.URL
	}
	if !isSameStrings(msg.Archive, t.Archive) {
		oMsg.Archive = msg.Archive
		nMsg.Archive = t.Archive
	}
	if !isSameRelations(msg.Relation, t.Relation) {
		oMsg.Relation = msg.Relation
		nMsg.Relation = t.Relation
	}
	if !isSameStrings(msg.ISSN, t.ISSN) {
		oMsg.ISSN = msg.ISSN
		nMsg.ISSN = t.ISSN
	}
	if !isSameIdentifiers(msg.IssnType, t.IssnType) {
		oMsg.IssnType = msg.IssnType
		nMsg.IssnType = t.IssnType
	}
	if !isSameStrings(msg.Subject, t.Subject) {
		oMsg.Subject = msg.Subject
		nMsg.Subject = t.Subject
	}
	if !msg.PublishedOther.IsSame(t.PublishedOther) {
		oMsg.PublishedOther = msg.PublishedOther
		nMsg.PublishedOther = t.PublishedOther
	}
	if !msg.Published.IsSame(t.Published) {
		oMsg.Published = msg.Published
		nMsg.Published = t.Published
	}
	if !isSameAssertions(msg.Assertion, t.Assertion) {
		oMsg.Assertion = msg.Assertion
		nMsg.Assertion = t.Assertion
	}
	return oMsg, nMsg
}

// DiffAsJSON performs a Diff and returns the results as a JSON array
// where the first element (index 0) is the old object's values and
// the second (index 1) is the updated values
func (msg *Message) DiffAsJSON(t *Message) ([]byte, error) {
	o, n := msg.Diff(t)
	return json.MarshalIndent([]*Message{o, n}, "", "    ")
}

// IsSame checks if two works object have the same content.
// NOTE: if both are nil then true is returned. Only compares
// the works' type and message attributes are compared.
func (work *Works) IsSame(t *Works) bool {
	if work == nil && t == nil {
		return true
	}
	if work == nil || t == nil {
		return false
	}
	if work.MessageVersion != t.MessageVersion {
	}
	if work.MessageType != t.MessageType {
		return false
	}
	return work.Message.IsSame(t.Message)
}

// Diff works returns the fields that differ
func (work *Works) Diff(t *Works) (*Works, *Works) {
	if work == nil && t == nil {
		return nil, nil
	}
	if work == nil && t != nil {
		return nil, t
	}
	if work != nil && t == nil {
		return work, nil
	}
	oWork := new(Works)
	nWork := new(Works)
	if work.Status != t.Status {
		oWork.Status = work.Status
		nWork.Status = t.Status
	}
	if work.MessageVersion != t.MessageVersion {
		oWork.MessageVersion = work.MessageVersion
		nWork.MessageVersion = t.MessageVersion
	}
	if work.MessageType != t.MessageType {
		oWork.MessageType = work.MessageType
		nWork.MessageType = t.MessageType
	}
	if !work.Message.IsSame(t.Message) {
		oWork.Message, nWork.Message = work.Message.Diff(t.Message)
	}
	return oWork, nWork
}

// DiffAsJSON performs a Diff and returns the results as a JSON array
// where the first element (index 0) is the old object's values and
// the second (index 1) is the updated values
func (work *Works) DiffAsJSON(t *Works) ([]byte, error) {
	o, n := work.Diff(t)
	return json.MarshalIndent([]*Works{o, n}, "", "    ")
}
