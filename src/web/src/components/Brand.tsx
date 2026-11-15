import { Link } from '@tanstack/react-router';
import logoDark from '../assets/brand/226os-logo-white.png';
import logoLight from '../assets/brand/226os-logo.png';
import { useResolvedTheme } from '../theme/ThemeProvider';

/**
 * The 226os logo, linking home (prototype `.brand`): the white variant on the dark theme. Used by
 * the app shell's top bar and drawer and above the account pages' card.
 */
export function Brand({ className = 'shell-brand' }: { className?: string }) {
  const theme = useResolvedTheme();
  return (
    <Link to="/" className={className}>
      <img
        src={theme === 'dark' ? logoDark : logoLight}
        alt="226os"
        width={92}
        height={32}
        className="shell-brand-logo"
      />
    </Link>
  );
}
