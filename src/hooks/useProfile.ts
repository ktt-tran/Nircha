"use client";

import { useCallback, useEffect, useState } from "react";
import { EMPTY_PROFILE, type Profile } from "@/src/components/profile/profile.types";

const STORAGE_KEY = "nircha:profile";

/*
 * Reads/writes the user's profile from localStorage, shared between the
 * home page's profile tab (writes) and the search page (reads, to
 * include in every search). localStorage rather than a query param or a
 * shared Context: it's meant to persist across page navigations and
 * browser sessions without needing a shared layout wrapper around both
 * pages.
 */
export function useProfile() {
  const [profile, setProfileState] = useState<Profile>(EMPTY_PROFILE);
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    // window/localStorage don't exist during server rendering - this
    // effect only runs client-side after the initial render, which is
    // exactly when it's safe to read them. Reading it in the render body
    // instead (rather than an effect) would throw during Next.js's
    // server-rendered pass of this Client Component.
    try {
      const stored = window.localStorage.getItem(STORAGE_KEY);
      if (stored) {
        setProfileState({ ...EMPTY_PROFILE, ...JSON.parse(stored) });
      }
    } catch {
      // Corrupt JSON, storage disabled (private browsing), fall back to 
      // the empty profile rather than crash the page over a non-essential feature.
    } finally {
      setLoaded(true);
    }
  }, []);

  const saveProfile = useCallback((next: Profile) => {
    setProfileState(next);
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch {
      // Same reasoning as above, the profile still updates for the rest
      // of this session even if it can't be persisted for next time.
    }
  }, []);

  return { profile, saveProfile, loaded };
}