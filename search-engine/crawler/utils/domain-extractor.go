package urltools

import (
	"errors"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ExtractDomain returns the registrable domain (eTLD+1) for a URL, e.g.
// "https://cs.stanford.edu/admissions" -> "stanford.edu".
func ExtractDomain(urlString string) (string, error) {

	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	host := u.Hostname() // Hostname() strips a ":port" suffix; u.Host does not.
	if host == "" {
		return "", errors.New("empty host")
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(strings.ToLower(host))
	if err != nil {
		return "", err
	}

	return domain, nil
}