package acme

import (
	"errors"
	"fmt"
	"net/http"
)

// ACME server response statuses used to describe Authorization and Challenge states.
const (
	StatusUnknown    = "unknown"
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusValid      = "valid"
	StatusInvalid    = "invalid"
	StatusRevoked    = "revoked"
)

// CRLReasonCode identifies the reason for a certificate revocation.
type CRLReasonCode int

// CRL reason codes as defined in RFC 5280.
const (
	CRLReasonUnspecified          CRLReasonCode = 0
	CRLReasonKeyCompromise        CRLReasonCode = 1
	CRLReasonCACompromise         CRLReasonCode = 2
	CRLReasonAffiliationChanged   CRLReasonCode = 3
	CRLReasonSuperseded           CRLReasonCode = 4
	CRLReasonCessationOfOperation CRLReasonCode = 5
	CRLReasonCertificateHold      CRLReasonCode = 6
	CRLReasonRemoveFromCRL        CRLReasonCode = 8
	CRLReasonPrivilegeWithdrawn   CRLReasonCode = 9
	CRLReasonAACompromise         CRLReasonCode = 10
)

// ErrUnsupportedKey is returned when an unsupported key type is encountered.
var ErrUnsupportedKey = errors.New("acme: unknown key type; only RSA and ECDSA are supported")

// Error is an ACME error, defined in Problem Details for HTTP APIs doc
// http://tools.ietf.org/html/draft-ietf-appsawg-http-problem.
type Error struct {
	// StatusCode is The HTTP status code generated by the origin server.
	StatusCode int
	// ProblemType is a URI reference that identifies the problem type,
	// typically in a "urn:acme:error:xxx" form.
	ProblemType string
	// Detail is a human-readable explanation specific to this occurrence of the problem.
	Detail string
	// Header is the original server error response headers.
	Header http.Header
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s: %s", e.StatusCode, e.ProblemType, e.Detail)
}

// Account is a user account. It is associated with a private key.
type Account struct {
	// URI is the account unique ID, which is also a URL used to retrieve
	// account data from the CA.
	URI string

	// Contact is a slice of contact info used during registration.
	Contact []string

	// The terms user has agreed to.
	// A value not matching CurrentTerms indicates that the user hasn't agreed
	// to the actual Terms of Service of the CA.
	AgreedTerms string

	// Actual terms of a CA.
	CurrentTerms string

	// Authz is the authorization URL used to initiate a new authz flow.
	Authz string

	// Authorizations is a URI from which a list of authorizations
	// granted to this account can be fetched via a GET request.
	Authorizations string

	// Certificates is a URI from which a list of certificates
	// issued for this account can be fetched via a GET request.
	Certificates string
}

// Directory is ACME server discovery data.
type Directory struct {
	// RegURL is an account endpoint URL, allowing for creating new
	// and modifying existing accounts.
	RegURL string

	// AuthzURL is used to initiate Identifier Authorization flow.
	AuthzURL string

	// CertURL is a new certificate issuance endpoint URL.
	CertURL string

	// RevokeURL is used to initiate a certificate revocation flow.
	RevokeURL string

	// Term is a URI identifying the current terms of service.
	Terms string

	// Website is an HTTP or HTTPS URL locating a website
	// providing more information about the ACME server.
	Website string

	// CAA consists of lowercase hostname elements, which the ACME server
	// recognises as referring to itself for the purposes of CAA record validation
	// as defined in RFC6844.
	CAA []string
}

// Challenge encodes a returned CA challenge.
type Challenge struct {
	// Type is the challenge type, e.g. "http-01", "tls-sni-02", "dns-01".
	Type string

	// URI is where a challenge response can be posted to.
	URI string

	// Token is a random value that uniquely identifies the challenge.
	Token string

	// Status identifies the status of this challenge.
	Status string
}

// Authorization encodes an authorization response.
type Authorization struct {
	// URI uniquely identifies a authorization.
	URI string

	// Status identifies the status of an authorization.
	Status string

	// Identifier is what the account is authorized to represent.
	Identifier AuthzID

	// Challenges that the client needs to fulfill in order to prove possession
	// of the identifier (for pending authorizations).
	// For final authorizations, the challenges that were used.
	Challenges []*Challenge

	// A collection of sets of challenges, each of which would be sufficient
	// to prove possession of the identifier.
	// Clients must complete a set of challenges that covers at least one set.
	// Challenges are identified by their indices in the challenges array.
	// If this field is empty, the client needs to complete all challenges.
	Combinations [][]int
}

// AuthzID is an identifier that an account is authorized to represent.
type AuthzID struct {
	Type  string // The type of identifier, e.g. "dns".
	Value string // The identifier itself, e.g. "example.org".
}

// wireAuthz is ACME JSON representation of Authorization objects.
type wireAuthz struct {
	Status       string
	Challenges   []wireChallenge
	Combinations [][]int
	Identifier   struct {
		Type  string
		Value string
	}
}

func (z *wireAuthz) authorization(uri string) *Authorization {
	a := &Authorization{
		URI:          uri,
		Status:       z.Status,
		Identifier:   AuthzID{Type: z.Identifier.Type, Value: z.Identifier.Value},
		Combinations: z.Combinations, // shallow copy
		Challenges:   make([]*Challenge, len(z.Challenges)),
	}
	for i, v := range z.Challenges {
		a.Challenges[i] = v.challenge()
	}
	return a
}

// wireChallenge is ACME JSON challenge representation.
type wireChallenge struct {
	URI    string `json:"uri"`
	Type   string
	Token  string
	Status string
}

func (c *wireChallenge) challenge() *Challenge {
	v := &Challenge{
		URI:    c.URI,
		Type:   c.Type,
		Token:  c.Token,
		Status: c.Status,
	}
	if v.Status == "" {
		v.Status = StatusPending
	}
	return v
}
