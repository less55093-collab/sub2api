import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '../client'
import {
  extractPlaygroundDeltaText,
  extractPlaygroundMessageText,
  getPlaygroundModels,
  normalizePlaygroundModels,
  PlaygroundSelectedAPIKeyIDHeader,
  sendPlaygroundChatCompletion,
} from '../playground'

describe('playground API helpers', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    apiClient.defaults.adapter = undefined
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    localStorage.clear()
  })

  it('normalizes OpenAI-style and simple model lists', () => {
    expect(normalizePlaygroundModels({
      object: 'list',
      data: [
        { id: 'gpt-5', object: 'model', owned_by: 'openai' },
        'claude-sonnet-4-5',
        { id: '' },
      ],
    })).toEqual([
      { id: 'gpt-5', object: 'model', owned_by: 'openai' },
      { id: 'claude-sonnet-4-5' },
    ])
  })

  it('extracts non-streaming assistant text', () => {
    expect(extractPlaygroundMessageText({
      choices: [
        {
          message: {
            role: 'assistant',
            content: [
              { type: 'text', text: 'hello' },
              { type: 'text', text: ' world' },
            ],
          },
        },
      ],
    })).toBe('hello world')
  })

  it('extracts streaming deltas', () => {
    expect(extractPlaygroundDeltaText({
      choices: [
        { delta: { content: 'hel' } },
        { delta: { content: [{ text: 'lo' }] } },
      ],
    })).toBe('hello')
  })

  it('sends selected key ID and group ID when loading models', async () => {
    const adapter = vi.fn().mockResolvedValue({
      status: 200,
      data: { object: 'list', data: ['gpt-5'] },
      headers: {},
      config: {},
      statusText: 'OK',
    })
    apiClient.defaults.adapter = adapter

    await expect(getPlaygroundModels(20, 7)).resolves.toEqual([{ id: 'gpt-5' }])

    const config = adapter.mock.calls[0][0]
    expect(config.params).toMatchObject({ group_id: 20 })
    expect(config.headers.get(PlaygroundSelectedAPIKeyIDHeader)).toBe('7')
  })

  it('sends selected key ID header for chat without putting it in the payload', async () => {
    localStorage.setItem('auth_token', 'jwt-token')
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({
        choices: [{ message: { role: 'assistant', content: 'pong' } }],
      }), { status: 200 }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await expect(sendPlaygroundChatCompletion(
      {
        model: 'gpt-5',
        messages: [{ role: 'user', content: 'ping' }],
        stream: false,
        group_id: 20,
      },
      { apiKeyId: 7 },
    )).resolves.toMatchObject({ content: 'pong' })

    const [, init] = fetchMock.mock.calls[0]
    const headers = init.headers as Record<string, string>
    const payload = JSON.parse(String(init.body)) as Record<string, unknown>
    expect(headers[PlaygroundSelectedAPIKeyIDHeader]).toBe('7')
    expect(headers.Authorization).toBe('Bearer jwt-token')
    expect(payload.group_id).toBe(20)
    expect(payload).not.toHaveProperty('api_key_id')
  })
})
