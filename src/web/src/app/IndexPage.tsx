import { HomePage } from '../features/home/HomePage';
import { LandingPage } from '../features/landing/LandingPage';
import { useSession } from './session';

/** `/`: the public landing page for visitors, the personal home (accueil) for members. */
export function IndexPage() {
  const session = useSession();
  return session.data?.authenticated ? <HomePage /> : <LandingPage />;
}
