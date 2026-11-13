import { LinkProvider } from '@astryxdesign/core/Link';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  type RouterHistory,
} from '@tanstack/react-router';
import { RouterAnchor } from '../components/RouterAnchor';
import { DiscoverPage } from '../features/discover/DiscoverPage';
import { ResourcePage } from '../features/discover/ResourcePage';
import { validateDiscoverSearch } from '../features/discover/routeSearch';
import { ForgotPasswordPage } from '../features/identity/ForgotPasswordPage';
import { ResetPasswordPage } from '../features/identity/ResetPasswordPage';
import { validateSignInSearch } from '../features/identity/routeSearch';
import { SignInPage } from '../features/identity/SignInPage';
import { SignUpPage } from '../features/identity/SignUpPage';
import { VerifyEmailPage } from '../features/identity/VerifyEmailPage';
import { WelcomePage } from '../features/identity/WelcomePage';
import { NotFound } from '../features/NotFound';
import { validateSearchSearch } from '../features/search/routeSearch';
import { SearchPage } from '../features/search/SearchPage';
import { SettingsPage } from '../features/settings/SettingsPage';
import { AuthLayout } from './AuthLayout';
import { IndexPage } from './IndexPage';
import { Shell } from './Shell';

function Root() {
  return (
    <LinkProvider component={RouterAnchor}>
      <Outlet />
    </LinkProvider>
  );
}

// Code-based routing: fully typed, no route-tree generator or router plugin, so no
// generated file lands in src (and none needs a coverage exclusion).
const rootRoute = createRootRoute({ component: Root, notFoundComponent: NotFound });

// Pages inside the app chrome.
const appLayout = createRoute({ getParentRoute: () => rootRoute, id: 'app', component: Shell });
const homeRoute = createRoute({ getParentRoute: () => appLayout, path: '/', component: IndexPage });
const settingsRoute = createRoute({
  getParentRoute: () => appLayout,
  path: '/settings',
  component: SettingsPage,
});
const discoverRoute = createRoute({
  getParentRoute: () => appLayout,
  path: '/discover',
  component: DiscoverPage,
  validateSearch: validateDiscoverSearch,
});
const resourceRoute = createRoute({
  getParentRoute: () => appLayout,
  path: '/discover/$resourceId',
  component: ResourcePage,
});
const searchRoute = createRoute({
  getParentRoute: () => appLayout,
  path: '/recherche',
  component: SearchPage,
  validateSearch: validateSearchSearch,
});

// Standalone account pages (no app chrome).
const authLayout = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  component: AuthLayout,
});
const signInRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/connexion',
  component: SignInPage,
  validateSearch: validateSignInSearch,
});
const signUpRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/inscription',
  component: SignUpPage,
});
const verifyEmailRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/verifier-email',
  component: VerifyEmailPage,
});
const forgotPasswordRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/mot-de-passe-oublie',
  component: ForgotPasswordPage,
});
const resetPasswordRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/reinitialiser',
  component: ResetPasswordPage,
});
const welcomeRoute = createRoute({
  getParentRoute: () => authLayout,
  path: '/bienvenue',
  component: WelcomePage,
});

const routeTree = rootRoute.addChildren([
  appLayout.addChildren([homeRoute, settingsRoute, discoverRoute, resourceRoute, searchRoute]),
  authLayout.addChildren([
    signInRoute,
    signUpRoute,
    verifyEmailRoute,
    forgotPasswordRoute,
    resetPasswordRoute,
    welcomeRoute,
  ]),
]);

export function createAppRouter(options?: { history?: RouterHistory }) {
  return createRouter({ routeTree, ...(options?.history ? { history: options.history } : {}) });
}

export type AppRouter = ReturnType<typeof createAppRouter>;

declare module '@tanstack/react-router' {
  interface Register {
    router: AppRouter;
  }
}
