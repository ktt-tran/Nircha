package document

import "testing"

func TestMatchesProfileNoProfileAlwaysMatches(t *testing.T) {
	l := Listing{DegreeLevels: []string{"phd"}, Location: "Boston, MA"}
	if !l.MatchesProfile(Profile{}) {
		t.Error("an empty profile should never exclude a listing")
	}
}

func TestMatchesProfileDegreeLevelMatch(t *testing.T) {
	l := Listing{DegreeLevels: []string{"undergraduate", "graduate"}}
	if !l.MatchesProfile(Profile{DegreeLevel: "undergraduate"}) {
		t.Error("expected a listed degree level to match")
	}
}

func TestMatchesProfileDegreeLevelMismatchExcludes(t *testing.T) {
	l := Listing{DegreeLevels: []string{"phd"}}
	if l.MatchesProfile(Profile{DegreeLevel: "undergraduate"}) {
		t.Error("expected an explicit degree-level conflict to exclude the listing")
	}
}

func TestMatchesProfileUnknownDegreeLevelPasses(t *testing.T) {
	l := Listing{} // DegreeLevels unset
	if !l.MatchesProfile(Profile{DegreeLevel: "undergraduate"}) {
		t.Error("a listing with no stated degree levels should not be excluded")
	}
}

func TestMatchesProfileLocationMatch(t *testing.T) {
	l := Listing{Location: "Boston, MA"}
	if !l.MatchesProfile(Profile{Location: "Boston"}) {
		t.Error("expected a substring location match to pass")
	}
}

func TestMatchesProfileLocationMismatchExcludes(t *testing.T) {
	l := Listing{Location: "Seattle, WA"}
	if l.MatchesProfile(Profile{Location: "Boston"}) {
		t.Error("expected an explicit location conflict to exclude the listing")
	}
}

func TestMatchesProfileRemoteAlwaysPasses(t *testing.T) {
	l := Listing{Location: "Seattle, WA", Remote: true}
	if !l.MatchesProfile(Profile{Location: "Boston"}) {
		t.Error("a remote listing should pass regardless of the profile's location")
	}
}

func TestMatchesProfileUnknownLocationPasses(t *testing.T) {
	l := Listing{} // Location unset
	if !l.MatchesProfile(Profile{Location: "Boston"}) {
		t.Error("a listing with no stated location should not be excluded")
	}
}

func TestMatchesProfileMajorNeverExcludes(t *testing.T) {
	l := Listing{FieldsOfStudy: []string{"biology"}}
	if !l.MatchesProfile(Profile{Major: "computer science"}) {
		t.Error("major should never hard-exclude a listing, only affect soft ranking")
	}
}