<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-card wide">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3 class="modal-title">🍽️ {{ wish.restaurant }} · 全部代吃回应</h3>
      <p class="modal-sub">心愿菜单「{{ wish.dish }}」 · 悬赏 🪙 {{ wish.bounty }} · {{ responses.length }} 份回应</p>

      <div v-if="loading" class="loading">加载中…</div>
      <div v-else-if="responses.length === 0" class="empty">
        <span class="emoji">🕊️</span>
        还没有人替 TA 去吃
      </div>

      <div v-for="r in responses" :key="r.id" class="resp">
        <div class="resp-head">
          <span class="avatar">{{ (r.taster?.nickname || '吃')[0] }}</span>
          <b>{{ r.taster?.nickname || '代吃官' }}</b>
          <span class="stars">{{ '★'.repeat(r.rating) }}{{ '☆'.repeat(5 - r.rating) }}</span>
          <span class="rtime">{{ formatTime(r.created_at) }}</span>
          <span v-if="wish.accepted_response_id === r.id" class="badge badge-settled">已采纳 ✅</span>
        </div>
        <div class="resp-photos">
          <img v-for="(p, i) in r.photos" :key="i" :src="p" @click="$emit('preview', p)" />
        </div>
        <p class="resp-review">{{ r.review }}</p>
        <div v-if="wish.mine && wish.status === 'open'" class="resp-actions">
          <button class="btn btn-primary btn-sm" :disabled="accepting === r.id || r.mine" @click="onAccept(r)">
            {{ accepting === r.id ? '结算中…' : `确认采纳，结算 🪙${wish.bounty}` }}
          </button>
          <span v-if="r.mine" class="self-tip">不能采纳自己的回应</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getWish, acceptWish } from '../api'
import { toast } from '../store/toast'

const props = defineProps({ wishId: { type: Number, required: true } })
const emit = defineEmits(['close', 'accepted', 'preview'])

const wish = ref({})
const responses = ref([])
const loading = ref(true)
const accepting = ref(0)

async function load() {
  loading.value = true
  try {
    const data = await getWish(props.wishId)
    wish.value = data.wish
    responses.value = data.responses
  } finally {
    loading.value = false
  }
}

async function onAccept(r) {
  accepting.value = r.id
  try {
    await acceptWish(props.wishId, r.id)
    toast('已采纳！悬赏尝鲜币已结算给对方 🎉')
    emit('accepted')
    await load()
  } catch (e) {
    toast(e.message)
  } finally {
    accepting.value = 0
  }
}

function formatTime(t) {
  const d = new Date(t)
  return `${d.getMonth() + 1}月${d.getDate()}日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

onMounted(load)
</script>

<style scoped>
.wide {
  max-width: 560px;
}
.loading {
  text-align: center;
  color: var(--ink-3);
  padding: 40px 0;
}
.resp {
  border: 1px solid var(--line);
  border-radius: 14px;
  padding: 14px;
  margin-top: 12px;
}
.resp-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}
.avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ffb08a, #ff6b35);
  color: #fff;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
}
.stars {
  color: var(--gold);
  font-size: 14px;
  letter-spacing: 1px;
}
.rtime {
  color: var(--ink-3);
  font-size: 12px;
}
.resp-photos {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}
.resp-photos img {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 10px;
  cursor: zoom-in;
  background: #f5ece5;
}
.resp-review {
  margin: 10px 0 0;
  font-size: 14px;
  color: var(--ink-2);
  white-space: pre-wrap;
}
.resp-actions {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.self-tip {
  font-size: 12px;
  color: var(--ink-3);
}
</style>
