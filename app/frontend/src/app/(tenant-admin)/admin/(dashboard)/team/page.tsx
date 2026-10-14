"use client"

import * as React from "react"
import {
  Users,
  UserPlus,
  ShieldCheck,
  Shield,
  Clock,
  Search,
  MoreHorizontal,
  Mail,
  User,
  Phone,
  Copy,
  Check,
  Trash2,
  AlertCircle,
  KeyRound,
  X,
  ChevronDown,
  RefreshCw,
} from "lucide-react"
import { useAuthStore } from "@/store/useAuthStore"
import {
  useStaffList,
  useRolesList,
  useInviteStaff,
  useUpdateStaffRole,
  useRemoveStaff,
  StaffMember,
} from "@/features/tenant-admin/api/staff"

function extractErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === "object" && "response" in err) {
    const res = (err as { response?: { data?: { error?: string } } }).response
    if (res?.data?.error) return res.data.error
  }
  if (err instanceof Error) return err.message
  return fallback
}

export default function TeamManagementPage() {
  const { user } = useAuthStore()

  const { data: staffMembers = [], isLoading: isLoadingStaff, refetch: refetchStaff } = useStaffList()
  const { data: roles = [] } = useRolesList()

  const inviteMutation = useInviteStaff()
  const updateRoleMutation = useUpdateStaffRole()
  const removeStaffMutation = useRemoveStaff()

  const [searchTerm, setSearchTerm] = React.useState("")
  const [selectedRoleFilter, setSelectedRoleFilter] = React.useState("all")
  const [selectedStatusFilter, setSelectedStatusFilter] = React.useState("all")

  const [isInviteModalOpen, setIsInviteModalOpen] = React.useState(false)
  const [isChangeRoleModalOpen, setIsChangeRoleModalOpen] = React.useState(false)
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState(false)

  const [activeMember, setActiveMember] = React.useState<StaffMember | null>(null)
  const [selectedNewRoleId, setSelectedNewRoleId] = React.useState("")

  const [inviteEmail, setInviteEmail] = React.useState("")
  const [inviteName, setInviteName] = React.useState("")
  const [invitePhone, setInvitePhone] = React.useState("")
  const [inviteRoleId, setInviteRoleId] = React.useState("")
  const [inviteError, setInviteError] = React.useState<string | null>(null)
  const [inviteSuccessResult, setInviteSuccessResult] = React.useState<StaffMember | null>(null)

  const [copiedLink, setCopiedLink] = React.useState(false)
  const [actionMenuOpenId, setActionMenuOpenId] = React.useState<string | null>(null)

  const defaultRoleId = React.useMemo(() => {
    return roles.find((r) => r.name.toLowerCase() !== "owner")?.id || roles[0]?.id || ""
  }, [roles])

  const effectiveInviteRoleId = inviteRoleId || defaultRoleId

  const resetInviteForm = () => {
    setInviteEmail("")
    setInviteName("")
    setInvitePhone("")
    setInviteRoleId("")
    setInviteError(null)
    setInviteSuccessResult(null)
    setCopiedLink(false)
  }

  const filteredMembers = React.useMemo(() => {
    return staffMembers.filter((member) => {
      const matchesSearch =
        (member.full_name || "").toLowerCase().includes(searchTerm.toLowerCase()) ||
        member.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (member.role_name || "").toLowerCase().includes(searchTerm.toLowerCase())

      const matchesRole =
        selectedRoleFilter === "all" ||
        member.role_id === selectedRoleFilter ||
        (selectedRoleFilter === "owner" && member.is_owner)

      const matchesStatus =
        selectedStatusFilter === "all" ||
        (selectedStatusFilter === "active" && member.status === "active") ||
        (selectedStatusFilter === "pending" && member.status === "pending_invite")

      return matchesSearch && matchesRole && matchesStatus
    })
  }, [staffMembers, searchTerm, selectedRoleFilter, selectedStatusFilter])

  const totalCount = staffMembers.length
  const activeCount = staffMembers.filter((m) => m.status === "active").length
  const pendingCount = staffMembers.filter((m) => m.status === "pending_invite").length
  const rolesCount = roles.length

  const handleOpenInvite = () => {
    resetInviteForm()
    setIsInviteModalOpen(true)
  }

  const handleSubmitInvite = async (e: React.FormEvent) => {
    e.preventDefault()
    setInviteError(null)

    if (!inviteEmail.trim()) {
      setInviteError("Please enter a valid email address.")
      return
    }
    if (!effectiveInviteRoleId) {
      setInviteError("Please select a role for this team member.")
      return
    }

    try {
      const inviteResponse = await inviteMutation.mutateAsync({
        email: inviteEmail.trim(),
        role_id: effectiveInviteRoleId,
        full_name: inviteName.trim() || undefined,
        phone: invitePhone.trim() || undefined,
      })
      setInviteSuccessResult(inviteResponse)
    } catch (err: unknown) {
      setInviteError(
        extractErrorMessage(err, "Failed to send invite. Please verify the role and email.")
      )
    }
  }

  const handleOpenChangeRole = (member: StaffMember) => {
    setActiveMember(member)
    setSelectedNewRoleId(member.role_id)
    setIsChangeRoleModalOpen(true)
    setActionMenuOpenId(null)
  }

  const handleSubmitRoleChange = async () => {
    if (!activeMember || !selectedNewRoleId) return
    try {
      await updateRoleMutation.mutateAsync({
        id: activeMember.id,
        role_id: selectedNewRoleId,
      })
      setIsChangeRoleModalOpen(false)
      setActiveMember(null)
    } catch (err: unknown) {
      alert(extractErrorMessage(err, "Failed to update role"))
    }
  }

  const handleOpenDelete = (member: StaffMember) => {
    setActiveMember(member)
    setIsDeleteModalOpen(true)
    setActionMenuOpenId(null)
  }

  const handleConfirmDelete = async () => {
    if (!activeMember) return
    try {
      await removeStaffMutation.mutateAsync(activeMember.id)
      setIsDeleteModalOpen(false)
      setActiveMember(null)
    } catch (err: unknown) {
      alert(extractErrorMessage(err, "Failed to remove staff member"))
    }
  }

  const handleCopyInviteLink = (token?: string) => {
    if (!token) return
    const inviteUrl = `${window.location.origin}/admin/invite?token=${token}`
    navigator.clipboard.writeText(inviteUrl)
    setCopiedLink(true)
    setTimeout(() => setCopiedLink(false), 2000)
  }

  return (
    <div className="flex flex-col gap-8 w-full max-w-6xl pb-16">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full border border-violet-500/20 bg-violet-500/10 px-2.5 py-0.5 text-xs font-medium text-violet-400">
              <ShieldCheck className="w-3.5 h-3.5" />
              Access & Governance
            </span>
          </div>
          <h1 className="text-3xl font-semibold tracking-tight text-zinc-50">Team Management</h1>
          <p className="text-zinc-400 text-sm">
            Manage your store&apos;s operators, assign granular permissions, and invite collaborators.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => refetchStaff()}
            title="Refresh list"
            className="flex h-10 w-10 items-center justify-center rounded-full border border-zinc-800 bg-zinc-950 text-zinc-400 hover:text-zinc-100 hover:border-zinc-700 transition-all active:scale-95"
          >
            <RefreshCw className={`w-4 h-4 ${isLoadingStaff ? "animate-spin text-zinc-200" : ""}`} />
          </button>
          <button
            onClick={handleOpenInvite}
            className="group inline-flex h-10 items-center justify-center gap-2 rounded-full bg-zinc-100 px-5 text-sm font-semibold text-zinc-900 transition-all hover:bg-white active:scale-95 shadow-[0_0_20px_rgba(255,255,255,0.15)]"
          >
            <UserPlus className="w-4 h-4 transition-transform group-hover:scale-110" />
            Invite Member
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
          <div className="flex flex-col justify-between p-5 rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] h-full">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium uppercase tracking-wider text-zinc-400">Total Members</span>
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-zinc-800/80 text-zinc-300">
                <Users className="w-4 h-4" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-semibold text-zinc-100">{totalCount}</span>
              <span className="text-xs text-zinc-500">accounts</span>
            </div>
          </div>
        </div>

        <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
          <div className="flex flex-col justify-between p-5 rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] h-full">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium uppercase tracking-wider text-zinc-400">Active Staff</span>
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <ShieldCheck className="w-4 h-4" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-semibold text-emerald-400">{activeCount}</span>
              <span className="text-xs text-zinc-500">operational</span>
            </div>
          </div>
        </div>

        <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
          <div className="flex flex-col justify-between p-5 rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] h-full">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium uppercase tracking-wider text-zinc-400">Pending Invites</span>
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <Clock className="w-4 h-4" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-semibold text-amber-400">{pendingCount}</span>
              <span className="text-xs text-zinc-500">awaiting response</span>
            </div>
          </div>
        </div>

        <div className="p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
          <div className="flex flex-col justify-between p-5 rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] h-full">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium uppercase tracking-wider text-zinc-400">Shop Roles</span>
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20">
                <KeyRound className="w-4 h-4" />
              </div>
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-semibold text-zinc-100">{rolesCount}</span>
              <span className="text-xs text-zinc-500">configured</span>
            </div>
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-4 p-1.5 rounded-3xl border border-white/[0.05] bg-white/[0.02]">
        <div className="rounded-2xl bg-zinc-900/50 shadow-[inset_0_1px_1px_rgba(255,255,255,0.05)] overflow-hidden">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-4 border-b border-white/[0.05]">
            <div className="flex flex-wrap items-center gap-3">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                <input
                  type="text"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  placeholder="Search by name, email, or role..."
                  className="h-9 w-64 sm:w-72 rounded-full bg-zinc-950 border border-zinc-800 pl-9 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-700 transition-all placeholder:text-zinc-600"
                />
              </div>

              <div className="relative flex items-center">
                <select
                  value={selectedRoleFilter}
                  onChange={(e) => setSelectedRoleFilter(e.target.value)}
                  className="h-9 appearance-none rounded-full bg-zinc-950 border border-zinc-800 pl-3 pr-8 text-xs font-medium text-zinc-300 focus:outline-none focus:ring-1 focus:ring-zinc-700 cursor-pointer"
                >
                  <option value="all">All Roles</option>
                  <option value="owner">Owner</option>
                  {roles.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
                <ChevronDown className="pointer-events-none absolute right-3 w-3.5 h-3.5 text-zinc-500" />
              </div>

              <div className="relative flex items-center">
                <select
                  value={selectedStatusFilter}
                  onChange={(e) => setSelectedStatusFilter(e.target.value)}
                  className="h-9 appearance-none rounded-full bg-zinc-950 border border-zinc-800 pl-3 pr-8 text-xs font-medium text-zinc-300 focus:outline-none focus:ring-1 focus:ring-zinc-700 cursor-pointer"
                >
                  <option value="all">All Statuses</option>
                  <option value="active">Active Only</option>
                  <option value="pending">Pending Invites</option>
                </select>
                <ChevronDown className="pointer-events-none absolute right-3 w-3.5 h-3.5 text-zinc-500" />
              </div>
            </div>

            <div className="text-xs text-zinc-500 self-center sm:self-auto">
              Showing <span className="font-medium text-zinc-300">{filteredMembers.length}</span> of{" "}
              <span className="font-medium text-zinc-300">{staffMembers.length}</span> members
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm whitespace-nowrap">
              <thead className="bg-zinc-950/60 text-zinc-500 text-xs uppercase tracking-wider">
                <tr>
                  <th className="px-6 py-3.5 font-medium">Team Member</th>
                  <th className="px-6 py-3.5 font-medium">Assigned Role</th>
                  <th className="px-6 py-3.5 font-medium">Status</th>
                  <th className="px-6 py-3.5 font-medium">Joined</th>
                  <th className="px-6 py-3.5 font-medium text-right pr-6 w-16">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/[0.05]">
                {isLoadingStaff ? (
                  [1, 2, 3, 4].map((i) => (
                    <tr key={i}>
                      <td colSpan={5} className="px-6 py-4">
                        <div className="h-12 bg-zinc-800/40 animate-pulse rounded-xl" />
                      </td>
                    </tr>
                  ))
                ) : filteredMembers.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-16 text-center">
                      <div className="flex flex-col items-center justify-center gap-3">
                        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-zinc-800/60 text-zinc-500">
                          <Users className="w-6 h-6" />
                        </div>
                        <p className="text-zinc-300 font-medium text-sm">No team members match your filter.</p>
                        <p className="text-zinc-500 text-xs max-w-sm">
                          Invite a colleague with specific role permissions to start delegating shop management.
                        </p>
                        <button
                          onClick={handleOpenInvite}
                          className="mt-2 inline-flex items-center gap-2 rounded-full bg-zinc-100 px-4 py-2 text-xs font-semibold text-zinc-900 hover:bg-white transition-all"
                        >
                          <UserPlus className="w-3.5 h-3.5" />
                          Invite Member
                        </button>
                      </div>
                    </td>
                  </tr>
                ) : (
                  filteredMembers.map((member) => {
                    const isSelf = user?.email && member.email.toLowerCase() === user.email.toLowerCase()
                    const initials = (member.full_name || member.email)
                      .substring(0, 2)
                      .toUpperCase()

                    return (
                      <tr key={member.id} className="hover:bg-white/[0.02] transition-colors group">
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-zinc-700 to-zinc-800 text-zinc-200 text-xs font-semibold border border-white/10 shadow-inner">
                              {initials}
                              {member.is_owner && (
                                <span
                                  title="Store Owner"
                                  className="absolute -bottom-0.5 -right-0.5 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-amber-500 text-[9px] text-zinc-950 font-bold ring-2 ring-zinc-950"
                                >
                                  ★
                                </span>
                              )}
                            </div>
                            <div className="flex flex-col">
                              <div className="flex items-center gap-2">
                                <span className="font-medium text-zinc-100">
                                  {member.full_name || "Invited User"}
                                </span>
                                {isSelf && (
                                  <span className="rounded-full bg-zinc-800 px-2 py-0.5 text-[10px] font-semibold text-zinc-300 border border-zinc-700">
                                    You
                                  </span>
                                )}
                              </div>
                              <span className="text-zinc-400 text-xs font-mono">{member.email}</span>
                            </div>
                          </div>
                        </td>

                        <td className="px-6 py-4">
                          {member.is_owner ? (
                            <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-300 border border-amber-500/20">
                              <Shield className="w-3 h-3 text-amber-400" />
                              Store Owner
                            </span>
                          ) : (
                            <div className="inline-flex items-center gap-1.5">
                              <span className="inline-flex items-center gap-1.5 rounded-full bg-zinc-800/80 px-3 py-1 text-xs font-medium text-zinc-200 border border-zinc-700/60">
                                <KeyRound className="w-3 h-3 text-zinc-400" />
                                {member.role_name || "Staff Member"}
                              </span>
                              <button
                                onClick={() => handleOpenChangeRole(member)}
                                title="Change Role"
                                className="opacity-0 group-hover:opacity-100 transition-opacity text-xs text-zinc-500 hover:text-zinc-300 underline underline-offset-2 ml-1"
                              >
                                Edit
                              </button>
                            </div>
                          )}
                        </td>

                        <td className="px-6 py-4">
                          {member.status === "active" ? (
                            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 px-2.5 py-1 text-xs font-medium text-emerald-400 border border-emerald-500/20">
                              <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                              Active
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-2.5 py-1 text-xs font-medium text-amber-400 border border-amber-500/20">
                              <Clock className="w-3 h-3" />
                              Pending Invite
                            </span>
                          )}
                        </td>

                        <td className="px-6 py-4 text-xs text-zinc-400">
                          {member.joined_at
                            ? new Date(member.joined_at).toLocaleDateString(undefined, {
                                year: "numeric",
                                month: "short",
                                day: "numeric",
                              })
                            : "—"}
                        </td>

                        <td className="px-6 py-4 text-right pr-6 relative">
                          <div className="relative inline-block text-left">
                            <button
                              onClick={() =>
                                setActionMenuOpenId(actionMenuOpenId === member.id ? null : member.id)
                              }
                              className="flex h-8 w-8 items-center justify-center rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
                            >
                              <MoreHorizontal className="w-4 h-4" />
                            </button>

                            {actionMenuOpenId === member.id && (
                              <div
                                onMouseLeave={() => setActionMenuOpenId(null)}
                                className="absolute right-0 mt-1 w-48 rounded-xl bg-zinc-950 border border-zinc-800 p-1 text-zinc-200 shadow-xl z-20"
                              >
                                {!member.is_owner && (
                                  <button
                                    onClick={() => handleOpenChangeRole(member)}
                                    className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800/80 hover:text-zinc-100 transition-colors"
                                  >
                                    <KeyRound className="w-3.5 h-3.5 text-zinc-400" />
                                    Change Role
                                  </button>
                                )}

                                {member.invite_token && (
                                  <button
                                    onClick={() => {
                                      handleCopyInviteLink(member.invite_token)
                                      setActionMenuOpenId(null)
                                    }}
                                    className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs text-zinc-300 hover:bg-zinc-800/80 hover:text-zinc-100 transition-colors"
                                  >
                                    <Copy className="w-3.5 h-3.5 text-zinc-400" />
                                    Copy Invite Link
                                  </button>
                                )}

                                {!member.is_owner && (
                                  <>
                                    <div className="my-1 border-t border-zinc-800/80" />
                                    <button
                                      onClick={() => handleOpenDelete(member)}
                                      className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs text-red-400 hover:bg-red-500/10 hover:text-red-300 transition-colors"
                                    >
                                      <Trash2 className="w-3.5 h-3.5 text-red-400" />
                                      Remove Member
                                    </button>
                                  </>
                                )}
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    )
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {isInviteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in-0 duration-200">
          <div className="relative w-full max-w-lg rounded-2xl border border-zinc-800 bg-zinc-950 p-6 text-zinc-100 shadow-2xl">
            <button
              onClick={() => setIsInviteModalOpen(false)}
              className="absolute top-5 right-5 text-zinc-400 hover:text-zinc-100 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>

            {inviteSuccessResult ? (
              <div className="flex flex-col gap-5 py-2">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <Check className="w-6 h-6" />
                </div>
                <div className="flex flex-col gap-1">
                  <h3 className="text-xl font-semibold text-zinc-50">Invitation Created!</h3>
                  <p className="text-sm text-zinc-400">
                    <span className="text-zinc-200 font-medium">{inviteSuccessResult.email}</span> has been
                    added as a <span className="text-zinc-200 font-medium">{inviteSuccessResult.role_name}</span>.
                  </p>
                </div>

                {inviteSuccessResult.invite_token && (
                  <div className="flex flex-col gap-2 rounded-xl border border-zinc-800 bg-zinc-900/60 p-4">
                    <span className="text-xs font-medium text-zinc-400 uppercase tracking-wider">
                      Direct Invite Link
                    </span>
                    <div className="flex items-center gap-2">
                      <input
                        readOnly
                        value={`${window.location.origin}/admin/invite?token=${inviteSuccessResult.invite_token}`}
                        className="flex-1 bg-transparent text-xs font-mono text-zinc-300 focus:outline-none select-all overflow-hidden text-ellipsis"
                      />
                      <button
                        onClick={() => handleCopyInviteLink(inviteSuccessResult.invite_token)}
                        className="inline-flex items-center gap-1.5 rounded-lg bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-200 hover:bg-zinc-700 transition-colors"
                      >
                        {copiedLink ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                        {copiedLink ? "Copied" : "Copy"}
                      </button>
                    </div>
                  </div>
                )}

                <div className="mt-4 flex justify-end">
                  <button
                    onClick={() => {
                      setIsInviteModalOpen(false)
                      resetInviteForm()
                    }}
                    className="rounded-full bg-zinc-100 px-6 py-2 text-sm font-semibold text-zinc-900 hover:bg-white transition-all"
                  >
                    Done
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSubmitInvite} className="flex flex-col gap-5">
                <div className="flex flex-col gap-1">
                  <div className="flex items-center gap-2 text-violet-400 mb-1">
                    <UserPlus className="w-5 h-5" />
                    <span className="text-xs font-semibold uppercase tracking-wider">New Collaborator</span>
                  </div>
                  <h3 className="text-xl font-semibold text-zinc-50">Invite Staff Member</h3>
                  <p className="text-sm text-zinc-400">
                    Send an invite to grant access to your shop dashboard with assigned permissions.
                  </p>
                </div>

                {inviteError && (
                  <div className="flex items-center gap-2 rounded-xl border border-red-500/20 bg-red-500/10 p-3 text-xs text-red-400">
                    <AlertCircle className="w-4 h-4 shrink-0" />
                    <span>{inviteError}</span>
                  </div>
                )}

                <div className="flex flex-col gap-4">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-zinc-300">
                      Email Address <span className="text-red-400">*</span>
                    </label>
                    <div className="relative">
                      <Mail className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                      <input
                        type="email"
                        required
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        placeholder="colleague@example.com"
                        className="h-10 w-full rounded-xl bg-zinc-900 border border-zinc-800 pl-10 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 transition-all placeholder:text-zinc-600"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-zinc-300">
                      Store Role <span className="text-red-400">*</span>
                    </label>
                    <div className="relative">
                      <KeyRound className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                      <select
                        value={effectiveInviteRoleId}
                        onChange={(e) => setInviteRoleId(e.target.value)}
                        required
                        className="h-10 w-full appearance-none rounded-xl bg-zinc-900 border border-zinc-800 pl-10 pr-10 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 cursor-pointer"
                      >
                        {roles
                          .filter((r) => r.name.toLowerCase() !== "owner")
                          .map((role) => (
                            <option key={role.id} value={role.id}>
                              {role.name}
                            </option>
                          ))}
                      </select>
                      <ChevronDown className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-zinc-300">
                      Full Name <span className="text-zinc-500 text-[11px]">(Optional)</span>
                    </label>
                    <div className="relative">
                      <User className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                      <input
                        type="text"
                        value={inviteName}
                        onChange={(e) => setInviteName(e.target.value)}
                        placeholder="Alex Morgan"
                        className="h-10 w-full rounded-xl bg-zinc-900 border border-zinc-800 pl-10 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 transition-all placeholder:text-zinc-600"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-zinc-300">
                      Phone Number <span className="text-zinc-500 text-[11px]">(Optional)</span>
                    </label>
                    <div className="relative">
                      <Phone className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                      <input
                        type="tel"
                        value={invitePhone}
                        onChange={(e) => setInvitePhone(e.target.value)}
                        placeholder="+20 100 123 4567"
                        className="h-10 w-full rounded-xl bg-zinc-900 border border-zinc-800 pl-10 pr-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 transition-all placeholder:text-zinc-600"
                      />
                    </div>
                  </div>
                </div>

                <div className="mt-4 flex items-center justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => setIsInviteModalOpen(false)}
                    className="h-10 rounded-full border border-zinc-800 px-5 text-sm font-medium text-zinc-400 hover:text-zinc-100 hover:bg-zinc-900 transition-all"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={inviteMutation.isPending}
                    className="inline-flex h-10 items-center justify-center gap-2 rounded-full bg-zinc-100 px-6 text-sm font-semibold text-zinc-900 hover:bg-white transition-all disabled:opacity-50"
                  >
                    {inviteMutation.isPending ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" />
                        Sending...
                      </>
                    ) : (
                      "Send Invitation"
                    )}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {isChangeRoleModalOpen && activeMember && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in-0 duration-200">
          <div className="relative w-full max-w-md rounded-2xl border border-zinc-800 bg-zinc-950 p-6 text-zinc-100 shadow-2xl">
            <button
              onClick={() => setIsChangeRoleModalOpen(false)}
              className="absolute top-5 right-5 text-zinc-400 hover:text-zinc-100 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>

            <div className="flex flex-col gap-4">
              <div className="flex items-center gap-2 text-blue-400">
                <KeyRound className="w-5 h-5" />
                <span className="text-xs font-semibold uppercase tracking-wider">Role Assignment</span>
              </div>
              <h3 className="text-xl font-semibold text-zinc-50">Change Member Role</h3>
              <p className="text-sm text-zinc-400">
                Update permissions for{" "}
                <span className="text-zinc-200 font-medium">{activeMember.full_name || activeMember.email}</span>.
              </p>

              <div className="mt-2 flex flex-col gap-2">
                <label className="text-xs font-medium text-zinc-300">Select Role</label>
                <div className="relative">
                  <select
                    value={selectedNewRoleId}
                    onChange={(e) => setSelectedNewRoleId(e.target.value)}
                    className="h-10 w-full appearance-none rounded-xl bg-zinc-900 border border-zinc-800 px-4 text-sm text-zinc-100 focus:outline-none focus:ring-1 focus:ring-zinc-600 cursor-pointer"
                  >
                    {roles
                      .filter((r) => r.name.toLowerCase() !== "owner")
                      .map((role) => (
                        <option key={role.id} value={role.id}>
                          {role.name}
                        </option>
                      ))}
                  </select>
                  <ChevronDown className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
                </div>
              </div>

              <div className="mt-4 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsChangeRoleModalOpen(false)}
                  className="h-10 rounded-full border border-zinc-800 px-5 text-sm font-medium text-zinc-400 hover:text-zinc-100 transition-all"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleSubmitRoleChange}
                  disabled={updateRoleMutation.isPending}
                  className="inline-flex h-10 items-center justify-center gap-2 rounded-full bg-zinc-100 px-6 text-sm font-semibold text-zinc-900 hover:bg-white transition-all disabled:opacity-50"
                >
                  {updateRoleMutation.isPending ? "Updating..." : "Save Role"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {isDeleteModalOpen && activeMember && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in-0 duration-200">
          <div className="relative w-full max-w-md rounded-2xl border border-zinc-800 bg-zinc-950 p-6 text-zinc-100 shadow-2xl">
            <button
              onClick={() => setIsDeleteModalOpen(false)}
              className="absolute top-5 right-5 text-zinc-400 hover:text-zinc-100 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>

            <div className="flex flex-col gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-red-500/10 text-red-400 border border-red-500/20">
                <Trash2 className="w-6 h-6" />
              </div>
              <h3 className="text-xl font-semibold text-zinc-50">Remove Team Member?</h3>
              <p className="text-sm text-zinc-400">
                Are you sure you want to remove{" "}
                <span className="text-zinc-200 font-medium">{activeMember.full_name || activeMember.email}</span> from
                your shop? They will immediately lose access to all store data.
              </p>

              <div className="mt-4 flex items-center justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsDeleteModalOpen(false)}
                  className="h-10 rounded-full border border-zinc-800 px-5 text-sm font-medium text-zinc-400 hover:text-zinc-100 transition-all"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={handleConfirmDelete}
                  disabled={removeStaffMutation.isPending}
                  className="inline-flex h-10 items-center justify-center gap-2 rounded-full bg-red-600 px-6 text-sm font-semibold text-white hover:bg-red-500 transition-all disabled:opacity-50"
                >
                  {removeStaffMutation.isPending ? "Removing..." : "Confirm Removal"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
