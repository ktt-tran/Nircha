import type { SearchResponse } from "@/src/components/search/search.types";
import type { Profile } from "@/src/components/profile/profile.types";

const SEARCH_API_URL = process.env.NEXT_PUBLIC_SEARCH_API_URL;

export class SearchRequestError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "SearchRequestError";
  }
}

export interface SearchOptions {
  page?: number;
  size?: number;
}

export async function searchOpportunities(
  query: string,
  profile: Profile,
  options: SearchOptions = {},
): Promise<SearchResponse> {
  const params = new URLSearchParams({ q: query });
  if (profile.major) params.set("major", profile.major);
  if (profile.degreeLevel) params.set("degree", profile.degreeLevel);
  if (profile.location) params.set("location", profile.location);
  if (options.page) params.set("page", String(options.page));
  if (options.size) params.set("size", String(options.size));

  const res = await fetch(`${SEARCH_API_URL}/api/search?${params.toString()}`);

  if (!res.ok) {
    throw new SearchRequestError(
      `search request failed: ${res.status} ${res.statusText}`,
      res.status,
    );
  }

  return (await res.json()) as SearchResponse;
}