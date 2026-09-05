export interface Profile {
  major: string;
  degreeLevel: string;
  location: string;
}

export const EMPTY_PROFILE: Profile = {
  major: "",
  degreeLevel: "",
  location: "",
};

export function isProfileEmpty(profile: Profile): boolean {
  return !profile.major && !profile.degreeLevel && !profile.location;
}