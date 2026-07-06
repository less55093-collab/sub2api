import axios from 'axios'
import { apiClient, buildGatewayUrl } from './client'

export type ImageQuality = 'auto' | 'low' | 'medium' | 'high'

export interface ImageGenerationRequest {
  prompt: string
  model: string
  size?: string
  quality?: ImageQuality
  n?: number
}

export interface ImageEditRequest {
  prompt: string
  model: string
  image: File
  size?: string
  quality?: ImageQuality
  n?: number
}

export interface ImageResultItem {
  b64_json?: string
  revised_prompt?: string
  url?: string
}

export interface ImageGenerationResponse {
  created: number
  data: ImageResultItem[]
}

export interface ImageRequestDebugInfo {
  endpoint: string
  hasBearer: boolean
  hasXApiKey: boolean
  groupId?: number
  model: string
  size?: string
  quality?: string
  count?: number
}

export class ImageRequestError extends Error {
  status?: number
  responseBody?: unknown
  debug?: ImageRequestDebugInfo

  constructor(message: string, options?: { status?: number; responseBody?: unknown; debug?: ImageRequestDebugInfo }) {
    super(message)
    this.name = 'ImageRequestError'
    this.status = options?.status
    this.responseBody = options?.responseBody
    this.debug = options?.debug
  }
}

export const ImageGroupHeader = 'X-FluxRouter-Group-ID'

export function buildImageGatewayEndpoint(path: '/v1/images/generations' | '/v1/images/edits'): string {
  return buildGatewayUrl(path)
}

export function maskImageAPIKey(apiKey: string): string {
  const trimmed = apiKey.trim()
  if (!trimmed) return '未输入 Key'
  if (trimmed.length <= 10) return '已输入 Key'
  return `已输入 (${trimmed.slice(0, 6)}...${trimmed.slice(-4)})`
}

function imageErrorMessage(error: any, fallback: string): string {
  const responseBody = error?.response?.data
  return responseBody?.error?.message || responseBody?.message || error?.message || fallback
}

function normalizeImageGroupId(groupId?: number | null): number | undefined {
  return typeof groupId === 'number' && Number.isInteger(groupId) && groupId > 0 ? groupId : undefined
}

function buildImageHeaders(apiKey: string, options?: { groupId?: number | null; json?: boolean }): Record<string, string | undefined> {
  const groupId = normalizeImageGroupId(options?.groupId)
  return {
    ...(options?.json ? { 'Content-Type': 'application/json' } : {}),
    'X-API-Key': apiKey,
    Authorization: `Bearer ${apiKey}`,
    ...(groupId ? { [ImageGroupHeader]: String(groupId) } : {}),
    'Accept-Language': apiClient.defaults.headers.common?.['Accept-Language'] as string | undefined,
  }
}

function buildDebugInfo(endpoint: string, payload: ImageGenerationRequest | ImageEditRequest, groupId?: number | null): ImageRequestDebugInfo {
  return {
    endpoint,
    hasBearer: true,
    hasXApiKey: true,
    groupId: normalizeImageGroupId(groupId),
    model: payload.model,
    size: payload.size,
    quality: payload.quality,
    count: payload.n,
  }
}

export async function generateImage(
  apiKey: string,
  payload: ImageGenerationRequest,
  options?: { signal?: AbortSignal; groupId?: number | null },
): Promise<ImageGenerationResponse> {
  const endpoint = buildImageGatewayEndpoint('/v1/images/generations')
  const debug = buildDebugInfo(endpoint, payload, options?.groupId)

  try {
    const { data } = await axios.post<ImageGenerationResponse>(endpoint, payload, {
      signal: options?.signal,
      withCredentials: true,
      headers: buildImageHeaders(apiKey, { groupId: options?.groupId, json: true }),
    })
    return data
  } catch (error: any) {
    const status = error?.response?.status as number | undefined
    const responseBody = error?.response?.data
    throw new ImageRequestError(imageErrorMessage(error, '生成失败，请稍后重试。'), { status, responseBody, debug })
  }
}

export async function editImage(
  apiKey: string,
  payload: ImageEditRequest,
  options?: { signal?: AbortSignal; groupId?: number | null },
): Promise<ImageGenerationResponse> {
  const endpoint = buildImageGatewayEndpoint('/v1/images/edits')
  const debug = buildDebugInfo(endpoint, payload, options?.groupId)
  const formData = new FormData()
  formData.append('prompt', payload.prompt)
  formData.append('model', payload.model)
  formData.append('image', payload.image)
  if (payload.size) formData.append('size', payload.size)
  if (payload.quality) formData.append('quality', payload.quality)
  if (typeof payload.n === 'number') formData.append('n', String(payload.n))

  try {
    const { data } = await axios.post<ImageGenerationResponse>(endpoint, formData, {
      signal: options?.signal,
      withCredentials: true,
      headers: buildImageHeaders(apiKey, { groupId: options?.groupId }),
    })
    return data
  } catch (error: any) {
    const status = error?.response?.status as number | undefined
    const responseBody = error?.response?.data
    throw new ImageRequestError(imageErrorMessage(error, '图生图失败，请稍后重试。'), { status, responseBody, debug })
  }
}

export async function proxyImage(sourceUrl: string, options?: { signal?: AbortSignal }): Promise<Blob> {
  const { data } = await apiClient.post<Blob>(
    '/user/image-proxy',
    { url: sourceUrl },
    {
      signal: options?.signal,
      responseType: 'blob',
      timeout: 60000,
    },
  )
  return data
}

export const imageAPI = {
  generateImage,
  editImage,
  proxyImage,
}

export default imageAPI
