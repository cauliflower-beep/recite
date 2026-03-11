<template>
  <div v-if="visible" class="dialog-overlay" @click.self="handleCancel">
    <!-- 彻底删除了背景颜色和放射线，只留下一个透明的拦截层 -->
    
    <div 
      class="manga-dialog" 
      :class="{ 
        'is-danger': type === 'confirm', 
        'is-info': type === 'info' 
      }"
    >
      <div class="dialog-header">
        <span v-if="type === 'confirm'">⚠️ 极度危险 (WARNING)</span>
        <span v-else-if="type === 'info'">📡 系统同步 (SYNC)</span>
        <span v-else>🖋️ 情报录入 (INPUT)</span>
      </div>
      
      <div class="dialog-body">
        <div class="dialog-title">{{ title }}</div>
        
        <input 
          v-if="type === 'input'" 
          ref="inputRef"
          v-model="inputValue" 
          class="manga-input" 
          type="text" 
          :placeholder="placeholder"
          @keyup.enter="handleConfirm"
        />
        
        <div v-else class="warning-text">
          {{ placeholder }}
        </div>
      </div>

      <div class="dialog-footer">
        <button v-if="type !== 'info'" class="btn-cancel" @click="handleCancel">撤退</button>
        <button class="btn-confirm" @click="handleConfirm">
          {{ type === 'confirm' ? '💥 强制抹杀' : (type === 'info' ? '了解 (OK)' : '确认装填') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';

const props = defineProps<{
  visible: boolean;
  type: 'input' | 'confirm' | 'info'; 
  title: string;
  placeholder?: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void;
  (e: 'confirm', payload?: string): void;
  (e: 'cancel'): void;
}>();

const inputValue = ref('');
const inputRef = ref<HTMLInputElement | null>(null);

watch(() => props.visible, (newVal) => {
  if (newVal) {
    inputValue.value = '';
    if (props.type === 'input') {
      nextTick(() => inputRef.value?.focus());
    }
  }
});

const handleConfirm = () => {
  if (props.type === 'input' && !inputValue.value.trim()) {
    inputRef.value?.classList.add('shake');
    setTimeout(() => inputRef.value?.classList.remove('shake'), 400);
    return;
  }
  emit('confirm', props.type === 'input' ? inputValue.value.trim() : undefined);
  emit('update:visible', false);
};

const handleCancel = () => {
  emit('cancel');
  emit('update:visible', false);
};
</script>

<style scoped>
/* --- 极简遮罩层：透明，仅用于阻断点击 --- */
.dialog-overlay {
  position: fixed;
  top: 0; left: 0; width: 100vw; height: 100vh;
  background: transparent; /* 彻底透明 */
  display: flex; justify-content: center; align-items: center;
  z-index: 9999;
}

/* --- 强化弹窗本体的“存在感” --- */
.manga-dialog {
  background: var(--manga-white);
  border: 6px solid var(--manga-black);
  width: 460px;
  transform: rotate(-1deg);
  /* 增加阴影的深度，让它在没有背景遮注的情况下依然能“跳”出来 */
  box-shadow: 25px 25px 0 var(--manga-black);
  animation: popDown 0.25s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  position: relative;
}

@keyframes popDown {
  0% { transform: scale(0.8) translateY(-50px) rotate(-5deg); opacity: 0; }
  100% { transform: scale(1) translateY(0) rotate(-1deg); opacity: 1; }
}

/* 颜色模式保持 */
.manga-dialog.is-danger { border-color: var(--manga-red); box-shadow: 25px 25px 0 var(--manga-red); }
.manga-dialog.is-info { border-color: var(--manga-blue); box-shadow: 25px 25px 0 var(--manga-blue); }

.dialog-header {
  background: var(--manga-black); color: var(--manga-yellow);
  padding: 8px 15px; font-size: 20px; font-weight: bold;
}
.is-danger .dialog-header { background: var(--manga-red); color: white; }
.is-info .dialog-header { background: var(--manga-blue); color: var(--manga-black); }

.dialog-body { padding: 25px; }
.dialog-title { font-size: 24px; font-weight: bold; color: var(--manga-black); margin-bottom: 15px; }

.manga-input {
  width: 100%; box-sizing: border-box; padding: 12px;
  font-size: 20px; font-family: inherit; border: 4px solid var(--manga-black);
  outline: none; background: #eee;
}
.manga-input:focus { background: white; border-color: var(--manga-yellow); }

.warning-text { font-size: 18px; color: var(--manga-black); font-weight: bold; line-height: 1.4; }

.dialog-footer {
  display: flex; justify-content: flex-end; gap: 15px;
  padding: 15px 25px; background: #f0f0f0; border-top: 3px solid var(--manga-black);
}

.btn-confirm, .btn-cancel {
  padding: 8px 20px; border: 3px solid var(--manga-black); font-family: inherit;
  font-weight: bold; cursor: pointer; transition: transform 0.1s;
}
.btn-confirm { background: var(--manga-yellow); }
.is-danger .btn-confirm { background: var(--manga-red); color: white; }
.is-info .btn-confirm { background: var(--manga-blue); }

.btn-cancel { background: white; }
.btn-confirm:active, .btn-cancel:active { transform: translate(2px, 2px); }

.shake { animation: shake 0.4s; }
@keyframes shake { 0%, 100% { transform: translateX(0); } 25% { transform: translateX(-8px); } 75% { transform: translateX(8px); } }
</style>