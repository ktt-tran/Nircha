package repository

import "search-engine/crawler/webpage"

type Repository interface {
	AddPage(page *webpage.WebPage) (bool, error)
	EnqueuePage(url string) (bool, error)
	DequeuePage() (string, error)
	GetPagesCount() int64
}