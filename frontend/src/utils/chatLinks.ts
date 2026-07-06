import type { ChatPreset, RawChatPreset } from '@/types'

export type ChatLinkType = 'web' | 'custom-protocol' | 'fluent'

export interface NormalizedChatPreset extends ChatPreset {
  id: string
  type: ChatLinkType
}

export interface ResolveChatUrlOptions {
  template: string
  apiKey?: string | null
  serverAddress?: string | null
}

const HTTP_URL_RE = /^https?:\/\//i
const PLACEHOLDER_KEY = '{key}'
const PLACEHOLDER_ADDRESS = '{address}'
const PLACEHOLDER_CHERRY_CONFIG = '{cherryConfig}'
const PLACEHOLDER_AIONUI_CONFIG = '{aionuiConfig}'
const PLACEHOLDER_DEEPCHAT_CONFIG = '{deepchatConfig}'

export function detectChatLinkType(url: string): ChatLinkType {
  const value = url.trim()
  if (HTTP_URL_RE.test(value)) return 'web'
  if (value.toLowerCase().startsWith('fluent')) return 'fluent'
  return 'custom-protocol'
}

export function chatLinkRequiresApiKey(url: string): boolean {
  return (
    url.includes(PLACEHOLDER_KEY) ||
    url.includes(PLACEHOLDER_CHERRY_CONFIG) ||
    url.includes(PLACEHOLDER_AIONUI_CONFIG) ||
    url.includes(PLACEHOLDER_DEEPCHAT_CONFIG)
  )
}

export function normalizeApiKey(apiKey?: string | null): string {
  const value = (apiKey ?? '').trim()
  if (!value) return ''
  return value.startsWith('sk-') ? value : `sk-${value}`
}

export function parseChatPresets(raw: unknown): NormalizedChatPreset[] {
  const parsed = parseRawConfig(raw)
  if (!Array.isArray(parsed)) return []

  const presets: NormalizedChatPreset[] = []
  parsed.forEach((entry, index) => {
    const preset = normalizePresetEntry(entry as RawChatPreset, index)
    if (preset) presets.push(preset)
  })
  return presets
}

export function resolveChatUrl({
  template,
  apiKey,
  serverAddress,
}: ResolveChatUrlOptions): string {
  let url = template
  const safeApiKey = normalizeApiKey(apiKey)
  const safeAddress = (serverAddress ?? '').trim()

  if (url.includes(PLACEHOLDER_CHERRY_CONFIG)) {
    return replaceAll(
      url,
      PLACEHOLDER_CHERRY_CONFIG,
      encodeConfigPayload({
        id: 'new-api',
        baseUrl: safeAddress,
        apiKey: safeApiKey,
      }),
    )
  }

  if (url.includes(PLACEHOLDER_AIONUI_CONFIG)) {
    return replaceAll(
      url,
      PLACEHOLDER_AIONUI_CONFIG,
      encodeConfigPayload({
        platform: 'new-api',
        baseUrl: safeAddress,
        apiKey: safeApiKey,
      }),
    )
  }

  if (url.includes(PLACEHOLDER_DEEPCHAT_CONFIG)) {
    return replaceAll(
      url,
      PLACEHOLDER_DEEPCHAT_CONFIG,
      encodeConfigPayload({
        id: 'new-api',
        baseUrl: safeAddress,
        apiKey: safeApiKey,
      }),
    )
  }

  if (safeAddress) {
    url = replaceAll(url, PLACEHOLDER_ADDRESS, encodeURIComponent(safeAddress))
  }
  if (safeApiKey) {
    url = replaceAll(url, PLACEHOLDER_KEY, safeApiKey)
  }
  return url
}

function parseRawConfig(raw: unknown): unknown {
  if (typeof raw !== 'string') return raw
  const value = raw.trim()
  if (!value) return []
  try {
    return JSON.parse(value)
  } catch {
    return []
  }
}

function normalizePresetEntry(entry: RawChatPreset, index: number): NormalizedChatPreset | null {
  if (!entry || typeof entry !== 'object' || Array.isArray(entry)) return null

  if ('name' in entry && 'url' in entry) {
    const name = String(entry.name ?? '').trim()
    const url = String(entry.url ?? '').trim()
    if (!name || !url) return null
    const id = String(entry.id ?? index).trim() || String(index)
    return { id, name, url, type: detectChatLinkType(url) }
  }

  const values = Object.entries(entry)
  if (values.length !== 1) return null
  const [rawName, rawUrl] = values[0]
  if (typeof rawUrl !== 'string') return null
  const name = rawName.trim()
  const url = rawUrl.trim()
  if (!name || !url) return null
  return { id: String(index), name, url, type: detectChatLinkType(url) }
}

function replaceAll(source: string, token: string, value: string): string {
  return source.split(token).join(value)
}

function encodeConfigPayload(payload: Record<string, string>): string {
  return encodeURIComponent(toBase64(JSON.stringify(payload)))
}

function toBase64(value: string): string {
  const globalObject = globalThis as typeof globalThis & {
    Buffer?: { from(input: string, encoding: BufferEncoding): { toString(encoding: BufferEncoding): string } }
  }
  if (globalObject.Buffer) {
    return globalObject.Buffer.from(value, 'utf8').toString('base64')
  }
  if (typeof window !== 'undefined' && typeof window.btoa === 'function') {
    const bytes = new TextEncoder().encode(value)
    let binary = ''
    bytes.forEach((byte) => {
      binary += String.fromCharCode(byte)
    })
    return window.btoa(binary)
  }
  return ''
}
