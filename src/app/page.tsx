"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FaSearch } from "react-icons/fa";
import styles from "../styles/home.module.css";
import ProfileTab from "../components/profile/ProfileTab";

export default function HomePage() {
  const router = useRouter();
  const [query, setQuery] = useState("");

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
    <main className={styles.home}>
      <ProfileTab />

      {/* HERO */}
      <section className={styles.hero}>
        <div className={styles.heroContent}>
          <h1>Here to Support your Career</h1>

          <div className={styles.searchArea}>
            <div className={styles.searchTitle}>
              <h1>Nircha Career Driven Search</h1>
                <img 
                  src="ICON.svg"
                  alt="Icon" 
                  style={{ width: '90px', height: '90px' }} 
                />
            </div>

            <form onSubmit={handleSearch}>
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
          </div>
        </div>
      </section>

      {/* DESCRIPTION */}
      <section className={styles.description}>
        <div className={styles.descriptionContent}>
          <h2>Nircha is here for you</h2>

          <p>
            Nircha is a full-text search engine built to serve people
            navigating one of the most consequential and least well-organized
            parts of their lives: figuring out what comes next in education and
            career. Unlike a conventional search engine: crawling sources,
            indexing content, and ranking results, but its distinguishing
            feature is a purpose-built vertical layered on top of general
            search: a dedicated index of university programs, internships, and
            research opportunities, ranked not just by relevance but by trust,
            timeliness, and fit. Where a general search engine treats a college
            scholarship page the same as a random blog post, this project
            treats it as a time-sensitive, high-stakes opportunity that
            deserves its own signals for quality and relevance.
          </p>

          <p>
            For students, that distinction matters in practice. Someone
            transitioning from high school into college often doesn&apos;t
            know what to search for, let alone which results to trust; someone
            looking for research experience or an internship is used to racing
            a deadline they may not even know exists yet. By combining
            relevance ranking with a curated sense of source trustworthiness
            and freshness, the engine surfaces legitimate opportunities for
            beginners, research programs, and career-focused searches while
            filtering out stale or unreliable listings that saves a student
            real time and prevents them from applying to something that&apos;s
            already closed. The goal isn&apos;t just to return more results,
            but to return the right results at the moment someone actually
            needs to act on them.
          </p>

          <p>
            Longer term, the ambition is broader than any one group of
            students. Career transitions aren&apos;t unique to eighteen-year-olds
            heading to college, they happen to career-changers, returning
            adults, first-generation students without a built-in support
            network, and anyone else trying to make sense of opportunities that
            are scattered across thousands of disconnected websites. The hope
            is for this project to grow into a resource that treats good
            information about education and career paths as something everyone
            deserves easy access to, regardless of what school they went to,
            what network they have, or how well they already know how to
            search.
          </p>
        </div>
      </section>
    </main>
  );
}