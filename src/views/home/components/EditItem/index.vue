<script setup lang="ts">
import { computed, defineEmits, defineProps, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NForm, NFormItem, NGrid, NGridItem, NInput, NInputGroup, NModal, NSelect, useMessage } from 'naive-ui'
import IconEditor from './IconEditor.vue'
import { edit, getSiteFavicon } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { t } from '@/locales'

interface Props {
  visible: boolean
  itemInfo: Panel.Info | null
  itemGroupId?: number
}

const props = defineProps<Props>()
const emit = defineEmits<Emit>()
const ms = useMessage()
const submitLoading = ref(false)
const getIconLoading = ref([false, false])
const itemIconGroupOptions = ref<{
  label: string
  value: number
}[]>([])

const restoreDefault: Panel.Info = {
  icon: null,
  title: '',
  url: '',
  lanUrl: '',
  description: '',
  openMethod: 2,
}

interface Emit {
  (e: 'update:visible', visible: boolean): void
  (e: 'done', item: Panel.Info): void// 创建完成
}

const model = ref<Panel.Info>(props.itemInfo ? { ...props.itemInfo } : { ...restoreDefault })
const formRef = ref<FormInst | null>(null)

const rules: FormRules = {
  title: {
    required: true,
    trigger: 'blur',
    message: t('form.required'),
  },
  // 修改url规则
  url: {
    trigger: 'blur',
    validator(rule, value, callback) {
      if (!value && !model.value.lanUrl) {
        callback(new Error(t('外网地址或内网地址必须填写其中一个')))
      } else {
        callback()
      }
    }
  },
  // 添加lanUrl规则
  lanUrl: {
    trigger: 'blur', 
    validator(rule, value, callback) {
      if (!value && !model.value.url) {
        callback(new Error(t('外网地址或内网地址必须填写其中一个')))
      } else {
        callback()
      }
    }
  },
  // itemIconGroupId: {
  //   required: true,
  //   trigger: ['blur', 'change'],
  //   message: t('form.required'),
  // },
}

const options = [
  {
    default: true,
    label: t('iconItem.currentPageOpen'),
    value: 1,
  },
  {
    label: t('iconItem.newWindowOpen'),
    value: 2,
  },
  {
    label: t('iconItem.currentPageLayerOpen'),
    value: 3,
  },
]

// 更新值父组件传来的值
const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
  },
})

async function editApi() {
  submitLoading.value = true
  try {
    const { code, data, msg } = await edit<Panel.ItemInfo>(model.value)
    if (code === 0) {
      show.value = false
      model.value = { ...restoreDefault }

      emit('done', data)
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  }
  catch (error) {
    ms.error(t('common.saveFail'))
  }
  submitLoading.value = false
}

const handleValidateButtonClick = (e: MouseEvent) => {
  e.preventDefault()
  formRef.value?.validate((errors) => {
    if (!errors)
      editApi()
  })
}

// 错误信息翻译函数
function translateErrorMessage(errorMsg: string): string {
  if (!errorMsg) return '未知错误'
  
  // 网络超时错误
  if (errorMsg.includes('context deadline exceeded') || errorMsg.includes('Client.Timeout exceeded')) {
    return '网络超时错误'
  }
  
  // 连接被拒绝
  if (errorMsg.includes('connection refused') || errorMsg.includes('connect: connection refused')) {
    return '连接被拒绝，目标服务器无法访问'
  }
  
  // DNS解析失败
  if (errorMsg.includes('no such host') || errorMsg.includes('dns resolution failed')) {
    return 'DNS解析失败，无法找到该域名'
  }
  
  // URL格式错误
  if (errorMsg.includes('Invalid URL') || errorMsg.includes('parse URL error')) {
    return 'URL格式错误，请检查网址是否正确'
  }
  
  // HTTP状态码错误
  if (errorMsg.includes('status code:')) {
    const statusMatch = errorMsg.match(/status code: (\d+)/)
    if (statusMatch) {
      const statusCode = statusMatch[1]
      switch (statusCode) {
        case '404':
          return '页面不存在(404错误)'
        case '403':
          return '访问被禁止(403错误)'
        case '500':
          return '服务器内部错误(500错误)'
        case '502':
          return '网关错误(502错误)'
        case '503':
          return '服务不可用(503错误)'
        default:
          return `HTTP错误(${statusCode})，服务器返回异常状态`
      }
    }
  }
  
  // Hello Favicon服务不可用
  if (errorMsg.includes('hello favicon service unavailable')) {
    return '图标获取服务暂时不可用，请稍后重试'
  }
  
  // 网站获取失败
  if (errorMsg.includes('Failed to fetch website')) {
    return '无法访问目标网站，请检查网址是否正确'
  }
  
  // 解析HTML失败
  if (errorMsg.includes('Failed to parse HTML')) {
    return '网页内容解析失败'
  }
  
  // 未找到图标
  if (errorMsg.includes('No favicon found')) {
    return '该网站未设置图标'
  }
  
  // 保存图标失败
  if (errorMsg.includes('save favicon failed')) {
    return '图标保存失败，请重试'
  }
  
  // SSL/TLS错误
  if (errorMsg.includes('tls:') || errorMsg.includes('certificate')) {
    return 'SSL证书错误，无法建立安全连接'
  }
  
  // 默认返回原始错误信息（如果没有匹配到特定模式）
  return errorMsg
}

async function getIconByUrl(url: string, loadingIndex: number) {
  getIconLoading.value[loadingIndex] = true
  try {
    const { code, data, msg } = await getSiteFavicon<{ iconUrl: string, title: string, description: string }>(url)
    if (code === 0) {
      model.value.icon = {
        itemType: 2,
        src: data.iconUrl,
        backgroundColor: '#ffffff',
      }
      // 如果标题为空，则不更新标题
      if (data.title && !model.value.title) {
        model.value.title = data.title
      }
      // 如果描述为空，则不更新描述
      if (data.description && !model.value.description) {
        model.value.description = data.description
      }
    }
    else {
      // 翻译并显示具体的错误信息
      const translatedError = translateErrorMessage(msg || '')
      ms.error(`图标获取失败: ${translatedError}`)
    }
  }
  catch (error: any) {
    // 尝试从错误对象中提取具体的错误信息并翻译
    let errorMsg = ''
    if (error?.response?.data?.msg) {
      errorMsg = error.response.data.msg
    } else if (error?.message) {
      errorMsg = error.message
    }
    
    const translatedError = translateErrorMessage(errorMsg)
    ms.error(`图标获取失败: ${translatedError}`)
  }
  getIconLoading.value[loadingIndex] = false
}

watch(() => props.visible, (newValue) => {
  if (newValue === true) {
    model.value = props.itemInfo ? { ...props.itemInfo } : { ...restoreDefault }
    if (props.itemGroupId)
      model.value.itemIconGroupId = props.itemGroupId
  }

  getGroupListOptions()
})

function getGroupListOptions() {
  getGroupList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ data, code, msg }) => {
    if (code === 0) {
      itemIconGroupOptions.value = []

      for (let i = 0; i < data.list.length; i++) {
        const element = data.list[i]
        if (i === 0 && !model.value.itemIconGroupId) {
          model.value.itemIconGroupId = element.id
          restoreDefault.itemIconGroupId = element.id
        }

        itemIconGroupOptions.value.push({
          value: element.id as number,
          label: element.title as string,
        })
      }
    }
    else {
      ms.error(`${t('iconItem.getGroupFail')}:${msg}`)
    }
  })
}
</script>

<template>
  <NModal v-model:show="show" preset="card" size="small" style="width: 600px;border-radius: 1rem;" :title="itemInfo ? t('iconItem.edit') : t('iconItem.add')">
    <div class="h-[600px] overflow-auto p-[5px]">
      <NForm ref="formRef" :model="model" :rules="rules">
        <NGrid cols="2" :x-gap="10" item-responsive>
          <NGridItem span="2 500:1">
            <NFormItem path="itemIconGroupId" :label="t('iconItem.iconGroup')">
              <NSelect v-model:value="model.itemIconGroupId" :options="itemIconGroupOptions" />
            </NFormItem>
          </NGridItem>
          <NGridItem span="2 500:1">
            <NFormItem path="title" :label="$t('common.title')">
              <NInput v-model:value="model.title" type="text" show-count :maxlength="20" />
            </NFormItem>
          </NGridItem>
        </NGrid>

        <NFormItem path="icon" :label="$t('common.icon')">
          <IconEditor v-model:item-icon="model.icon" />
        </NFormItem>
        <NFormItem path="url" :label="$t('iconItem.url')">
          <!-- <NSelect :style="{ width: '100px' }" :options="urlProtocolOptions" /> -->
          <NInputGroup>
            <NInput v-model:value="model.url" type="text" :maxlength="1000" placeholder="http(s)://" />
            <NButton :disabled="!model.url" :loading="getIconLoading[0]" @click="getIconByUrl(model.url, 0)">
              {{ $t('iconItem.getIcon') }}
            </NButton>
          </NInputGroup>
        </NFormItem>
        <NFormItem path="lanUrl" :label="$t('iconItem.lanUrl')">
          <NInputGroup>
            <NInput v-model:value="model.lanUrl" type="text" :maxlength="1000" :placeholder="$t('iconItem.lanUrlInputPlaceholder')" />
            <NButton :disabled="!model.lanUrl" :loading="getIconLoading[1]" @click="getIconByUrl(model.lanUrl || '', 1)">
              {{ $t('iconItem.getIcon') }}
            </NButton>
          </NInputGroup>
        </NFormItem>
        <NFormItem path="description" :label="$t('common.description')">
          <NInput v-model:value="model.description" type="text" show-count :maxlength="100" />
        </NFormItem>
        <NFormItem path="openMethod" :label="$t('iconItem.openMethod')">
          <NSelect v-model:value="model.openMethod" :options="options" />
        </NFormItem>
      </NForm>
    </div>

    <template #footer>
      <NButton type="success" :loading="submitLoading" style="float: right;" @click="handleValidateButtonClick">
        {{ $t('common.save') }}
      </NButton>
    </template>
  </NModal>
</template>
