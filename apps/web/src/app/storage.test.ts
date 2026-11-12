import { describe, expect, it, vi } from 'vitest';
import { readStorage, writeStorage } from './storage';

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
