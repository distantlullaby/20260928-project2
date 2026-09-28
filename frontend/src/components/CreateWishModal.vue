<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3>🙏 发布"求代吃"</h3>
      <p class="muted small" style="margin:-8px 0 18px">
        发布即冻结悬赏尝鲜币，有人替吃并被你采纳后结算给对方；无人响应时可取消并全额退回。
      </p>

      <form @submit.prevent="submit">
        <div class="field">
          <label>餐厅 / 小摊名称</label>
          <input v-model="form.restaurant_name" maxlength="64" placeholder="例如：巷子里小面" required />
        </div>
        <div class="field">
          <label>地址</label>
          <input v-model="form.address" maxlength="128" placeholder="例如：朝阳区幸福三村北街 12 号" required />
        </div>
        <div class="field">
          <label>心愿菜单</label>
          <input v-model="form.dish_list" maxlength="255" placeholder="最想让别人替你尝的菜，用顿号分隔" required />
        </div>
        <div class="field">
          <label>种草理由</label>
          <textarea v-model="form.reason" maxlength="512" placeholder="为什么种草这家？最想知道什么（分量 / 辣度 / 性价比）？" required></textarea>
        </div>
        <div class="field">
          <label>悬赏尝鲜币</label>
          <input v-model.number="form.reward_coins" type="number" min="1" :max="auth.user?.coins" required />
          <div class="hint">
            当前可用 <span class="coin">{{ auth.user?.coins }}</span>，冻结中 {{ auth.user?.frozen_coins }} 币
          </div>
        </div>

        <button class="btn btn-primary btn-block" :disabled="submitting">
          {{ submitting ? '发布中…' : `冻结 ${form.reward_coins || 0} 币并发布` }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import http, { errMsg } from '../api'
import { auth } from '../store'
import { toast } from '../toast'

const emit = defineEmits(['close', 'created'])
const submitting = ref(false)
const form = reactive({
  restaurant_name: '',
  address: '',
  dish_list: '',
  reason: '',
  reward_coins: 10
})

async function submit() {
  if (form.reward_coins > auth.user.coins) {
    toast.show('尝鲜币余额不足，先去帮别人代吃赚一点吧', 'err')
    return
  }
  submitting.value = true
  try {
    const { data } = await http.post('/api/wishes', form)
    toast.show('心愿已发布，悬赏已冻结 🧊', 'ok')
    emit('created', data.wish)
  } catch (e) {
    toast.show(errMsg(e), 'err')
  } finally {
    submitting.value = false
  }
}
</script>
