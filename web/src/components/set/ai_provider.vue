<template>
  <ProviderModelPanel ref="panel" mode="set" :show-copy="true" :show-request-log="true" />
</template>

<script>
import { ref } from 'vue'
import ProviderModelPanel from '@/components/ProviderModelPanel.vue'

export default {
  name: 'AiProvider',
  components: { ProviderModelPanel },
  setup() {
    const panel = ref(null)
    // Set.vue 切换 AI tab 时调用，转发到共享子组件的 reload
    const LoadProviderList = function () {
      if (panel.value && typeof panel.value.reload === 'function') {
        panel.value.reload()
      }
    }
    return { panel, LoadProviderList }
  },
}
</script>
