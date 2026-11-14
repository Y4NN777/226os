import { screen, waitFor, type Matcher } from '@testing-library/react';

/**
 * Astryx may also announce a message in its aria-live status region, depending on timing. These
 * helpers match the visible copy only and still require exactly one element, as getByText does.
 */
export function getVisibleText(matcher: Matcher): HTMLElement {
  const visible = screen
    .getAllByText(matcher)
    .filter((element) => !element.closest('[data-astryx-live-region]'));
  if (visible.length !== 1) {
    throw new Error(
      `Expected exactly one visible element for ${String(matcher)}, found ${String(visible.length)}`,
    );
  }
  return visible[0] as HTMLElement;
}

export async function findVisibleText(matcher: Matcher): Promise<HTMLElement> {
  return waitFor(() => getVisibleText(matcher));
}
