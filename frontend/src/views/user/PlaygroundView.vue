<template>
  <AppLayout>
    <div class="flex h-[calc(100vh-7rem)] min-h-[640px] flex-col gap-4 lg:flex-row">
      <aside class="flex w-full flex-shrink-0 flex-col gap-4 lg:w-80">
        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4">
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
              {{ t('playground.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('playground.description') }}
            </p>
          </div>

          <div class="space-y-4">
            <div>
              <label class="input-label">{{ t('playground.apiKey') }}</label>
              <select v-model="selectedKeyValue" class="input" :disabled="keysLoading">
                <option value="" disabled>
                  {{ keysLoading ? t('common.loading') : t('playground.noKeys') }}
                </option>
                <option v-for="option in apiKeyOptions" :key="option.value" :value="String(option.value)">
                  {{ option.label }}
                </option>
              </select>
            </div>

            <div>
              <label class="input-label">{{ t('playground.group') }}</label>
              <select v-model="selectedGroupValue" class="input">
                <option value="" disabled>{{ t('playground.noKeyGroups') }}</option>
                <option v-for="group in selectedKeyGroups" :key="group.id" :value="String(group.id)">
                  {{ group.name }} · {{ group.platform }}
                </option>
              </select>
            </div>

            <div>
              <div class="mb-1 flex items-center justify-between gap-2">
                <label class="input-label mb-0">{{ t('playground.model') }}</label>
                <button
                  type="button"
                  class="text-xs font-medium text-primary-600 hover:text-primary-700 disabled:opacity-50 dark:text-primary-400"
                  :disabled="modelsLoading || !selectedKeyId"
                  @click="loadModels"
                >
                  {{ t('playground.reloadModels') }}
                </button>
              </div>
              <input
                v-model="selectedModel"
                class="input font-mono text-sm"
                list="playground-models"
                :placeholder="t('playground.modelPlaceholder')"
              />
              <datalist id="playground-models">
                <option v-for="model in models" :key="model.id" :value="model.id" />
              </datalist>
              <p class="input-hint">
                {{ models.length ? t('playground.modelHint') : t('playground.noModels') }}
              </p>
            </div>

            <div>
              <label class="input-label">{{ t('playground.systemPrompt') }}</label>
              <textarea
                v-model="systemPrompt"
                rows="4"
                class="input resize-none text-sm"
                :placeholder="t('playground.systemPromptPlaceholder')"
              />
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('playground.temperature') }}</label>
                <input v-model.number="temperature" type="number" min="0" max="2" step="0.1" class="input" />
              </div>
              <div>
                <label class="input-label">{{ t('playground.maxTokens') }}</label>
                <input v-model.number="maxTokens" type="number" min="1" step="1" class="input" />
              </div>
            </div>

            <label class="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-600">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('playground.stream') }}
              </span>
              <input v-model="streamEnabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
            </label>
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex gap-2">
            <button type="button" class="btn-secondary flex-1" :disabled="isSending || messages.length === 0" @click="clearConversation">
              {{ t('playground.clear') }}
            </button>
            <button v-if="isSending" type="button" class="btn-secondary flex-1" @click="stopGeneration">
              {{ t('playground.stop') }}
            </button>
          </div>
        </section>
      </aside>

      <main class="flex min-h-0 flex-1 flex-col rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center justify-between gap-3 border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <div>
            <h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
              {{ t('playground.messages') }}
            </h2>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-500">
              {{ selectedModel || t('playground.modelPlaceholder') }}
            </p>
          </div>
          <Icon v-if="isSending" name="sync" size="md" class="animate-spin text-primary-500" />
        </div>

        <div ref="messagesEl" class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
          <div v-if="errorMessage" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-900/20 dark:text-red-200">
            {{ errorMessage }}
          </div>

          <div v-if="messages.length === 0" class="flex h-full min-h-[260px] flex-col items-center justify-center text-center">
            <Icon name="chat" size="xl" class="text-gray-300 dark:text-dark-500" />
            <h3 class="mt-3 text-base font-semibold text-gray-900 dark:text-white">
              {{ t('playground.emptyTitle') }}
            </h3>
            <p class="mt-1 max-w-md text-sm text-gray-500 dark:text-gray-400">
              {{ t('playground.emptyDescription') }}
            </p>
          </div>

          <article
            v-for="message in messages"
            :key="message.id"
            class="flex gap-3"
            :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-[min(760px,85%)] rounded-lg px-4 py-3 text-sm shadow-sm"
              :class="message.role === 'user'
                ? 'bg-primary-600 text-white'
                : 'border border-gray-200 bg-gray-50 text-gray-800 dark:border-dark-600 dark:bg-dark-900 dark:text-gray-100'"
            >
              <div class="mb-1 text-xs font-semibold uppercase opacity-70">
                {{ t(`playground.roles.${message.role}`) }}
              </div>
              <div class="whitespace-pre-wrap break-words leading-6">
                {{ message.content || (message.pending ? t('playground.thinking') : '') }}
              </div>
            </div>
          </article>
        </div>

        <form class="border-t border-gray-200 p-4 dark:border-dark-700" @submit.prevent="sendMessage">
          <div class="flex flex-col gap-3">
            <textarea
              v-model="draft"
              rows="4"
              class="input resize-none"
              :placeholder="t('playground.inputPlaceholder')"
              :disabled="isSending"
              @keydown.meta.enter.prevent="sendMessage"
              @keydown.ctrl.enter.prevent="sendMessage"
            />
            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('playground.submitHint') }}
              </p>
              <button type="submit" class="btn-primary inline-flex items-center justify-center gap-2" :disabled="isSending">
                <Icon name="play" size="sm" />
                {{ isSending ? t('playground.sending') : t('playground.send') }}
              </button>
            </div>
          </div>
        </form>
      </main>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import keysAPI from '@/api/keys'
import playgroundAPI, {
  type PlaygroundChatMessage,
  type PlaygroundModel,
} from '@/api/playground'
import userGroupsAPI from '@/api/groups'
import type { ApiKey, Group } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  buildApiKeySelectionOptions,
  getSelectableApiKeyGroups,
  resolveDefaultApiKeySelection,
} from '@/utils/apiKeySelection'

interface DisplayMessage extends PlaygroundChatMessage {
  id: string
  pending?: boolean
}

const { t } = useI18n()
const appStore = useAppStore()

const availableGroups = ref<Group[]>([])
const apiKeys = ref<ApiKey[]>([])
const models = ref<PlaygroundModel[]>([])
const selectedKeyValue = ref('')
const selectedGroupValue = ref('')
const selectedModel = ref('')
const systemPrompt = ref('')
const temperature = ref(0.7)
const maxTokens = ref<number | null>(1024)
const streamEnabled = ref(true)
const draft = ref('')
const messages = ref<DisplayMessage[]>([])
const errorMessage = ref('')
const isSending = ref(false)
const modelsLoading = ref(false)
const keysLoading = ref(false)
const messagesEl = ref<HTMLElement | null>(null)

let abortController: AbortController | null = null

const apiKeyOptions = computed(() => buildApiKeySelectionOptions(apiKeys.value, availableGroups.value))

const selectedKeyId = computed(() => {
  const value = Number(selectedKeyValue.value)
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

const selectedGroupId = computed(() => {
  const value = Number(selectedGroupValue.value)
  return Number.isFinite(value) && value > 0 ? value : null
})

function nextMessageId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

async function loadInitialData(): Promise<void> {
  await Promise.all([loadGroups(), loadApiKeys()])
  syncDefaultSelection()
  await loadModels()
}

async function loadApiKeys(): Promise<void> {
  keysLoading.value = true
  try {
    const response = await keysAPI.list(1, 1000, {
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'asc',
    })
    apiKeys.value = response.items
  } catch (error: unknown) {
    apiKeys.value = []
    appStore.showError(extractApiErrorMessage(error, t('playground.loadKeysFailed')))
  } finally {
    keysLoading.value = false
  }
}

async function loadGroups(): Promise<void> {
  try {
    availableGroups.value = await userGroupsAPI.getAvailable()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('playground.loadGroupsFailed')))
  }
}

async function loadModels(): Promise<void> {
  if (!selectedKeyId.value) {
    models.value = []
    return
  }
  modelsLoading.value = true
  try {
    const list = await playgroundAPI.getModels(selectedGroupId.value, selectedKeyId.value)
    models.value = list
    if (!selectedModel.value && list[0]) {
      selectedModel.value = list[0].id
    } else if (list.length > 0 && !list.some((model) => model.id === selectedModel.value)) {
      selectedModel.value = list[0].id
    }
  } catch (error: unknown) {
    models.value = []
    appStore.showError(extractApiErrorMessage(error, t('playground.loadModelsFailed')))
  } finally {
    modelsLoading.value = false
  }
}

function syncDefaultSelection(): void {
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

function requestMessagesWith(userText: string): PlaygroundChatMessage[] {
  const history = messages.value
    .filter((message) => message.content.trim())
    .map((message) => ({ role: message.role, content: message.content }))
  const system = systemPrompt.value.trim()
  return [
    ...(system ? [{ role: 'system' as const, content: system }] : []),
    ...history,
    { role: 'user', content: userText },
  ]
}

async function sendMessage(): Promise<void> {
  if (isSending.value) return
  const text = draft.value.trim()
  errorMessage.value = ''

  if (!selectedModel.value.trim()) {
    errorMessage.value = t('playground.modelRequired')
    return
  }
  if (!selectedKeyId.value) {
    errorMessage.value = t('playground.keyRequired')
    return
  }
  if (!selectedGroupId.value) {
    errorMessage.value = t('playground.groupRequired')
    return
  }
  if (!text) {
    errorMessage.value = t('playground.messageRequired')
    return
  }

  const requestMessages = requestMessagesWith(text)
  const userMessage: DisplayMessage = { id: nextMessageId(), role: 'user', content: text }
  const assistantMessage: DisplayMessage = { id: nextMessageId(), role: 'assistant', content: '', pending: true }
  messages.value.push(userMessage, assistantMessage)
  draft.value = ''
  isSending.value = true
  abortController = new AbortController()
  await scrollToBottom()

  try {
    const result = await playgroundAPI.sendChatCompletion(
      {
        model: selectedModel.value.trim(),
        messages: requestMessages,
        stream: streamEnabled.value,
        group_id: selectedGroupId.value,
        temperature: temperature.value,
        max_tokens: maxTokens.value,
      },
      {
        signal: abortController.signal,
        apiKeyId: selectedKeyId.value,
        onDelta: (delta) => {
          assistantMessage.pending = false
          assistantMessage.content += delta
          void scrollToBottom()
        },
      },
    )
    assistantMessage.pending = false
    if (!streamEnabled.value) {
      assistantMessage.content = result.content || t('playground.emptyResponse')
    } else if (!assistantMessage.content) {
      assistantMessage.content = result.content || t('playground.emptyResponse')
    }
  } catch (error: unknown) {
    assistantMessage.pending = false
    if (error instanceof DOMException && error.name === 'AbortError') {
      errorMessage.value = t('playground.aborted')
    } else {
      errorMessage.value = extractApiErrorMessage(error, t('playground.requestFailed'))
    }
    if (!assistantMessage.content) {
      messages.value = messages.value.filter((message) => message.id !== assistantMessage.id)
    }
  } finally {
    isSending.value = false
    abortController = null
    await scrollToBottom()
  }
}

function stopGeneration(): void {
  abortController?.abort()
}

function clearConversation(): void {
  messages.value = []
  errorMessage.value = ''
}

async function scrollToBottom(): Promise<void> {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

watch(selectedKeyValue, () => {
  const groupChanged = syncGroupSelectionForKey()
  if (!groupChanged) {
    void loadModels()
  }
})

watch(selectedGroupValue, () => {
  void loadModels()
})

onMounted(() => {
  void loadInitialData()
})

onBeforeUnmount(() => {
  abortController?.abort()
})
</script>
