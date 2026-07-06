import { apiClient, buildGatewayUrl } from './client'
import { getLocale } from '@/i18n'

export type PlaygroundChatRole = 'system' | 'user' | 'assistant'

export interface PlaygroundChatMessage {
  role: PlaygroundChatRole
  content: string
}

export interface PlaygroundModel {
  id: string
  object?: string
  created?: number
  owned_by?: string
}

export interface PlaygroundChatCompletionRequest {
  model: string
  messages: PlaygroundChatMessage[]
  stream?: boolean
  group_id?: number | null
  temperature?: number
  max_tokens?: number | null
}

export interface PlaygroundChatCompletionResult {
  content: string
  raw: unknown
}

export interface PlaygroundStreamOptions {
  signal?: AbortSignal
  apiKeyId?: number | null
  onDelta?: (text: string, event: unknown) => void
  onEvent?: (event: unknown) => void
}

export class PlaygroundAPIError extends Error {
  status: number
  code?: string

  constructor(message: string, status: number, code?: string) {
    super(message)
    this.name = 'PlaygroundAPIError'
    this.status = status
    this.code = code
  }
}

export const PlaygroundSelectedAPIKeyIDHeader = 'X-FluxRouter-API-Key-ID'

export async function getPlaygroundModels(
  groupId?: number | null,
  apiKeyId?: number | null,
): Promise<PlaygroundModel[]> {
  const { data } = await apiClient.get<unknown>('/playground/models', {
    params: groupId ? { group_id: groupId } : undefined,
    headers: buildSelectedAPIKeyHeaders(apiKeyId),
  })
  return normalizePlaygroundModels(data)
}

export async function sendPlaygroundChatCompletion(
  request: PlaygroundChatCompletionRequest,
  options: PlaygroundStreamOptions = {},
): Promise<PlaygroundChatCompletionResult> {
  const stream = Boolean(request.stream)
  const response = await fetch(buildGatewayUrl('/pg/chat/completions'), {
    method: 'POST',
    headers: buildPlaygroundHeaders(stream, options.apiKeyId),
    body: JSON.stringify(cleanPlaygroundPayload(request)),
    signal: options.signal,
  })

  if (!response.ok) {
    const error = await readPlaygroundError(response)
    throw new PlaygroundAPIError(error.message, response.status, error.code)
  }

  if (stream) {
    const content = await readPlaygroundStream(response, options)
    return { content, raw: null }
  }

  const raw = await response.json()
  return {
    content: extractPlaygroundMessageText(raw),
    raw,
  }
}

export function normalizePlaygroundModels(raw: unknown): PlaygroundModel[] {
  const data = Array.isArray(raw)
    ? raw
    : Array.isArray((raw as { data?: unknown })?.data)
      ? (raw as { data: unknown[] }).data
      : []

  return data
    .map((item): PlaygroundModel | null => {
      if (typeof item === 'string') {
        const id = item.trim()
        return id ? { id } : null
      }
      if (!item || typeof item !== 'object') return null
      const record = item as Record<string, unknown>
      const id = String(record.id ?? '').trim()
      if (!id) return null
      const model: PlaygroundModel = { id }
      if (typeof record.object === 'string') model.object = record.object
      if (typeof record.created === 'number') model.created = record.created
      if (typeof record.owned_by === 'string') model.owned_by = record.owned_by
      return model
    })
    .filter((item): item is PlaygroundModel => item !== null)
}

export function extractPlaygroundMessageText(raw: unknown): string {
  const record = asRecord(raw)
  const choices = Array.isArray(record?.choices) ? record.choices : []
  for (const choice of choices) {
    const choiceRecord = asRecord(choice)
    const message = asRecord(choiceRecord?.message)
    const messageContent = textFromContent(message?.content)
    if (messageContent) return messageContent
    const directText = textFromContent(choiceRecord?.text)
    if (directText) return directText
  }
  return textFromContent(record?.output_text)
}

export function extractPlaygroundDeltaText(raw: unknown): string {
  const record = asRecord(raw)
  const choices = Array.isArray(record?.choices) ? record.choices : []
  let out = ''
  for (const choice of choices) {
    const choiceRecord = asRecord(choice)
    const delta = asRecord(choiceRecord?.delta)
    out += textFromContent(delta?.content)
    out += textFromContent(choiceRecord?.text)
    if (!delta) {
      const message = asRecord(choiceRecord?.message)
      out += textFromContent(message?.content)
    }
  }
  return out
}

function buildSelectedAPIKeyHeaders(apiKeyId?: number | null): Record<string, string> | undefined {
  if (typeof apiKeyId !== 'number' || !Number.isInteger(apiKeyId) || apiKeyId <= 0) {
    return undefined
  }
  return { [PlaygroundSelectedAPIKeyIDHeader]: String(apiKeyId) }
}

function buildPlaygroundHeaders(stream: boolean, apiKeyId?: number | null): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: stream ? 'text/event-stream' : 'application/json',
    'Accept-Language': getLocale(),
  }
  const selectedKeyHeaders = buildSelectedAPIKeyHeaders(apiKeyId)
  if (selectedKeyHeaders) {
    Object.assign(headers, selectedKeyHeaders)
  }
  const token = localStorage.getItem('auth_token')
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  return headers
}

function cleanPlaygroundPayload(request: PlaygroundChatCompletionRequest): Record<string, unknown> {
  const payload: Record<string, unknown> = {
    model: request.model,
    messages: request.messages,
    stream: Boolean(request.stream),
  }
  if (request.group_id) payload.group_id = request.group_id
  if (typeof request.temperature === 'number' && Number.isFinite(request.temperature)) {
    payload.temperature = request.temperature
  }
  if (typeof request.max_tokens === 'number' && Number.isFinite(request.max_tokens) && request.max_tokens > 0) {
    payload.max_tokens = Math.floor(request.max_tokens)
  }
  return payload
}

async function readPlaygroundStream(
  response: Response,
  options: PlaygroundStreamOptions,
): Promise<string> {
  const reader = response.body?.getReader()
  if (!reader) {
    throw new PlaygroundAPIError('No response body', response.status)
  }

  const decoder = new TextDecoder()
  let buffer = ''
  let fullText = ''
  let doneSignal = false

  const consumeLine = (line: string) => {
    const trimmed = line.trimEnd()
    if (!trimmed.startsWith('data:')) return
    const data = trimmed.slice(5).trim()
    if (!data) return
    if (data === '[DONE]') {
      doneSignal = true
      return
    }
    try {
      const event = JSON.parse(data)
      options.onEvent?.(event)
      const delta = extractPlaygroundDeltaText(event)
      if (delta) {
        fullText += delta
        options.onDelta?.(delta, event)
      }
    } catch {
      // Ignore malformed SSE lines and keep reading; upstreams may emit comments.
    }
  }

  while (!doneSignal) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() ?? ''
    lines.forEach(consumeLine)
  }

  if (buffer) {
    consumeLine(buffer)
  }

  return fullText
}

async function readPlaygroundError(response: Response): Promise<{ message: string; code?: string }> {
  const text = await response.text().catch(() => '')
  if (!text) {
    return { message: `Request failed with status ${response.status}` }
  }
  try {
    const parsed = JSON.parse(text) as Record<string, unknown>
    const errorRecord = asRecord(parsed.error)
    const message =
      stringValue(parsed.message) ||
      stringValue(parsed.detail) ||
      stringValue(errorRecord?.message) ||
      text
    const code = stringValue(parsed.code) || stringValue(errorRecord?.code)
    return { message, code: code || undefined }
  } catch {
    return { message: text }
  }
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : null
}

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function textFromContent(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value)) {
    return value.map(textFromContent).join('')
  }
  const record = asRecord(value)
  if (!record) return ''
  return (
    textString(record.text) ||
    textString(record.content) ||
    textString(record.value)
  )
}

function textString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

export const playgroundAPI = {
  getModels: getPlaygroundModels,
  sendChatCompletion: sendPlaygroundChatCompletion,
}

export default playgroundAPI
