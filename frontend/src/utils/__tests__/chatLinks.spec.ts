import { describe, expect, it } from 'vitest'
import {
  chatLinkRequiresApiKey,
  detectChatLinkType,
  normalizeApiKey,
  parseChatPresets,
  resolveChatUrl,
} from '../chatLinks'

function decodeConfigParam(value: string): Record<string, string> {
  return JSON.parse(Buffer.from(decodeURIComponent(value), 'base64').toString('utf8'))
}

describe('chatLinks', () => {
  it('normalizes typed and new-api legacy chat preset shapes', () => {
    expect(parseChatPresets([
      { id: 'typed', name: ' Cherry ', url: ' https://chat.example.com/{key} ' },
      { Fluent: 'fluent://chat?key={key}' },
      { invalid: '' },
    ])).toEqual([
      { id: 'typed', name: 'Cherry', url: 'https://chat.example.com/{key}', type: 'web' },
      { id: '1', name: 'Fluent', url: 'fluent://chat?key={key}', type: 'fluent' },
    ])
  })

  it('parses JSON string settings and ignores malformed input', () => {
    expect(parseChatPresets('[{"DeepChat":"deepchat://import?config={deepchatConfig}"}]')).toEqual([
      {
        id: '0',
        name: 'DeepChat',
        url: 'deepchat://import?config={deepchatConfig}',
        type: 'custom-protocol',
      },
    ])
    expect(parseChatPresets('not-json')).toEqual([])
    expect(parseChatPresets(null)).toEqual([])
  })

  it('detects link types and api-key placeholders', () => {
    expect(detectChatLinkType('https://chat.example.com')).toBe('web')
    expect(detectChatLinkType('fluent://chat')).toBe('fluent')
    expect(detectChatLinkType('cherry-studio://import')).toBe('custom-protocol')
    expect(chatLinkRequiresApiKey('https://chat.example.com?key={key}')).toBe(true)
    expect(chatLinkRequiresApiKey('https://chat.example.com?address={address}')).toBe(false)
  })

  it('resolves key and address placeholders', () => {
    const url = resolveChatUrl({
      template: 'https://chat.example.com?base={address}&key={key}',
      apiKey: 'abc',
      serverAddress: 'https://api.example.com/v1',
    })

    expect(url).toBe('https://chat.example.com?base=https%3A%2F%2Fapi.example.com%2Fv1&key=sk-abc')
    expect(normalizeApiKey('sk-existing')).toBe('sk-existing')
  })

  it('resolves config placeholders using new-api compatible payloads', () => {
    const cherryUrl = resolveChatUrl({
      template: 'cherry://import?config={cherryConfig}',
      apiKey: 'abc',
      serverAddress: 'https://api.example.com',
    })
    const cherryConfig = decodeConfigParam(new URL(cherryUrl).searchParams.get('config') ?? '')
    expect(cherryConfig).toEqual({
      id: 'new-api',
      baseUrl: 'https://api.example.com',
      apiKey: 'sk-abc',
    })

    const aionUrl = resolveChatUrl({
      template: 'aionui://import?config={aionuiConfig}',
      apiKey: 'sk-def',
      serverAddress: 'https://api.example.com',
    })
    const aionConfig = decodeConfigParam(new URL(aionUrl).searchParams.get('config') ?? '')
    expect(aionConfig).toEqual({
      platform: 'new-api',
      baseUrl: 'https://api.example.com',
      apiKey: 'sk-def',
    })
  })
})
