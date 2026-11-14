import { within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

/**
 * Selects one segment of an Astryx SegmentedControl by keyboard (APG radiogroup: selection follows
 * focus). A pointer click is not deterministic in jsdom: Astryx commits a press to the segment
 * nearest the pointer, and without layout that choice races with the click and can pick the first
 * segment under load.
 */
export async function selectSegment(radio: HTMLElement): Promise<void> {
  const group = radio.closest<HTMLElement>('[role="radiogroup"]');
  if (!group) throw new Error('The segment is not inside a radiogroup');
  const radios = within(group).getAllByRole('radio');
  const from = radios.findIndex((item) => item.getAttribute('aria-checked') === 'true');
  const to = radios.indexOf(radio);
  if (to < 0 || from === to) return;
  (radios[from >= 0 ? from : 0] as HTMLElement).focus();
  const key = to > from ? '{ArrowRight}' : '{ArrowLeft}';
  await userEvent.setup({ delay: null }).keyboard(key.repeat(Math.abs(to - Math.max(from, 0))));
}
