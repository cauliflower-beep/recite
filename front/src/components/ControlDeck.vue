<script setup lang="ts">
/**
 * ControlDeck.vue - 命运控制台
 * 负责：抽奖触发、结算判定
 */
defineProps<{
  gameState: 'IDLE' | 'ROLLING' | 'JUDGING';
}>();

defineEmits<{
  (e: 'roll'): void;
  (e: 'judge', result: 'PASS' | 'FAIL'): void;
}>();
</script>

<template>
  <div class="control-deck">
    <!-- 1. 抽取阶段：显示命运抽取按钮 -->
    <button 
      v-if="gameState !== 'JUDGING'" 
      class="manga-btn btn-draw" 
      @click="$emit('roll')" 
      :disabled="gameState === 'ROLLING'"
    >
      {{ gameState === 'ROLLING' ? '⚡ 锁定中...' : '💥 命运抽取 (ROLL)' }}
    </button>

    <!-- 2. 结算阶段：显示判定按钮组 -->
    <div v-else class="judgment-zone">
      <button class="manga-btn btn-pass" @click="$emit('judge', 'PASS')">
        ✅ 完美吟唱 (PASS)
      </button>
      <button class="manga-btn btn-fail" @click="$emit('judge', 'FAIL')">
        💀 处刑 (FAIL)
      </button>
    </div>
  </div>
</template>

<style scoped>
.control-deck {
  /* 容器保持透明，仅负责布局 */
  height: 180px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 100;
}

.judgment-zone {
  display: flex;
  gap: 40px;
  /* 结算区域弹出的入场动画 */
  animation: popUp 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

/* --- 核心：统一的漫画风按钮基础样式 --- */
.manga-btn {
  /* 基础形状：斜切、粗黑边 */
  font-family: inherit;
  font-size: 45px;
  font-weight: 900;
  padding: 15px 60px;
  border: 6px solid #000;
  cursor: pointer;
  transform: skewX(-10deg);
  
  /* 默认状态：黄色投影 */
  box-shadow: 10px 10px 0 var(--manga-yellow);
  transition: all 0.1s;
  position: relative;
  outline: none;
  user-select: none;
}

/* 1. 抽取按钮：红色背景，白色文字 */
.btn-draw {
  background: var(--manga-red);
  color: var(--manga-white);
}

/* 2. 完美吟唱：鲜绿色背景，黑色文字 */
.btn-pass {
  background: #00FF66;
  color: #000;
}

/* 3. 处刑：深灰色背景，白色文字 */
.btn-fail {
  background: #666;
  color: #fff;
}

/* --- 悬浮交互：按钮扩大，投影变黑 --- */
.manga-btn:hover:not(:disabled) {
  transform: skewX(-10deg) scale(1.05);
  box-shadow: 15px 15px 0 #000; /* 投影变黑变厚 */
}

/* --- 点击交互：向右下按压感 --- */
.manga-btn:active:not(:disabled) {
  transform: skewX(-10deg) translate(5px, 5px);
  box-shadow: 5px 5px 0 #000;
}

/* 禁用状态（抽取中） */
.manga-btn:disabled {
  background: #555;
  color: #888;
  border-color: #333;
  box-shadow: 8px 8px 0 #222;
  cursor: not-allowed;
  transform: skewX(-10deg);
}

@keyframes popUp {
  0% { transform: translateY(50px) scale(0.5); opacity: 0; }
  100% { transform: translateY(0) scale(1); opacity: 1; }
}
</style>