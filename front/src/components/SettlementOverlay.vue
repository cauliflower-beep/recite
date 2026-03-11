<script setup lang="ts">
import { onMounted } from 'vue';
import confetti from 'canvas-confetti';

const props = defineProps<{
  type: 'PASS' | 'FAIL';
  student: string;
}>();

const emit = defineEmits(['close']);

const successQuotes = ["全场最强！", "背诵之神下凡！", "优雅，实在是优雅！", "SSS级学霸归位！"];
const failQuotes = ["脑路瞬间短路...", "感受到老师的凝视了吗？", "抄写 3 遍警告！", "这波...只能说尽力了。"];

const randomQuote = (quotes: string[]) => quotes[Math.floor(Math.random() * quotes.length)];

onMounted(() => {
  if (props.type === 'PASS') {
    // 成功时撒花！
    confetti({
      particleCount: 150,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#FFD700', '#FF2A55', '#00FF66', '#00E5FF']
    });
  }
});
</script>

<template>
  <div class="settlement-overlay" :class="type.toLowerCase()" @click="emit('close')">
    <!-- 1. 成功画面 -->
    <template v-if="type === 'PASS'">
      <div class="speed-lines"></div>
      <div class="content-box bounce-in">
        <div class="manga-title">LEVEL UP!</div>
        <div class="result-text success-glow">{{ student }}</div>
        <div class="sub-quote">“{{ randomQuote(successQuotes) }}”</div>
        <div class="stamp-good">合格</div>
      </div>
    </template>

    <!-- 2. 失败画面 -->
    <template v-else>
      <div class="rain-lines"></div>
      <div class="content-box shake-in">
        <div class="manga-title gray">WASTED.</div>
        <div class="result-text fail-dim">{{ student }}</div>
        <div class="sub-quote">“{{ randomQuote(failQuotes) }}”</div>
        <div class="stamp-bad">不合格</div>
      </div>
    </template>

    <div class="click-hint">- 点击屏幕回归现实 -</div>
  </div>
</template>

<style scoped>
.settlement-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  overflow: hidden;
}

/* --- 成功配色：金红 --- */
.pass {
  background: rgba(255, 215, 0, 0.95);
  border: 20px solid #000;
}

/* --- 失败配色：黑白灰 --- */
.fail {
  background: rgba(30, 30, 30, 0.98);
  border: 20px solid #FF2A55;
}

/* 漫画装饰线 */
.speed-lines {
  position: absolute;
  inset: 0;
  background-image: repeating-conic-gradient(from 0deg, transparent 0deg 2deg, rgba(255,255,255,0.3) 2deg 4deg);
  opacity: 0.5;
  animation: rotateLines 10s linear infinite;
}

.rain-lines {
  position: absolute;
  inset: 0;
  background-image: repeating-linear-gradient(135deg, transparent 0 20px, rgba(255,255,255,0.05) 20px 22px);
  animation: rain 0.5s linear infinite;
}

/* 文字内容 */
.content-box {
  position: relative;
  z-index: 10;
  text-align: center;
}

.manga-title {
  font-size: 120px;
  font-weight: 900;
  font-style: italic;
  color: #000;
  text-shadow: 8px 8px 0 #fff;
  margin-bottom: -20px;
}
.manga-title.gray { color: #fff; text-shadow: 8px 8px 0 var(--manga-red); }

.result-text {
  font-size: 180px;
  font-weight: 900;
  letter-spacing: -5px;
}

.success-glow {
  color: #fff;
  text-shadow: 
    -6px -6px 0 #000, 6px -6px 0 #000, -6px 6px 0 #000, 6px 6px 0 #000,
    15px 15px 0 var(--manga-red);
}

.fail-dim {
  color: #444;
  filter: grayscale(1);
  text-shadow: 0 0 30px rgba(255,255,255,0.2);
}

.sub-quote {
  font-size: 30px;
  background: #000;
  color: #fff;
  padding: 10px 30px;
  transform: rotate(-2deg);
  display: inline-block;
  margin-top: 20px;
}

/* 印章效果 */
.stamp-good, .stamp-bad {
  position: absolute;
  right: -100px;
  bottom: -50px;
  font-size: 80px;
  border: 10px solid;
  padding: 10px 30px;
  border-radius: 20px;
  transform: rotate(-30deg);
  font-weight: 900;
  opacity: 0;
  animation: stampIt 0.4s forwards 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}
.stamp-good { color: var(--manga-red); border-color: var(--manga-red); }
.stamp-bad { color: #fff; border-color: #fff; }

.click-hint {
  position: absolute;
  bottom: 50px;
  color: #fff;
  font-size: 20px;
  animation: blink 1s infinite;
}

/* 动画库 */
@keyframes rotateLines { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@keyframes rain { from { transform: translateY(-100px); } to { transform: translateY(100px); } }
@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
@keyframes stampIt { 
  0% { transform: scale(5) rotate(-30deg); opacity: 0; } 
  100% { transform: scale(1) rotate(-30deg); opacity: 1; } 
}
.bounce-in { animation: bounceIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
.shake-in { animation: shakeIn 0.5s; }

@keyframes bounceIn { 0% { transform: scale(0); } 100% { transform: scale(1); } }
@keyframes shakeIn { 
  0%, 100% { transform: translateX(0); } 
  25% { transform: translateX(-20px); } 
  75% { transform: translateX(20px); } 
}
</style>