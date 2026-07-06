import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ModelMarketAdminView from '../ModelMarketAdminView.vue'

const {
  getAdminModelMarketConfig,
  updateAdminModelMarketConfig,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getAdminModelMarketConfig: vi.fn(),
  updateAdminModelMarketConfig: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/modelMarket', () => ({
  getAdminModelMarketConfig,
  updateAdminModelMarketConfig,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

vi.mock('@/utils/apiError', () => ({
  extractApiErrorMessage: (_error: unknown, fallback: string) => fallback,
}))

const AppLayoutStub = { template: '<div><slot /></div>' }
const RouterLinkStub = { props: ['to'], template: '<a><slot /></a>' }

function mountView() {
  return mount(ModelMarketAdminView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: true,
        RouterLink: RouterLinkStub,
      },
    },
  })
}

describe('admin ModelMarketAdminView', () => {
  beforeEach(() => {
    getAdminModelMarketConfig.mockReset()
    updateAdminModelMarketConfig.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    updateAdminModelMarketConfig.mockImplementation(async (payload) => payload)
  })

  it('normalizes legacy configs before rendering model fields', async () => {
    getAdminModelMarketConfig.mockResolvedValue({
      config: {
        enabled: true,
        visible_group_ids: null,
        models: [
          {
            id: 'legacy-image',
            provider: 'openai',
            endpoint_types: null,
            input_modalities: null,
            output_modalities: null,
            tags: null,
            prices: null,
          },
        ],
      },
      groups: null,
    })

    const wrapper = mountView()
    await flushPromises()

    expect(showError).not.toHaveBeenCalled()
    expect((wrapper.find<HTMLInputElement>('input[placeholder="展示名称"]').element).value).toBe('legacy-image')
    expect((wrapper.find<HTMLInputElement>('input[placeholder="gpt-5.5"]').element).value).toBe('legacy-image')

    wrapper.unmount()
  })

  it('saves a normalized payload for partial configs', async () => {
    getAdminModelMarketConfig.mockResolvedValue({
      config: {
        enabled: true,
        visible_group_ids: [2, 2, 0, '3'],
        models: [
          {
            id: 'legacy-chat',
            platform: 'openai',
            endpoint_types: undefined,
            input_modalities: undefined,
            output_modalities: undefined,
            tags: undefined,
            prices: undefined,
          },
        ],
      },
      groups: [
        {
          id: 2,
          name: 'OpenAI',
          platform: 'openai',
          rate_multiplier: 0.2,
          subscription_type: 'standard',
          sort_order: 1,
        },
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    const saveButton = wrapper.findAll('button').find((button) => button.text().includes('保存'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    expect(updateAdminModelMarketConfig).toHaveBeenCalledWith(expect.objectContaining({
      visible_group_ids: [2, 3],
      models: [
        expect.objectContaining({
          id: 'legacy-chat',
          endpoint_types: ['openai-response'],
          input_modalities: ['文本'],
          output_modalities: ['文本'],
          tags: [],
          prices: [],
        }),
      ],
    }))
    expect(showSuccess).toHaveBeenCalledWith('模型广场配置已保存')

    wrapper.unmount()
  })
})
