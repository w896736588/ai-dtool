// utils/eventBus.js
import { ref } from 'vue'

const errorCount = ref(0)
const listeners = []

export const eventBus = {
    // 更新错误计数
    updateErrorCount(count) {
        errorCount.value = count
        listeners.forEach(listener => listener(count))
    },

    // 监听错误计数变化
    onErrorCountChange(callback) {
        listeners.push(callback)
        // 返回取消监听函数
        return () => {
            const index = listeners.indexOf(callback)
            if (index > -1) {
                listeners.splice(index, 1)
            }
        }
    },

    // 获取当前错误计数
    getErrorCount() {
        return errorCount.value
    }
}