<template>
  <div class="card-wrap">
    <div :class="['flip-scene', { flipped: flipped }]">
      <div class="flip-inner">
        <!-- ============ 正面：心愿菜单 + 种草理由 ============ -->
        <section class="face face-front">
          <div class="front-head">
            <div class="avatar">{{ wish.user.avatar_emoji }}</div>
            <div class="who">
              <strong>{{ wish.user.nickname }}</strong>
              <span class="time">{{ timeText }}</span>
            </div>
            <span :class="['tag', 'tag-' + wish.status]">{{ statusText }}</span>
          </div>

          <h2 class="restaurant">
            🏮 {{ wish.restaurant_name }}
            <small v-if="wish.address">📍 {{ wish.address }}</small>
          </h2>

          <div class="menu-box">
            <div class="box-title"><span>📝</span> 心愿菜单</div>
            <p class="menu-list">{{ wish.dish_list }}</p>
          </div>

          <div class="reason-box">
            <div class="box-title"><span>🌱</span> 种草理由</div>
            <p class="reason-text">{{ wish.reason }}</p>
          </div>

          <div class="front-foot">
            <div class="reward">
              <span class="muted small">悬赏</span>
              <span class="coin reward-num">{{ wish.reward_coins }}</span>
            </div>
            <div class="front-actions">
              <span v-if="wish.response_count" class="resp-hint" @click="flip">
                👅 {{ wish.response_count }} 人替吃 · 翻看现场
              </span>
              <button class="btn btn-ghost btn-sm flip-btn" @click="flip">
                {{ wish.response_count ? '翻到现场' : '查看现场' }} 🔄
              </button>
            </div>
          </div>
        </section>

        <!-- ============ 背面：现场实拍 + 口味点评 ============ -->
        <section class="face face-back">
          <div class="back-head">
            <button class="btn btn-plain btn-sm" @click="flip">‹ 翻回心愿</button>
            <span class="back-title">📷 别人替你吃的现场</span>
            <span class="resp-count">{{ wish.response_count || 0 }} 份回应</span>
          </div>

          <div v-if="!responses.length" class="no-response">
            <span class="emoji">🥢</span>
            <p>还没有人替 TA 去吃</p>
            <small class="muted">成为第一个替 TA 尝鲜的人，悬赏币等你拿</small>
          </div>

          <div v-else class="resp-list">
            <article v-for="r in responses" :key="r.id" :class="['resp-item', { adopted: r.is_adopted }]">              <div class="resp-user">
                <span class="avatar sm">{{ r.user.avatar_emoji }}</span>
                <strong>{{ r.user.nickname }}</strong>
                <span class="stars">{{ '⭐'.repeat(r.rating) }}</span>
                <span v-if="r.is_adopted" class="tag tag-settled adopt-badge">✓ 已采纳</span>
              </div>
              <div class="photos">
                <img v-for="(p, i) in r.photos" :key="i" :src="p" :alt="'实拍' + i" loading="lazy" />
              </div>
              <p class="comment">“{{ r.comment }}”</p>
              <div class="resp-foot">
                <span class="muted small">{{ relTime(r.created_at) }}</span>
                <button
                  v-if="canSettle && !r.is_adopted"
                  class="btn btn-primary btn-sm"
                  @click="$emit('settle', r.id)"
                >
                  采纳并结算 {{ wish.reward_coins }} 🪙
                </button>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>

    <!-- 卡片外操作区（不参与翻转） -->
    <div class="card-actions">
      <button v-if="canTaste" class="btn btn-primary taste-btn" @click="$emit('taste')">
        👅 替 TA 去吃
      </button>
      <button v-if="isOwner && wish.status === 'open'" class="btn btn-ghost btn-sm" @click="$emit('append')">
        ➕ 追加悬赏
      </button>
      <button v-if="isOwner && wish.status === 'open'" class="btn btn-plain btn-sm" @click="$emit('cancel')">
        取消并退款
      </button>
      <span v-if="!isOwner && wish.status === 'open' && !auth.isLoggedIn" class="muted small login-tip">
        <RouterLink to="/login">登录</RouterLink> 后即可替 TA 去吃
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { auth } from '../store'

const props = defineProps({
  wish: { type: Object, required: true },
  responses: { type: Array, default: () => [] },
  // 当前用户是否为心愿发起人
  isOwner: { type: Boolean, default: false }
})
const emit = defineEmits(['settle', 'taste', 'append', 'cancel', 'flip'])

const flipped = ref(false)

const statusMap = { open: '求代吃中', settled: '已结算', canceled: '已取消' }
const statusText = computed(() => statusMap[props.wish.status] || props.wish.status)

const timeText = computed(() => relTime(props.wish.created_at))

const canSettle = computed(
  () => props.isOwner && props.wish.status === 'open' && props.responses.length > 0
)
const canTaste = computed(
  () => auth.isLoggedIn && !props.isOwner && props.wish.status === 'open'
)

function flip() {
  flipped.value = !flipped.value
  if (flipped.value) emit('flip')
}

function relTime(t) {
  const d = new Date(t)
  const diff = (Date.now() - d.getTime()) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  if (diff < 86400 * 7) return Math.floor(diff / 86400) + ' 天前'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

// 心愿切换时回到正面
watch(() => props.wish.id, () => (flipped.value = false))
</script>

<style scoped>
.card-wrap {
  position: relative;
  background: var(--card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  margin-bottom: 20px;
  overflow: hidden;
  border: 1px solid rgba(240, 228, 216, 0.8);
}

/* 3D 翻转 */
.flip-scene { perspective: 1600px; }
.flip-inner {
  position: relative;
  width: 100%;
  transform-style: preserve-3d;
  transition: transform 0.65s cubic-bezier(0.4, 0.2, 0.2, 1);
}
.flipped .flip-inner { transform: rotateY(180deg); }
.face {
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
  padding: 20px;
  width: 100%;
}
.face-back {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  transform: rotateY(180deg);
  background: linear-gradient(180deg, #fff9f4, #fff);
  overflow-y: auto;
}

/* 正面 */
.front-head { display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
.avatar {
  width: 42px; height: 42px; border-radius: 50%;
  background: var(--brand-soft);
  display: flex; align-items: center; justify-content: center;
  font-size: 23px; flex-shrink: 0;
}
.avatar.sm { width: 30px; height: 30px; font-size: 16px; }
.who { display: flex; flex-direction: column; line-height: 1.3; flex: 1; }
.who small, .time { font-size: 12px; color: var(--ink-soft); }
.time { font-size: 12px; color: var(--ink-soft); }

.restaurant {
  margin: 0 0 14px;
  font-size: 20px;
  line-height: 1.35;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.restaurant small { font-size: 12.5px; font-weight: 500; color: var(--ink-soft); }

.menu-box {
  background: var(--brand-soft);
  border-radius: 14px;
  padding: 12px 14px;
  margin-bottom: 12px;
  border-left: 4px solid var(--brand);
}
.box-title { font-size: 13px; font-weight: 700; color: var(--brand-dark); margin-bottom: 5px; }
.menu-list { margin: 0; font-size: 15px; font-weight: 600; letter-spacing: 0.3px; }
.reason-box {
  background: #faf5ee;
  border-radius: 14px;
  padding: 12px 14px;
  border-left: 4px solid var(--accent);
}
.reason-text { margin: 0; font-size: 14px; line-height: 1.7; color: #5b4a3c; }

.front-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  gap: 10px;
}
.reward { display: flex; align-items: baseline; gap: 6px; }
.reward-num { font-size: 22px; }
.front-actions { display: flex; align-items: center; gap: 10px; }
.resp-hint { font-size: 12.5px; color: var(--brand-dark); cursor: pointer; font-weight: 600; }

/* 背面 */
.back-head { display: flex; align-items: center; gap: 8px; margin-bottom: 14px; }
.back-title { font-weight: 700; font-size: 14px; flex: 1; }
.resp-count { font-size: 12px; color: var(--ink-soft); }
.no-response { text-align: center; padding: 48px 20px; }
.no-response .emoji { font-size: 44px; display: block; margin-bottom: 10px; }
.no-response p { margin: 0 0 6px; font-weight: 700; }

.resp-list { display: flex; flex-direction: column; gap: 16px; }
.resp-item {
  border: 1.5px solid var(--line);
  border-radius: 16px;
  padding: 14px;
  background: #fff;
}
.resp-item.adopted { border-color: var(--ok); background: var(--ok-soft); }
.resp-user { display: flex; align-items: center; gap: 7px; margin-bottom: 10px; flex-wrap: wrap; }
.stars { font-size: 12px; letter-spacing: -1px; }
.adopt-badge { margin-left: auto; }
.photos { display: grid; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); gap: 8px; margin-bottom: 10px; }
.photos img {
  width: 100%; height: 120px;
  object-fit: cover;
  border-radius: 10px;
  background: #f3e9df;
  cursor: zoom-in;
}
.comment { margin: 0 0 10px; font-size: 14px; line-height: 1.7; color: #4a3c30; }
.resp-foot { display: flex; align-items: center; justify-content: space-between; gap: 8px; }

.card-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px dashed var(--line);
  background: #fffdfa;
  flex-wrap: wrap;
}
.taste-btn { flex: 1; min-width: 140px; }
.login-tip a { color: var(--brand); font-weight: 700; }
</style>
