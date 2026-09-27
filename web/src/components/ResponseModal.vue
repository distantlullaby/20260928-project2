<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-card">
      <button class="modal-close" @click="$emit('close')">✕</button>
      <h3 class="modal-title">替 TA 去吃 🥢</h3>
      <p class="modal-sub">
        为 <b>{{ wish.restaurant }}</b> 的「{{ wish.dish }}」拍现场、写点评，被采纳即可获得
        <b class="coin-inline">🪙 {{ wish.bounty }}</b>
      </p>

      <label class="label">现场实拍 *（至少 1 张，可多选）</label>
      <div class="uploader">
        <label v-for="(img, i) in images" :key="i" class="thumb">
          <img :src="img.preview" />
          <span v-if="img.uploading" class="mask">上传中…</span>
          <button class="del" @click.prevent="removeImg(i)">✕</button>
        </label>
        <label v-if="images.length < 6" class="thumb add">
          <input type="file" accept="image/*" multiple hidden @change="onPick" />
          <span>＋<br />上传照片</span>
        </label>
      </div>

      <label class="label">口味打分</label>
      <div class="stars">
        <span
          v-for="n in 5"
          :key="n"
          :class="{ on: n <= form.rating }"
          @click="form.rating = n"
        >★</span>
      </div>

      <label class="label">口味点评 *</label>
      <textarea v-model="form.review" class="textarea" placeholder="味道、份量、价格、排队、值不值得亲自跑一趟……" maxlength="2048"></textarea>

      <p v-if="error" class="err">{{ error }}</p>

      <button class="btn btn-primary btn-block" style="margin-top: 18px" :disabled="loading || uploading" @click="submit">
        {{ loading ? '提交中…' : uploading ? '图片上传中…' : '提交代吃回应' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { createResponse, uploadImage } from '../api'
import { toast } from '../store/toast'

const props = defineProps({ wish: { type: Object, required: true } })
const emit = defineEmits(['close', 'submitted'])

const form = reactive({ rating: 5, review: '' })
const images = ref([])
const error = ref('')
const loading = ref(false)
const uploading = ref(false)

async function onPick(e) {
  const files = Array.from(e.target.files || [])
  for (const file of files) {
    if (images.value.length >= 6) break
    const item = { preview: URL.createObjectURL(file), uploading: true, url: '' }
    images.value.push(item)
    uploading.value = true
    try {
      const { url } = await uploadImage(file)
      item.url = url
    } catch (err) {
      toast(err.message)
      images.value.pop()
    } finally {
      item.uploading = false
      uploading.value = images.value.some((i) => i.uploading)
    }
  }
  e.target.value = ''
}

function removeImg(i) {
  images.value.splice(i, 1)
}

async function submit() {
  error.value = ''
  const photos = images.value.map((i) => i.url).filter(Boolean)
  if (photos.length === 0) {
    error.value = '请至少上传一张现场实拍'
    return
  }
  if (!form.review.trim()) {
    error.value = '请填写口味点评'
    return
  }
  loading.value = true
  try {
    await createResponse(props.wish.id, { photos, review: form.review.trim(), rating: form.rating })
    toast('回应已提交，等待 TA 确认采纳 🤞')
    emit('submitted')
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
.uploader {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}
.thumb {
  position: relative;
  width: 92px;
  height: 92px;
  border-radius: 12px;
  overflow: hidden;
  display: block;
}
.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.thumb.add {
  border: 2px dashed var(--line);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--ink-3);
  font-size: 13px;
  text-align: center;
  cursor: pointer;
  background: #faf6f2;
}
.thumb.add:hover {
  border-color: var(--brand);
  color: var(--brand);
}
.mask {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 11px;
  line-height: 1;
}
.stars {
  font-size: 28px;
  color: #e7ded8;
  letter-spacing: 4px;
  user-select: none;
}
.stars span {
  cursor: pointer;
}
.stars span.on {
  color: var(--gold);
}
</style>
