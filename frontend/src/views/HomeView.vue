<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <main v-else ref="homeRoot" class="fluxrouter-home station-shell">
    <a class="skip-link" href="#gateway">跳到主要内容</a>
    <div class="cyber-noise" aria-hidden="true"></div>
    <div class="starfield" aria-hidden="true"></div>
    <div ref="rainRoot" class="data-rain" aria-hidden="true"></div>
    <div class="beam beam-a" aria-hidden="true"></div>
    <div class="beam beam-b" aria-hidden="true"></div>
    <div class="nebula nebula-left" aria-hidden="true"></div>
    <div class="nebula nebula-right" aria-hidden="true"></div>
    <div class="scanline" aria-hidden="true"></div>

    <header class="topbar">
      <RouterLink class="brand" to="/" aria-label="FluxRouter 首页">
        <span class="brand-mark"></span>
        <span>
          <strong>FLUXROUTER</strong>
          <small>FluxRouter</small>
        </span>
      </RouterLink>

      <nav class="nav-links" aria-label="主页导航">
        <a href="#gateway" aria-current="page">入口</a>
        <a href="#plans">Token</a>
        <a href="#features">节点</a>
        <RouterLink to="/models">模型</RouterLink>
      </nav>
    </header>

    <section class="hero" id="gateway">
      <div class="hud hud-left" aria-hidden="true">
        <span class="hud-number">98.2%</span>
        <span class="hud-label">TOKEN SUPPLY</span>
        <i></i>
      </div>

      <div class="hud hud-right" aria-hidden="true">
        <span class="hud-number">56.39%</span>
        <span class="hud-label">API TRAFFIC</span>
        <i></i>
      </div>

      <div class="hero-copy">
        <p class="eyebrow">FLUXROUTER TOKEN BLACKBOX</p>
        <h1 class="glitch-title" data-text="FluxRouter">
          Flux
          <span>Router</span>
        </h1>
        <p class="hero-text">
          API Token 售卖与中转入口。稳定额度、极速转发、即买即用，把模型调用变成一条发光的数据高速路。
        </p>

        <div class="hero-stats" aria-label="服务状态">
          <span><b>24H</b> 在线</span>
          <span><b>LOW</b> 延迟</span>
          <span><b>API</b> Ready</span>
        </div>

        <div class="hero-actions">
        <RouterLink class="station-btn station-btn-primary" :to="primaryActionPath">
          <span>{{ primaryActionLabel }}</span>
        </RouterLink>
        <RouterLink class="station-btn station-btn-ghost" to="/models">
          <span>模型广场</span>
        </RouterLink>
        </div>
      </div>

      <div class="portal-wrap" aria-hidden="true">
        <div class="terminal-card">
          <span class="terminal-dot"></span>
          <span class="terminal-dot"></span>
          <span class="terminal-dot"></span>
          <p>&gt; AUTH_CHANNEL: ONLINE</p>
          <p>&gt; TOKEN_POOL: READY</p>
          <p>&gt; RELAY_NODE: FLUX-01</p>
        </div>

        <div class="token-orb token-top">API</div>
        <div class="token-orb token-mid">KEY</div>
        <div class="portal">
          <div class="portal-ring ring-outer"></div>
          <div class="portal-ring ring-mid"></div>
          <div class="portal-ring ring-inner"></div>
          <div class="portal-core">
            <span></span>
          </div>
        </div>
        <div class="data-runway">
          <span></span>
          <span></span>
          <span></span>
        </div>
      </div>
    </section>

    <section class="signal-grid" id="plans" aria-label="Token 套餐">
      <article class="glass-card status-card">
        <span class="card-kicker">STARTER</span>
        <h2>轻量体验</h2>
        <p>适合个人测试、Demo、轻量应用调试，低门槛开始使用 API Token。</p>
        <div class="meter" aria-hidden="true">
          <span style="--value: 82%"></span>
        </div>
      </article>

      <article class="glass-card status-card featured-card">
        <span class="card-kicker">HOT / PRO</span>
        <h2>稳定中转</h2>
        <p>面向日常高频调用，突出稳定可用、响应迅速、额度清晰。</p>
        <div class="meter" aria-hidden="true">
          <span style="--value: 94%"></span>
        </div>
      </article>

      <article class="glass-card status-card">
        <span class="card-kicker">TEAM</span>
        <h2>团队接入</h2>
        <p>适合工作室、团队项目和多应用调用，后续可对接真实套餐。</p>
        <div class="meter" aria-hidden="true">
          <span style="--value: 72%"></span>
        </div>
      </article>
    </section>

    <section class="launch-panel glass-card" id="features">
      <div>
        <span class="card-kicker">WHY FLUXROUTER</span>
        <h2>为 API 调用而生</h2>
        <p>纯前端展示页，已经接入 FluxRouter 首页。主按钮会根据登录状态进入控制台或登录页。</p>
      </div>

      <div class="module-list">
        <RouterLink :to="purchasePath" class="module-card">
          <span>01</span>
          <strong>Token 购买</strong>
          <small>Token Store</small>
        </RouterLink>
        <RouterLink to="/models" class="module-card">
          <span>02</span>
          <strong>模型广场</strong>
          <small>Model Market</small>
        </RouterLink>
        <RouterLink :to="playgroundPath" class="module-card">
          <span>03</span>
          <strong>调试沙盒</strong>
          <small>Playground</small>
        </RouterLink>
        <RouterLink :to="imageStudioPath" class="module-card">
          <span>04</span>
          <strong>生图工作台</strong>
          <small>Image Studio</small>
        </RouterLink>
        <RouterLink to="/available-channels" class="module-card">
          <span>05</span>
          <strong>高速转发</strong>
          <small>Fast Relay</small>
        </RouterLink>
        <RouterLink to="/usage" class="module-card">
          <span>06</span>
          <strong>余额透明</strong>
          <small>Usage Monitor</small>
        </RouterLink>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore, useAppStore } from '@/stores'
import './fluxrouter-home.css'

const authStore = useAuthStore()
const appStore = useAppStore()

const homeRoot = ref<HTMLElement | null>(null)
const rainRoot = ref<HTMLElement | null>(null)

const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const primaryActionPath = computed(() => (isAuthenticated.value ? dashboardPath.value : '/login'))
const primaryActionLabel = computed(() => (isAuthenticated.value ? '进入 FluxRouter' : '登录 FluxRouter'))
const purchasePath = computed(() => (isAuthenticated.value ? '/purchase' : '/login'))
const playgroundPath = computed(() => (isAuthenticated.value ? '/playground' : '/login'))
const imageStudioPath = computed(() => (isAuthenticated.value ? '/image-generator' : '/login'))

let revealObserver: IntersectionObserver | null = null
let removePointerListener: (() => void) | null = null

const rainTokens = ['API', 'KEY', 'TOK', '200', 'GPT', 'JSON', 'SSL', '01']

function createDataRain() {
  const rain = rainRoot.value
  if (!rain) return

  rain.textContent = ''
  Array.from({ length: 42 }, (_, index) => {
    const stream = document.createElement('span')
    stream.textContent = Array.from(
      { length: 12 },
      () => rainTokens[Math.floor(Math.random() * rainTokens.length)]
    ).join(' ')
    stream.style.left = `${(index * 2.45 + Math.random() * 1.6).toFixed(2)}%`
    stream.style.animationDuration = `${(6 + Math.random() * 8).toFixed(2)}s`
    stream.style.animationDelay = `${(-Math.random() * 10).toFixed(2)}s`
    stream.style.opacity = (0.22 + Math.random() * 0.58).toFixed(2)
    rain.appendChild(stream)
  })
}

function setupRevealAnimation() {
  const root = homeRoot.value
  if (!root) return

  revealObserver?.disconnect()

  const cards = root.querySelectorAll<HTMLElement>('.glass-card, .module-card')
  if (!('IntersectionObserver' in window)) {
    cards.forEach((card) => card.classList.add('is-visible'))
    return
  }

  revealObserver = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return

        entry.target.classList.add('is-visible')
        revealObserver?.unobserve(entry.target)
      })
    },
    { threshold: 0.18 }
  )

  cards.forEach((card, index) => {
    card.style.transitionDelay = `${index * 80}ms`
    revealObserver?.observe(card)
  })
}

function setupPointerTracking() {
  const root = homeRoot.value
  if (!root) return

  const onPointerMove = (event: PointerEvent) => {
    const x = (event.clientX / window.innerWidth - 0.5).toFixed(3)
    const y = (event.clientY / window.innerHeight - 0.5).toFixed(3)

    root.style.setProperty('--cursor-x', x)
    root.style.setProperty('--cursor-y', y)
  }

  window.addEventListener('pointermove', onPointerMove)
  removePointerListener = () => window.removeEventListener('pointermove', onPointerMove)
}

function setupStationHome() {
  if (homeContent.value) return

  nextTick(() => {
    createDataRain()
    setupRevealAnimation()
    setupPointerTracking()
  })
}

onMounted(() => {
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  setupStationHome()
})

watch(homeContent, () => {
  if (!homeContent.value) {
    setupStationHome()
  }
})

onBeforeUnmount(() => {
  revealObserver?.disconnect()
  removePointerListener?.()
})
</script>
