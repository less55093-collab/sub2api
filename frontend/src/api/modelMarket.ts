import { apiClient } from './client'

export interface ModelMarketPrice {
  key: string
  label: string
  unit: string
  base_price: number
}

export interface ModelMarketModel {
  id: string
  display_name: string
  provider: string
  platform: string
  category: string
  billing_mode: string
  description: string
  endpoint_types: string[]
  input_modalities: string[]
  output_modalities: string[]
  tags: string[]
  prices: ModelMarketPrice[]
  docs_url: string
  context_window: string
  max_output: string
  enabled: boolean
  sort_order: number
}

export interface ModelMarketGroup {
  id: number
  name: string
  platform: string
  rate_multiplier: number
  subscription_type: string
  sort_order: number
}

export interface ModelMarketConfig {
  enabled: boolean
  visible_group_ids: number[]
  models: ModelMarketModel[]
}

export interface ModelMarketPublicResponse {
  enabled: boolean
  groups: ModelMarketGroup[]
  models: ModelMarketModel[]
}

export interface ModelMarketAdminResponse {
  config: ModelMarketConfig
  groups: ModelMarketGroup[]
}

export async function getModelMarket(options?: { signal?: AbortSignal }): Promise<ModelMarketPublicResponse> {
  const { data } = await apiClient.get<ModelMarketPublicResponse>('/model-market', {
    signal: options?.signal,
  })
  return data
}

export async function getAdminModelMarketConfig(options?: { signal?: AbortSignal }): Promise<ModelMarketAdminResponse> {
  const { data } = await apiClient.get<ModelMarketAdminResponse>('/admin/model-market/config', {
    signal: options?.signal,
  })
  return data
}

export async function updateAdminModelMarketConfig(config: ModelMarketConfig): Promise<ModelMarketConfig> {
  const { data } = await apiClient.put<ModelMarketConfig>('/admin/model-market/config', config)
  return data
}

export default {
  getModelMarket,
  getAdminModelMarketConfig,
  updateAdminModelMarketConfig,
}
