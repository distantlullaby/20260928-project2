import { reactive } from 'vue'

// 全局轻提示：toast.show('内容') / toast.show('成功', 'ok')
export const toast = reactive({
  text: '',
  type: '',
  visible: false,
  _timer: null,
  show(text, type = '') {
    this.text = text
    this.type = type
    this.visible = true
    clearTimeout(this._timer)
    this._timer = setTimeout(() => (this.visible = false), 2400)
  }
})
