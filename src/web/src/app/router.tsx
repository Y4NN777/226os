import {
  createRootRoute,
  createRoute,
  createRouter,
  type RouterHistory,
} from '@tanstack/react-router';
import { HomePage } from '../features/health/HomePage';
import { NotFound } from '../features/NotFound';
import { SettingsPage } from '../features/settings/SettingsPage';
import { Shell } from './Shell';

// Code-based routing: fully typed, no route-tree generator or router plugin, so no
// generated file lands in src (and none needs a coverage exclusion).
const rootRoute = createRootRoute({ component: Shell, notFoundComponent: NotFound });
const homeRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: HomePage });
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
});

const routeTree = rootRoute.addChildren([homeRoute, settingsRoute]);

export function createAppRouter(options?: { history?: RouterHistory }) {
  return createRouter({ routeTree, ...(options?.history ? { history: options.history } : {}) });
}

export type AppRouter = ReturnType<typeof createAppRouter>;

declare module '@tanstack/react-router' {
  interface Register {
    router: AppRouter;
  }
}
