export interface SearchResult {
  title: string;
  abstract: string;
  url: string;
  organization: string;
  trustTier: number; // 1 = most trusted, higher = less trusted.
  postedAt: string;
  lastCrawledAt: string;
  deadlineAt?: string;
  fieldsOfStudy?: string[];
  degreeLevels?: string[];
  location?: string;
  remote: boolean;
  score: number; // This specific query+profile's score. Computed fresh per request.
}

export interface SearchResponse {
  query: string;
  pages: SearchResult[];
  totalResults: number;
  page: number;
  pageSize: number;
  pagesCount: number;
  timeMs: number;
}