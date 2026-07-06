<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-center">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">模型广场配置</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理公开模型广场展示的模型、价格、标签和可选择分组。
          </p>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/models" class="btn btn-secondary">
            <Icon name="externalLink" size="sm" />
            预览
          </router-link>
          <button class="btn btn-secondary" :disabled="loading" @click="loadConfig">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button class="btn btn-primary" :disabled="saving || loading" @click="saveConfig">
            <Icon name="check" size="sm" />
            保存
          </button>
        </div>
      </div>

      <div v-if="loading" class="card flex min-h-[260px] items-center justify-center text-gray-500">
        <Icon name="refresh" size="lg" class="animate-spin" />
      </div>

      <template v-else>
        <section class="card p-5">
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">公开状态</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                关闭后公开接口仍返回空状态，页面会显示暂未启用。
              </p>
            </div>
            <label class="inline-flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-200">
              <input v-model="draft.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
              启用模型广场
            </label>
          </div>
        </section>

        <section class="card p-5">
          <div class="mb-4 flex items-start justify-between gap-4">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">公开可选分组</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                不勾选任何分组时显示全部公开活跃分组；专属分组不会在公开接口展示。
              </p>
            </div>
            <button class="btn btn-secondary btn-sm" @click="draft.visible_group_ids = []">清空限制</button>
          </div>
          <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            <label
              v-for="group in configurableGroups"
              :key="group.id"
              class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 text-sm dark:border-dark-700"
            >
              <input
                type="checkbox"
                class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                :checked="isGroupVisible(group.id)"
                @change="toggleGroup(group.id, $event)"
              />
              <span class="min-w-0">
                <span class="block truncate font-medium text-gray-900 dark:text-white">{{ group.name }}</span>
                <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">
                  {{ group.platform || '全部平台' }} · ×{{ group.rate_multiplier }}
                </span>
              </span>
            </label>
          </div>
        </section>

        <section class="space-y-4">
          <div class="flex items-center justify-between gap-3">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">展示模型</h2>
            <button class="btn btn-secondary btn-sm" @click="addModel">
              <Icon name="plus" size="sm" />
              添加模型
            </button>
          </div>

          <div
            v-for="(model, modelIndex) in draft.models"
            :key="`${model.id}-${modelIndex}`"
            class="card p-5"
          >
            <div class="mb-4 flex flex-col justify-between gap-3 lg:flex-row lg:items-start">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <input v-model="model.display_name" class="input max-w-sm font-mono font-semibold" placeholder="展示名称" />
                  <label class="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                    <input v-model="model.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    显示
                  </label>
                </div>
                <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  价格为官方基准价，公开页按所选分组倍率实时计算。
                </p>
              </div>
              <div class="flex items-center gap-2">
                <button class="btn btn-secondary btn-sm" :disabled="modelIndex === 0" @click="moveModel(modelIndex, -1)">
                  <Icon name="arrowUp" size="sm" />
                </button>
                <button class="btn btn-secondary btn-sm" :disabled="modelIndex === draft.models.length - 1" @click="moveModel(modelIndex, 1)">
                  <Icon name="arrowDown" size="sm" />
                </button>
                <button class="btn btn-danger btn-sm" @click="removeModel(modelIndex)">
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>

            <div class="grid gap-4 lg:grid-cols-4">
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">模型 ID</span>
                <input v-model="model.id" class="input font-mono" placeholder="gpt-5.5" />
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">厂商</span>
                <select v-model="model.provider" class="input">
                  <option value="openai">OpenAI</option>
                  <option value="anthropic">Anthropic</option>
                </select>
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">平台</span>
                <select v-model="model.platform" class="input">
                  <option value="openai">openai</option>
                  <option value="anthropic">anthropic</option>
                </select>
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">排序</span>
                <input v-model.number="model.sort_order" type="number" class="input" />
              </label>
            </div>

            <div class="mt-4 grid gap-4 lg:grid-cols-3">
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">类型</span>
                <select v-model="model.category" class="input">
                  <option value="chat">对话</option>
                  <option value="image">图像</option>
                  <option value="video">视频</option>
                  <option value="audio">音频</option>
                </select>
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">计费模式</span>
                <input v-model="model.billing_mode" class="input" placeholder="按量计费" />
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">文档链接</span>
                <input v-model="model.docs_url" class="input" placeholder="https://..." />
              </label>
            </div>

            <label class="mt-4 block space-y-1 text-sm">
              <span class="text-gray-600 dark:text-gray-300">描述</span>
              <textarea v-model="model.description" rows="2" class="input min-h-[76px]" />
            </label>

            <div class="mt-4 grid gap-4 lg:grid-cols-4">
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">端点类型</span>
                <input :value="model.endpoint_types.join(', ')" class="input" @input="updateArray(model, 'endpoint_types', $event)" />
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">输入能力</span>
                <input :value="model.input_modalities.join(', ')" class="input" @input="updateArray(model, 'input_modalities', $event)" />
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">输出能力</span>
                <input :value="model.output_modalities.join(', ')" class="input" @input="updateArray(model, 'output_modalities', $event)" />
              </label>
              <label class="space-y-1 text-sm">
                <span class="text-gray-600 dark:text-gray-300">标签</span>
                <input :value="model.tags.join(', ')" class="input" @input="updateArray(model, 'tags', $event)" />
              </label>
            </div>

            <div class="mt-5">
              <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">价格</h3>
                <button class="btn btn-secondary btn-sm" @click="addPrice(model)">
                  <Icon name="plus" size="sm" />
                  添加价格项
                </button>
              </div>
              <div class="space-y-2">
                <div
                  v-for="(price, priceIndex) in model.prices"
                  :key="`${price.key}-${priceIndex}`"
                  class="grid gap-2 rounded-lg border border-gray-200 p-3 dark:border-dark-700 md:grid-cols-[1fr_1fr_1fr_120px_auto]"
                >
                  <input v-model="price.key" class="input font-mono" placeholder="input" />
                  <input v-model="price.label" class="input" placeholder="输入" />
                  <input v-model="price.unit" class="input" placeholder="/ 1M" />
                  <input v-model.number="price.base_price" type="number" min="0" step="0.01" class="input" />
                  <button class="btn btn-danger btn-sm" @click="model.prices.splice(priceIndex, 1)">
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import {
  getAdminModelMarketConfig,
  updateAdminModelMarketConfig,
  type ModelMarketConfig,
  type ModelMarketGroup,
  type ModelMarketModel,
} from '@/api/modelMarket'
import { extractApiErrorMessage } from '@/utils/apiError'

const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const groups = ref<ModelMarketGroup[]>([])
const draft = ref<ModelMarketConfig>(emptyModelMarketConfig())

const configurableGroups = computed(() => groups.value.filter((group) => group.id > 0))

onMounted(loadConfig)

async function loadConfig() {
  loading.value = true
  try {
    const data = await getAdminModelMarketConfig()
    draft.value = cloneConfig(data?.config)
    groups.value = normalizeGroups(data?.groups)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '加载模型广场配置失败'))
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    const saved = await updateAdminModelMarketConfig(cloneConfig(draft.value))
    draft.value = cloneConfig(saved)
    appStore.showSuccess('模型广场配置已保存')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, '保存模型广场配置失败'))
  } finally {
    saving.value = false
  }
}

function emptyModelMarketConfig(): ModelMarketConfig {
  return {
    enabled: true,
    visible_group_ids: [],
    models: [],
  }
}

function cloneConfig(config: unknown): ModelMarketConfig {
  try {
    return normalizeConfig(JSON.parse(JSON.stringify(config ?? {})))
  } catch {
    return normalizeConfig(config)
  }
}

function normalizeConfig(value: unknown): ModelMarketConfig {
  const raw = isRecord(value) ? value : {}
  return {
    enabled: typeof raw.enabled === 'boolean' ? raw.enabled : true,
    visible_group_ids: normalizeNumberArray(raw.visible_group_ids),
    models: Array.isArray(raw.models) ? raw.models.map(normalizeModel) : [],
  }
}

function normalizeModel(value: unknown, index: number): ModelMarketModel {
  const raw = isRecord(value) ? value : {}
  const id = stringValue(raw.id)
  const provider = stringValue(raw.provider) || providerFromPlatform(raw.platform) || 'openai'
  const platform = stringValue(raw.platform) || platformFromProvider(provider)
  const isAnthropic = provider === 'anthropic' || platform === 'anthropic'
  return {
    id,
    display_name: stringValue(raw.display_name) || id,
    provider,
    platform,
    category: stringValue(raw.category) || 'chat',
    billing_mode: stringValue(raw.billing_mode) || '按量计费',
    description: stringValue(raw.description),
    endpoint_types: normalizeStringArray(raw.endpoint_types, isAnthropic ? ['anthropic-messages'] : ['openai-response']),
    input_modalities: normalizeStringArray(raw.input_modalities, ['文本']),
    output_modalities: normalizeStringArray(raw.output_modalities, ['文本']),
    tags: normalizeStringArray(raw.tags, []),
    prices: normalizePrices(raw.prices),
    docs_url: stringValue(raw.docs_url),
    context_window: stringValue(raw.context_window),
    max_output: stringValue(raw.max_output),
    enabled: typeof raw.enabled === 'boolean' ? raw.enabled : true,
    sort_order: numberValue(raw.sort_order, (index + 1) * 10),
  }
}

function normalizeGroups(value: unknown): ModelMarketGroup[] {
  if (!Array.isArray(value)) return []
  return value.map((item) => {
    const raw = isRecord(item) ? item : {}
    return {
      id: numberValue(raw.id, 0),
      name: stringValue(raw.name) || '未命名分组',
      platform: stringValue(raw.platform),
      rate_multiplier: numberValue(raw.rate_multiplier, 1),
      subscription_type: stringValue(raw.subscription_type) || 'standard',
      sort_order: numberValue(raw.sort_order, 0),
    }
  })
}

function normalizePrices(value: unknown): ModelMarketModel['prices'] {
  if (!Array.isArray(value)) return []
  return value.map((item) => {
    const raw = isRecord(item) ? item : {}
    return {
      key: stringValue(raw.key),
      label: stringValue(raw.label) || stringValue(raw.key),
      unit: stringValue(raw.unit) || '/ 1M',
      base_price: numberValue(raw.base_price, 0),
    }
  })
}

function normalizeStringArray(value: unknown, fallback: string[]): string[] {
  if (Array.isArray(value)) {
    return uniqueStrings(value.map(stringValue))
  }
  if (typeof value === 'string') {
    return uniqueStrings(splitComma(value))
  }
  return [...fallback]
}

function normalizeNumberArray(value: unknown): number[] {
  if (!Array.isArray(value)) return []
  const seen = new Set<number>()
  const ids: number[] = []
  value.forEach((item) => {
    const id = numberValue(item, 0)
    if (id <= 0 || seen.has(id)) return
    seen.add(id)
    ids.push(id)
  })
  return ids.sort((a, b) => a - b)
}

function uniqueStrings(values: string[]): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  values.forEach((value) => {
    const trimmed = value.trim()
    const key = trimmed.toLowerCase()
    if (!trimmed || seen.has(key)) return
    seen.add(key)
    out.push(trimmed)
  })
  return out
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function numberValue(value: unknown, fallback: number): number {
  const numeric = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(numeric) ? numeric : fallback
}

function providerFromPlatform(value: unknown): string {
  const platform = stringValue(value).toLowerCase()
  if (platform === 'openai') return 'openai'
  if (platform === 'anthropic') return 'anthropic'
  return platform
}

function platformFromProvider(value: string): string {
  const provider = value.toLowerCase()
  if (provider === 'openai') return 'openai'
  if (provider === 'anthropic') return 'anthropic'
  return provider
}

function isGroupVisible(id: number): boolean {
  return draft.value.visible_group_ids.includes(id)
}

function toggleGroup(id: number, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  const ids = new Set(draft.value.visible_group_ids)
  if (checked) {
    ids.add(id)
  } else {
    ids.delete(id)
  }
  draft.value.visible_group_ids = Array.from(ids).sort((a, b) => a - b)
}

function updateArray(
  model: ModelMarketModel,
  field: 'endpoint_types' | 'input_modalities' | 'output_modalities' | 'tags',
  event: Event,
) {
  model[field] = splitComma((event.target as HTMLInputElement).value)
}

function splitComma(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function addModel() {
  const order = draft.value.models.length > 0
    ? Math.max(...draft.value.models.map((model) => model.sort_order || 0)) + 10
    : 10
  draft.value.models.push({
    id: '',
    display_name: '',
    provider: 'openai',
    platform: 'openai',
    category: 'chat',
    billing_mode: '按量计费',
    description: '',
    endpoint_types: ['openai-response'],
    input_modalities: ['文本'],
    output_modalities: ['文本'],
    tags: ['OpenAI', '对话', '按量'],
    prices: [
      { key: 'input', label: '输入', unit: '/ 1M', base_price: 0 },
      { key: 'output', label: '输出', unit: '/ 1M', base_price: 0 },
    ],
    docs_url: '',
    context_window: '',
    max_output: '',
    enabled: true,
    sort_order: order,
  })
}

function addPrice(model: ModelMarketModel) {
  model.prices = normalizePrices(model.prices)
  model.prices.push({ key: '', label: '', unit: '/ 1M', base_price: 0 })
}

function removeModel(index: number) {
  draft.value.models.splice(index, 1)
  normalizeOrder()
}

function moveModel(index: number, delta: number) {
  const next = index + delta
  if (next < 0 || next >= draft.value.models.length) return
  const models = draft.value.models
  const [item] = models.splice(index, 1)
  models.splice(next, 0, item)
  normalizeOrder()
}

function normalizeOrder() {
  draft.value.models.forEach((model, index) => {
    model.sort_order = (index + 1) * 10
  })
}
</script>
