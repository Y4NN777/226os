import { Card } from '@astryxdesign/core/Card';
import { VStack } from '@astryxdesign/core/VStack';
import { Outlet } from '@tanstack/react-router';
import { Brand } from '../components/Brand';

/**
 * Standalone account pages (prototype `layout-account`): no app chrome, the 226os logo centred
 * above one 400px card, a way home. The prototype's links under the card (privacy, code of
 * conduct) have no document yet, so there are none.
 */
export function AuthLayout() {
  return (
    <main className="auth-layout">
      <VStack gap={6} width="100%" maxWidth={400} hAlign="center">
        <Brand className="shell-brand auth-brand" />
        <Card padding={6} width="100%" className="auth-card">
          <Outlet />
        </Card>
      </VStack>
    </main>
  );
}
