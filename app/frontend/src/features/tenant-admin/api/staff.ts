import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { customInstance } from '@/lib/axios'

export interface StaffMember {
  id: string
  shop_id: string
  user_id: string
  role_id: string
  role_name: string
  is_owner: boolean
  joined_at: string
  email: string
  full_name: string
  phone: string
  status: 'active' | 'pending_invite' | 'banned' | string
  invite_token?: string
}

export interface ShopRole {
  id: string
  name: string
  is_system_role?: boolean
  permissions?: string[] | Record<string, unknown>
}

export interface InviteStaffPayload {
  email: string
  role_id: string
  full_name?: string
  phone?: string
}

export interface UpdateStaffRolePayload {
  id: string
  role_id: string
}

export interface Envelope<T> {
  data: T
  error?: string
}

export async function fetchStaffList(): Promise<StaffMember[]> {
  const res = await customInstance<Envelope<StaffMember[]>>({
    url: '/api/v1/staff',
    method: 'GET',
  })
  return res.data || []
}

export async function inviteStaff(payload: InviteStaffPayload): Promise<StaffMember> {
  const res = await customInstance<Envelope<StaffMember>>({
    url: '/api/v1/staff/invite',
    method: 'POST',
    data: payload,
  })
  return res.data
}

export async function updateStaffRole({ id, role_id }: UpdateStaffRolePayload): Promise<StaffMember> {
  const res = await customInstance<Envelope<StaffMember>>({
    url: `/api/v1/staff/${id}/role`,
    method: 'PATCH',
    data: { role_id },
  })
  return res.data
}

export async function removeStaff(id: string): Promise<void> {
  await customInstance<void>({
    url: `/api/v1/staff/${id}`,
    method: 'DELETE',
  })
}

export async function fetchRoles(): Promise<ShopRole[]> {
  const res = await customInstance<Envelope<ShopRole[]>>({
    url: '/api/v1/roles',
    method: 'GET',
  })
  return res.data || []
}

export const STAFF_QUERY_KEY = ['tenant-admin', 'staff']
export const ROLES_QUERY_KEY = ['tenant-admin', 'roles']

export function useStaffList() {
  return useQuery({
    queryKey: STAFF_QUERY_KEY,
    queryFn: fetchStaffList,
  })
}

export function useRolesList() {
  return useQuery({
    queryKey: ROLES_QUERY_KEY,
    queryFn: fetchRoles,
  })
}

export function useInviteStaff() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: inviteStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFF_QUERY_KEY })
    },
  })
}

export function useUpdateStaffRole() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateStaffRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFF_QUERY_KEY })
    },
  })
}

export function useRemoveStaff() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: removeStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFF_QUERY_KEY })
    },
  })
}
