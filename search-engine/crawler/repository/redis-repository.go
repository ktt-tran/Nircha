package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"search-engine/crawler/webpage"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	rh    *redis.Client
	ctx   *context.Context
	mutex *sync.Mutex
}

var keyPrefixes = struct {
	Word            string
	Pages           string
	PageToId        string
	IdToPage        string
	OutboundLinks   string
	PageQueue       string
	VisitedPages    string
	WordFrequencies string
	NextPageId      string
}{
	Word:            "w:",
	Pages:           "pages",
	PageToId:        "p-i:",
	IdToPage:        "i-p:",
	OutboundLinks:   "out:",
	PageQueue:       "queue",
	VisitedPages:    "visited",
	WordFrequencies: "words-freq",
	NextPageId:      "next-page-id",
}

func (r *RedisRepository) AddPage(page *webpage.WebPage) (bool, error) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	pageId, err := r.getOrCreatePageId(page.URL)
	if err != nil {
		return false, err
	}

	alreadySaved := r.rh.HExists(*r.ctx, keyPrefixes.Pages, strconv.FormatInt(pageId, 10)).Val()
	if alreadySaved {
		return false, nil
	}

	if err := r.savePageData(pageId, page); err != nil {
		return false, err
	}

	if err := r.savePageLinks(pageId, page); err != nil {
		return false, err
	}

	if err := r.saveWordIndex(pageId, page); err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisRepository) GetPagesCount() int64 {
	return r.rh.HLen(*r.ctx, keyPrefixes.Pages).Val()
}

func (r *RedisRepository) savePageData(pageId int64, page *webpage.WebPage) error {

	info := map[string]interface{}{
		"url":      	 page.URL,
		"title":    	 page.Title,
		"abstract": 	 page.Abstract,
		"content": 	     page.Content,
		"organization":  page.Organization,
		"trustTier":     page.TrustTier,
		"postedAt":      page.PostedAt,
		"lastCrawledAt": page.LastCrawledAt,
		"deadlineAt":    page.DeadlineAt,
		"fieldsOfStudy": page.FieldsOfStudy,
		"degreeLevels":  page.DegreeLevels,
		"location":      page.Location,
		"remote":        page.Remote,
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return r.rh.HSet(*r.ctx, keyPrefixes.Pages, pageId, string(jsonData)).Err()
}

func (r *RedisRepository) savePageLinks(pageId int64, page *webpage.WebPage) error {
	if len(page.OutboundLinks) == 0 {
		return nil
	}

	key := fmt.Sprintf("%s%d", keyPrefixes.OutboundLinks, pageId)
	for _, link := range page.OutboundLinks {
		linkId, err := r.getOrCreatePageId(link)
		if err != nil {
			return err
		}
		if err := r.rh.HSet(*r.ctx, key, linkId, linkId).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisRepository) saveWordIndex(pageId int64, page *webpage.WebPage) error {
	if len(page.Words) == 0 {
		return nil
	}

	_, err := r.rh.Pipelined(*r.ctx, func(pipe redis.Pipeliner) error {
		for _, word := range page.Words {
			key := keyPrefixes.Word + word.Word
			scorePosition := fmt.Sprintf("%d:%d", word.Score, word.Position)
			pipe.HSet(*r.ctx, key, pageId, scorePosition)
			pipe.HIncrBy(*r.ctx, keyPrefixes.WordFrequencies, word.Word, 1)
		}
		return nil
	})
	return err
}

func (r *RedisRepository) getOrCreatePageId(url string) (int64, error) {

	if idStr, err := r.rh.HGet(*r.ctx, keyPrefixes.PageToId, url).Result(); err == nil {
		return strconv.ParseInt(idStr, 10, 64)
	} else if err != redis.Nil {
		return 0, err
	}

	newId, err := r.rh.Incr(*r.ctx, keyPrefixes.NextPageId).Result()
	if err != nil {
		return 0, err
	}

	won, err := r.rh.HSetNX(*r.ctx, keyPrefixes.PageToId, url, newId).Result()
	if err != nil {
		return 0, err
	}
	if !won {
		idStr, err := r.rh.HGet(*r.ctx, keyPrefixes.PageToId, url).Result()
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(idStr, 10, 64)
	}

	if err := r.rh.HSet(*r.ctx, keyPrefixes.IdToPage, newId, url).Err(); err != nil {
		return 0, err
	}

	return newId, nil
}

func (r *RedisRepository) EnqueuePage(url string) (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	alreadyVisited := r.rh.HExists(*r.ctx, keyPrefixes.VisitedPages, url).Val()
	if alreadyVisited {
		return false, nil
	}

	err := r.rh.LPush(*r.ctx, keyPrefixes.PageQueue, url).Err()
	if err != nil {
		return false, err
	}

	err = r.rh.HSet(*r.ctx, keyPrefixes.VisitedPages, url, true).Err()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisRepository) DequeuePage() (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	url, err := r.rh.RPop(*r.ctx, keyPrefixes.PageQueue).Result()
	if err != nil {
		return "", err
	}

	return url, nil
}

func NewRedisRepository() (*RedisRepository, error) {
	connectionAddress := os.Getenv("REDIS_CONNECTION_ADDRESS")
	password := os.Getenv("REDIS_CONNECTION_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     connectionAddress,
		Password: password,
		DB:       0,
	})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to redis at %q: %w", connectionAddress, err)
	}

	mu := new(sync.Mutex)
	return &RedisRepository{rdb, &ctx, mu}, nil
}