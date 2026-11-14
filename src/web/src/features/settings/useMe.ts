import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/http';
import { useSession } from '../../app/session';

/** Key root is not `session`, so forgetPrivateData drops it whenever the user changes. */
export const meQueryKey = ['me'] as const;

/** GET /me for the signed-in member; shared by the name form and the visibility setting. */
export function useMe() {
  const signedIn = useSession().data?.authenticated === true;
  return useQuery({
    queryKey: meQueryKey,
    queryFn: () => api.call(() => api.client.GET('/me')),
    enabled: signedIn,
    retry: false,
  });
}
