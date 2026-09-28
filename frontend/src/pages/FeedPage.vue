<template>
  <div class="feed-page">
    <!-- 顶部欢迎区 -->
    <section class="hero">
      <div class="hero-text">
        <h1>先看「别人替你吃的现场」<br />再决定要不要亲自去</h1>
        <p>忙碌的城市青年，把想吃又怕踩雷的餐厅与街角小吃交给同城味蕾代尝</p>
      </div>
      <button v-if="auth.isLoggedIn" class="btn btn-primary hero-btn" @click="showCreate = true">
        🙏 发个求代吃
      </button>
      <RouterLink v-else to="/login" class="btn btn-primary hero-btn">登录后发求代吃</RouterLink>
    </section>

    <!-- 状态筛选 -->
    <div class="filters">
      <button
        v-for="f in filters"
        :key="f.value"
        :class="['filter', { active: status === f.value }]"
        @click="changeFilter(f.value)"
      >{{ f.label }}</button>
    </div>

    <!-- Feed 列表 -->
    <div v-loading="loading">
      <div v-if="loading && !wishes.length" class="loading-box">
        <div class="spinner"></div><p class="muted">正在翻找好吃的…</p>
      </div>

      <WishCard
        v-for="w in wishes"
        :key="w.id"
        :wish="w"
        :responses="detailMap[w.id]?.responses || []"
        :is-owner="auth.user?.id === w.user.id"
        @flip="ensureDetail(w)"
        @taste="openTaste(w)"
        @append="appendTarget = w"
        @cancel="onCancel(w)"
        @settle="(rid) => onSettle(w, rid)"
      />

      <div v-if="!loading && !wishes.length" class="empty">
        <span class="emoji">🍽️</span>
        <p>这里还没有求代吃心愿</p>
        <button v-if="auth.isLoggedIn" class="btn btn-ghost" @click="showCreate = true">我来发第一个</button>
      </div>

      <button
        v-if="hasMore && !loading"
        class="btn btn-plain btn-block load-more"
        @click="loadWishes"
      >加载更多</button>
    </div>

    <!-- 弹层 -->
    <CreateWishModal
      v-if="showCreate"
      @close="showCreate = false"
      @created="onCreated"
    />
    <TasteModal
      v-if="tasteTarget"
      :wish="tasteTarget"
      @close="tasteTarget = null"
      @created="onResponseCreated"
    />
    <AppendModal
      v-if="appendTarget"
      :wish="appendTarget"
      @close="appendTarget = null"
      @done="onAppended"
    />
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import http, { errMsg } from '../api'
import { auth } from '../store'
import { toast } from '../toast'
import WishCard from '../components/WishCard.vue'
import CreateWishModal from '../components/CreateWishModal.vue'
import TasteModal from '../components/TasteModal.vue'
import AppendModal from '../components/AppendModal.vue'

const filters = [
  { value: '', label: '全部' },
  { value: 'open', label: '🔥 求代吃中' },
  { value: 'settled', label: '✅ 已结算' },
  { value: 'canceled', label: '已取消' }
]

const wishes = ref([])
const detailMap = reactive({})
const status = ref('')
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const hasMore = ref(false)

const showCreate = ref(false)
const tasteTarget = ref(null)
const appendTarget = ref(null)

async function loadWishes(reset = false) {
  if (reset) {
    page.value = 1
    wishes.value = []
  }
  loading.value = true
  try {
    const params = { page: page.value, size: 10 }
    if (status.value) params.status = status.value
    const { data } = await http.get('/api/wishes', { params })
    if (reset) wishes.value = data.items
    else wishes.value.push(...data.items)
    total.value = data.total
    hasMore.value = wishes.value.length < total.value
    page.value += 1

    // 预加载已有点赞回应的心愿背面，翻面即可见
    data.items.filter((w) => w.response_count > 0).forEach(loadDetail)
  } catch (e) {
    toast.show(errMsg(e, '加载失败'), 'err')
  } finally {
    loading.value = false
  }
}

// 加载单条心愿详情（含回应列表）
async function loadDetail(w) {
  try {
    const { data } = await http.get(`/api/wishes/${w.id}`)
    detailMap[w.id] = data.wish
  } catch {
    /* 忽略单条加载失败 */
  }
}

// 卡片翻面/需要操作时懒加载详情（含回应列表）
async function ensureDetail(w) {
  if (detailMap[w.id]) return
  await loadDetail(w)
}

function changeFilter(v) {
  status.value = v
  loadWishes(true)
}

function openTaste(w) {
  if (!auth.isLoggedIn) return
  tasteTarget.value = w
}

function onCreated() {
  showCreate.value = false
  refreshMe()
  loadWishes(true)
}

async function onResponseCreated() {
  const w = tasteTarget.value
  tasteTarget.value = null
  if (w) {
    delete detailMap[w.id]
    await ensureDetail(w)
    const idx = wishes.value.findIndex((x) => x.id === w.id)
    if (idx >= 0 && detailMap[w.id]) {
      wishes.value[idx].response_count = detailMap[w.id].responses.length
    }
  }
  refreshMe()
}

async function onAppended(updated) {
  appendTarget.value = null
  refreshMe()
  const idx = wishes.value.findIndex((x) => x.id === updated.id)
  if (idx >= 0) wishes.value[idx] = { ...wishes.value[idx], ...updated }
}

async function onCancel(w) {
  if (!confirm(`确定取消「${w.restaurant_name}」吗？已冻结的 ${w.frozen_coins || w.reward_coins} 币将全额退回。`)) return
  try {
    await http.post(`/api/wishes/${w.id}/cancel`, {})
    toast.show('已取消，悬赏币退回钱包 💰', 'ok')
    refreshMe()
    loadWishes(true)
  } catch (e) {
    toast.show(errMsg(e), 'err')
  }
}

async function onSettle(w, responseId) {
  if (!confirm(`采纳这份代吃并结算 ${w.reward_coins} 尝鲜币？结算后不可撤回。`)) return
  try {
    await http.post(`/api/wishes/${w.id}/settle`, { response_id: responseId })
    toast.show(`已结算 ${w.reward_coins} 币，感谢代吃者 🎉`, 'ok')
    refreshMe()
    loadWishes(true)
  } catch (e) {
    toast.show(errMsg(e), 'err')
  }
}

async function refreshMe() {
  try {
    const { data } = await http.get('/api/me')
    auth.setUser(data.user)
  } catch { /* ignore */ }
}

onMounted(() => loadWishes(true))
</script>

<style scoped>
.hero {
  background: linear-gradient(135deg, #fff4ec 0%, #ffe6d8 100%);
  border-radius: 22px;
  padding: 26px 24px;
  margin: 18px 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  border: 1px solid #ffe0d0;
}
.hero-text h1 {
  margin: 0 0 8px;
  font-size: 22px;
  line-height: 1.45;
  color: var(--brand-dark);
}
.hero-text p { margin: 0; color: var(--ink-soft); font-size: 13.5px; }
.hero-btn { flex-shrink: 0; }

.filters {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  overflow-x: auto;
  padding-bottom: 4px;
}
.filter {
  border: 1.5px solid var(--line);
  background: #fff;
  border-radius: 999px;
  padding: 7px 16px;
  font-size: 13.5px;
  color: var(--ink-soft);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}
.filter.active {
  background: var(--brand);
  border-color: var(--brand);
  color: #fff;
  font-weight: 700;
}

.loading-box { text-align: center; padding: 60px 0; }
.spinner {
  width: 34px; height: 34px;
  border: 3px solid var(--brand-soft);
  border-top-color: var(--brand);
  border-radius: 50%;
  margin: 0 auto 12px;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.load-more { margin-bottom: 10px; }
</style>
