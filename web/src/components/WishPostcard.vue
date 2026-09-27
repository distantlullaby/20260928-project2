<template>
  <div class="postcard" :class="{ flipped }">
    <div class="pc-inner">
      <!-- 正面：心愿菜单 + 种草理由 -->
      <div class="pc-face pc-front">
        <div class="pc-head">
          <div class="pc-user">
            <span class="avatar">{{ avatarText(wish.poster?.nickname) }}</span>
            <div>
              <div class="uname">{{ wish.poster?.nickname || '吃货用户' }}</div>
              <div class="utime">{{ formatTime(wish.created_at) }}</div>
            </div>
          </div>
          <span :class="['badge', statusBadge(wish.status)]">{{ statusText(wish.status) }}</span>
        </div>

        <div class="pc-shop">
          <span class="pin">📍</span>
          <div>
            <div class="shop-name">{{ wish.restaurant }}</div>
            <div v-if="wish.address" class="shop-addr">{{ wish.address }}</div>
          </div>
        </div>

        <div class="pc-dish">
          <span class="dish-emoji">🍽️</span>
          <span class="dish-name">{{ wish.dish }}</span>
        </div>

        <p class="pc-reason">{{ wish.reason }}</p>

        <div class="pc-foot">
          <span class="coin-tag">悬赏 🪙 {{ wish.bounty }}</span>
          <span class="resp-count">👅 {{ wish.response_count }} 人替吃</span>
          <button
            v-if="hasBack"
            class="btn btn-ghost btn-sm flip-btn"
            @click="flipped = !flipped"
          >
            {{ flipped ? '翻回心愿面' : '看现场 →' }}
          </button>
        </div>

        <!-- 展开后的操作区 -->
        <div v-if="expanded" class="pc-actions">
          <template v-if="wish.status === 'open'">
            <button v-if="!wish.mine" class="btn btn-primary btn-sm" @click="$emit('goEat', wish)">
              👅 替 TA 去吃
            </button>
            <template v-else>
              <button class="btn btn-outline btn-sm" @click="$emit('addBounty', wish)">+ 追加悬赏</button>
              <button class="btn btn-outline btn-sm" @click="$emit('cancel', wish)">取消并退回</button>
            </template>
            <button class="btn btn-ghost btn-sm" @click="$emit('viewAll', wish)">
              全部回应 ({{ wish.response_count }})
            </button>
          </template>
          <button v-if="!hasBack" class="btn btn-ghost btn-sm" @click="$emit('viewAll', wish)">
            查看回应 ({{ wish.response_count }})
          </button>
        </div>

        <!-- 收起/展开提示条 -->
        <button class="expand-bar" @click.stop="toggleExpand">
          {{ expanded ? '▲ 收起' : '▼ 展开操作' }}
        </button>
      </div>

      <!-- 背面：现场实拍 + 口味点评（采纳的回应） -->
      <div v-if="hasBack" class="pc-face pc-back">
        <div class="back-stamp">
          现场实拍 · {{ wish.response.taster?.nickname || '代吃官' }}
        </div>
        <div class="back-photos">
          <img
            v-for="(p, i) in wish.response.photos"
            :key="i"
            :src="p"
            alt="现场实拍"
            @click="$emit('preview', p)"
          />
        </div>
        <div class="back-rating">{{ '★'.repeat(wish.response.rating) }}{{ '☆'.repeat(5 - wish.response.rating) }}</div>
        <p class="back-review">{{ wish.response.review }}</p>
        <div class="pc-foot">
          <span class="settled-tip">✅ 发起人已采纳，悬赏已结算 🪙 {{ wish.bounty }}</span>
          <button class="btn btn-ghost btn-sm" @click="flipped = !flipped">↶ 翻回</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  wish: { type: Object, required: true },
  expanded: { type: Boolean, default: false }
})
const emit = defineEmits(['goEat', 'addBounty', 'cancel', 'viewAll', 'preview', 'toggleExpand'])

const flipped = ref(false)
const hasBack = computed(() => !!props.wish.response)

function toggleExpand() {
  emit('toggleExpand')
}

function avatarText(name) {
  return (name || '吃')[0]
}
function statusText(s) {
  return { open: '求代吃中', settled: '已采纳', canceled: '已取消' }[s] || s
}
function statusBadge(s) {
  return { open: 'badge-open', settled: 'badge-settled', canceled: 'badge-canceled' }[s] || ''
}
function formatTime(t) {
  const d = new Date(t)
  const diff = (Date.now() - d.getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  return `${d.getMonth() + 1}月${d.getDate()}日`
}
</script>

<style scoped>
.postcard {
  perspective: 1400px;
}
.pc-inner {
  position: relative;
  transform-style: preserve-3d;
  transition: transform 0.65s cubic-bezier(0.4, 0.2, 0.2, 1);
}
.flipped .pc-inner {
  transform: rotateY(180deg);
}
.pc-face {
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
  background: var(--card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 18px;
}
.pc-back {
  position: absolute;
  inset: 0;
  transform: rotateY(180deg);
  background: linear-gradient(160deg, #fff9f4, #fff1ea);
}
.pc-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.pc-user {
  display: flex;
  align-items: center;
  gap: 9px;
}
.avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ffb08a, #ff6b35);
  color: #fff;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}
.uname {
  font-size: 14px;
  font-weight: 700;
}
.utime {
  font-size: 12px;
  color: var(--ink-3);
}
.pc-shop {
  display: flex;
  gap: 8px;
  margin-top: 14px;
  align-items: flex-start;
}
.pin {
  font-size: 18px;
}
.shop-name {
  font-size: 17px;
  font-weight: 800;
}
.shop-addr {
  font-size: 12px;
  color: var(--ink-3);
}
.pc-dish {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--brand-soft);
  color: var(--brand-dark);
  font-weight: 700;
  font-size: 14px;
  padding: 5px 12px;
  border-radius: 999px;
  margin-top: 12px;
}
.pc-reason {
  color: var(--ink-2);
  font-size: 14px;
  margin: 12px 0 0;
  white-space: pre-wrap;
}
.pc-foot {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  flex-wrap: wrap;
}
.resp-count {
  font-size: 13px;
  color: var(--ink-3);
}
.flip-btn {
  margin-left: auto;
}
.pc-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--line);
  flex-wrap: wrap;
}
.expand-bar {
  display: block;
  width: 100%;
  margin-top: 10px;
  background: none;
  color: var(--ink-3);
  font-size: 12px;
  padding: 4px;
}
.expand-bar:hover {
  color: var(--brand);
}
.back-stamp {
  display: inline-block;
  font-size: 12px;
  font-weight: 700;
  color: var(--brand-dark);
  background: #fff;
  border: 1.5px dashed var(--brand);
  padding: 3px 10px;
  border-radius: 8px;
}
.back-photos {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
  margin-top: 12px;
}
.back-photos img {
  width: 100%;
  height: 130px;
  object-fit: cover;
  border-radius: 12px;
  cursor: zoom-in;
  background: #f5ece5;
}
.back-rating {
  color: var(--gold);
  font-size: 18px;
  letter-spacing: 2px;
  margin-top: 12px;
}
.back-review {
  color: var(--ink);
  font-size: 14px;
  margin: 8px 0 0;
  white-space: pre-wrap;
}
.settled-tip {
  font-size: 12px;
  color: var(--green);
  font-weight: 600;
}
.pc-back .pc-foot {
  position: absolute;
  bottom: 16px;
  left: 18px;
  right: 18px;
}
.pc-back .btn {
  margin-left: auto;
}
</style>
