import { describe, expect, it, vi } from 'vitest';
import { readStorage, removeLegacyDisplayName, writeStorage } from './storage';

describe('storage', () => {
  it('round-trips a value', () => {
    writeStorage('k', 'v');
    expect(readStorage('k')).toBe('v');
    expect(readStorage('missing')).toBeNull();
  });

  it('never throws when storage is unavailable (private mode, quota)', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('denied');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota');
    });
    expect(readStorage('k')).toBeNull();
    expect(() => {
      writeStorage('k', 'v');
    }).not.toThrow();
  });
});

describe('removeLegacyDisplayName', () => {
  it('deletes the display name an earlier version kept on this device', () => {
    localStorage.setItem('226os.displayName', 'Awa');
    removeLegacyDisplayName();
    expect(localStorage.getItem('226os.displayName')).toBeNull();
  });

  it('does nothing when the key is absent, and never throws when storage is unavailable', () => {
    expect(() => {
      removeLegacyDisplayName();
    }).not.toThrow();
    vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
      throw new Error('denied');
    });
    expect(() => {
      removeLegacyDisplayName();
    }).not.toThrow();
  });
});
