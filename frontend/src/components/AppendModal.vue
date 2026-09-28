<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal modal-sm">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3>➕ 追加悬赏</h3>
      <p class="muted small" style="margin:-8px 0 18px">
        当前「{{ wish.restaurant_name }}」悬赏 <span class="coin">{{ wish.reward_coins }}</span>，
        追加会立即从你的可用余额冻结。
      </p>

      <form @submit.prevent="submit">
        <div class="field">
          <label>追加尝鲜币</label>
          <input v-model.number="amount" type="number" min="1" :max="auth.user?.coins" required autofocus />
          <div class="hint">可用 <span class="coin">{{ auth.user?.coins }}</span></div>
        </div>
        <button class="btn btn-primary btn-block" :disabled="submitting">
          {{ submitting ? '提交中…' : `追加并冻结 ${amount || 0} 币` }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import http, { errMsg } from '../api'
import { auth } from '../store'
import { toast } from '../toast'

const props = defineProps({ wish: { type: Object, required: true } })
const emit = defineEmits(['close', 'done'])
const amount = ref(5)
const submitting = ref(false)

async function submit() {
  if (!amount.value || amount.value <= 0) {
    toast.show('追加数量必须大于 0', 'err')
    return
  }
  submitting.value = true
  try {
    const { data } = await http.post(`/api/wishes/${props.wish.id}/append`, { amount: amount.value })
    toast.show(`已追加 ${amount.value} 币，总悬赏 ${data.wish.reward_coins} 币`, 'ok')
    emit('done', data.wish)
  } catch (e) {
    toast.show(errMsg(e), 'err')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.modal-sm { max-width: 420px; }
</style>
