import { describe, expect, it } from 'vitest'
import type { ApiKey, Group } from '@/types'
import {
  buildApiKeySelectionLabel,
  buildApiKeySelectionOptions,
  formatApiKeyGroupSummary,
  getApiKeyGroupIds,
  getSelectableApiKeyGroups,
  getSelectableApiKeys,
  isSelectableApiKey,
  resolveApiKeyGroups,
  resolveDefaultApiKeySelection,
} from '@/utils/apiKeySelection'

const makeGroup = (overrides: Partial<Group> & Pick<Group, 'id' | 'name'>): Group => ({
  id: overrides.id,
  name: overrides.name,
  description: overrides.description ?? null,
  platform: overrides.platform ?? 'openai',
  rate_multiplier: overrides.rate_multiplier ?? 1,
  rpm_limit: overrides.rpm_limit ?? 0,
  is_exclusive: overrides.is_exclusive ?? false,
  status: overrides.status ?? 'active',
  subscription_type: overrides.subscription_type ?? 'standard',
  daily_limit_usd: overrides.daily_limit_usd ?? null,
  weekly_limit_usd: overrides.weekly_limit_usd ?? null,
  monthly_limit_usd: overrides.monthly_limit_usd ?? null,
  allow_image_generation: overrides.allow_image_generation ?? true,
  image_rate_independent: overrides.image_rate_independent ?? false,
  image_rate_multiplier: overrides.image_rate_multiplier ?? 1,
  image_price_1k: overrides.image_price_1k ?? null,
  image_price_2k: overrides.image_price_2k ?? null,
  image_price_4k: overrides.image_price_4k ?? null,
  peak_rate_enabled: overrides.peak_rate_enabled ?? false,
  peak_start: overrides.peak_start ?? '',
  peak_end: overrides.peak_end ?? '',
  peak_rate_multiplier: overrides.peak_rate_multiplier ?? 1,
  claude_code_only: overrides.claude_code_only ?? false,
  fallback_group_id: overrides.fallback_group_id ?? null,
  fallback_group_id_on_invalid_request: overrides.fallback_group_id_on_invalid_request ?? null,
  allow_messages_dispatch: overrides.allow_messages_dispatch,
  default_mapped_model: overrides.default_mapped_model,
  messages_dispatch_model_config: overrides.messages_dispatch_model_config,
  require_oauth_only: overrides.require_oauth_only ?? false,
  require_privacy_set: overrides.require_privacy_set ?? false,
  created_at: overrides.created_at ?? '2026-01-01T00:00:00Z',
  updated_at: overrides.updated_at ?? '2026-01-01T00:00:00Z',
})

const makeKey = (overrides: Partial<ApiKey> & Pick<ApiKey, 'id' | 'key' | 'name'>): ApiKey => ({
  id: overrides.id,
  user_id: overrides.user_id ?? 1,
  key: overrides.key,
  name: overrides.name,
  group_id: overrides.group_id ?? null,
  group_ids: overrides.group_ids ?? [],
  status: overrides.status ?? 'active',
  ip_whitelist: overrides.ip_whitelist ?? [],
  ip_blacklist: overrides.ip_blacklist ?? [],
  last_used_at: overrides.last_used_at ?? null,
  quota: overrides.quota ?? 0,
  quota_used: overrides.quota_used ?? 0,
  expires_at: overrides.expires_at ?? null,
  created_at: overrides.created_at ?? '2026-01-01T00:00:00Z',
  updated_at: overrides.updated_at ?? '2026-01-01T00:00:00Z',
  group: overrides.group,
  groups: overrides.groups,
  rate_limit_5h: overrides.rate_limit_5h ?? 0,
  rate_limit_1d: overrides.rate_limit_1d ?? 0,
  rate_limit_7d: overrides.rate_limit_7d ?? 0,
  usage_5h: overrides.usage_5h ?? 0,
  usage_1d: overrides.usage_1d ?? 0,
  usage_7d: overrides.usage_7d ?? 0,
  window_5h_start: overrides.window_5h_start ?? null,
  window_1d_start: overrides.window_1d_start ?? null,
  window_7d_start: overrides.window_7d_start ?? null,
  reset_5h_at: overrides.reset_5h_at ?? null,
  reset_1d_at: overrides.reset_1d_at ?? null,
  reset_7d_at: overrides.reset_7d_at ?? null,
})

describe('api key selection helpers', () => {
  it('normalizes group IDs from group_ids, groups, legacy group_id, and group', () => {
    const embeddedGroup = makeGroup({ id: 7, name: 'embedded' })
    const key = makeKey({
      id: 1,
      key: 'sk-test-1234567890abcdef',
      name: 'primary',
      group_ids: [2, 3, 2],
      groups: [makeGroup({ id: 3, name: 'duplicate' }), makeGroup({ id: 5, name: 'team' })],
      group_id: 6,
      group: embeddedGroup,
    })

    expect(getApiKeyGroupIds(key)).toEqual([2, 3, 5, 6, 7])
  })

  it('resolves groups from available and embedded key group payloads', () => {
    const availableGroups = [
      makeGroup({ id: 2, name: 'OpenAI shared' }),
      makeGroup({ id: 6, name: 'legacy' }),
    ]
    const key = makeKey({
      id: 1,
      key: 'sk-test-1234567890abcdef',
      name: 'primary',
      group_ids: [2],
      groups: [makeGroup({ id: 5, name: 'OpenAI private' })],
      group_id: 6,
    })

    expect(resolveApiKeyGroups(key, availableGroups).map((group) => group.name)).toEqual([
      'OpenAI shared',
      'OpenAI private',
      'legacy',
    ])
  })

  it('treats only usable active keys as selectable', () => {
    const active = makeKey({ id: 1, key: 'sk-active-1234567890', name: 'active' })
    const inactive = makeKey({ id: 2, key: 'sk-inactive-1234567890', name: 'inactive', status: 'inactive' })
    const expired = makeKey({ id: 3, key: 'sk-expired-1234567890', name: 'expired', status: 'expired' })
    const exhaustedStatus = makeKey({
      id: 4,
      key: 'sk-exhausted-1234567890',
      name: 'exhausted',
      status: 'quota_exhausted',
    })
    const exhaustedQuota = makeKey({
      id: 5,
      key: 'sk-quota-1234567890',
      name: 'quota',
      quota: 10,
      quota_used: 10,
    })

    expect(isSelectableApiKey(active)).toBe(true)
    expect(getSelectableApiKeys([active, inactive, expired, exhaustedStatus, exhaustedQuota])).toEqual([active])
  })

  it('filters inactive groups for selected group defaults without hiding all labels', () => {
    const activeGroup = makeGroup({ id: 1, name: 'active group' })
    const inactiveGroup = makeGroup({ id: 2, name: 'inactive group', status: 'inactive' })
    const key = makeKey({
      id: 1,
      key: 'sk-test-1234567890abcdef',
      name: 'primary',
      group_ids: [2, 1],
    })

    expect(resolveApiKeyGroups(key, [activeGroup, inactiveGroup]).map((group) => group.name)).toEqual([
      'inactive group',
      'active group',
    ])
    expect(getSelectableApiKeyGroups(key, [activeGroup, inactiveGroup])).toEqual([activeGroup])
  })

  it('builds safe labels and options without exposing full key material', () => {
    const group = makeGroup({ id: 1, name: 'openai-default' })
    const key = makeKey({
      id: 1,
      key: 'sk-test-1234567890abcdef',
      name: 'production',
      group_ids: [1],
    })

    const label = buildApiKeySelectionLabel(key, [group])
    const [option] = buildApiKeySelectionOptions([key], [group])

    expect(label).toContain('production')
    expect(label).toContain('sk-tes')
    expect(label).toContain('cdef')
    expect(label).toContain('openai-default')
    expect(label).not.toContain('1234567890ab')
    expect(option).toMatchObject({
      value: 1,
      maskedKey: 'sk-tes...cdef',
      groupIds: [1],
      groups: [group],
    })
  })

  it('resolves a default key and active group selection', () => {
    const inactiveGroup = makeGroup({ id: 1, name: 'inactive', status: 'inactive' })
    const activeGroup = makeGroup({ id: 2, name: 'active' })
    const inactiveKey = makeKey({
      id: 1,
      key: 'sk-inactive-1234567890',
      name: 'inactive',
      status: 'inactive',
      group_ids: [2],
    })
    const activeKey = makeKey({
      id: 2,
      key: 'sk-active-1234567890',
      name: 'active',
      group_ids: [1, 2],
    })

    expect(resolveDefaultApiKeySelection([inactiveKey, activeKey], [inactiveGroup, activeGroup])).toEqual({
      key: activeKey,
      group: activeGroup,
      groupId: 2,
    })
  })

  it('summarizes long group lists compactly', () => {
    expect(formatApiKeyGroupSummary([])).toBe('未绑定分组')
    expect(formatApiKeyGroupSummary([
      makeGroup({ id: 1, name: 'a' }),
      makeGroup({ id: 2, name: 'b' }),
      makeGroup({ id: 3, name: 'c' }),
    ])).toBe('a / b +1')
  })
})
