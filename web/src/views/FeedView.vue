<template>
  <div>
    <!-- Hero -->
    <section class="hero">
      <h1>让别人的味蕾，先替你出发 🍜</h1>
      <p>刷到想吃的店却不敢探店？发布心愿、悬赏尝鲜币，路过的吃货替你拍现场、写点评</p>
      <div class="hero-actions">
        <button v-if="auth.isLoggedIn" class="btn btn-primary" @click="showForm = true">＋ 发布求代吃</button>
        <button v-else class="btn btn-primary" @click="ui.openAuth('login')">登录后发布心愿</button>
      </div>
    </section>

    <!-- 过滤栏 -->
    <div class="filter-bar">
      <button
        v-for="f in filters"
        :key="f.value"
        :class="['filter', { active: status === f.value }]"
        @click="changeFilter(f.value)"
      >
        {{ f.label }}
      </button>
      <span v-if="auth.isLoggedIn" class="mine-toggle" @click="toggleMine">
        <input type="checkbox" :checked="mine" readonly /> 只看我发的
      </span>
    </div>

    <!-- 卡片列表 -->
    <div v-if="loading && list.length === 0" class="empty"><span class="emoji">🍱</span>加载中…</div>
    <div v-else-if="list.length === 0" class="empty">
      <span class="emoji">🛎️</span>
      {{ mine ? '你还没有发布过心愿' : '暂时没有人求代吃，来当第一个许愿的人吧' }}
    </div>

    <div class="feed">
      <WishPostcard
        v-for="(w, i) in list"
        :key="w.id"
        :wish="w"
        :expanded="expandedId === w.id"
        @toggle-expand="toggleExpand(w.id)"
        @goEat="onGoEat"
        @addBounty="onAddBounty"
        @cancel="onCancel"
        @viewAll="openDetail"
        @preview="previewImage"
      />
    </div>

    <div v-if="list.length > 0" class="pager">
      <button class="btn btn-outline btn-sm" :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span class="page-info">第 {{ page }} 页 / 共 {{ totalPages }} 页</span>
      <button class="btn btn-outline btn-sm" :disabled="page >= totalPages" @click="changePage(page + 1)">下一页</button>
    </div>

    <!-- 弹窗们 -->
    <WishFormModal v-if="showForm" @close="showForm = false" @created="onCreated" />
    <ResponseModal v-if="goEatWish" :wish="goEatWish" @close="goEatWish = null" @submitted="onResponseSubmitted" />
    <WishDetailModal
      v-if="detailWishId"
      :wish-id="detailWishId"
      @close="detailWishId = null"
      @accepted="reload"
      @preview="previewImage"
    />
    <div v-if="preview" class="lightbox" @click="preview = ''">
      <img :src="preview" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { listWishes, addBounty, cancelWish, getMe } from '../api'
import { useAuth } from '../store/auth'
import { useUI } from '../store/ui'
import { toast } from '../store/toast'
import WishPostcard from '../components/WishPostcard.vue'
import WishFormModal from '../components/WishFormModal.vue'
import ResponseModal from '../components/ResponseModal.vue'
import WishDetailModal from '../components/WishDetailModal.vue'

const auth = useAuth()
const ui = useUI()
const list = ref([])
const total = ref(0)
const page = ref(1)
const size = 10
const loading = ref(false)
const status = ref('')
const mine = ref(false)
const expandedId = ref(null)

const showForm = ref(false)
const goEatWish = ref(null)
const detailWishId = ref(null)
const preview = ref('')

const filters = [
  { label: '全部', value: '' },
  { label: '🟢 求代吃中', value: 'open' },
  { label: '✅ 已采纳', value: 'settled' },
  { label: '已取消', value: 'canceled' }
]
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size)))

async function reload() {
  loading.value = true
  try {
    const params = { page: page.value, size }
    if (status.value) params.status = status.value
    if (mine.value) params.mine = 1
    const data = await listWishes(params)
    list.value = data.list
    total.value = data.total
    if (auth.isLoggedIn) {
      const { user } = await getMe()
      auth.setUser(user)
    }
  } catch (e) {
    toast(e.message)
  } finally {
    loading.value = false
  }
}

function changeFilter(v) {
  status.value = v
  page.value = 1
  reload()
}
function changePage(p) {
  page.value = p
  reload()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
function toggleMine() {
  if (!auth.isLoggedIn) {
    ui.openAuth('login')
    return
  }
  mine.value = !mine.value
  page.value = 1
  reload()
}
function toggleExpand(id) {
  expandedId.value = expandedId.value === id ? null : id
}

function onCreated() {
  showForm.value = false
  status.value = ''
  mine.value = false
  page.value = 1
  reload()
}
function onResponseSubmitted() {
  goEatWish.value = null
  reload()
}

function onGoEat(wish) {
  if (!auth.isLoggedIn) {
    ui.openAuth('login')
    return
  }
  goEatWish.value = wish
}
function openDetail(wish) {
  detailWishId.value = wish.id
}
function previewImage(url) {
  preview.value = url
}

async function onAddBounty(wish) {
  const amount = Number(prompt(`当前悬赏 🪙${wish.bounty}，追加多少枚尝鲜币？（可用 ${auth.user?.coin_balance ?? 0}）`, '10'))
  if (!amount || amount < 1) return
  try {
    await addBounty(wish.id, amount)
    toast(`已追加冻结 🪙${amount}`)
    reload()
  } catch (e) {
    toast(e.message)
  }
}
async function onCancel(wish) {
  if (!confirm(`取消后冻结的 🪙${wish.bounty} 将全额退回（已有回应时不可取消）。确认取消？`)) return
  try {
    await cancelWish(wish.id)
    toast('心愿已取消，悬赏币已退回')
    reload()
  } catch (e) {
    toast(e.message)
  }
}

onMounted(reload)
</script>

<style scoped>
.hero {
  background: linear-gradient(135deg, #ff8a50 0%, #ff6b35 55%, #f5a623 130%);
  border-radius: 22px;
  color: #fff;
  padding: 30px 26px;
  text-align: center;
  box-shadow: 0 12px 30px rgba(255, 107, 53, 0.28);
}
.hero h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 800;
}
.hero p {
  margin: 10px auto 18px;
  font-size: 14px;
  opacity: 0.92;
  max-width: 460px;
}
.hero-actions .btn {
  background: #fff;
  color: var(--brand-dark);
  font-size: 15px;
  padding: 10px 24px;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 18px 2px 14px;
  flex-wrap: wrap;
}
.filter {
  background: #fff;
  border: 1px solid var(--line);
  color: var(--ink-2);
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
}
.filter.active {
  background: var(--brand);
  border-color: var(--brand);
  color: #fff;
}
.mine-toggle {
  margin-left: auto;
  font-size: 13px;
  color: var(--ink-2);
  cursor: pointer;
  user-select: none;
}
.feed {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  margin-top: 22px;
}
.page-info {
  font-size: 13px;
  color: var(--ink-3);
}
.lightbox {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.82);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 30px;
  cursor: zoom-out;
}
.lightbox img {
  max-width: 100%;
  max-height: 100%;
  border-radius: 12px;
}
</style>
