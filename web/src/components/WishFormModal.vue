<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-card">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3 class="modal-title">发布求代吃心愿</h3>
      <p class="modal-sub">
        冻结悬赏尝鲜币，让路过的吃货替你先尝一口 · 当前可用
        <b class="coin-inline">🪙 {{ auth.user?.coin_balance ?? 0 }}</b>
      </p>

      <label class="label">心仪餐厅 / 街角摊位 *</label>
      <input v-model="form.restaurant" class="input" placeholder="例如：巷口老陈麻辣烫" maxlength="128" />
      <label class="label">心愿菜单 *</label>
      <input v-model="form.dish" class="input" placeholder="最想吃的那道菜，例如：招牌毛血旺" maxlength="128" />
      <label class="label">地址（可选）</label>
      <input v-model="form.address" class="input" placeholder="帮代吃官找到它" maxlength="256" />
      <label class="label">种草理由 *</label>
      <textarea v-model="form.reason" class="textarea" placeholder="为什么想吃？刷到过、听朋友说、还是馋了很久？" maxlength="1024"></textarea>
      <label class="label">悬赏尝鲜币 *</label>
      <input v-model.number="form.bounty" class="input" type="number" min="1" :max="auth.user?.coin_balance ?? 0" />

      <p v-if="error" class="err">{{ error }}</p>

      <button class="btn btn-primary btn-block" style="margin-top: 18px" :disabled="loading" @click="submit">
        {{ loading ? '冻结悬赏中…' : `冻结 🪙${form.bounty || 0} 并发布` }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { createWish } from '../api'
import { useAuth } from '../store/auth'
import { toast } from '../store/toast'

const emit = defineEmits(['close', 'created'])
const auth = useAuth()
const form = reactive({ restaurant: '', dish: '', reason: '', address: '', bounty: 20 })
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  if (!form.restaurant.trim() || !form.dish.trim() || !form.reason.trim()) {
    error.value = '请填完餐厅、心愿菜单和种草理由'
    return
  }
  if (!form.bounty || form.bounty < 1) {
    error.value = '悬赏至少 1 枚尝鲜币'
    return
  }
  if (form.bounty > (auth.user?.coin_balance ?? 0)) {
    error.value = '可用尝鲜币不足'
    return
  }
  loading.value = true
  try {
    const { wish } = await createWish({
      restaurant: form.restaurant.trim(),
      dish: form.dish.trim(),
      reason: form.reason.trim(),
      address: form.address.trim(),
      bounty: Number(form.bounty)
    })
    toast('心愿已发布，悬赏已冻结 ❄️')
    emit('created', wish)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.err {
  color: var(--red);
  font-size: 13px;
  margin: 10px 0 0;
}
.coin-inline {
  color: #9a6b00;
}
</style>
