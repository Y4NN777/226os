import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useListView } from './useListView';

describe('useListView', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('defaults to cards', () => {
    const { result } = renderHook(() => useListView('build', undefined));
    expect(result.current.view).toBe('cards');
  });

  it('uses the stored choice when the address has none', () => {
    localStorage.setItem('226os.view.build', 'list');
    const { result } = renderHook(() => useListView('build', undefined));
    expect(result.current.view).toBe('list');
  });

  it('lets the address win over the stored choice', () => {
    localStorage.setItem('226os.view.build', 'list');
    const { result } = renderHook(() => useListView('build', 'cards'));
    expect(result.current.view).toBe('cards');
  });

  it('ignores a stored value it does not know', () => {
    localStorage.setItem('226os.view.build', 'grid');
    const { result } = renderHook(() => useListView('build', undefined));
    expect(result.current.view).toBe('cards');
  });

  it('remembers the choice per page', () => {
    const { result } = renderHook(() => useListView('discover', undefined));
    act(() => {
      result.current.remember('list');
    });
    expect(localStorage.getItem('226os.view.discover')).toBe('list');
    expect(localStorage.getItem('226os.view.build')).toBeNull();
  });
});
