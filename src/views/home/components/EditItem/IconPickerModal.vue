<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NButton, NEmpty, NInput, NModal } from 'naive-ui'
import { presetIconNames } from './presetIconNames'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', show: boolean): void
  (e: 'confirm', src: string): void
}>()

interface PresetIconItem {
  name: string
  url: string
}

// 内置预设图标（public/presetIcons/ 下的静态资源，根路径直接引用，避免被 vite 打包内联）
const icons = computed<PresetIconItem[]>(() => {
  return presetIconNames.map(name => ({
    name,
    url: `/presetIcons/${name}.webp`,
  }))
})

const showValue = computed<boolean>({
  get: () => props.show,
  set: (value: boolean) => emit('update:show', value),
})

const keyword = ref('')
const selected = ref<string>('')

const filteredIcons = computed<PresetIconItem[]>(() => {
  const kw = keyword.value.trim().toLowerCase()
  return kw ? icons.value.filter(item => item.name.toLowerCase().includes(kw)) : icons.value
})

function handleSelect(url: string) {
  selected.value = url
}

function handleConfirm() {
  if (!selected.value)
    return
  emit('confirm', selected.value)
  showValue.value = false
}

function handleCancel() {
  showValue.value = false
}

watch(() => props.show, (value) => {
  if (value) {
    keyword.value = ''
    selected.value = ''
  }
})
</script>

<template>
  <NModal
    v-model:show="showValue"
    preset="card"
    :title="$t('iconItem.presetIcon')"
    style="width: 760px; max-width: 90vw; height: min(560px, calc(100vh - 40px));"
    size="small"
    content-class="flex-1 min-h-0 overflow-auto"
  >
    <NInput
      v-model:value="keyword"
      :placeholder="$t('iconItem.presetIconSearchPlaceholder')"
      clearable
      class="mb-[10px]"
    />
    <NEmpty
      v-if="filteredIcons.length === 0"
      :description="$t('common.noData')"
      class="pt-[40px]"
    />
    <div v-else class="grid grid-cols-8 lg:grid-cols-10 gap-[6px] p-[2px]">
      <div
        v-for="item in filteredIcons"
        :key="item.name"
        class="flex flex-col items-center rounded-lg border p-[4px] cursor-pointer select-none transition-colors"
        :class="item.url === selected ? 'border-blue-500 bg-blue-100' : 'border-gray-200 hover:border-blue-400'"
        :title="item.name"
        @click="handleSelect(item.url)"
      >
        <img :src="item.url" class="w-10 h-10 object-contain" alt="">
        <span class="text-[10px] leading-[14px] w-full text-center truncate text-gray-600">{{ item.name }}</span>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <NButton size="small" class="mr-[8px]" @click="handleCancel">
          {{ $t('common.cancel') }}
        </NButton>
        <NButton size="small" type="primary" :disabled="!selected" @click="handleConfirm">
          {{ $t('common.confirm') }}
        </NButton>
      </div>
    </template>
  </NModal>
</template>
