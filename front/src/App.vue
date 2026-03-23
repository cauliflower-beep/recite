<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, computed } from 'vue';
import LotterySlot from './components/LotterySlot.vue';
import ControlDeck from './components/ControlDeck.vue';
import ConfigPanel from './components/ConfigPanel.vue';

// --- 类型定义 ---
interface Mission {
  student: string;
  poem: string;
  author: string;
  status: 'PENDING' | 'PASS' | 'FAIL';
}

type GameState = 'IDLE' | 'ROLLING' | 'BATTLE' | 'SUMMARY';

// --- 基础配置与响应式状态 ---
const scale = ref(1);
const baseWidth = 1920;
const baseHeight = 1080;

const gameState = ref<GameState>('IDLE');
const showConfig = ref(false);
const selectionCount = ref(6); // 默认抽选 6 人
const activeIdx = ref(0); // 当前正在背诵的索引
const missionQueue = ref<Mission[]>([]);

const students = ref<string[]>([]);
const poems = ref<{ title: string; author: string }[]>([]);

// 当前选中的任务（用于大屏显示）
const activeMission = computed(() => missionQueue.value[activeIdx.value] || null);

// --- 窗口缩放逻辑 ---
const updateScale = () => {
  scale.value = Math.min(window.innerWidth / baseWidth, window.innerHeight / baseHeight);
};

// --- 数据加载 ---
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

// --- 核心业务逻辑：批量滚动抽选 ---
const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const startBatchRoll = async () => {
  if (students.value.length < selectionCount.value) return alert("学生人数不足！");
  if (poems.value.length < selectionCount.value) return alert("课文储备不足！");

  // 1. 初始化
  gameState.value = 'ROLLING';
  missionQueue.value = [];
  activeIdx.value = 0;

  // 用于去重的集合
  const usedStudents = new Set<string>();
  const usedPoems = new Set<string>();

  // 2. 依次锁定每个人
  for (let i = 0; i < selectionCount.value; i++) {
    // 设置当前焦点，确保大屏展示正在滚动的这一项
    activeIdx.value = i;
    
    const currentMission = reactive<Mission>({
      student: '???',
      poem: '???',
      author: '???',
      status: 'PENDING'
    });
    missionQueue.value.push(currentMission);

    // 视觉上的滚动效果
    const rollDuration = 800; // 每个人滚动 0.8 秒
    const rollStart = Date.now();
    
    while (Date.now() - rollStart < rollDuration) {
      // 随机池过滤已选中的
      const availableStudents = students.value.filter(s => !usedStudents.has(s));
      const availablePoems = poems.value.filter(p => !usedPoems.has(p.title));

      const randomS = availableStudents[Math.floor(Math.random() * availableStudents.length)];
      const randomP = availablePoems[Math.floor(Math.random() * availablePoems.length)];
      
      if (randomS) currentMission.student = randomS;
      if (randomP) {
        currentMission.poem = randomP.title;
        currentMission.author = randomP.author;
      }
      
      await sleep(50); // 闪烁频率
    }
    
    // 锁定最终结果并加入去重集合
    usedStudents.add(currentMission.student);
    usedPoems.add(currentMission.poem);
    
    await sleep(200); // 定格停顿感
  }

  // 3. 抽选完成，重置焦点到第一位，进入背诵环节
  activeIdx.value = 0;
  gameState.value = 'BATTLE';
};

// --- 判定与推进逻辑 ---
const handleJudge = (result: 'PASS' | 'FAIL') => {
  if (missionQueue.value[activeIdx.value]) {
    missionQueue.value[activeIdx.value].status = result;
  }
  
  // 如果是最后一位，进入总结
  if (activeIdx.value === missionQueue.value.length - 1) {
    gameState.value = 'SUMMARY';
  } else {
    // 自动切到下一位
    activeIdx.value++;
  }
};

const resetGame = () => {
  gameState.value = 'IDLE';
  missionQueue.value = [];
  activeIdx.value = 0;
};

// --- 生命周期 ---
onMounted(() => {
  loadData();
  updateScale();
  window.addEventListener('resize', updateScale);
});

onUnmounted(() => {
  window.removeEventListener('resize', updateScale);
});
</script>

<template>
  <div class="app-container">
    <div 
      class="main-scaler" 
      :style="{ 
        transform: `scale(${scale})`, 
        width: `${baseWidth}px`, 
        height: `${baseHeight}px` 
      }"
    >
      <!-- 顶栏：配置与人数设置 -->
      <div class="top-bar">
        <div class="count-selector" v-if="gameState === 'IDLE'">
            <span class="label">作战人数:</span>
            <input type="range" min="1" max="10" v-model.number="selectionCount">
            <span class="value">{{ selectionCount }}</span>
        </div>
        <button class="btn-config" @click="showConfig = true" :disabled="gameState === 'ROLLING'">
          ⚙️ 战术配置 (CONFIG)
        </button>
      </div>

      <!-- 主战场：Manga 分镜布局 -->
      <div class="battle-arena">
        <!-- A. 实时对战分镜 -->
        <div class="display-main comic-panel">
          <div class="panel-tag">LIVE BATTLE / 战况实时</div>
          
          <template v-if="gameState === 'IDLE'">
            <div class="idle-placeholder">READY TO BATTLE</div>
          </template>
          
          <template v-else-if="gameState === 'SUMMARY'">
            <div class="summary-board">
              <div class="board-title">MISSION SUMMARY / 作战战报</div>
              <div class="summary-list">
                <div v-for="(m, i) in missionQueue" :key="i" class="summary-item" :class="m.status">
                    <span class="idx">{{ i+1 }}</span>
                    <span class="name">{{ m.student }}</span>
                    <span class="poem">《{{ m.poem }}》</span>
                    <span class="status-tag">{{ m.status }}</span>
                </div>
              </div>
              <button class="btn-reset" @click="resetGame">RETURN / 撤退</button>
            </div>
          </template>

          <template v-else>
            <div class="focus-container">
                <LotterySlot 
                    side="left" 
                    label="📜 MISSION / 绝密任务" 
                    :subtext="activeMission?.author || '未知'"
                    :text="activeMission ? `《${activeMission.poem}》` : '???'" 
                    :isRolling="gameState === 'ROLLING' && activeIdx === missionQueue.indexOf(activeMission!)" 
                />
                <div class="vs-text">VS</div>
                <LotterySlot 
                    side="right" 
                    label="🎯 TARGET / 锁定目标" 
                    :text="activeMission?.student || '???'" 
                    :isRolling="gameState === 'ROLLING' && activeIdx === missionQueue.indexOf(activeMission!)" 
                />
            </div>
          </template>
        </div>

        <!-- B. 任务队列分镜 -->
        <div class="mission-log comic-panel">
          <div class="panel-tag log-tag">MISSION QUEUE / 作战序列</div>
          <div class="log-list">
            <div 
              v-for="(mission, index) in missionQueue" 
              :key="index" 
              class="log-card"
              :class="{ 'is-active': index === activeIdx && gameState !== 'SUMMARY', [mission.status]: true }"
            >
              <div class="log-idx">{{ index + 1 }}</div>
              <div class="log-body">
                <div class="log-name">{{ mission.student }}</div>
                <div class="log-poem">《{{ mission.poem }}》- {{ mission.author }}</div>
              </div>
              <div class="log-status">{{ mission.status }}</div>
            </div>
            <div v-if="missionQueue.length === 0" class="log-empty">等待情报载入...</div>
          </div>
        </div>
      </div>

      <!-- 控制台 -->
      <ControlDeck 
        :game-state="gameState" 
        @roll="startBatchRoll" 
        @judge="handleJudge" 
        @reset="resetGame"
      />

      <!-- 战术配置弹窗 -->
      <div v-if="showConfig" class="modal-overlay">
        <ConfigPanel @close="showConfig = false" @saved="loadData" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  width: 100vw; height: 100vh;
  overflow: hidden;
  background-color: #e0e0e0;
  background-image: radial-gradient(rgba(0,0,0,0.15) 15%, transparent 16%), 
                    radial-gradient(rgba(0,0,0,0.15) 15%, transparent 16%);
  background-size: 15px 15px;
  background-position: 0 0, 7.5px 7.5px;
  display: flex; justify-content: center; align-items: center;
}

.main-scaler {
  position: relative; flex-shrink: 0;
  display: flex; flex-direction: column;
  transform-origin: center center;
  background: transparent;
}

.top-bar { 
  height: 100px; display: flex; justify-content: flex-end; align-items: center; 
  padding: 0 60px; z-index: 10; gap: 30px;
}

.count-selector {
    background: #fff; border: 4px solid #000; padding: 10px 25px;
    display: flex; align-items: center; gap: 20px; font-weight: 900;
    box-shadow: 6px 6px 0 #000;
}
.count-selector .label { font-size: 18px; letter-spacing: 1px; }
.count-selector .value { 
    font-size: 32px; color: var(--manga-red); min-width: 40px; text-align: center;
    text-shadow: 2px 2px 0 #000;
}

/* 自定义滑块样式 */
.count-selector input[type=range] {
  -webkit-appearance: none; width: 150px; background: transparent;
}
.count-selector input[type=range]:focus { outline: none; }

/* 轨道 */
.count-selector input[type=range]::-webkit-slider-runnable-track {
  width: 100%; height: 10px; cursor: pointer;
  background: #000; border-radius: 5px;
}
/* 滑块钮 (Chrome) */
.count-selector input[type=range]::-webkit-slider-thumb {
  height: 30px; width: 30px; border-radius: 50%;
  background: var(--manga-yellow); border: 4px solid #000;
  cursor: pointer; -webkit-appearance: none; margin-top: -12px;
  box-shadow: 3px 3px 0 #000; transition: transform 0.1s;
}
.count-selector input[type=range]::-webkit-slider-thumb:hover { transform: scale(1.2); }
.count-selector input[type=range]::-webkit-slider-thumb:active { background: var(--manga-red); }

/* Firefox 滑块 */
.count-selector input[type=range]::-moz-range-track {
  width: 100%; height: 10px; background: #000; border-radius: 5px;
}
.count-selector input[type=range]::-moz-range-thumb {
  height: 25px; width: 25px; background: var(--manga-yellow); border: 4px solid #000;
  border-radius: 50%; cursor: pointer; box-shadow: 3px 3px 0 #000;
}

.btn-config { 
  background: #fff; border: 4px solid #000; padding: 10px 30px; 
  font-size: 24px; font-weight: bold; cursor: pointer; 
  box-shadow: 8px 8px 0 #000; transition: 0.2s; 
}
.btn-config:hover { background: var(--manga-yellow); transform: scale(1.05); }

.battle-arena {
  flex: 1; display: flex; gap: 40px; padding: 30px 60px; z-index: 5;
}

/* 共享的分镜样式 */
.comic-panel {
  background: #fff; border: 6px solid #000; position: relative;
  box-shadow: 15px 15px 0 rgba(0,0,0,0.1); display: flex; flex-direction: column;
}

.panel-tag {
  position: absolute; top: -20px; left: 30px;
  background: #000; color: var(--manga-yellow);
  padding: 4px 20px; font-weight: 900; font-size: 16px; z-index: 10;
}

.log-tag { left: auto; right: 30px; background: var(--manga-red); color: #fff; }

/* 左侧主显示区 */
.display-main {
  flex: 2.2; display: flex; align-items: center; justify-content: center;
  background: radial-gradient(circle at 30% 30%, #fff 0%, #f0f0f0 100%);
}

.idle-placeholder {
    font-size: 100px; font-weight: 900; color: #ddd;
    text-shadow: 5px 5px 0 var(--manga-black);
    letter-spacing: 5px; transform: rotate(-2deg);
}

.focus-container {
    width: 100%; display: flex; align-items: center; justify-content: space-around;
    animation: slideIn 0.3s cubic-bezier(0.25, 1, 0.5, 1);
}

@keyframes slideIn {
  0% { transform: translateX(50px); opacity: 0; }
  100% { transform: translateX(0); opacity: 1; }
}

.vs-text {
  font-size: 120px; font-weight: 900; font-style: italic;
  color: var(--manga-yellow); text-shadow: 10px 10px 0px #000;
  animation: vs-kick 0.6s infinite alternate cubic-bezier(0.68, -0.55, 0.27, 1.55);
}

@keyframes vs-kick {
  from { transform: scale(1) rotate(-10deg); }
  to { transform: scale(1.1) rotate(10deg); }
}

/* 右侧日志区 */
.mission-log {
  flex: 1; background: #fafafa;
}

.log-list {
    flex: 1; padding: 40px 20px 20px; display: flex; flex-direction: column; gap: 15px;
    overflow-y: auto;
}

.log-card {
    background: #fff; border: 3px solid #000; padding: 12px;
    display: flex; align-items: center; gap: 15px; position: relative;
    transition: all 0.3s;
}

.log-card.is-active {
    border-color: var(--manga-red); border-width: 5px;
    transform: scale(1.05) translateX(-5px);
    box-shadow: 8px 8px 0 var(--manga-red);
    z-index: 5;
}

.log-card.PASS { background: #e8f5e9; border-color: #2e7d32; }
.log-card.FAIL { background: #ffebee; border-color: #c62828; }

.log-idx {
    width: 35px; height: 35px; background: #000; color: #fff;
    display: flex; justify-content: center; align-items: center; font-weight: bold;
}

.log-body { flex: 1; }
.log-name { font-size: 20px; font-weight: 900; }
.log-poem { font-size: 14px; color: #666; }

.log-status { font-weight: 900; font-style: italic; font-size: 18px; }
.PASS .log-status { color: #2e7d32; }
.FAIL .log-status { color: #c62828; }

.log-empty {
    height: 100%; display: flex; align-items: center; justify-content: center;
    color: #999; font-weight: bold; font-style: italic; font-size: 20px;
}

/* 战报样式 */
.summary-board {
    width: 90%; background: #fff; border: 8px solid #000; padding: 40px;
    box-shadow: 20px 20px 0 #000; display: flex; flex-direction: column; align-items: center;
    z-index: 100;
}
.board-title { font-size: 48px; font-weight: 900; margin-bottom: 30px; border-bottom: 6px solid #000; }
.summary-list { width: 100%; display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 40px; }
.summary-item { 
    display: flex; align-items: center; gap: 15px; padding: 15px; border: 3px solid #000; font-weight: bold;
}
.summary-item.PASS { border-color: #2e7d32; background: #e8f5e9; }
.summary-item.FAIL { border-color: #c62828; background: #ffebee; }
.btn-reset {
    background: var(--manga-black); color: var(--manga-yellow); border: none;
    padding: 15px 60px; font-size: 28px; font-weight: 900; cursor: pointer;
    box-shadow: 6px 6px 0 var(--manga-red);
}
.btn-reset:active { transform: translate(3px, 3px); box-shadow: 3px 3px 0 var(--manga-red); }

.modal-overlay {
  position: absolute; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0, 0, 0, 0.05); backdrop-filter: blur(10px);
  display: flex; justify-content: center; align-items: center; z-index: 9999;
}
</style>