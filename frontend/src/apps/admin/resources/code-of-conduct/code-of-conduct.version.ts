const versionPattern = /^(\d+)\.(\d+)\.(\d+)$/;

export const validateVersion = (value: string) => {
  if (!value) return "A version is required";
  if (!versionPattern.test(value)) return "Use a semantic version like 1.0.0";

  return undefined;
};

// nextPatchVersion derives the default version of a new draft from the last
// published one: 1.2.3 -> 1.2.4.
export const nextPatchVersion = (version?: string) => {
  const match = versionPattern.exec(version ?? "");
  if (!match) return "1.0.0";

  return `${match[1]}.${match[2]}.${Number(match[3]) + 1}`;
};
