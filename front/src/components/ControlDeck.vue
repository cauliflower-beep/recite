<script setup lang="ts">
/**
 * ControlDeck.vue - 命运控制台
 * 负责：抽奖触发、结算判定
 */
defineProps<{
  gameState: 'IDLE' | 'ROLLING' | 'BATTLE' | 'SUMMARY';
}>();

defineEmits<{
  (e: 'roll'): void;
  (e: 'judge', result: 'PASS' | 'FAIL'): void;
  (e: 'reset'): void;
}>();
</script>

<template>
  <div class="control-deck">
    <!-- 1. 初始阶段：显示命运抽取按钮 -->
    <button 
      v-if="gameState === 'IDLE' || gameState === 'ROLLING'" 
      class="manga-btn btn-draw" 
      @click="$emit('roll')" 
      :disabled="gameState === 'ROLLING'"
    >
      {{ gameState === 'ROLLING' ? '⚡ 锁定中...' : '💥 命运抽取 (ROLL)' }}
    </button>

    <!-- 2. 背诵阶段：显示判定按钮组 -->
    <div v-else-if="gameState === 'BATTLE'" class="judgment-zone">
      <button class="manga-btn btn-pass" @click="$emit('judge', 'PASS')">
        ✅ 完美吟唱 (PASS)
      </button>
      <button class="manga-btn btn-fail" @click="$emit('judge', 'FAIL')">
        💀 处刑 (FAIL)
      </button>
    </div>

    <!-- 3. 总结阶段：可以显示一个快速返回按钮（可选） -->
    <button 
      v-else-if="gameState === 'SUMMARY'" 
      class="manga-btn btn-reset-mini" 
      @click="$emit('reset')"
    >
      🔄 重新整备
    </button>
  </div>
</template>

<style scoped>
.control-deck {
  display: flex;
  justify-content: center;
  gap: 30px;
  z-index: 50;
  position: relative;
}

.judgment-zone {
  display: flex;
  gap: 40px;
}

.btn-action {
  display: flex;
  gap: 40px;
  animation: popUp 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.manga-btn {
  font-family: inherit;
  font-size: 45px;
  font-weight: 900;
  padding: 15px 60px;
  border: 6px solid #000;
  cursor: pointer;
  transform: skewX(-10deg);
  box-shadow: 10px 10px 0 var(--manga-yellow);
  transition: all 0.1s;
  position: relative;
  outline: none;
  user-select: none;
}

.btn-draw { background: var(--manga-red); color: var(--manga-white); }
.btn-pass { background: #00FF66; color: #000; }
.btn-fail { background: #666; color: #fff; }
.btn-reset-mini { background: #fff; color: #000; font-size: 30px; }

.manga-btn:hover:not(:disabled) {
  transform: skewX(-10deg) scale(1.05);
  box-shadow: 15px 15px 0 #000;
}

.manga-btn:active:not(:disabled) {
  transform: skewX(-10deg) translate(5px, 5px);
  box-shadow: 5px 5px 0 #000;
}

.manga-btn:disabled {
  background: #555; color: #888; border-color: #333;
  box-shadow: 8px 8px 0 #222; cursor: not-allowed;
  transform: skewX(-10deg);
}

@keyframes popUp {
  0% { transform: translateY(50px) scale(0.5); opacity: 0; }
  100% { transform: translateY(0) scale(1); opacity: 1; }
}
</style>