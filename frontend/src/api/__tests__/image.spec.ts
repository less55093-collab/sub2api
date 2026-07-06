import axios from 'axios'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { buildImageGatewayEndpoint, editImage, generateImage, ImageGroupHeader, maskImageAPIKey } from '../image'

describe('image API helpers', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('builds gateway image endpoints outside /api/v1', () => {
    expect(buildImageGatewayEndpoint('/v1/images/generations')).toMatch(/\/v1\/images\/generations$/)
    expect(buildImageGatewayEndpoint('/v1/images/edits')).toMatch(/\/v1\/images\/edits$/)
  })

  it('masks API keys before display or persistence', () => {
    const masked = maskImageAPIKey('sk-test-1234567890abcdef')
    expect(masked).toContain('sk-tes')
    expect(masked).toContain('cdef')
    expect(masked).not.toContain('1234567890ab')
    expect(maskImageAPIKey('')).toBe('未输入 Key')
  })

  it('sends selected group header for image generation', async () => {
    const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: { created: 0, data: [] } })

    await generateImage('sk-test-1234567890abcdef', {
      prompt: 'ping',
      model: 'gpt-image-2',
      n: 1,
    }, { groupId: 20 })

    const [, payload, config] = post.mock.calls[0]
    expect(payload).toMatchObject({ prompt: 'ping', model: 'gpt-image-2' })
    expect(config?.headers).toMatchObject({
      'X-API-Key': 'sk-test-1234567890abcdef',
      Authorization: 'Bearer sk-test-1234567890abcdef',
      [ImageGroupHeader]: '20',
      'Content-Type': 'application/json',
    })
  })

  it('sends selected group header for image edits', async () => {
    const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: { created: 0, data: [] } })
    const file = new File(['image'], 'source.png', { type: 'image/png' })

    await editImage('sk-test-1234567890abcdef', {
      prompt: 'edit',
      model: 'gpt-image-2',
      image: file,
    }, { groupId: 21 })

    const [, payload, config] = post.mock.calls[0]
    expect(payload).toBeInstanceOf(FormData)
    expect(config?.headers).toMatchObject({
      'X-API-Key': 'sk-test-1234567890abcdef',
      Authorization: 'Bearer sk-test-1234567890abcdef',
      [ImageGroupHeader]: '21',
    })
    expect(config?.headers).not.toHaveProperty('Content-Type')
  })
})
