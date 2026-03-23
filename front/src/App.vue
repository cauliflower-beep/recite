<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive } from 'vue'; // 引入 onUnmounted 销毁监听
import LotterySlot from './components/LotterySlot.vue';
import ControlDeck from './components/ControlDeck.vue';
import ConfigPanel from './components/ConfigPanel.vue';
import SettlementOverlay from './components/SettlementOverlay.vue';

// --- 方案 B：自适应缩放核心逻辑 ---
const scale = ref(1); // 缩放比例
const baseWidth = 1920; // 设计稿基准宽
const baseHeight = 1080; // 设计稿基准高

const updateScale = () => {
  // 计算当前窗口与基准尺寸的比例，取较小值以确保内容完全显示在屏幕内（Contain模式）
  const w = window.innerWidth / baseWidth;
  const h = window.innerHeight / baseHeight;
  scale.value = Math.min(w, h);
};

// --- 类型与状态 ---
type GameState = 'IDLE' | 'ROLLING' | 'JUDGING';
const showConfig = ref(false);
const gameState = ref<GameState>('IDLE');

const settlement = reactive({
  show: false,
  type: 'PASS' as 'PASS' | 'FAIL',
  student: ''
});

const resetGame = () => {
  settlement.show = false;
  gameState.value = 'IDLE';
  currentStudent.value = '???';
  currentAuthor.value = '未知';
  currentPoem.value = '???';
};

const students = ref<string[]>([]);
const poems = ref<{ title: string; author: string }[]>([]);
const currentStudent = ref('???');
const currentAuthor = ref('未知');
const currentPoem = ref('???');

const loadData = async () => {
  try {
    const res = await fetch('/api/config');
    const json = await res.json();
    if (json.code === 200) {
      students.value = json.data.students || [];
      poems.value = json.data.poems || [];
    }
  } catch (e) {
    console.error("数据加载失败");
  }
};

let rollInterval: any = null;
const startRoll = () => {
  if (!students.value.length || !poems.value.length) return alert("配置库空虚！请先录入情报！");

  gameState.value = 'ROLLING';
  rollInterval = setInterval(() => {
    const randomStudent = students.value[Math.floor(Math.random() * students.value.length)];
    if (randomStudent) currentStudent.value = randomStudent;
    
    const m = poems.value[Math.floor(Math.random() * poems.value.length)];
    if (m) {
      currentAuthor.value = m.author;
      currentPoem.value = m.title;
    }
  }, 50);

  setTimeout(() => {
    clearInterval(rollInterval);
    gameState.value = 'JUDGING';
  }, 2500);
};

const handleJudge = (result: 'PASS' | 'FAIL') => {
  settlement.type = result;
  settlement.student = currentStudent.value;
  settlement.show = true;
};

// 生命周期钩子
onMounted(() => {
  loadData();
  updateScale(); // 挂载时计算一次缩放
  window.addEventListener('resize', updateScale); // 监听窗口大小变化
});

onUnmounted(() => {
  window.removeEventListener('resize', updateScale); // 销毁时移除监听
});
</script>

<template>
  <div class="app-container">
    <!-- 这里的 .main-scaler 是核心，所有的内容都放在它里面等比缩放 -->
    <div 
      class="main-scaler" 
      :style="{ 
        transform: `scale(${scale})`, 
        width: `${baseWidth}px`, 
        height: `${baseHeight}px` 
      }"
    >
      <!-- 背景装饰：巨型黑色斜切带 (现在它也会跟着缩放了) -->
      <div class="bg-slash"></div>

      <div class="top-bar">
        <button class="btn-config" @click="showConfig = true" :disabled="gameState === 'ROLLING'">
          ⚙️ 战术配置 (CONFIG)
        </button>
      </div>

      <div class="battle-arena">
        <LotterySlot 
          side="left" 
          label="🎯 TARGET / 锁定目标" 
          :text="currentStudent" 
          :isRolling="gameState === 'ROLLING'" 
        />

        <div class="vs-container">
          <div class="vs-text">VS</div>
          <div class="vs-slash"></div>
        </div>

        <LotterySlot 
          side="right" 
          label="📜 MISSION / 绝密任务" 
          :subtext="currentAuthor"
          :text="currentPoem" 
          :isRolling="gameState === 'ROLLING'" 
        />
      </div>

      <ControlDeck :game-state="gameState" @roll="startRoll" @judge="handleJudge" />

      <!-- 弹窗部分 -->
      <div v-if="showConfig" class="modal-overlay">
        <ConfigPanel @close="showConfig = false" @saved="loadData" />
      </div>

      <!-- 结算组件 -->
      <Transition name="fade">
        <SettlementOverlay 
          v-if="settlement.show" 
          :type="settlement.type" 
          :student="settlement.student"
          @close="resetGame"
        />
      </Transition>
    </div>
  </div>
</template>

<style scoped>
/* 外部容器：永远占满全屏，负责背景底色 */
.app-container {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #e0e0e0;
  background-image: radial-gradient(rgba(0,0,0,0.15) 15%, transparent 16%), 
                    radial-gradient(rgba(0,0,0,0.15) 15%, transparent 16%);
  background-size: 15px 15px;
  background-position: 0 0, 7.5px 7.5px;
  
  /* 居中缩放后的内容 */
  display: flex;
  justify-content: center;
  align-items: center;
}

/* 内部缩放器：它的宽高是固定的 1920x1080，但通过 transform: scale 适配屏幕 */
.main-scaler {
  position: relative;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  /* 必须设置缩放中心为中心，或者左上角 */
  transform-origin: center center;
  /* 保证内部元素能够正常绝对定位 */
  background: transparent;
}

/* 所有的子元素现在都是基于 1920x1080 的 px 布局，不用动！ */

.control-deck {
  flex-shrink: 0;
  height: 180px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  z-index: 10;
  padding-bottom: 20px;
}

.modal-overlay {
  position: absolute; /* 在 scaler 内部使用 absolute 即可 */
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.05); /* 极淡的底色增强层次感 */
  backdrop-filter: blur(10px); /* 保持模糊效果 */
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
}

.bg-slash {
  position: absolute;
  width: 200%; /* 稍微加长点 */
  height: 250px;
  background: #1a1a1a;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-12deg);
  z-index: 1;
  border-top: 8px solid #FF2A55;
  border-bottom: 8px solid #FF2A55;
}

.battle-arena {
  flex: 1;
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 0 50px;
  z-index: 5;
}

.vs-container {
  position: relative;
  width: 200px;
  height: 200px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.vs-text {
  font-size: 140px;
  font-weight: 900;
  font-style: italic;
  color: var(--manga-yellow);
  z-index: 2;
  text-shadow: 10px 10px 0px #000;
  animation: vs-kick 0.6s infinite alternate cubic-bezier(0.68, -0.55, 0.27, 1.55);
}

.vs-slash {
  position: absolute;
  width: 300px;
  height: 80px;
  background: var(--manga-red);
  transform: rotate(-45deg);
  z-index: 1;
  clip-path: polygon(10% 0%, 100% 0%, 90% 100%, 0% 100%);
}

@keyframes vs-kick {
  from { transform: scale(1) rotate(-10deg); }
  to { transform: scale(1.2) rotate(10deg); }
}

.top-bar { 
  height: 120px; /* 稍微给 top-bar 多点空间 */
  display: flex; 
  justify-content: flex-end; 
  align-items: center; 
  padding-right: 60px; 
  z-index: 10; 
}
.btn-config { 
  background: #fff; 
  border: 4px solid #000; 
  padding: 10px 30px; 
  font-size: 24px;
  font-family: inherit; 
  font-weight: bold; 
  cursor: pointer; 
  box-shadow: 8px 8px 0 #000; 
  transform: rotate(2deg); 
  transition: 0.2s; 
}
.btn-config:hover { background: var(--manga-yellow); transform: rotate(-2deg) scale(1.1); }

/* 结算动画 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.5s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>