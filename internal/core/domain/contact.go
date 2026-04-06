package domain

import "time"

// ContactID is the primary key for a contact.
type ContactID = CompanyID // UUID

// Contact is a LinkedIn connection imported by a user.
type Contact struct {
	ID                    ContactID
	UserID                UserID
	FirstName             string
	LastName              string
	FullName              string // generated: first_name || ' ' || last_name
	Email                 string
	Title                 string
	LinkedInURL           string
	ConnectedOn           *time.Time
	CompanyName           string     // raw value from LinkedIn CSV
	NormalizedCompanyName string     // normalized for matching
	CompanyID             *CompanyID // linked after company matching; nil if unmatched
	CreatedAt             time.Time
}

// ImportResult summarises a CSV import operation.
type ImportResult struct {
	ContactsImported    int
	CompaniesLinked     int // matched to existing companies
	CompaniesRegistered int // newly discovered via ATS probe
	CompaniesUnmatched  int // could not match or discover
}
