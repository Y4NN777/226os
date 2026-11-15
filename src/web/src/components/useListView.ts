import { readStorage, writeStorage } from '../app/storage';

export type ListView = 'cards' | 'list';

/** The pages that offer the switch; each remembers its own choice. */
export type ViewPage = 'build' | 'discover' | 'learn';

/** A value from the address or from storage: anything but the two views is dropped. */
export function parseView(raw: unknown): ListView | undefined {
  return raw === 'cards' || raw === 'list' ? raw : undefined;
}

const storageKey = (page: ViewPage) => `226os.view.${page}`;

/**
 * The view a list page shows: the address first (a shared link keeps its view), then the choice
 * remembered in this browser, then cards. `remember` stores the choice; the page also writes it
 * to the address, so the two never disagree on screen.
 */
export function useListView(page: ViewPage, fromAddress: ListView | undefined) {
  const view = fromAddress ?? parseView(readStorage(storageKey(page))) ?? 'cards';
  const remember = (next: ListView) => {
    writeStorage(storageKey(page), next);
  };
  return { view, remember };
}
