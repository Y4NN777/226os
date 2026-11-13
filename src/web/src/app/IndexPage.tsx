import { LoadingState } from '../components/LoadingState';
import { HomePage } from '../features/home/HomePage';
import { LandingPage } from '../features/landing/LandingPage';
import { useSession } from './session';

/**
 * `/`: the public landing page for visitors, the personal home (accueil) for members.
 * Nothing is chosen until the session is known, so a member never sees the landing flash.
 * If the session cannot be read, the visitor view is the safe default.
 */
export function IndexPage() {
  const session = useSession();
  if (session.isPending) return <LoadingState />;
  return session.data?.authenticated ? <HomePage /> : <LandingPage />;
}
