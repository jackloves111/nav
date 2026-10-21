<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const rawUrl = typeof route.query.url === 'string' ? route.query.url : ''
const title = typeof route.query.title === 'string' ? route.query.title : ''
const mode = route.query.mode === 'wan' ? 'wan' : 'lan'

function normalizeUrl(url: string): string | null {
  try {
    const parsed = new URL(url)
    if (parsed.protocol === 'http:' || parsed.protocol === 'https:')
      return parsed.href
  }
  catch {
    // 非法 URL，返回 null
  }
  return null
}

const target = normalizeUrl(rawUrl)

onMounted(() => {
  if (!target)
    return
  // 先让目标地址渲染出来，再立即跳转
  setTimeout(() => window.location.replace(target), 150)
})
</script>

<template>
  <div class="jump-page">
    <div class="jump-card">
      <span class="badge" :class="{ wan: mode === 'wan' }">
        {{ mode === 'wan' ? '外网' : '内网' }}
      </span>
      <div v-if="title" class="app-title">
        {{ title }}
      </div>
      <div class="target-url">
        {{ target || rawUrl || '地址为空' }}
      </div>
      <div v-if="target" class="status">
        <span class="spinner" />
        <span>正在打开…</span>
      </div>
      <div v-else class="status error">
        地址无效，无法跳转
      </div>
    </div>
  </div>
</template>

<style scoped>
.jump-page {
  width: 100%;
  height: 100%;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(900px 420px at 25% 10%, #2b3550 0%, transparent 60%),
    radial-gradient(700px 400px at 85% 90%, #22303a 0%, transparent 55%),
    #363636;
  overflow: hidden;
}

.jump-card {
  width: min(720px, 100%);
  padding: 40px 44px 34px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, .1);
  background: rgba(24, 25, 27, .55);
  backdrop-filter: blur(18px);
  box-shadow: 0 24px 70px rgba(0, 0, 0, .45);
  text-align: center;
  color: #fff;
}

.badge {
  display: inline-block;
  padding: 6px 12px;
  margin-bottom: 20px;
  border-radius: 99px;
  border: 1px solid rgba(91, 140, 255, .38);
  background: rgba(91, 140, 255, .18);
  color: #b9ccff;
  font-size: 12px;
  line-height: 1;
}

.badge.wan {
  border-color: rgba(58, 176, 120, .38);
  background: rgba(58, 176, 120, .16);
  color: #b7e8cd;
}

.app-title {
  margin-bottom: 14px;
  font-size: 20px;
  font-weight: 500;
  word-break: break-all;
}

.target-url {
  padding: 14px 18px;
  margin-bottom: 22px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, .07);
  background: rgba(0, 0, 0, .28);
  font-family: Consolas, "Cascadia Mono", monospace;
  font-size: 22px;
  line-height: 1.5;
  word-break: break-all;
}

.status {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  color: #a8adb5;
  font-size: 13px;
}

.status.error {
  color: #f3b7b3;
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, .18);
  border-top-color: rgba(255, 255, 255, .78);
  border-radius: 50%;
  animation: jump-spin .8s linear infinite;
}

@keyframes jump-spin {
  to { transform: rotate(360deg); }
}
</style>
