<template>
  <AppLayout v-if="isAuthenticated">
    <div class="model-market-shell model-market-shell-auth">
      <ModelMarketBody />
    </div>
  </AppLayout>

  <div v-else class="min-h-screen bg-[#f6f7fb] text-gray-900 dark:bg-dark-950 dark:text-white">
    <header class="border-b border-gray-200/80 bg-white/90 px-5 py-3 backdrop-blur dark:border-dark-800 dark:bg-dark-900/90">
      <div class="mx-auto flex max-w-[1600px] items-center justify-between">
        <router-link to="/home" class="flex items-center gap-3">
          <span class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-lg bg-primary-600 shadow-sm">
            <img v-if="siteLogo" :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
            <Icon v-else name="sparkles" size="sm" class="text-white" />
          </span>
          <span class="text-sm font-semibold">{{ siteName }}</span>
        </router-link>
        <div class="flex items-center gap-2">
          <router-link to="/models" class="rounded-md bg-primary-50 px-3 py-1.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
            模型广场
          </router-link>
          <router-link to="/login" class="rounded-md bg-gray-900 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-white dark:text-dark-950">
            登录
          </router-link>
        </div>
      </div>
    </header>
    <main class="mx-auto max-w-[1600px] px-5 py-6">
      <div class="model-market-shell">
        <ModelMarketBody />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import { getModelMarket, type ModelMarketGroup, type ModelMarketModel, type ModelMarketPrice } from '@/api/modelMarket'
import { extractApiErrorMessage } from '@/utils/apiError'

const authStore = useAuthStore()
const appStore = useAppStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'FluxRouter')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')

const ModelMarketBody = defineComponent({
  name: 'ModelMarketBody',
  setup() {
    const { copyToClipboard } = useClipboard()
    const loading = ref(false)
    const enabled = ref(true)
    const models = ref<ModelMarketModel[]>([])
    const groups = ref<ModelMarketGroup[]>([])
    const selectedGroupId = ref(0)
    const activeProvider = ref('all')
    const activeCategory = ref('all')
    const searchQuery = ref('')
    const selectedModel = ref<ModelMarketModel | null>(null)
    const selectedEndpointType = ref('')

    const providerMeta: Record<string, { label: string; dot: string }> = {
      openai: { label: 'OpenAI', dot: 'bg-emerald-500' },
      anthropic: { label: 'Anthropic', dot: 'bg-orange-500' },
    }

    const categoryFilters = [
      { value: 'all', label: '全部' },
      { value: 'image', label: '图像' },
      { value: 'chat', label: '对话' },
      { value: 'video', label: '视频' },
      { value: 'audio', label: '音频' },
    ]

    const selectedGroup = computed(() => {
      return groups.value.find((group) => group.id === selectedGroupId.value) ?? groups.value[0] ?? null
    })

    const providerFilters = computed(() => {
      const seen = new Set<string>()
      const items = [{ value: 'all', label: '全部', dot: 'bg-gray-400' }]
      for (const model of models.value) {
        const provider = model.provider || model.platform
        if (!provider || seen.has(provider)) continue
        seen.add(provider)
        items.push({
          value: provider,
          label: providerMeta[provider]?.label ?? provider,
          dot: providerMeta[provider]?.dot ?? 'bg-blue-500',
        })
      }
      return items
    })

    const filteredModels = computed(() => {
      const query = searchQuery.value.trim().toLowerCase()
      return models.value.filter((model) => {
        if (!modelMatchesGroup(model)) return false
        if (activeProvider.value !== 'all' && model.provider !== activeProvider.value) return false
        if (activeCategory.value !== 'all' && model.category !== activeCategory.value) return false
        if (!query) return true
        return [
          model.id,
          model.display_name,
          model.description,
          model.provider,
          ...model.tags,
        ].some((value) => value.toLowerCase().includes(query))
      })
    })

    const gatewayBaseURL = computed(() => {
      const raw = appStore.cachedPublicSettings?.api_base_url?.trim()
      const fallback = `${window.location.origin}/v1`
      const base = raw || fallback
      return base.replace(/\/$/, '')
    })

    watch(selectedModel, (model) => {
      selectedEndpointType.value = model?.endpoint_types?.[0] ?? ''
    })

    onMounted(() => {
      loadMarket()
    })

    async function loadMarket() {
      loading.value = true
      try {
        const data = await getModelMarket()
        enabled.value = data.enabled
        models.value = data.models
        groups.value = data.groups.length > 0 ? data.groups : [{
          id: 0,
          name: '官方基准价',
          platform: '',
          rate_multiplier: 1,
          subscription_type: 'standard',
          sort_order: -1,
        }]
        if (!groups.value.some((group) => group.id === selectedGroupId.value)) {
          selectedGroupId.value = groups.value[0]?.id ?? 0
        }
      } catch (error) {
        appStore.showError(extractApiErrorMessage(error, '加载模型广场失败'))
      } finally {
        loading.value = false
      }
    }

    function selectProvider(provider: string) {
      activeProvider.value = provider
      if (provider === 'all') return
      const matchingGroup = groups.value.find((group) => group.platform === provider)
      if (matchingGroup) {
        selectedGroupId.value = matchingGroup.id
      }
    }

    function modelMatchesGroup(model: ModelMarketModel): boolean {
      const group = selectedGroup.value
      return !group || !group.platform || group.platform === model.platform
    }

    function providerLabel(provider: string): string {
      return providerMeta[provider]?.label ?? provider
    }

    function providerDot(provider: string): string {
      return providerMeta[provider]?.dot ?? 'bg-blue-500'
    }

    function categoryLabel(category: string): string {
      return categoryFilters.find((item) => item.value === category)?.label ?? category
    }

    function multipliedPrice(price: ModelMarketPrice): number {
      const multiplier = selectedGroup.value?.rate_multiplier ?? 1
      return price.base_price * multiplier
    }

    function formatPrice(price: ModelMarketPrice): string {
      return `$${multipliedPrice(price).toFixed(2)}`
    }

    function summaryPrices(model: ModelMarketModel): ModelMarketPrice[] {
      const keys = model.category === 'image'
        ? ['input_text', 'input_image', 'output_image']
        : ['input', 'output', 'cache_read', 'cache_5m']
      return keys
        .map((key) => model.prices.find((price) => price.key === key))
        .filter((price): price is ModelMarketPrice => Boolean(price))
        .slice(0, 3)
    }

    function endpointLabel(type: string): string {
      switch (type) {
        case 'openai-response':
          return 'OpenAI Responses'
        case 'openai-chat':
          return 'OpenAI Chat'
        case 'openai-images-generations':
          return 'OpenAI Images'
        case 'openai-images-edits':
          return 'OpenAI Image Edits'
        case 'anthropic-messages':
          return 'Claude Messages'
        default:
          return type
      }
    }

    function endpointPath(type: string): string {
      switch (type) {
        case 'openai-response':
          return 'POST /v1/responses'
        case 'openai-chat':
          return 'POST /v1/chat/completions'
        case 'openai-images-generations':
          return 'POST /v1/images/generations'
        case 'openai-images-edits':
          return 'POST /v1/images/edits'
        case 'anthropic-messages':
          return 'POST /v1/messages'
        default:
          return 'POST /v1/chat/completions'
      }
    }

    function endpointURL(type: string): string {
      const base = gatewayBaseURL.value.endsWith('/v1')
        ? gatewayBaseURL.value
        : `${gatewayBaseURL.value}/v1`
      switch (type) {
        case 'openai-response':
          return `${base}/responses`
        case 'openai-chat':
          return `${base}/chat/completions`
        case 'openai-images-generations':
          return `${base}/images/generations`
        case 'openai-images-edits':
          return `${base}/images/edits`
        case 'anthropic-messages':
          return `${base}/messages`
        default:
          return `${base}/chat/completions`
      }
    }

    function requestExample(model: ModelMarketModel, endpointType: string): string {
      if (endpointType === 'openai-images-generations') {
        return `curl ${endpointURL(endpointType)} \\
  -X POST \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${model.id}",
    "prompt": "一张未来城市夜景海报，霓虹灯，电影感构图。",
    "size": "1024x1024",
    "quality": "high",
    "n": 1
  }'`
      }
      if (endpointType === 'openai-images-edits') {
        return `curl ${endpointURL(endpointType)} \\
  -X POST \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -F "model=${model.id}" \\
  -F "prompt=把背景替换成极光天空，保留主体。" \\
  -F "image=@input.png" \\
  -F "size=1024x1024" \\
  -F "quality=high"`
      }
      if (endpointType === 'openai-response') {
        return `curl ${endpointURL(endpointType)} \\
  -X POST \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${model.id}",
    "input": "用一句话介绍这个模型。"
  }'`
      }
      if (endpointType === 'anthropic-messages') {
        return `curl ${endpointURL(endpointType)} \\
  -X POST \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${model.id}",
    "max_tokens": 1024,
    "messages": [{ "role": "user", "content": "用一句话介绍这个模型。" }]
  }'`
      }
      return `curl ${endpointURL(endpointType)} \\
  -X POST \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${model.id}",
    "messages": [{ "role": "user", "content": "用一句话介绍这个模型。" }]
  }'`
    }

    function copy(text: string) {
      copyToClipboard(text, '已复制')
    }

    function copyEndpointValue(value: string) {
      copy(value)
    }

    function openModel(model: ModelMarketModel) {
      selectedModel.value = model
    }

    function closeDrawer() {
      selectedModel.value = null
    }

    return () => h('div', { class: 'model-market-content' }, [
      h('div', { class: 'market-heading' }, [
        h('div', [
          h('h1', '模型广场'),
          h('p', '可用模型与计费价格，支持按厂商、类型、分组筛选。'),
        ]),
        h('button', {
          class: ['market-refresh', loading.value ? 'is-loading' : ''],
          disabled: loading.value,
          title: '刷新',
          onClick: loadMarket,
        }, [h(Icon, { name: 'refresh', size: 'sm' })]),
      ]),
      h('div', { class: 'market-filters' }, [
        h('div', { class: 'market-chip-row' }, providerFilters.value.map((item) =>
          h('button', {
            key: item.value,
            class: ['market-chip', activeProvider.value === item.value ? 'market-chip-active' : ''],
            onClick: () => selectProvider(item.value),
          }, [
            item.value !== 'all' ? h('span', { class: ['market-dot', item.dot] }) : null,
            h('span', item.label),
          ]),
        )),
        h('div', { class: 'market-chip-row market-type-row' }, categoryFilters.map((item) =>
          h('button', {
            key: item.value,
            class: ['market-chip', activeCategory.value === item.value ? 'market-chip-active' : ''],
            onClick: () => { activeCategory.value = item.value },
          }, item.label),
        )),
        h('div', { class: 'market-control-row' }, [
          h('select', {
            class: 'market-select',
            value: selectedGroupId.value,
            onChange: (event: Event) => {
              selectedGroupId.value = Number((event.target as HTMLSelectElement).value)
            },
          }, groups.value.map((group) =>
            h('option', { key: group.id, value: group.id }, `${group.name}${group.platform ? ` · ${providerLabel(group.platform)}` : ''} ×${group.rate_multiplier}`),
          )),
          h('div', { class: 'market-search' }, [
            h(Icon, { name: 'search', size: 'sm' }),
            h('input', {
              value: searchQuery.value,
              placeholder: '搜索模型...',
              onInput: (event: Event) => {
                searchQuery.value = (event.target as HTMLInputElement).value
              },
            }),
          ]),
        ]),
      ]),
      loading.value
        ? h('div', { class: 'market-empty' }, [h(Icon, { name: 'refresh', size: 'lg', class: 'animate-spin' }), h('span', '加载中...')])
        : !enabled.value
          ? h('div', { class: 'market-empty' }, [h(Icon, { name: 'inbox', size: 'lg' }), h('span', '模型广场暂未启用')])
          : filteredModels.value.length === 0
            ? h('div', { class: 'market-empty' }, [h(Icon, { name: 'inbox', size: 'lg' }), h('span', '没有匹配的模型')])
            : h('div', { class: 'market-grid' }, filteredModels.value.map((model) =>
              h('button', {
                key: model.id,
                type: 'button',
                class: 'market-card',
                onClick: () => openModel(model),
              }, [
                h('div', { class: 'market-card-top' }, [
                  h('div', [
                    h('h2', model.display_name || model.id),
                    h('div', { class: 'market-provider' }, [
                      h('span', { class: ['market-dot', providerDot(model.provider)] }),
                      h('span', providerLabel(model.provider)),
                    ]),
                  ]),
                  h('span', { class: 'market-category' }, categoryLabel(model.category)),
                ]),
                h('div', { class: 'market-price-strip' }, summaryPrices(model).map((price) =>
                  h('div', { key: price.key }, [
                    h('span', `${price.label} ${price.unit}`),
                    h('strong', formatPrice(price)),
                  ]),
                )),
                h('div', { class: 'market-tags' }, model.tags.slice(0, 4).map((tag) =>
                  h('span', { key: tag }, tag),
                )),
                h('div', { class: 'market-io-grid' }, [
                  h('div', [h('span', '输入'), h('strong', model.input_modalities.join('、') || '-')]),
                  h('div', [h('span', '输出'), h('strong', model.output_modalities.join('、') || '-')]),
                ]),
                model.description ? h('p', { class: 'market-description' }, model.description) : null,
                h('span', { class: 'market-detail-link' }, '详情页 →'),
              ]),
            )),
      selectedModel.value ? h('div', { class: 'market-drawer-wrap' }, [
        h('div', { class: 'market-drawer-mask', onClick: closeDrawer }),
        h('aside', { class: 'market-drawer' }, [
          h('button', { class: 'market-drawer-close', title: '关闭', onClick: closeDrawer }, [h(Icon, { name: 'x', size: 'sm' })]),
          h('div', { class: 'market-drawer-title' }, [
            h('div', { class: 'market-provider' }, [
              h('span', { class: ['market-dot', providerDot(selectedModel.value.provider)] }),
              h('span', providerLabel(selectedModel.value.provider)),
              h('span', { class: 'market-badge-soft' }, categoryLabel(selectedModel.value.category)),
            ]),
            h('h2', selectedModel.value.display_name || selectedModel.value.id),
          ]),
          h('section', { class: 'drawer-section' }, [
            h('h3', '连接信息'),
            h('div', { class: 'endpoint-tabs' }, selectedModel.value.endpoint_types.map((type) =>
              h('button', {
                key: type,
                class: selectedEndpointType.value === type ? 'endpoint-tab-active' : '',
                onClick: () => { selectedEndpointType.value = type },
              }, endpointLabel(type)),
            )),
            h('div', { class: 'connection-list' }, [
              h('div', [h('span', 'SDK Base URL'), h('code', gatewayBaseURL.value), h('button', { onClick: () => copyEndpointValue(gatewayBaseURL.value) }, [h(Icon, { name: 'copy', size: 'xs' })])]),
              h('div', [h('span', 'Endpoint'), h('code', endpointPath(selectedEndpointType.value)), h('button', { onClick: () => copyEndpointValue(endpointPath(selectedEndpointType.value)) }, [h(Icon, { name: 'copy', size: 'xs' })])]),
              h('div', [h('span', '鉴权'), h('code', 'Authorization: Bearer sk-your-api-key')]),
              h('div', [h('span', '模型参数'), h('code', selectedModel.value.id), h('button', { onClick: () => copyEndpointValue(selectedModel.value?.id ?? '') }, [h(Icon, { name: 'copy', size: 'xs' })])]),
            ]),
          ]),
          h('section', { class: 'drawer-section' }, [
            h('h3', '模型信息'),
            h('div', { class: 'drawer-info-grid' }, [
              h('div', [h('span', '计费模式'), h('strong', selectedModel.value.billing_mode)]),
              h('div', [h('span', '可用分组'), h('strong', selectedGroup.value?.name ?? '-')]),
              h('div', [h('span', '端点类型'), h('strong', selectedModel.value.endpoint_types.map(endpointLabel).join('、'))]),
              h('div', [h('span', '支持输入'), h('strong', selectedModel.value.input_modalities.join('、') || '-')]),
              h('div', [h('span', '支持输出'), h('strong', selectedModel.value.output_modalities.join('、') || '-')]),
              h('div', [h('span', '上下文'), h('strong', selectedModel.value.context_window || '-')]),
            ]),
            selectedModel.value.description ? h('p', { class: 'drawer-description' }, selectedModel.value.description) : null,
          ]),
          h('section', { class: 'drawer-section' }, [
            h('h3', '价格'),
            h('div', { class: 'drawer-price-list' }, selectedModel.value.prices.map((price) =>
              h('div', { key: price.key }, [
                h('span', `${price.label} ${price.unit}`),
                h('strong', formatPrice(price)),
              ]),
            )),
          ]),
          h('section', { class: 'drawer-section' }, [
            h('div', { class: 'drawer-code-title' }, [
              h('h3', '请求示例'),
              h('button', { onClick: () => selectedModel.value && copy(requestExample(selectedModel.value, selectedEndpointType.value)) }, [
                h(Icon, { name: 'copy', size: 'xs' }),
                h('span', '复制'),
              ]),
            ]),
            h('pre', { class: 'drawer-code' }, requestExample(selectedModel.value, selectedEndpointType.value)),
          ]),
        ]),
      ]) : null,
    ])
  },
})
</script>

<style scoped>
.model-market-shell {
  min-height: calc(100vh - 96px);
}

.model-market-shell-auth {
  min-height: calc(100vh - 128px);
}

:deep(.model-market-content) {
  --market-border: rgba(226, 232, 240, 0.95);
  --market-soft: rgba(255, 255, 255, 0.86);
  --market-muted: #64748b;
  color: #111827;
}

:deep(.dark .model-market-content) {
  --market-border: rgba(51, 65, 85, 0.95);
  --market-soft: rgba(15, 23, 42, 0.86);
  --market-muted: #94a3b8;
  color: #f8fafc;
}

:deep(.market-heading) {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

:deep(.market-heading h1) {
  margin: 0 0 6px;
  font-size: 28px;
  font-weight: 800;
  letter-spacing: 0;
}

:deep(.market-heading p) {
  margin: 0;
  color: var(--market-muted);
  font-size: 14px;
}

:deep(.market-refresh) {
  display: inline-flex;
  height: 34px;
  width: 34px;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  background: var(--market-soft);
  color: #475569;
  transition: 160ms ease;
}

:deep(.market-refresh:hover) {
  border-color: #c4b5fd;
  color: #5b21b6;
}

:deep(.market-refresh.is-loading svg) {
  animation: spin 1s linear infinite;
}

:deep(.market-filters) {
  display: grid;
  gap: 12px;
  margin-bottom: 24px;
}

:deep(.market-chip-row) {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

:deep(.market-chip) {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 34px;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  background: var(--market-soft);
  padding: 0 14px;
  color: #475569;
  font-size: 13px;
  font-weight: 600;
  transition: 160ms ease;
}

:deep(.market-chip:hover),
:deep(.market-chip-active) {
  border-color: #8b7cf6;
  color: #4c1d95;
  box-shadow: 0 0 0 3px rgba(139, 124, 246, 0.12);
}

:deep(.market-dot) {
  display: inline-flex;
  height: 8px;
  width: 8px;
  flex: 0 0 auto;
  border-radius: 999px;
}

:deep(.market-control-row) {
  display: flex;
  max-width: 620px;
  flex-wrap: wrap;
  gap: 10px;
}

:deep(.market-select),
:deep(.market-search) {
  height: 40px;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  background: var(--market-soft);
}

:deep(.market-select) {
  min-width: 230px;
  padding: 0 12px;
  color: #334155;
  font-size: 13px;
}

:deep(.market-search) {
  display: flex;
  min-width: 260px;
  flex: 1;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  color: #94a3b8;
}

:deep(.market-search input) {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  font-size: 13px;
  outline: 0;
}

:deep(.market-grid) {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(252px, 1fr));
  gap: 16px;
}

:deep(.market-card) {
  position: relative;
  min-height: 232px;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  background: var(--market-soft);
  padding: 18px;
  text-align: left;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.04);
  transition: border-color 160ms ease, transform 160ms ease, box-shadow 160ms ease;
}

:deep(.market-card:hover) {
  transform: translateY(-2px);
  border-color: #c4b5fd;
  box-shadow: 0 18px 46px rgba(79, 70, 229, 0.12);
}

:deep(.market-card-top) {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

:deep(.market-card h2) {
  margin: 0 0 12px;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 15px;
  font-weight: 800;
}

:deep(.market-provider) {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  color: var(--market-muted);
  font-size: 12px;
  font-weight: 600;
}

:deep(.market-category) {
  flex: 0 0 auto;
  border-radius: 7px;
  background: #eff6ff;
  padding: 3px 7px;
  color: #2563eb;
  font-size: 12px;
  font-weight: 700;
}

:deep(.market-price-strip) {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin: 18px 0 12px;
}

:deep(.market-price-strip div) {
  min-width: 0;
}

:deep(.market-price-strip span),
:deep(.market-io-grid span) {
  display: block;
  overflow: hidden;
  color: var(--market-muted);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.market-price-strip strong) {
  display: block;
  margin-top: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
}

:deep(.market-tags) {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

:deep(.market-tags span) {
  border-radius: 6px;
  border: 1px solid var(--market-border);
  padding: 2px 7px;
  color: #64748b;
  font-size: 11px;
}

:deep(.market-io-grid) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

:deep(.market-io-grid div) {
  border-radius: 7px;
  border: 1px solid var(--market-border);
  padding: 8px 10px;
}

:deep(.market-io-grid strong) {
  display: block;
  margin-top: 3px;
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.market-description) {
  margin: 14px 0 0;
  color: var(--market-muted);
  font-size: 12px;
  line-height: 1.6;
}

:deep(.market-detail-link) {
  display: inline-flex;
  margin-top: 12px;
  color: #4f46e5;
  font-size: 13px;
  font-weight: 700;
}

:deep(.market-empty) {
  display: flex;
  min-height: 260px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 10px;
  border: 1px dashed var(--market-border);
  color: var(--market-muted);
}

:deep(.market-drawer-wrap) {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  justify-content: flex-end;
}

:deep(.market-drawer-mask) {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.52);
}

:deep(.market-drawer) {
  position: relative;
  z-index: 1;
  width: min(460px, 100vw);
  height: 100vh;
  overflow-y: auto;
  background: #fbfbfd;
  padding: 22px;
  box-shadow: -24px 0 60px rgba(15, 23, 42, 0.22);
}

:deep(.dark .market-drawer) {
  background: #0f172a;
}

:deep(.market-drawer-close) {
  position: absolute;
  right: 18px;
  top: 18px;
  display: inline-flex;
  height: 30px;
  width: 30px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid var(--market-border);
  color: #64748b;
}

:deep(.market-drawer-title) {
  padding-right: 36px;
}

:deep(.market-drawer-title h2) {
  margin: 10px 0 24px;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0;
}

:deep(.market-badge-soft) {
  border-radius: 999px;
  background: #eef2ff;
  padding: 2px 7px;
  color: #4f46e5;
}

:deep(.drawer-section) {
  margin-top: 22px;
}

:deep(.drawer-section h3) {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 800;
}

:deep(.endpoint-tabs) {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

:deep(.endpoint-tabs button) {
  border-radius: 8px;
  border: 1px solid var(--market-border);
  padding: 7px 10px;
  color: #64748b;
  font-size: 12px;
  font-weight: 700;
}

:deep(.endpoint-tabs .endpoint-tab-active) {
  border-color: #8b7cf6;
  color: #5b21b6;
  box-shadow: 0 0 0 3px rgba(139, 124, 246, 0.12);
}

:deep(.connection-list) {
  display: grid;
  gap: 10px;
}

:deep(.connection-list > div) {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 6px;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  background: rgba(255, 255, 255, 0.7);
  padding: 11px 12px;
}

:deep(.connection-list span) {
  grid-column: 1 / -1;
  color: var(--market-muted);
  font-size: 12px;
}

:deep(.connection-list code) {
  min-width: 0;
  overflow: hidden;
  color: #0f172a;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.connection-list button) {
  color: #64748b;
}

:deep(.drawer-info-grid) {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

:deep(.drawer-info-grid div) {
  min-width: 0;
  border-radius: 8px;
  border: 1px solid var(--market-border);
  padding: 10px;
}

:deep(.drawer-info-grid span) {
  display: block;
  color: var(--market-muted);
  font-size: 12px;
}

:deep(.drawer-info-grid strong) {
  display: block;
  margin-top: 6px;
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.drawer-description) {
  margin: 12px 0 0;
  color: var(--market-muted);
  font-size: 13px;
  line-height: 1.6;
}

:deep(.drawer-price-list) {
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid var(--market-border);
}

:deep(.drawer-price-list div) {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--market-border);
  padding: 11px 12px;
  font-size: 13px;
}

:deep(.drawer-price-list div:last-child) {
  border-bottom: 0;
}

:deep(.drawer-code-title) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

:deep(.drawer-code-title button) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 7px;
  border: 1px solid var(--market-border);
  padding: 5px 9px;
  color: #475569;
  font-size: 12px;
  font-weight: 700;
}

:deep(.drawer-code) {
  overflow-x: auto;
  border-radius: 8px;
  background: #0f172a;
  padding: 14px;
  color: #dbeafe;
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 720px) {
  :deep(.market-heading) {
    align-items: stretch;
  }

  :deep(.market-control-row),
  :deep(.market-search),
  :deep(.market-select) {
    width: 100%;
  }

  :deep(.market-grid) {
    grid-template-columns: 1fr;
  }

  :deep(.drawer-info-grid) {
    grid-template-columns: 1fr;
  }
}
</style>
