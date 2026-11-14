import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, ApiError } from '../../api/http';
import type { components } from '../../api/schema';
import { meQueryKey } from '../settings/useMe';

export type PublicProfile = components['schemas']['PublicProfile'];
export type ProfileUpdate = components['schemas']['ProfileUpdate'];
export type MyProfile = components['schemas']['MyProfile'];

export const profileKeys = { detail: (username: string) => ['profiles', username] as const };

/** A hidden, private or unknown profile (404) and a malformed name (422) all read as "not found". */
export function isProfileMissing(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 404 || error.status === 422);
}

export function useProfile(username: string, enabled: boolean) {
  return useQuery({
    queryKey: profileKeys.detail(username),
    enabled,
    retry: (failures, error) =>
      failures < 1 && error instanceof ApiError && error.code === 'NETWORK',
    queryFn: () =>
      api.call(() => api.client.GET('/profiles/{username}', { params: { path: { username } } })),
  });
}

/** PATCH /me/profile for one checklist item; the answer replaces the cached own profile. */
export function useSaveProfileField(username: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: ProfileUpdate) => api.call(() => api.client.PATCH('/me/profile', { body })),
    onSuccess: (profile) => {
      queryClient.setQueryData(
        meQueryKey,
        (old: { account: unknown; profile: MyProfile } | undefined) =>
          old ? { ...old, profile } : old,
      );
      return queryClient.invalidateQueries({ queryKey: profileKeys.detail(username) });
    },
  });
}
