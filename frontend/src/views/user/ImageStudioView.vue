<template>
  <AppLayout>
    <div class="space-y-4">
      <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wide text-primary-600 dark:text-primary-400">
              FluxRouter Image Studio
            </p>
            <h1 class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
              生图工作台
            </h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              选择已有 API Key 和分组，直接测试文生图、图生图、下载和最近结果。
            </p>
          </div>
          <div class="grid grid-cols-3 gap-2 text-center text-xs text-gray-500 dark:text-gray-400 sm:min-w-80">
            <div class="rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
              <strong class="block text-base text-gray-900 dark:text-white">{{ generationMode === 'text' ? '文生图' : '图生图' }}</strong>
              当前模式
            </div>
            <div class="rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
              <strong class="block text-base text-gray-900 dark:text-white">{{ selectedRatio }}</strong>
              画面比例
            </div>
            <div class="rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
              <strong class="block text-base text-gray-900 dark:text-white">{{ recentResults.length }}</strong>
              最近结果
            </div>
          </div>
        </div>
      </section>

      <section id="create" class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(360px,0.82fr)]">
        <form class="space-y-4 rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800" @submit.prevent="handleGenerate">
          <div class="grid grid-cols-2 gap-2 rounded-lg bg-gray-100 p-1 dark:bg-dark-900">
            <button
              type="button"
              class="rounded-md px-3 py-2 text-left text-sm font-medium transition"
              :class="generationMode === 'text' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
              @click="setMode('text')"
            >
              文生图
              <span class="block text-xs font-normal opacity-70">输入描述生成图片</span>
            </button>
            <button
              type="button"
              class="rounded-md px-3 py-2 text-left text-sm font-medium transition"
              :class="generationMode === 'image' ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
              @click="setMode('image')"
            >
              图生图
              <span class="block text-xs font-normal opacity-70">上传参考图后二次创作</span>
            </button>
          </div>

          <div v-if="generationMode === 'image'" class="rounded-lg border border-dashed border-gray-300 p-3 dark:border-dark-600">
            <div class="mb-2 flex items-center justify-between gap-3">
              <div>
                <label class="input-label mb-0" for="reference-image">参考图</label>
                <p class="input-hint">{{ referenceImageName || '支持 PNG / JPG / WEBP' }}</p>
              </div>
              <div class="flex gap-2">
                <button type="button" class="btn-secondary text-xs" @click="triggerReferencePicker">选择图片</button>
                <button type="button" class="btn-secondary text-xs" :disabled="!referenceImageFile" @click="clearReferenceImage">清除</button>
              </div>
            </div>
            <label class="flex min-h-48 cursor-pointer items-center justify-center overflow-hidden rounded-lg bg-gray-50 text-center dark:bg-dark-900" for="reference-image">
              <input id="reference-image" class="sr-only" type="file" accept="image/png,image/jpeg,image/webp" @change="handleReferenceImageChange" />
              <img v-if="referencePreviewUrl" :src="referencePreviewUrl" alt="reference-preview" class="h-full max-h-72 w-full object-contain" />
              <span v-else class="text-sm text-gray-500 dark:text-gray-400">点击上传参考图</span>
            </label>
          </div>

          <div>
            <div class="mb-1 flex items-center justify-between gap-2">
              <label class="input-label mb-0" for="image-prompt">提示词</label>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ prompt.length }} 字</span>
            </div>
            <textarea
              id="image-prompt"
              v-model="prompt"
              rows="5"
              class="input resize-none"
              :placeholder="promptPlaceholder"
            />
            <div class="mt-2 flex flex-wrap gap-2">
              <button v-for="tag in quickPrompts" :key="tag" type="button" class="btn-secondary text-xs" @click="appendPrompt(tag)">
                {{ tag }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 lg:grid-cols-2">
            <label class="lg:col-span-2">
              <span class="input-label">API Key</span>
              <select v-model="selectedKeyValue" class="input" :disabled="keysLoading">
                <option value="" disabled>{{ keysLoading ? '加载中...' : '暂无可用 API Key' }}</option>
                <option v-for="option in apiKeyOptions" :key="option.value" :value="String(option.value)">
                  {{ option.label }}
                </option>
              </select>
              <span class="input-hint">当前 Key：{{ selectedKeyLabel }}。完整 Key 仅用于本次请求，不写入历史记录。</span>
            </label>
            <label class="lg:col-span-2">
              <span class="input-label">分组</span>
              <select v-model="selectedGroupValue" class="input">
                <option value="" disabled>暂无可用分组</option>
                <option v-for="group in selectedKeyGroups" :key="group.id" :value="String(group.id)">
                  {{ group.name }} · {{ group.platform }}
                </option>
              </select>
            </label>
            <label>
              <span class="input-label">模型</span>
              <select v-model="selectedModel" class="input">
                <option v-for="model in modelOptions" :key="model.value" :value="model.value">{{ model.label }}</option>
              </select>
            </label>
            <label>
              <span class="input-label">比例</span>
              <select v-model="selectedRatio" class="input">
                <option v-for="ratio in ratioOptions" :key="ratio.value" :value="ratio.value">{{ ratio.label }}</option>
              </select>
              <span class="input-hint">{{ selectedRatioHint }}</span>
            </label>
            <label>
              <span class="input-label">张数</span>
              <select v-model.number="selectedCount" class="input">
                <option v-for="count in countOptions" :key="count" :value="count">{{ count }} 张</option>
              </select>
            </label>
            <label>
              <span class="input-label">质量</span>
              <select v-model="selectedQuality" class="input">
                <option v-for="quality in qualityOptions" :key="quality.value" :value="quality.value">{{ quality.label }}</option>
              </select>
            </label>
          </div>

          <div class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-900/20 dark:text-amber-200">
            请确认这个 Key 所属分组已开启生图权限；图生图会调用 `/v1/images/edits`，需要上游支持图片编辑。
          </div>
          <div v-if="errorMessage" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-900/20 dark:text-red-200">
            {{ errorMessage }}
          </div>
          <button type="submit" class="btn-primary w-full justify-center" :disabled="generateDisabled">
            {{ generating ? '生成中...' : generationMode === 'text' ? '立即生成' : '开始图生图' }}
          </button>
        </form>

        <aside class="flex min-h-[520px] flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
          <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-700">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">生成结果</h2>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ selectedModel }} · {{ selectedRatio }}</p>
              </div>
              <span class="h-2.5 w-2.5 rounded-full bg-primary-500 shadow-[0_0_0_6px_rgba(14,165,233,0.12)]" />
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto p-4">
            <div v-if="!resultImages.length" class="flex h-full min-h-[360px] flex-col items-center justify-center rounded-lg border border-dashed border-gray-300 text-center dark:border-dark-600">
              <div class="mb-3 flex h-16 w-16 items-center justify-center rounded-lg bg-primary-50 text-sm font-bold text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
                IMG
              </div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">等待第一张作品</h3>
              <p class="mt-1 max-w-xs text-sm text-gray-500 dark:text-gray-400">
                {{ generationMode === 'text' ? '选择 Key、分组并输入提示词后即可测试真实文生图。' : '选择 Key、分组，上传参考图并输入要求后即可测试图生图。' }}
              </p>
            </div>
            <div v-else class="grid gap-3 sm:grid-cols-2">
              <article v-for="(image, index) in resultImages" :key="image.url" class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
                <img :src="image.url" :alt="`generated-image-${index + 1}`" class="aspect-square w-full object-cover" />
                <div class="flex gap-2 p-2">
                  <button type="button" class="btn-secondary flex-1 text-xs" @click="openImage(image)">打开</button>
                  <button type="button" class="btn-secondary flex-1 text-xs" @click="downloadImage(image, index)">下载</button>
                </div>
              </article>
            </div>
          </div>
          <div class="flex gap-2 border-t border-gray-200 p-4 dark:border-dark-700">
            <button type="button" class="btn-secondary flex-1" :disabled="!prompt" @click="copyPrompt">复制提示词</button>
            <button type="button" class="btn-secondary flex-1" :disabled="!resultImages.length" @click="clearResults">清空结果</button>
          </div>
        </aside>
      </section>

      <section class="grid gap-4 xl:grid-cols-[0.85fr_1.15fr]">
        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">灵感提示词</h2>
          <div class="mt-3 space-y-3">
            <button
              v-for="card in inspirationCards"
              :key="card.title"
              type="button"
              class="w-full rounded-lg border border-gray-200 p-3 text-left transition hover:border-primary-300 dark:border-dark-700 dark:hover:border-primary-700"
              @click="usePrompt(card.prompt)"
            >
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ card.title }}</span>
              <span class="mt-1 block text-sm leading-6 text-gray-500 dark:text-gray-400">{{ card.prompt }}</span>
            </button>
          </div>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">最近生成</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">仅在当前浏览器保留 1 小时，不保存完整 API Key。</p>
            </div>
            <button type="button" class="btn-secondary text-xs" :disabled="!recentResults.length" @click="clearHistory">清空</button>
          </div>
          <div v-if="!recentResults.length" class="mt-4 rounded-lg border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
            暂无生成记录。
          </div>
          <div v-else class="mt-4 grid gap-3 md:grid-cols-2">
            <button
              v-for="item in recentResults"
              :key="item.id"
              type="button"
              class="overflow-hidden rounded-lg border border-gray-200 text-left transition hover:border-primary-300 dark:border-dark-700 dark:hover:border-primary-700"
              @click="restoreHistory(item)"
            >
              <img v-if="item.images[0]?.url" :src="item.images[0].url" alt="history-preview" class="aspect-video w-full object-cover" />
              <div class="space-y-1 p-3">
                <strong class="block text-sm text-gray-900 dark:text-white">{{ item.model }}</strong>
                <span class="block text-xs text-primary-600 dark:text-primary-400">{{ item.mode === 'image' ? '图生图' : '文生图' }} · {{ item.ratio }} · {{ item.count }} 张</span>
                <p class="line-clamp-2 text-sm text-gray-500 dark:text-gray-400">{{ item.prompt }}</p>
              </div>
            </button>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { imageAPI, type ImageQuality } from '@/api'
import keysAPI from '@/api/keys'
import userGroupsAPI from '@/api/groups'
import AppLayout from '@/components/layout/AppLayout.vue'
import type { ApiKey, Group } from '@/types'
import {
  buildApiKeySelectionOptions,
  getSelectableApiKeyGroups,
  resolveDefaultApiKeySelection,
} from '@/utils/apiKeySelection'

type GenerationMode = 'text' | 'image'

interface ModelOption {
  value: string
  label: string
}

interface RatioOption {
  value: string
  label: string
  size: string
  hint: string
}

interface ResultImage {
  url: string
  sourceUrl?: string
  objectUrl?: boolean
}

interface HistoryItem {
  id: string
  prompt: string
  model: string
  ratio: string
  count: number
  quality: ImageQuality
  keyLabel: string
  images: ResultImage[]
  createdAt: number
  mode: GenerationMode
}

interface PersistedHistoryPayload {
  version: number
  items: HistoryItem[]
}

const HISTORY_STORAGE_KEY = 'image-studio-recent-results'
const HISTORY_RETENTION_MS = 60 * 60 * 1000
const HISTORY_STORAGE_VERSION = 4

const generationMode = ref<GenerationMode>('text')
const prompt = ref('')
const apiKeys = ref<ApiKey[]>([])
const availableGroups = ref<Group[]>([])
const selectedKeyValue = ref('')
const selectedGroupValue = ref('')
const keysLoading = ref(false)
const generating = ref(false)
const errorMessage = ref('')
const selectedModel = ref('gpt-image-2')
const selectedRatio = ref('16:9')
const selectedCount = ref(1)
const selectedQuality = ref<ImageQuality>('auto')
const resultImages = ref<ResultImage[]>([])
const recentResults = ref<HistoryItem[]>([])
const referenceImageFile = ref<File | null>(null)
const referencePreviewUrl = ref('')
const referenceImageName = ref('')

const quickPrompts = ['电影级', '产品摄影', '国风幻想', '3D 潮玩', '极简海报']
const countOptions = [1, 2, 3, 4]
const qualityOptions = [
  { value: 'auto', label: 'Auto' },
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
] as const
const modelOptions: ModelOption[] = [
  { value: 'gpt-image-2', label: 'gpt-image-2' },
  { value: 'gpt-image-1.5', label: 'gpt-image-1.5' },
  { value: 'gpt-image-1', label: 'gpt-image-1' },
]
const ratioOptions: RatioOption[] = [
  { value: '1:1', label: '1:1 方图', size: '1024x1024', hint: '适合头像、社媒配图和商品图。' },
  { value: '16:9', label: '16:9 横版', size: '1536x1024', hint: '适合封面、海报和产品展示。' },
  { value: '9:16', label: '9:16 竖版', size: '1024x1536', hint: '适合手机海报和短视频封面。' },
  { value: '4:5', label: '4:5 竖图', size: '1024x1280', hint: '适合详情图、人物和运营内容。' },
]
const inspirationCards = [
  { title: '商业产品图', prompt: '未来感智能硬件产品，白色无缝背景，柔和棚拍光，高级材质，干净构图，高清商业广告图。' },
  { title: '封面海报', prompt: '夜晚城市天台，远处霓虹灯和雨后反光，电影级镜头语言，强视觉中心，16:9 横版封面。' },
  { title: '人物头像', prompt: '半身人物肖像，柔和侧光，细腻皮肤质感，简洁背景，高级杂志摄影风格，竖版构图。' },
]

const selectedRatioOption = computed(() => ratioOptions.find((item) => item.value === selectedRatio.value) ?? ratioOptions[0])
const selectedRatioHint = computed(() => selectedRatioOption.value.hint)
const apiKeyOptions = computed(() => buildApiKeySelectionOptions(apiKeys.value, availableGroups.value))
const selectedKeyId = computed(() => {
  const value = Number(selectedKeyValue.value)
  return Number.isFinite(value) && value > 0 ? value : null
})
const selectedGroupId = computed(() => {
  const value = Number(selectedGroupValue.value)
  return Number.isFinite(value) && value > 0 ? value : null
})
const selectedKey = computed(() => {
  const id = selectedKeyId.value
  if (!id) return null
  return apiKeys.value.find((key) => key.id === id) || null
})
const selectedKeyGroups = computed(() =>
  selectedKey.value ? getSelectableApiKeyGroups(selectedKey.value, availableGroups.value) : []
)
const selectedKeyLabel = computed(() => {
  const option = apiKeyOptions.value.find((item) => String(item.value) === selectedKeyValue.value)
  return option?.label || '未选择 Key'
})
const promptPlaceholder = computed(() =>
  generationMode.value === 'text'
    ? '例如：一座漂浮在海上的未来能源塔，玻璃穹顶、金色晨光、电影级构图、高清细节，16:9 比例'
    : '例如：保留主体轮廓，把画面改成赛博朋克夜景风格，增加霓虹灯、雨夜倒影、电影级光影和高细节质感',
)
const generateDisabled = computed(() => {
  if (generating.value || !prompt.value.trim() || !selectedKey.value?.key || !selectedGroupId.value) return true
  return generationMode.value === 'image' && !referenceImageFile.value
})

async function loadKeySelectionData() {
  keysLoading.value = true
  try {
    const [groups, keysResponse] = await Promise.all([
      userGroupsAPI.getAvailable(),
      keysAPI.list(1, 1000, {
        status: 'active',
        sort_by: 'created_at',
        sort_order: 'asc',
      }),
    ])
    availableGroups.value = groups
    apiKeys.value = keysResponse.items
    syncDefaultSelection()
  } catch (error: any) {
    availableGroups.value = []
    apiKeys.value = []
    errorMessage.value = error?.message || '加载 API Key 失败。'
  } finally {
    keysLoading.value = false
  }
}

function syncDefaultSelection() {
  const existingKeyStillSelectable = apiKeyOptions.value.some((option) => String(option.value) === selectedKeyValue.value)
  if (!existingKeyStillSelectable) {
    const selection = resolveDefaultApiKeySelection(apiKeys.value, availableGroups.value)
    selectedKeyValue.value = selection.key ? String(selection.key.id) : ''
    selectedGroupValue.value = selection.groupId ? String(selection.groupId) : ''
    return
  }
  syncGroupSelectionForKey()
}

function syncGroupSelectionForKey(): boolean {
  const groups = selectedKeyGroups.value
  const current = selectedGroupId.value
  if (current && groups.some((group) => group.id === current)) {
    return false
  }

  const nextGroup = groups[0] || null
  const nextValue = nextGroup ? String(nextGroup.id) : ''
  const changed = selectedGroupValue.value !== nextValue
  selectedGroupValue.value = nextValue
  return changed
}

function setMode(mode: GenerationMode) {
  generationMode.value = mode
}

function revokeImageObjectUrl(image: ResultImage) {
  if (image.objectUrl && image.url.startsWith('blob:')) {
    URL.revokeObjectURL(image.url)
  }
}

function revokeImageObjectUrls(images: ResultImage[]) {
  images.forEach(revokeImageObjectUrl)
}

async function localizeImageResult(image: ResultImage): Promise<ResultImage> {
  if (image.url.startsWith('data:')) return image
  if (image.sourceUrl) {
    const blob = await imageAPI.proxyImage(image.sourceUrl)
    return {
      ...image,
      url: URL.createObjectURL(blob),
      objectUrl: true,
    }
  }
  if (image.url.startsWith('blob:')) return image
  return image
}

async function localizeImageResults(images: ResultImage[]): Promise<ResultImage[]> {
  const localized: ResultImage[] = []
  try {
    for (const image of images) {
      localized.push(await localizeImageResult(image))
    }
    return localized
  } catch (error) {
    revokeImageObjectUrls(localized)
    throw error
  }
}

function pruneExpiredHistory(items: HistoryItem[], now = Date.now()) {
  return items.filter((item) => now - item.createdAt < HISTORY_RETENTION_MS)
}

function persistHistory(items: HistoryItem[]) {
  if (typeof window === 'undefined') return
  const payload: PersistedHistoryPayload = {
    version: HISTORY_STORAGE_VERSION,
    items: items.map((item) => ({
      ...item,
      keyLabel: item.keyLabel || '已输入 Key',
      images: item.images.map((image) => image.sourceUrl ? { url: image.sourceUrl, sourceUrl: image.sourceUrl } : { url: image.url }),
    })),
  }
  window.localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(payload))
}

function loadPersistedHistory() {
  if (typeof window === 'undefined') return
  const raw = window.localStorage.getItem(HISTORY_STORAGE_KEY)
  if (!raw) return

  try {
    const parsed = JSON.parse(raw) as PersistedHistoryPayload | HistoryItem[]
    const items = Array.isArray(parsed) ? parsed : parsed.items
    if (!Array.isArray(items)) {
      window.localStorage.removeItem(HISTORY_STORAGE_KEY)
      return
    }
    const validItems = items.filter((item): item is HistoryItem => {
      return Boolean(
        item &&
          typeof item.id === 'string' &&
          typeof item.prompt === 'string' &&
          typeof item.model === 'string' &&
          typeof item.ratio === 'string' &&
          typeof item.count === 'number' &&
          typeof item.quality === 'string' &&
          typeof item.keyLabel === 'string' &&
          typeof item.createdAt === 'number' &&
          (item.mode === 'text' || item.mode === 'image') &&
          Array.isArray(item.images),
      )
    })
    const recoverableItems = validItems
      .map((item) => ({
        ...item,
        images: item.images.filter((image) => image.url.startsWith('data:') || Boolean(image.sourceUrl)),
      }))
      .filter((item) => item.images.length > 0)
    recentResults.value = pruneExpiredHistory(recoverableItems).slice(0, 10)
    persistHistory(recentResults.value)
  } catch {
    window.localStorage.removeItem(HISTORY_STORAGE_KEY)
  }
}

watch(
  recentResults,
  (items) => {
    persistHistory(pruneExpiredHistory(items).slice(0, 10))
  },
  { deep: true },
)

watch(generationMode, () => {
  revokeImageObjectUrls(resultImages.value)
  resultImages.value = []
  errorMessage.value = ''
})

watch(selectedKeyValue, () => {
  syncGroupSelectionForKey()
})

onMounted(() => {
  loadPersistedHistory()
  void loadKeySelectionData()
})

onBeforeUnmount(() => {
  revokeImageObjectUrls(resultImages.value)
  recentResults.value.forEach((item) => revokeImageObjectUrls(item.images))
  if (referencePreviewUrl.value) {
    URL.revokeObjectURL(referencePreviewUrl.value)
  }
})

function appendPrompt(tag: string) {
  prompt.value = prompt.value ? `${prompt.value}，${tag}` : tag
}

function usePrompt(value: string) {
  prompt.value = value
  window.location.hash = '#create'
}

async function copyPrompt() {
  if (!prompt.value) return
  await navigator.clipboard.writeText(prompt.value)
}

function triggerReferencePicker() {
  document.getElementById('reference-image')?.click()
}

function clearReferenceImage() {
  if (referencePreviewUrl.value) {
    URL.revokeObjectURL(referencePreviewUrl.value)
  }
  referenceImageFile.value = null
  referencePreviewUrl.value = ''
  referenceImageName.value = ''
}

function handleReferenceImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (referencePreviewUrl.value) {
    URL.revokeObjectURL(referencePreviewUrl.value)
  }
  referenceImageFile.value = file
  referenceImageName.value = file.name
  referencePreviewUrl.value = URL.createObjectURL(file)
  errorMessage.value = ''
}

function isLocalImageUrl(url: string) {
  return url.startsWith('blob:') || url.startsWith('data:')
}

async function getFreshLocalImageUrl(image: ResultImage): Promise<{ url: string; revoke: boolean }> {
  if (image.sourceUrl) {
    const blob = await imageAPI.proxyImage(image.sourceUrl)
    return { url: URL.createObjectURL(blob), revoke: true }
  }
  if (isLocalImageUrl(image.url)) return { url: image.url, revoke: false }
  throw new Error('图片地址已失效，请重新生成。')
}

async function openImage(image: ResultImage) {
  try {
    const localImage = await getFreshLocalImageUrl(image)
    window.open(localImage.url, '_blank', 'noopener,noreferrer')
    if (localImage.revoke) {
      window.setTimeout(() => URL.revokeObjectURL(localImage.url), 60 * 1000)
    }
  } catch (error: any) {
    errorMessage.value = error?.message || '打开失败，请重新生成。'
  }
}

function triggerImageDownload(url: string, index: number) {
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `image-${Date.now()}-${index + 1}.png`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
}

async function downloadImage(image: ResultImage, index: number) {
  try {
    const localImage = await getFreshLocalImageUrl(image)
    triggerImageDownload(localImage.url, index)
    if (localImage.revoke) {
      window.setTimeout(() => URL.revokeObjectURL(localImage.url), 60 * 1000)
    }
  } catch (error: any) {
    errorMessage.value = error?.message || '下载失败，请重试。'
  }
}

function convertImageResult(item: { url?: string; b64_json?: string }): ResultImage | null {
  if (item.url) return { url: item.url, sourceUrl: item.url }
  if (item.b64_json) return { url: `data:image/png;base64,${item.b64_json}` }
  return null
}

function resolveImageErrorMessage(error: any, mode: GenerationMode): string {
  const responseBody = error?.responseBody
  const nestedMessage = responseBody?.error?.message || responseBody?.message
  const status = error?.status ?? error?.response?.status
  const rawMessage = nestedMessage || error?.message || ''

  if (status === 404) {
    if (rawMessage.includes('Images API is not supported for this platform')) {
      return mode === 'image'
        ? '当前 Key 所属分组不是 OpenAI/Grok 生图分组，或不支持 /images/edits。'
        : '当前 Key 所属分组不支持 /images/generations。'
    }
    return mode === 'image' ? '图生图接口返回 404：当前 Key、分组或上游未提供 /images/edits。' : '生图接口返回 404：当前 Key、分组或上游未提供 /images/generations。'
  }

  if (status === 403) {
    return rawMessage || '当前 Key 所属分组未开启生图权限。'
  }

  return rawMessage || (mode === 'text' ? '生成失败，请稍后重试。' : '图生图失败，请稍后重试。')
}

function pushHistory(images: ResultImage[]) {
  recentResults.value = pruneExpiredHistory([
    {
      id: `${Date.now()}`,
      prompt: prompt.value,
      model: selectedModel.value,
      ratio: selectedRatio.value,
      count: selectedCount.value,
      quality: selectedQuality.value,
      keyLabel: selectedKeyLabel.value,
      images: images.map((image) => image.sourceUrl ? { url: image.sourceUrl, sourceUrl: image.sourceUrl } : { url: image.url }),
      createdAt: Date.now(),
      mode: generationMode.value,
    },
    ...recentResults.value,
  ]).slice(0, 10)
}

async function restoreHistory(item: HistoryItem) {
  generationMode.value = item.mode ?? 'text'
  prompt.value = item.prompt
  selectedModel.value = item.model
  selectedRatio.value = item.ratio
  selectedCount.value = item.count
  selectedQuality.value = item.quality
  revokeImageObjectUrls(resultImages.value)
  resultImages.value = await localizeImageResults(item.images)
  window.location.hash = '#create'
}

function clearResults() {
  revokeImageObjectUrls(resultImages.value)
  resultImages.value = []
  errorMessage.value = ''
}

function clearHistory() {
  recentResults.value.forEach((item) => revokeImageObjectUrls(item.images))
  recentResults.value = []
  window.localStorage.removeItem(HISTORY_STORAGE_KEY)
}

async function handleGenerate() {
  const apiKey = selectedKey.value?.key?.trim()
  if (!apiKey || !selectedGroupId.value || !prompt.value.trim()) return
  if (generationMode.value === 'image' && !referenceImageFile.value) {
    errorMessage.value = '请先上传参考图。'
    return
  }

  generating.value = true
  errorMessage.value = ''
  try {
    const response = generationMode.value === 'text'
      ? await imageAPI.generateImage(apiKey, {
          prompt: prompt.value.trim(),
          model: selectedModel.value,
          size: selectedRatioOption.value.size,
          quality: selectedQuality.value,
          n: selectedCount.value,
        }, {
          groupId: selectedGroupId.value,
        })
      : await imageAPI.editImage(apiKey, {
          prompt: prompt.value.trim(),
          model: selectedModel.value,
          image: referenceImageFile.value as File,
          size: selectedRatioOption.value.size,
          quality: selectedQuality.value,
          n: selectedCount.value,
        }, {
          groupId: selectedGroupId.value,
        })

    const sourceImages = response.data.map(convertImageResult).filter((item): item is ResultImage => Boolean(item))
    const images = await localizeImageResults(sourceImages)
    clearResults()
    resultImages.value = images
    pushHistory(images)
    if (!images.length) errorMessage.value = '接口已返回，但没有拿到图片结果。'
  } catch (error: any) {
    errorMessage.value = resolveImageErrorMessage(error, generationMode.value)
  } finally {
    generating.value = false
  }
}
</script>
