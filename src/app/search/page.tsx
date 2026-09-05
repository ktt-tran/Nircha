"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import styles from "../../styles/search.module.css";
import { SearchResult } from "@/src/components/search/search.types";
import { FaSearch } from "react-icons/fa";
import { searchOpportunities, SearchRequestError } from "@/src/lib/searchEngine";
import { useProfile } from "@/src/hooks/useProfile";

function badgeLabel(result: SearchResult): string {
  return (
    result.fieldsOfStudy?.[0] ??
    result.degreeLevels?.[0] ??
    (result.remote ? "Remote" : "Opportunity")
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={null}>
      <SearchPageContent />
    </Suspense>
  );
}

function SearchPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialQuery = searchParams.get("q") ?? "";
  const [query, setQuery] = useState(initialQuery);
  const { profile, loaded: profileLoaded } = useProfile();
  const [results, setResults] = useState<SearchResult[]>([]);
  const [totalResults, setTotalResults] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!profileLoaded) return;

    const trimmed = initialQuery.trim();
    if (!trimmed) {
      setResults([]);
      setTotalResults(0);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);

    searchOpportunities(trimmed, profile)
      .then((data) => {
        if (cancelled) return;
        setResults(data.pages);
        setTotalResults(data.totalResults);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(
          err instanceof SearchRequestError
            ? err.message
            : "Something went wrong reaching the search service.",
        );
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    // Cleanup guards against a slow, stale request.
    return () => {
      cancelled = true;
    };
  }, [initialQuery, profile, profileLoaded]);

  const handleSearch: React.SubmitEventHandler<
    HTMLFormElement
  > = (event) => {
    event.preventDefault();

    const trimmedQuery = query.trim();

    if (!trimmedQuery) {
      return;
    }

    router.push(
      `/search?q=${encodeURIComponent(trimmedQuery)}`
    );
  };

  return (
    <main className={styles.resultsPage}>
      <header className={styles.header}>
        <button
          className={styles.brand}
          type="button"
          onClick={() => router.push("/")}
        >
          Nircha
        </button>

        <form
          className={styles.headerSearch}
          onSubmit={handleSearch}
        >
          <div className={styles.searchBox}>
            <input
              type="text"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Where to?"
              aria-label="Search Nircha"
            />

            <button
              type="submit"
              className={styles.searchButton}
              aria-label="Search"
            >
              <FaSearch />
            </button>
          </div>
        </form>

        <div className={styles.headerLogo}>
          <img 
            src='ICON.svg'
            alt="Icon" 
            style={{ width: '90px', height: '90px' }} 
          />
        </div>
      </header>

      <section className={styles.resultsContainer}>
        <div className={styles.resultsHeading}>
          <p className={styles.resultsLabel}>
            Search results
          </p>

          <h1>
            Results for{" "}
            <span>
              &quot;{initialQuery}&quot;
            </span>
          </h1>

          <p className={styles.resultsCount}>
            {loading
              ? "Searching…"
              : error
                ? "We couldn't complete this search."
                : `${totalResults} result${totalResults === 1 ? "" : "s"} related to your search`}
          </p>
        </div>

        {error && (
          <p className={styles.resultsCount} role="alert">
            {error}
          </p>
        )}

        {!loading && !error && results.length === 0 && (
          <p className={styles.resultsCount}>
            No matches. If a profile is saved, try loosening the degree
            level or location - they exclude listings, major only
            re-ranks.
          </p>
        )}

        <div className={styles.resultsList}>
          {results.map((result) => (
            <article
              className={styles.resultCard}
              key={result.url}
            >
              <span className={styles.resultType}>
                {badgeLabel(result)}
              </span>

              <h2>{result.title}</h2>

              <p className={styles.organization}>
                {result.organization}
              </p>

              <p className={styles.resultDescription}>
                {result.abstract}
              </p>

              <a
                href={result.url}
                target="_blank"
                rel="noopener noreferrer"
                className={styles.resultButton}
              >
                View
              </a>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}