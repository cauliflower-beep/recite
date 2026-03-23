<script setup lang="ts">
/**
 * LotterySlot.vue - 进化版漫画分镜框
 * 包含：多层阴影、动态网点、拟声词特效、对角线裁剪
 */
defineProps<{
  side: 'left' | 'right';
  label: string;
  subtext?: string;
  text: string;
  isRolling: boolean;
}>();
</script>

<template>
  <div class="slot-container" :class="[side, { 'is-rolling': isRolling }]">
    
    <!-- 1. 背景装饰层：三层影分身 -->
    <div class="shadow-layer layer-3"></div>
    <div class="shadow-layer layer-2"></div>
    <div class="shadow-layer layer-1"></div>
    
    <!-- 2. 主框体 -->
    <div class="manga-bg-frame">
      <!-- 框内动态网点背景 -->
      <div class="halftone-overlay"></div>
      
      <!-- 抽取时的拟声词浮现 -->
      <div v-if="isRolling" class="sfx-text sfx-1">ドドド</div>
      <div v-if="isRolling" class="sfx-text sfx-2">ゴゴゴ</div>
    </div>

    <!-- 3. 标签与文字内容 -->
    <div class="slot-content">
      <div class="slot-label">{{ label }}</div>
      <div class="text-wrapper">
        <div class="main-text">{{ text }}</div>
        <div v-if="subtext" class="book-tag">—— {{ subtext }}</div>
      </div>
    </div>

    <!-- 4. 装饰：边缘的破损线条 -->
    <div class="manga-scratches"></div>
  </div>
</template>

<style scoped>
.slot-container {
  position: relative;
  width: 800px;
  height: 280px;
  transition: transform 0.2s;
}

/* 依旧保持对角线布局 */
.left { align-self: center; transform: translate(-20px, 0) rotate(-1deg); }
.right { align-self: center; transform: translate(20px, 0) rotate(1deg); }

/* --- 1. 三层影分身设计 --- */
.shadow-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  background: var(--manga-black);
}
.layer-1 { transform: translate(12px, 12px); opacity: 1; }
.layer-2 { transform: translate(24px, 24px); opacity: 0.4; }
.layer-3 { transform: translate(36px, 36px); opacity: 0.1; }

/* 统一裁剪形状 (取消不规则边框) */
/* .left .shadow-layer, .left .manga-bg-frame { clip-path: polygon(0% 15%, 100% 0%, 92% 85%, 8% 100%); }
.right .shadow-layer, .right .manga-bg-frame { clip-path: polygon(8% 0%, 92% 15%, 100% 85%, 0% 100%); } */

/* --- 2. 主框体与动态背景 --- */
.manga-bg-frame {
  position: absolute;
  inset: 0;
  background: var(--manga-white);
  border: 10px solid var(--manga-black);
  z-index: 5;
  overflow: hidden;
}

.halftone-overlay {
  position: absolute;
  inset: -50%;
  background-image: radial-gradient(rgba(0,0,0,0.05) 15%, transparent 16%);
  background-size: 10px 10px;
  opacity: 0.5;
  z-index: 1;
}

/* 滚动时的狂暴效果 */
.is-rolling .manga-bg-frame {
  border-color: var(--manga-red);
  animation: frame-flash 0.1s infinite;
}

/* --- 3. 拟声词 (SFX) --- */
.sfx-text {
  position: absolute;
  font-size: 40px;
  font-weight: 900;
  color: var(--manga-red);
  z-index: 6;
  opacity: 0;
  animation: sfx-pop 0.4s infinite alternate;
}
.sfx-1 { top: 20px; left: 40px; transform: rotate(-15deg); }
.sfx-2 { bottom: 20px; right: 40px; transform: rotate(15deg); animation-delay: 0.2s; }

/* --- 4. 标签与主文字优化 --- */
.slot-content {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.text-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.main-text {
  font-size: 75px;
  font-weight: 900;
  text-align: center;
  word-break: break-all;
  padding: 0 20px;
  text-shadow: 
    -4px -4px 0 #000, 4px -4px 0 #000, 
    -4px 4px 0 #000, 4px 4px 0 #000,
    10px 10px 0px rgba(0,0,0,0.3);
}

.left .main-text {
  color: var(--manga-blue, #00E5FF);
}

.right .main-text {
  color: var(--manga-red, #FF2A55);
}

.is-rolling .main-text {
  color: #fff;
  animation: glitch 0.1s infinite;
}

/* --- 重点修改：标签配色 --- */
.slot-label {
  position: absolute;
  top: 0px;
  border: 4px solid var(--manga-black);
  padding: 5px 30px;
  font-size: 24px;
  font-weight: 900;
  box-shadow: 6px 6px 0 rgba(0,0,0,0.2);
}

/* 左侧：换成你要求的蓝色（弹药填充色） */
.left .slot-label { 
  left: -40px; 
  transform: rotate(-8deg); 
  background: var(--manga-blue); 
  color: var(--manga-black); 
}

/* 右侧：保持热血红 */
.right .slot-label { 
  right: -40px; 
  transform: rotate(8deg); 
  background: var(--manga-red); 
  color: white; 
}

.book-tag {
  align-self: flex-end;
  margin-right: 80px;
  margin-top: 10px;
  font-size: 32px;
  font-weight: bold;
  font-style: italic;
  text-shadow: 
    -2px -2px 0 #000, 2px -2px 0 #000, 
    -2px 2px 0 #000, 2px 2px 0 #000,
    6px 6px 0px rgba(0,0,0,0.3);
}

.left .book-tag {
  color: var(--manga-blue, #00E5FF);
}

.right .book-tag {
  color: var(--manga-red, #FF2A55);
}

.is-rolling .book-tag {
  color: #fff;
}

/* --- 动画库 --- */
@keyframes frame-flash {
  0%, 100% { background: #fff; }
  50% { background: #fdfdfd; }
}
@keyframes sfx-pop {
  0% { transform: scale(0.8) rotate(-15deg); opacity: 0; }
  100% { transform: scale(1.2) rotate(-10deg); opacity: 1; }
}
@keyframes glitch {
  0% { transform: translate(0); }
  33% { transform: translate(-3px, 2px); text-shadow: 2px -2px #00E5FF, -2px 2px #FF2A55; }
  66% { transform: translate(3px, -2px); text-shadow: -2px 2px #00E5FF, 2px -2px #FF2A55; }
  100% { transform: translate(0); }
}
.is-rolling { animation: total-shake 0.1s infinite; }
@keyframes total-shake {
  0% { transform: translate(-2px, 2px); }
  50% { transform: translate(2px, -2px); }
  100% { transform: translate(-2px, -2px); }
}
</style>