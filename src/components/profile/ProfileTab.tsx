"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import styles from "@/src/styles/profile.module.css";
import { useProfile } from "@/src/hooks/useProfile";
import { isProfileEmpty, type Profile } from "@/src/components/profile/profile.types";


export default function ProfileTab() {
  const { profile, saveProfile, loaded } = useProfile();
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<Profile>(profile);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      setDraft(profile);
    }
  }, [profile, open]);

  useEffect(() => {
    if (!open) return;

    function handleClickOutside(event: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [open]);

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    saveProfile(draft);
    setOpen(false);
  }

  function handleClear() {
    const cleared: Profile = { major: "", degreeLevel: "", location: "" };
    setDraft(cleared);
    saveProfile(cleared);
  }

  const hasProfile = loaded && !isProfileEmpty(profile);

  return (
    <div className={styles.container} ref={panelRef}>
      <button
        type="button"
        className={hasProfile ? styles.tabActive : styles.tab}
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="dialog"
        aria-expanded={open}
      >
        Profile{hasProfile && <span className={styles.dot} aria-hidden="true" />}
      </button>

      {open && (
        <div className={styles.panel} role="dialog" aria-label="Your profile">
          <p className={styles.panelHint}>
            Optional - narrows and re-ranks results for you. Degree level and
            location exclude clearly ineligible listings; major only
            re-ranks, it never hides results.
          </p>

          <form onSubmit={handleSubmit} className={styles.form}>
            <label className={styles.field}>
              <span>Major</span>
              <input
                type="text"
                value={draft.major}
                onChange={(e) => setDraft({ ...draft, major: e.target.value })}
                placeholder="e.g. Computer Science"
              />
            </label>

            <label className={styles.field}>
              <span>Degree level</span>
              <select
                value={draft.degreeLevel}
                onChange={(e) => setDraft({ ...draft, degreeLevel: e.target.value })}
              >
                <option value="">Any</option>
                <option value="undergraduate">Undergraduate</option>
                <option value="graduate">Graduate</option>
                <option value="phd">PhD</option>
              </select>
            </label>

            <label className={styles.field}>
              <span>Location</span>
              <input
                type="text"
                value={draft.location}
                onChange={(e) => setDraft({ ...draft, location: e.target.value })}
                placeholder="e.g. Boston"
              />
            </label>

            <div className={styles.actions}>
              <button type="button" className={styles.clearButton} onClick={handleClear}>
                Clear
              </button>
              <button type="submit" className={styles.saveButton}>
                Save
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}