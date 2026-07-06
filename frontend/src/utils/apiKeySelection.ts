import type { ApiKey, Group } from '@/types'
import { maskApiKey } from '@/utils/maskApiKey'

export interface ApiKeySelectionOption {
  value: number
  label: string
  key: ApiKey
  maskedKey: string
  groupIds: number[]
  groups: Group[]
}

export interface ApiKeyDefaultSelection {
  key: ApiKey | null
  group: Group | null
  groupId: number | null
}

const selectableStatuses = new Set<ApiKey['status']>(['active'])

export const normalizeApiKeyGroupIds = (
  ids: Array<number | string | null | undefined>
): number[] =>
  Array.from(
    new Set(
      ids
        .map((id) => {
          if (typeof id === 'number') return id
          if (typeof id === 'string' && /^\d+$/.test(id.trim())) return Number(id)
          return null
        })
        .filter((id): id is number => typeof id === 'number' && Number.isInteger(id) && id > 0)
    )
  )

export const getApiKeyGroupIds = (key: ApiKey | null | undefined): number[] => {
  if (!key) return []

  return normalizeApiKeyGroupIds([
    ...(key.group_ids || []),
    ...(key.groups || []).map((group) => group.id),
    key.group_id,
    key.group?.id
  ])
}

export const resolveApiKeyGroups = (
  key: ApiKey | null | undefined,
  availableGroups: Group[] = []
): Group[] => {
  if (!key) return []

  const byID = new Map<number, Group>()
  for (const group of availableGroups) byID.set(group.id, group)
  for (const group of key.groups || []) byID.set(group.id, group)
  if (key.group) byID.set(key.group.id, key.group)

  return getApiKeyGroupIds(key)
    .map((id) => byID.get(id))
    .filter((group): group is Group => Boolean(group))
}

export const getSelectableApiKeyGroups = (
  key: ApiKey | null | undefined,
  availableGroups: Group[] = []
): Group[] => resolveApiKeyGroups(key, availableGroups).filter((group) => group.status === 'active')

export const isSelectableApiKey = (key: ApiKey | null | undefined): key is ApiKey => {
  if (!key) return false
  if (!selectableStatuses.has(key.status)) return false
  if (key.quota > 0 && key.quota_used >= key.quota) return false
  return true
}

export const getSelectableApiKeys = (keys: ApiKey[] = []): ApiKey[] =>
  keys.filter((key) => isSelectableApiKey(key))

export const formatApiKeyGroupSummary = (groups: Group[]): string => {
  if (groups.length === 0) return '未绑定分组'
  if (groups.length <= 2) return groups.map((group) => group.name).join(' / ')
  return `${groups.slice(0, 2).map((group) => group.name).join(' / ')} +${groups.length - 2}`
}

export const buildApiKeySelectionLabel = (
  key: ApiKey,
  availableGroups: Group[] = []
): string => {
  const name = key.name?.trim() || '未命名 Key'
  const maskedKey = maskApiKey(key.key)
  const groupSummary = formatApiKeyGroupSummary(resolveApiKeyGroups(key, availableGroups))

  return `${name} · ${maskedKey} · ${groupSummary}`
}

export const buildApiKeySelectionOptions = (
  keys: ApiKey[] = [],
  availableGroups: Group[] = []
): ApiKeySelectionOption[] =>
  getSelectableApiKeys(keys).map((key) => ({
    value: key.id,
    label: buildApiKeySelectionLabel(key, availableGroups),
    key,
    maskedKey: maskApiKey(key.key),
    groupIds: getApiKeyGroupIds(key),
    groups: resolveApiKeyGroups(key, availableGroups)
  }))

export const resolveDefaultApiKeySelection = (
  keys: ApiKey[] = [],
  availableGroups: Group[] = []
): ApiKeyDefaultSelection => {
  const key = getSelectableApiKeys(keys)[0] || null
  const group = key ? getSelectableApiKeyGroups(key, availableGroups)[0] || null : null

  return {
    key,
    group,
    groupId: group?.id ?? null
  }
}
