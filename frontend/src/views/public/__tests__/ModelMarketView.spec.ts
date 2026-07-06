import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ModelMarketView from '../ModelMarketView.vue'

const {
  getModelMarket,
  showError,
  copyToClipboard,
} = vi.hoisted(() => ({
  getModelMarket: vi.fn(),
  showError: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('@/api/modelMarket', () => ({
  getModelMarket,
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
  }),
  useAppStore: () => ({
    cachedPublicSettings: {
      api_base_url: 'https://gateway.example.com/v1',
      site_logo: '',
      site_name: 'FluxRouter',
    },
    siteName: 'FluxRouter',
    siteLogo: '',
    showError,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

const RouterLinkStub = { props: ['to'], template: '<a><slot /></a>' }

function mountView() {
  return mount(ModelMarketView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        Icon: true,
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('public ModelMarketView', () => {
  beforeEach(() => {
    getModelMarket.mockReset()
    showError.mockReset()
    copyToClipboard.mockReset()

    getModelMarket.mockResolvedValue({
      enabled: true,
      groups: [
        {
          id: 0,
          name: '官方基准价',
          platform: '',
          rate_multiplier: 1,
          subscription_type: 'standard',
          sort_order: -1,
        },
      ],
      models: [
        {
          id: 'gpt-image-2',
          display_name: 'gpt-image-2',
          provider: 'openai',
          platform: 'openai',
          category: 'image',
          billing_mode: '按量计费',
          description: 'OpenAI 生图模型。',
          endpoint_types: ['openai-images-generations', 'openai-images-edits'],
          input_modalities: ['文本', '图片'],
          output_modalities: ['图片'],
          tags: ['OpenAI', '生图'],
          prices: [
            { key: 'input_text', label: '文本输入', unit: '/ 1M tokens', base_price: 5 },
            { key: 'input_image', label: '图片输入', unit: '/ 1M image tokens', base_price: 8 },
            { key: 'output_image', label: '图片输出', unit: '/ 1M image tokens', base_price: 30 },
          ],
          docs_url: '',
          context_window: '-',
          max_output: '-',
          enabled: true,
          sort_order: 25,
        },
      ],
    })
  })

  it('renders image endpoints and examples for gpt-image-2', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('文本输入 / 1M tokens')
    expect(wrapper.text()).toContain('图片输入 / 1M image tokens')
    expect(wrapper.text()).toContain('图片输出 / 1M image tokens')

    await wrapper.find('button.market-card').trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('OpenAI Images')
    expect(wrapper.text()).toContain('POST /v1/images/generations')
    expect(wrapper.text()).toContain('https://gateway.example.com/v1/images/generations')
    expect(wrapper.text()).toContain('"model": "gpt-image-2"')
    expect(wrapper.text()).toContain('"prompt": "一张未来城市夜景海报，霓虹灯，电影感构图。"')

    const editsTab = wrapper.findAll('button').find((button) => button.text().includes('OpenAI Image Edits'))
    expect(editsTab).toBeTruthy()
    await editsTab!.trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('POST /v1/images/edits')
    expect(wrapper.text()).toContain('https://gateway.example.com/v1/images/edits')
    expect(wrapper.text()).toContain('-F "model=gpt-image-2"')
    expect(wrapper.text()).toContain('-F "image=@input.png"')

    wrapper.unmount()
  })
})
