<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3>👅 替 TA 去吃：{{ wish.restaurant_name }}</h3>

      <form @submit.prevent="submit">
        <div class="field">
          <label>现场实拍（1~6 张）</label>
          <div class="upload-grid">
            <div v-for="(p, i) in form.photos" :key="i" class="up-thumb">
              <img :src="p" alt="实拍" />
              <button type="button" class="up-del" @click="remove(i)">✕</button>
            </div>
            <label v-if="form.photos.length < 6" class="up-add">
              <input type="file" accept="image/*" hidden @change="onPick" />
              <span v-if="!uploading">＋<br /><small>添加实拍</small></span>
              <span v-else class="muted small">上传中…</span>
            </label>
          </div>
          <div class="hint">拍下食物与现场，发起人会根据你的实拍决定是否采纳</div>
        </div>

        <div class="field">
          <label>口味点评</label>
          <textarea
            v-model="form.comment"
            maxlength="1024"
            placeholder="味道、分量、辣度/甜度、性价比、推荐指数……帮 TA 决定值不值得亲自跑一趟（至少 5 个字）"
            required
          ></textarea>
        </div>

        <div class="field">
          <label>综合星级</label>
          <div class="star-picker">
            <button
              v-for="n in 5"
              :key="n"
              type="button"
              :class="{ on: n <= form.rating }"
              @click="form.rating = n"
            >⭐</button>
          </div>
        </div>

        <button class="btn btn-primary btn-block" :disabled="submitting || uploading || !form.photos.length">
          {{ submitting ? '提交中…' : '提交代吃回应' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import http, { errMsg } from '../api'
import { toast } from '../toast'

const props = defineProps({ wish: { type: Object, required: true } })
const emit = defineEmits(['close', 'created'])
const submitting = ref(false)
const uploading = ref(false)
const form = reactive({ photos: [], comment: '', rating: 5 })

async function onPick(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  if (file.size > 5 * 1024 * 1024) {
    toast.show('图片不能超过 5MB', 'err')
    return
  }
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    const { data } = await http.post('/api/upload', fd, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    form.photos.push(data.url)
  } catch (e2) {
    toast.show(errMsg(e2, '图片上传失败'), 'err')
  } finally {
    uploading.value = false
  }
}

function remove(i) {
  form.photos.splice(i, 1)
}

async function submit() {
  if (form.comment.trim().length < 5) {
    toast.show('口味点评至少写 5 个字哦', 'err')
    return
  }
  submitting.value = true
  try {
    const { data } = await http.post(`/api/wishes/${props.wish.id}/responses`, form)
    toast.show('已替 TA 品尝，等待发起人采纳 🤞', 'ok')
    emit('created', data.response)
  } catch (e) {
    toast.show(errMsg(e), 'err')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.upload-grid { display: flex; flex-wrap: wrap; gap: 10px; }
.up-thumb { position: relative; width: 92px; height: 92px; }
.up-thumb img { width: 100%; height: 100%; object-fit: cover; border-radius: 12px; }
.up-del {
  position: absolute; top: -6px; right: -6px;
  width: 22px; height: 22px; border-radius: 50%;
  border: none; background: rgba(43,33,24,0.8); color: #fff;
  font-size: 11px; cursor: pointer; line-height: 1;
}
.up-add {
  width: 92px; height: 92px;
  border: 2px dashed var(--line);
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  text-align: center;
  cursor: pointer;
  color: var(--ink-soft);
  font-size: 22px;
  transition: border-color 0.15s, color 0.15s;
}
.up-add:hover { border-color: var(--brand); color: var(--brand); }
.up-add small { font-size: 11px; }

.star-picker button {
  background: none; border: none; font-size: 30px; cursor: pointer;
  filter: grayscale(1) opacity(0.45);
  padding: 2px;
  transition: transform 0.1s, filter 0.1s;
}
.star-picker button.on { filter: none; }
.star-picker button:hover { transform: scale(1.15); }
</style>
