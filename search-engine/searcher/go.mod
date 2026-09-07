module search-engine/searcher

go 1.23

require (
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.17.3
	github.com/tebeka/snowball v0.8.0
	search-engine/ranker v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace search-engine/ranker => ../ranker
