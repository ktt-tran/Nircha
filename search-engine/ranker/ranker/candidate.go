package ranker

import "search-engine/ranker/document"

type Candidate struct {
	Listing   document.Listing
	BM25Score float64
}