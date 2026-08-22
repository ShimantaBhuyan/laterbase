import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  listLinks,
  getLink,
  createLink,
  updateLink,
  deleteLink,
  reprocessLink,
} from '../api/client';

export function useLinks(params?: {
  group_id?: string;
  status?: string;
  search?: string;
  sort?: string;
  order?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: ['links', params],
    queryFn: () => listLinks(params),
  });
}

export function useLink(id: string) {
  return useQuery({
    queryKey: ['links', id],
    queryFn: () => getLink(id),
    enabled: !!id,
  });
}

export function useCreateLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (url: string) => createLink(url),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['links'] });
      queryClient.invalidateQueries({ queryKey: ['stats'] });
    },
  });
}

export function useUpdateLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { title?: string; source_type?: string; group_ids?: string[]; notes?: string; is_interesting?: boolean } }) =>
      updateLink(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['links'] });
      queryClient.invalidateQueries({ queryKey: ['stats'] });
    },
  });
}

export function useDeleteLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteLink(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['links'] });
      queryClient.invalidateQueries({ queryKey: ['stats'] });
    },
  });
}

export function useReprocessLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => reprocessLink(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['links'] });
      queryClient.invalidateQueries({ queryKey: ['links', id] });
    },
  });
}
