<script setup lang="ts">
import { NAvatar, NImage } from 'naive-ui'
import { computed, ref, withDefaults } from 'vue'
import { SvgIconOnline } from '@/components/common'

interface Prop {
  itemIcon?: Panel.ItemIcon | null
  size?: number // 默认70
  forceBackground?: string // 强制背景色
}

const props = withDefaults(defineProps<Prop>(), { size: 70 })
const defaultBackground = '#ffffff'
const defaultStyle = ref({
  width: `${props.size}px`,
  height: `${props.size}px`,
})
const iconExt = computed(() => {
  return props.itemIcon?.src?.split('.').pop()
})

// 文字图标的系统配色（Material Design 600 系，中饱和，配白色文字对比度足够，类似手机 App 图标）
const textIconColors = [
  '#E53935', // Red 600
  '#D81B60', // Pink 600
  '#8E24AA', // Purple 600
  '#5E35B1', // Deep Purple 600
  '#3949AB', // Indigo 600
  '#1E88E5', // Blue 600
  '#00ACC1', // Cyan 600
  '#00897B', // Teal 600
  '#43A047', // Green 600
  '#F4511E', // Deep Orange 600
  '#6D4C41', // Brown 600
  '#546E7A', // Blue Grey 600
]

// 按文字内容稳定分配一个彩色背景，同一文字每次显示颜色固定、避免白底白字同色问题
function pickTextIconColor(text?: string): string {
  const key = text || ''
  let hash = 5381
  for (let i = 0; i < key.length; i++)
    hash = ((hash << 5) + hash + key.charCodeAt(i)) | 0
  const index = Math.abs(hash) % textIconColors.length
  return textIconColors[index]
}

// 文字图标的字号：字符越少字越大，做出手机 App 图标的大字感觉；超长文本由 NAvatar 自动缩放兜底
const textFontSize = computed(() => {
  const len = props.itemIcon?.text?.length || 0
  const base = props.size
  if (len === 1)
    return `${Math.round(base * 0.55)}px`
  if (len === 2)
    return `${Math.round(base * 0.48)}px`
  if (len <= 4)
    return `${Math.round(base * 0.36)}px`
  return `${Math.round(base * 0.3)}px`
})
</script>

<template>
  <div class="item-icon" :style="defaultStyle">
    <slot>
      <template v-if="itemIcon">
        <template v-if="itemIcon?.itemType === 1">
          <NAvatar
            :size="props.size"
            :style="{
              backgroundColor: pickTextIconColor(itemIcon.text),
              fontSize: textFontSize,
              fontWeight: 600,
            }"
          >
            {{ itemIcon.text }}
          </NAvatar>
        </template>

        <template v-else-if="itemIcon?.itemType === 2">
          <div v-if="iconExt === 'svg'" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle }" class="flex justify-center items-center">
            <img :src="itemIcon?.src" class="w-[35px] h-[35px]">
          </div>
          <NImage v-else :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground, ...defaultStyle }" :src="itemIcon?.src" preview-disabled />
        </template>

        <template v-else-if="itemIcon?.itemType === 3">
          <NAvatar :size="props.size" :style="{ backgroundColor: (forceBackground ?? itemIcon?.backgroundColor) || defaultBackground }">
            <SvgIconOnline style="font-size: 35px;" :icon="itemIcon.text" />
          </NAvatar>
        </template>
      </template>

      <template v-else>
        <NAvatar :size="props.size" />
      </template>
    </slot>
  </div>
</template>
