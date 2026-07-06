<template>
  <AppLayout>
    <div class="flex h-[calc(100vh-7rem)] min-h-[520px] flex-col gap-4">
      <div class="flex flex-col gap-3 border-b border-gray-200 pb-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
        <div class="min-w-0">
          <h1 class="truncate text-2xl font-semibold text-gray-900 dark:text-white">
            {{ currentPreset?.name || t('chat.title') }}
          </h1>
        </div>
        <div class="flex flex-wrap gap-2">
          <button
            v-if="resolvedUrl"
            type="button"
            class="btn-secondary"
            @click="openResolvedUrl"
          >
            {{ t(isWebLink ? 'chat.openInNewTab' : 'chat.openExternal') }}
          </button>
        </div>
      </div>

      <div
        v-if="isLoading"
        class="flex flex-1 items-center justify-center rounded-lg border border-gray-200 bg-white text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400"
      >
        {{ t('chat.loading') }}
      </div>

      <div
        v-else-if="errorMessage"
        class="flex flex-1 items-center justify-center rounded-lg border border-amber-200 bg-amber-50 px-6 text-center text-sm text-amber-700 dark:border-amber-800/60 dark:bg-amber-900/20 dark:text-amber-200"
      >
        {{ errorMessage }}
      </div>

      <iframe
        v-else-if="isWebLink && resolvedUrl"
        :src="resolvedUrl"
        :title="currentPreset?.name || t('chat.title')"
        class="min-h-0 flex-1 rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900"
        allow="clipboard-read; clipboard-write; microphone; camera"
      />

      <div
        v-else
        class="flex flex-1 flex-col items-center justify-center gap-4 rounded-lg border border-gray-200 bg-white px-6 text-center dark:border-dark-700 dark:bg-dark-800"
      >
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ externalOpened ? t('chat.externalOpened') : t('chat.externalReady') }}
        </p>
        <button type="button" class="btn-primary" :disabled="!resolvedUrl" @click="openResolvedUrl">
          {{ t('chat.openExternal') }}
        </button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { keysAPI } from '@/api/keys'
import { useAppStore } from '@/stores'
import {
  chatLinkRequiresApiKey,
  parseChatPresets,
  resolveChatUrl,
} from '@/utils/chatLinks'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()

const apiKey = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const externalOpened = ref(false)
const lastOpenedUrl = ref('')
let prepareSequence = 0

const chatId = computed(() => String(route.params.chatId ?? ''))
const chatPresets = computed(() => parseChatPresets(appStore.cachedPublicSettings?.chats ?? []))
const currentPreset = computed(() => chatPresets.value.find((item) => item.id === chatId.value) ?? null)
const needsApiKey = computed(() => currentPreset.value ? chatLinkRequiresApiKey(currentPreset.value.url) : false)
const serverAddress = computed(() => {
  const configured = appStore.cachedPublicSettings?.api_base_url?.trim() || appStore.apiBaseUrl.trim()
  if (configured) return configured
  return typeof window !== 'undefined' ? window.location.origin : ''
})
const resolvedUrl = computed(() => {
  const preset = currentPreset.value
  if (!preset) return ''
  if (needsApiKey.value && !apiKey.value) return ''
  return resolveChatUrl({
    template: preset.url,
    apiKey: apiKey.value,
    serverAddress: serverAddress.value,
  })
})
const isWebLink = computed(() => currentPreset.value?.type === 'web')

async function prepareChatLink(): Promise<void> {
  const sequence = ++prepareSequence
  errorMessage.value = ''
  externalOpened.value = false
  lastOpenedUrl.value = ''

  if (!currentPreset.value) {
    if (!appStore.publicSettingsLoaded) {
      isLoading.value = true
      return
    }
    isLoading.value = false
    errorMessage.value = t('chat.notFound')
    return
  }

  if (!needsApiKey.value) {
    apiKey.value = ''
    isLoading.value = false
    return
  }

  isLoading.value = true
  try {
    const response = await keysAPI.list(1, 100, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'asc',
    })
    if (sequence !== prepareSequence) return
    const firstKey = response.items.find((item) => item.status === 'active' && item.key.trim())
    apiKey.value = firstKey?.key ?? ''
    if (!apiKey.value) {
      errorMessage.value = t('chat.noApiKey')
    }
  } catch {
    if (sequence !== prepareSequence) return
    errorMessage.value = t('chat.loadKeyFailed')
  } finally {
    if (sequence === prepareSequence) {
      isLoading.value = false
    }
  }
}

function openResolvedUrl(): void {
  if (!resolvedUrl.value) return
  if (isWebLink.value) {
    window.open(resolvedUrl.value, '_blank', 'noopener,noreferrer')
    return
  }
  try {
    window.location.assign(resolvedUrl.value)
    externalOpened.value = true
    lastOpenedUrl.value = resolvedUrl.value
  } catch {
    errorMessage.value = t('chat.openFailed')
  }
}

watch(
  [currentPreset, needsApiKey, () => appStore.publicSettingsLoaded],
  () => {
    void prepareChatLink()
  },
  { immediate: true },
)

watch(resolvedUrl, (url) => {
  if (!url || isWebLink.value || url === lastOpenedUrl.value) return
  openResolvedUrl()
})

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})
</script>
