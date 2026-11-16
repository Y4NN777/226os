/** localStorage that never throws (private mode, disabled storage, quota). */
export function readStorage(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function writeStorage(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    // Persistence is best effort; the in-memory state stays authoritative.
  }
}

const LEGACY_DISPLAY_NAME_KEY = '226os.displayName';

/**
 * Earlier versions kept the display name in this browser. The name now lives on the account
 * only, so the old copy is deleted: it must not outlive a sign-out on a shared device.
 */
export function removeLegacyDisplayName(): void {
  try {
    localStorage.removeItem(LEGACY_DISPLAY_NAME_KEY);
  } catch {
    // Nothing to clean when storage is unavailable.
  }
}
