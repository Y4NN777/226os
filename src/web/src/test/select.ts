import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

/**
 * Picks one option of an Astryx Selector. Two things keep it fast, which matters with the 54
 * countries: the trigger is opened by keyboard (focus, Enter) so the list is in the document at
 * once instead of waiting for the popover to settle, and the option is looked up inside the
 * listbox by its text, not by an accessible-name query over the whole page (about 600 ms per pick).
 */
export async function selectOption(trigger: HTMLElement, option: string): Promise<void> {
  const user = userEvent.setup({ delay: null });
  trigger.focus();
  await user.keyboard('{Enter}');
  const item = within(screen.getByRole('listbox')).getByText(option, {
    selector: '[role="option"], [role="option"] *',
  });
  await user.click(item);
}
