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

// 反馈特效状态
const feedback = reactive({
  show: false,
  type: 'PASS' as 'PASS' | 'FAIL',
  text: '',
});

// --- 文案库 ---
const passSlogans = [
  "完美吟唱！不愧是你！",
  "战斗力突破天际！",
  "这等咒语也想难倒我？",
  "毫无破绽的绝对领域！",
  "神级背诵，全场膜拜！",
  "这种程度，不过是热身罢了！"
];

const failSlogans = [
  "卡壳了？回家再练一百年吧！",
  "就这？就这？",
  "大脑内存已清空...",
  "吟唱中断，魔法反噬！",
  "连这种常识都忘了？太逊了！",
  "被自己的结界困住了吗？"
];

const triggerFeedback = (type: 'PASS' | 'FAIL') => {
  feedback.type = type;
  const slogans = type === 'PASS' ? passSlogans : failSlogans;
  const idx = Math.floor(Math.random() * slogans.length);
  feedback.text = slogans[idx] ?? '';
  feedback.show = true;
  
  // 播放对应状态音效 + 沉重砸下音效
  playSound('smash');
  setTimeout(() => {
    playSound(type === 'PASS' ? 'pass' : 'fail');
  }, 100); // 稍微错开一点点，让音效有层次感
  
  // 1.5秒后自动隐藏
  setTimeout(() => {
    feedback.show = false;
  }, 1500);
};

// --- 音效资源初始化 ---
// 由于 freesound 等外链容易被浏览器拦截或失效，这里换用一套更稳定的、来自维基媒体及常用公开图床/CDN的提示音效
const audioContext = {
  roll: new Audio('https://upload.wikimedia.org/wikipedia/commons/3/3d/Tick_tock.ogg'),        // 经典的滴答滴答声，稳定且清脆
  lock: new Audio('https://upload.wikimedia.org/wikipedia/commons/b/bd/Camera_click.ogg'),     // 相机快门般的清脆锁定声
  pass: new Audio('https://upload.wikimedia.org/wikipedia/commons/e/ec/Success_1.wav'),        // 经典的成功提示音
  fail: new Audio('https://upload.wikimedia.org/wikipedia/commons/1/15/Buzzer.ogg'),           // 刺耳的错误蜂鸣声
  smash: new Audio('https://upload.wikimedia.org/wikipedia/commons/d/d4/Drum_snare_hit.ogg'),  // 沉重的鼓击声（模拟砸印章）
};

// 预加载设置并调低一些音量防炸麦
Object.values(audioContext).forEach(audio => {
  audio.volume = 0.5;
  audio.preload = 'auto';
});
// 恢复滚动音量，因为这次的木琴音效本身比较轻柔
audioContext.roll.volume = 0.6;

// 播放工具函数
const playSound = (type: keyof typeof audioContext) => {
  const audio = audioContext[type];
  audio.currentTime = 0;
  audio.play().catch(e => console.log('浏览器限制自动播放，需用户先交互:', e));
};

const students = ref<string[]>([]);
const poems = ref<{ title: string; author: string }[]>([]);

// 当前选中的任务（用于大屏显示）
const activeMission = computed(() => missionQueue.value[activeIdx.value] || null);

// 任务进度计算
const missionProgress = computed(() => {
  if (missionQueue.value.length === 0) return 0;
  if (gameState.value === 'SUMMARY') return 100;
  return (activeIdx.value / missionQueue.value.length) * 100;
});

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
    
    // 开始滚动时播放循环音效
    playSound('roll');
    // 让滚动音效循环播放
    audioContext.roll.loop = true;
    
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
    
    // 停止滚动音效
    audioContext.roll.pause();
    audioContext.roll.currentTime = 0;
    
    // 锁定最终结果并加入去重集合，播放锁定音效
    usedStudents.add(currentMission.student);
    usedPoems.add(currentMission.poem);
    playSound('lock');
    
    await sleep(300); // 定格停顿感稍微加长一点配合音效
  }

  // 3. 抽选完成，重置焦点到第一位，进入背诵环节
  activeIdx.value = 0;
  gameState.value = 'BATTLE';
};

// --- 判定与推进逻辑 ---
const handleJudge = (result: 'PASS' | 'FAIL') => {
  const currentMission = missionQueue.value[activeIdx.value];
  if (currentMission) {
    currentMission.status = result;
    triggerFeedback(result);
  }
  
  // 延迟一小会儿再进入下一个或总结，给反馈特效留出展示时间
  setTimeout(() => {
    // 如果是最后一位，进入总结
    if (activeIdx.value === missionQueue.value.length - 1) {
      gameState.value = 'SUMMARY';
    } else {
      // 自动切到下一位
      activeIdx.value++;
    }
  }, 1500);
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
        <!-- A. 实时对战分镜 & 控制台 -->
        <div class="display-main">
          
          <template v-if="gameState === 'IDLE'">
            <div class="idle-view">
              <div class="graffiti-slogan">
                READY TO BATTLE
              </div>
              
              <div class="idle-action">
                <ControlDeck 
                  :game-state="gameState" 
                  @roll="startBatchRoll" 
                  @judge="handleJudge" 
                  @reset="resetGame"
                />
              </div>
            </div>
          </template>
          
          <template v-else-if="gameState !== 'SUMMARY'">
            <div class="focus-container">
                <LotterySlot 
                    side="left" 
                    label="🎯 TARGET / 锁定目标" 
                    :text="activeMission?.student || '???'" 
                    :isRolling="gameState === 'ROLLING' && activeIdx === missionQueue.indexOf(activeMission!)" 
                />
                <div class="vs-text">VS</div>
                <LotterySlot 
                    side="right" 
                    label="📜 MISSION / 绝密任务" 
                    :subtext="activeMission?.author || '未知'"
                    :text="activeMission ? `《${activeMission.poem}》` : '???'" 
                    :isRolling="gameState === 'ROLLING' && activeIdx === missionQueue.indexOf(activeMission!)" 
                />
            </div>
            
            <!-- 战斗状态下的控制台（判定按钮等）移至左侧下方 -->
            <div class="control-deck-wrapper">
              <ControlDeck 
                :game-state="gameState" 
                @roll="startBatchRoll" 
                @judge="handleJudge" 
                @reset="resetGame"
              />
            </div>
          </template>
        </div>

        <!-- 酷炫分割线 -->
        <div class="arena-divider" v-if="gameState !== 'SUMMARY'">
          <div class="divider-line"></div>
          <div class="divider-text">WARNING</div>
          <div class="divider-line"></div>
        </div>

        <!-- B. 任务队列分镜 -->
        <div class="mission-log comic-panel" v-if="gameState !== 'SUMMARY'">
          <div class="panel-tag log-tag">MISSION QUEUE / 作战序列</div>
          
          <!-- 任务进度条 -->
          <div class="progress-container" v-if="gameState !== 'IDLE'">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: `${missionProgress}%` }"></div>
            </div>
            <div class="progress-text">
              <span>{{ activeIdx }} / {{ missionQueue.length }}</span>
            </div>
          </div>

          <div class="log-list">
            <div 
              v-for="(mission, index) in missionQueue" 
              :key="index" 
              class="log-card"
              :class="{ 'is-active': index === activeIdx, [mission.status]: true }"
            >
              <!-- 漫画风序号标牌 -->
              <div class="log-idx-wrapper">
                <div class="log-idx-bg"></div>
                <div class="log-idx">No.{{ index + 1 }}</div>
              </div>
              
              <div class="log-body">
                <div class="log-name">{{ mission.student }}</div>
                <div class="log-poem">
                  <span class="poem-title">《{{ mission.poem }}》</span>
                  <span class="poem-author">{{ mission.author }}</span>
                </div>
              </div>
              
              <!-- 漫画风状态印章 -->
              <div class="log-status-stamp" :class="mission.status">
                <div class="stamp-inner">{{ mission.status }}</div>
              </div>
            </div>
            <div v-if="missionQueue.length === 0" class="log-empty">
              <div class="empty-text">NO MISSIONS</div>
              <div class="empty-subtext">等待情报载入...</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 战术配置弹窗 -->
      <div v-if="showConfig" class="modal-overlay">
        <ConfigPanel @close="showConfig = false" @saved="loadData" />
      </div>

      <!-- 判定反馈全屏特效 -->
      <div v-if="feedback.show" class="feedback-overlay" :class="feedback.type">
        <!-- 激进的放射状背景冲击波 -->
        <div class="impact-lines"></div>
        
        <div class="feedback-content">
          <!-- 判定结果印章 -->
          <div class="feedback-stamp-wrapper">
            <div class="feedback-stamp">{{ feedback.type }}</div>
            <!-- 碎裂或发光特效粒子 -->
            <div class="particles"></div>
          </div>
          <!-- 嘲讽/燃向文案 -->
          <div class="feedback-text">{{ feedback.text }}</div>
        </div>
      </div>

      <!-- 独立的全屏战报面板（不在 battle-arena 内部） -->
      <div v-if="gameState === 'SUMMARY'" class="summary-board-radical">
        <div class="summary-bg-decor"></div>
        <div class="summary-title-huge">
          <div class="title-bg">REPORT</div>
          <div class="title-fg">MISSION<br/>SUMMARY</div>
        </div>
        
        <div class="radical-list-container">
          <div 
            v-for="(m, i) in missionQueue" 
            :key="i" 
            class="radical-item" 
            :class="[m.status, i % 2 === 0 ? 'tilt-left' : 'tilt-right']"
          >
            <div class="r-idx">0{{ i+1 }}</div>
            <div class="r-content">
              <div class="r-name">{{ m.student }}</div>
              <div class="r-poem">《{{ m.poem }}》</div>
            </div>
            <div class="r-stamp">{{ m.status }}</div>
            <div class="r-tape"></div>
          </div>
        </div>
        
        <button class="btn-radical-return" @click="resetGame">
          <span>RESTART // 再次开战</span>
        </button>
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
  overflow: hidden; /* 防止内部元素撑破容器 */
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
  flex: 2.2; display: flex; flex-direction: column; align-items: center; justify-content: center;
  position: relative;
}

.control-deck-wrapper {
  margin-top: 40px;
  animation: slideUp 0.4s cubic-bezier(0.25, 1, 0.5, 1);
}

@keyframes slideUp {
  0% { transform: translateY(30px); opacity: 0; }
  100% { transform: translateY(0); opacity: 1; }
}

/* 酷炫分割线 */
.arena-divider {
  width: 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20px;
  position: relative;
}

.divider-line {
  flex: 1;
  width: 6px;
  background: repeating-linear-gradient(
    0deg,
    #000,
    #000 15px,
    transparent 15px,
    transparent 30px
  );
  position: relative;
}

.divider-line::after {
  content: '';
  position: absolute;
  top: 0;
  left: 10px;
  width: 2px;
  height: 100%;
  background: var(--manga-red);
  opacity: 0.5;
}

.divider-text {
  writing-mode: vertical-rl;
  text-orientation: mixed;
  font-size: 24px;
  font-weight: 900;
  color: var(--manga-yellow);
  background: #000;
  padding: 20px 5px;
  border: 4px solid var(--manga-red);
  letter-spacing: 10px;
  text-shadow: 2px 2px 0 var(--manga-red);
  box-shadow: 4px 4px 0 rgba(0,0,0,0.3);
}

.idle-view {
    width: 100%; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center;
    gap: 60px; animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.graffiti-slogan {
    font-size: 120px;
    font-weight: 900;
    font-family: "Impact", "Arial Black", sans-serif;
    color: transparent;
    -webkit-text-stroke: 4px var(--manga-black);
    text-transform: uppercase;
    text-align: center;
    transform: rotate(-5deg);
    position: relative;
    letter-spacing: 2px;
}

.graffiti-slogan::before {
    content: 'READY TO BATTLE';
    position: absolute;
    top: 8px;
    left: 8px;
    color: var(--manga-yellow);
    -webkit-text-stroke: 0;
    z-index: -1;
}

.graffiti-slogan::after {
    content: 'READY TO BATTLE';
    position: absolute;
    top: -6px;
    left: -6px;
    color: var(--manga-red);
    -webkit-text-stroke: 0;
    z-index: -2;
    opacity: 0.8;
}

.idle-action {
    transform: scale(1.1);
}

.focus-container {
    width: 100%; display: flex; flex-direction: column; align-items: center; justify-content: space-around;
    gap: 20px;
    animation: slideIn 0.3s cubic-bezier(0.25, 1, 0.5, 1);
}

@keyframes slideIn {
  0% { transform: translateY(50px); opacity: 0; }
  100% { transform: translateY(0); opacity: 1; }
}

.vs-text {
  font-size: 80px; font-weight: 900; font-style: italic;
  color: var(--manga-yellow); text-shadow: 6px 6px 0px #000;
  animation: vs-kick 0.6s infinite alternate cubic-bezier(0.68, -0.55, 0.27, 1.55);
}

@keyframes vs-kick {
  from { transform: scale(1) rotate(-10deg); }
  to { transform: scale(1.1) rotate(10deg); }
}

/* 右侧日志区 */
.mission-log {
  flex: 1; background: #fafafa;
  background-image: repeating-linear-gradient(45deg, #f0f0f0 25%, transparent 25%, transparent 75%, #f0f0f0 75%, #f0f0f0), repeating-linear-gradient(45deg, #f0f0f0 25%, #fafafa 25%, #fafafa 75%, #f0f0f0 75%, #f0f0f0);
  background-position: 0 0, 10px 10px;
  background-size: 20px 20px;
  display: flex;
  flex-direction: column;
}

/* 进度条设计 */
.progress-container {
  padding: 30px 20px 20px;
  display: flex;
  align-items: center;
  gap: 15px;
  z-index: 10;
  background: var(--manga-black);
  border-bottom: 4px solid #000;
  box-shadow: 0 4px 0 rgba(0,0,0,0.1);
}

.progress-bar {
  flex: 1;
  height: 24px;
  background: #333;
  border: 3px solid #000;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);
  transform: skewX(-15deg);
}

.progress-fill {
  height: 100%;
  background: var(--manga-yellow);
  background-image: repeating-linear-gradient(
    -45deg,
    rgba(255,255,255,0.4),
    rgba(255,255,255,0.4) 15px,
    transparent 15px,
    transparent 30px
  );
  transition: width 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  border-right: 3px solid #000;
  box-shadow: 0 0 10px rgba(255, 229, 0, 0.5);
}

.progress-text {
  font-family: "Impact", "Arial Black", sans-serif;
  font-size: 24px;
  font-style: italic;
  font-weight: 900;
  color: var(--manga-yellow);
  text-shadow: 2px 2px 0 var(--manga-red);
  min-width: 70px;
  text-align: right;
  letter-spacing: 2px;
}

.log-list {
    flex: 1; padding: 20px 20px; display: flex; flex-direction: column; gap: 20px;
    overflow-y: auto;
    overflow-x: hidden;
    min-height: 0; /* 关键：允许 flex 子项收缩，从而触发滚动条 */
}

.log-card {
    background: #fff; border: 4px solid #000; padding: 15px;
    display: flex; align-items: center; gap: 15px; position: relative;
    box-shadow: 6px 6px 0 rgba(0,0,0,0.1);
    transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
    clip-path: polygon(2% 0, 100% 0, 98% 100%, 0 100%);
}

.log-card:hover {
    transform: translateY(-2px);
    box-shadow: 8px 8px 0 rgba(0,0,0,0.2);
}

.log-card.is-active {
    border-color: var(--manga-red); border-width: 5px;
    transform: scale(1.05) translateX(-10px);
    box-shadow: 10px 10px 0 var(--manga-red);
    z-index: 5;
    clip-path: none;
    background: repeating-linear-gradient(
      -45deg,
      #fff,
      #fff 10px,
      #fff9f9 10px,
      #fff9f9 20px
    );
}

.log-card.PASS { background: #f0fdf4; border-color: #059669; }
.log-card.FAIL { background: #fef2f2; border-color: #dc2626; }

/* 序号标牌设计 */
.log-idx-wrapper {
    position: relative; width: 50px; height: 50px;
    display: flex; justify-content: center; align-items: center;
    transform: rotate(-5deg);
}
.log-idx-bg {
    position: absolute; inset: 0; background: var(--manga-yellow);
    border: 3px solid #000; clip-path: polygon(10% 0, 100% 10%, 90% 100%, 0 90%);
}
.is-active .log-idx-bg { background: var(--manga-red); }
.log-idx {
    position: relative; z-index: 1; font-weight: 900; font-size: 16px;
    color: #000; text-shadow: 1px 1px 0 #fff;
}
.is-active .log-idx { color: #fff; text-shadow: 2px 2px 0 #000; font-size: 18px; }

.log-body { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.log-name { font-size: 24px; font-weight: 900; letter-spacing: 1px; }
.log-poem { display: flex; align-items: baseline; gap: 5px; }
.poem-title { font-size: 16px; font-weight: bold; color: #333; }
.poem-author { font-size: 12px; color: #888; background: #eee; padding: 2px 6px; border-radius: 4px; font-style: italic; }

/* 状态印章设计 */
.log-status-stamp {
    width: 60px; height: 60px; display: flex; justify-content: center; align-items: center;
    font-weight: 900; font-style: italic; font-size: 16px;
    border: 3px solid #ccc; color: #ccc; border-radius: 50%;
    transform: rotate(10deg); opacity: 0.5;
}
.stamp-inner { border: 1px solid currentColor; border-radius: 50%; width: 48px; height: 48px; display: flex; align-items: center; justify-content: center; }

.PASS .log-status-stamp { 
    color: #059669; border-color: #059669; opacity: 1; transform: rotate(-5deg) scale(1.1);
    box-shadow: 2px 2px 0 rgba(5, 150, 105, 0.2);
}
.FAIL .log-status-stamp { 
    color: #dc2626; border-color: #dc2626; opacity: 1; transform: rotate(15deg) scale(1.1);
    box-shadow: 2px 2px 0 rgba(220, 38, 38, 0.2);
}

.log-empty {
    height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center;
    color: #bbb; gap: 10px;
}
.empty-text { font-size: 36px; font-weight: 900; font-style: italic; text-shadow: 2px 2px 0 #eee; }
.empty-subtext { font-weight: bold; font-size: 16px; letter-spacing: 2px; }

/* 判定反馈特效 */
.feedback-overlay {
  position: absolute;
  top: 0; left: 0;
  width: 100%; height: 100%;
  z-index: 9000;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  animation: flash-bg 1.5s ease-out forwards;
  overflow: hidden;
}

@keyframes flash-bg {
  0% { opacity: 0; }
  5% { opacity: 1; background: rgba(255, 255, 255, 0.8); }
  10% { opacity: 1; background: rgba(0, 0, 0, 0.6); }
  80% { opacity: 1; }
  100% { opacity: 0; }
}

/* 放射状冲击波背景 */
.impact-lines {
  position: absolute;
  top: 50%; left: 50%;
  width: 200vw; height: 200vw;
  transform: translate(-50%, -50%);
  background: repeating-conic-gradient(
    from 0deg,
    transparent 0deg,
    transparent 5deg,
    rgba(255, 255, 255, 0.1) 5deg,
    rgba(255, 255, 255, 0.1) 10deg
  );
  animation: spin-impact 10s linear infinite;
  z-index: 1;
}

.feedback-overlay.PASS .impact-lines {
  background: repeating-conic-gradient(
    from 0deg, transparent 0deg, transparent 5deg, rgba(0, 229, 255, 0.15) 5deg, rgba(0, 229, 255, 0.15) 10deg
  );
}

.feedback-overlay.FAIL .impact-lines {
  background: repeating-conic-gradient(
    from 0deg, transparent 0deg, transparent 5deg, rgba(255, 42, 85, 0.15) 5deg, rgba(255, 42, 85, 0.15) 10deg
  );
}

@keyframes spin-impact {
  from { transform: translate(-50%, -50%) rotate(0deg); }
  to { transform: translate(-50%, -50%) rotate(360deg); }
}

.feedback-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 30px;
  transform: rotate(-5deg);
  z-index: 10;
  animation: stamp-smash 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards, screen-shake 0.4s 0.1s ease-in-out;
}

/* 强烈的屏幕震动特效 */
@keyframes screen-shake {
  0%, 100% { transform: rotate(-5deg) translate(0, 0); }
  10% { transform: rotate(-6deg) translate(-20px, 20px); }
  30% { transform: rotate(-4deg) translate(20px, -20px); }
  50% { transform: rotate(-7deg) translate(-20px, -20px); }
  70% { transform: rotate(-3deg) translate(20px, 20px); }
  90% { transform: rotate(-5deg) translate(-10px, 10px); }
}

@keyframes stamp-smash {
  0% { transform: scale(4) rotate(30deg); opacity: 0; filter: blur(20px); }
  100% { transform: scale(1) rotate(-5deg); opacity: 1; filter: blur(0); }
}

.feedback-stamp-wrapper {
  position: relative;
}

.feedback-stamp {
  font-family: "Impact", sans-serif;
  font-size: 180px;
  font-weight: 900;
  font-style: italic;
  padding: 10px 80px;
  border: 20px solid currentColor;
  border-radius: 20px;
  line-height: 1;
  text-transform: uppercase;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(10px);
}

/* 印章砸下时的粒子爆发 */
.particles {
  position: absolute;
  inset: -50px;
  background-image: 
    radial-gradient(circle at 50% 50%, currentColor 10%, transparent 10%),
    radial-gradient(circle at 20% 80%, currentColor 10%, transparent 10%),
    radial-gradient(circle at 80% 20%, currentColor 10%, transparent 10%),
    radial-gradient(circle at 10% 20%, currentColor 10%, transparent 10%),
    radial-gradient(circle at 90% 80%, currentColor 10%, transparent 10%);
  background-size: 20px 20px, 15px 15px, 25px 25px, 10px 10px, 30px 30px;
  background-repeat: no-repeat;
  opacity: 0;
  animation: particle-burst 0.5s 0.2s ease-out forwards;
}

@keyframes particle-burst {
  0% { transform: scale(0.5); opacity: 1; }
  100% { transform: scale(2); opacity: 0; }
}

.feedback-text {
  font-size: 60px;
  font-weight: 900;
  color: #fff;
  background: #000;
  padding: 20px 50px;
  border: 8px solid currentColor;
  text-align: center;
  max-width: 90%;
  white-space: pre-wrap;
  box-shadow: 20px 20px 0 rgba(0,0,0,0.8);
  letter-spacing: 4px;
  transform: skewX(-10deg);
}

.feedback-overlay.PASS .feedback-stamp {
  color: #00E5FF;
  text-shadow: 10px 10px 0 #000;
  box-shadow: inset 0 0 50px rgba(0, 229, 255, 0.8), 0 0 50px rgba(0, 229, 255, 0.8);
}
.feedback-overlay.PASS .particles { color: #00E5FF; }
.feedback-overlay.PASS .feedback-text {
  border-color: #00E5FF;
  color: #00E5FF;
}

.feedback-overlay.FAIL .feedback-stamp {
  color: #FF2A55;
  text-shadow: 10px 10px 0 #000;
  box-shadow: inset 0 0 50px rgba(255, 42, 85, 0.8), 0 0 50px rgba(255, 42, 85, 0.8);
}
.feedback-overlay.FAIL .particles { color: #FF2A55; }
.feedback-overlay.FAIL .feedback-text {
  border-color: #FF2A55;
  color: #FF2A55;
}

/* 激进战报样式重构 */
.summary-board-radical {
  width: 100vw;
  height: 100vh;
  position: fixed;
  top: 0; left: 0;
  background: var(--manga-black);
  z-index: 9999;
  display: flex;
  overflow: hidden;
  animation: smash-in 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes smash-in {
  0% { transform: scale(3) rotate(20deg); opacity: 0; filter: blur(10px); }
  100% { transform: scale(1) rotate(0deg); opacity: 1; filter: blur(0); }
}

.summary-bg-decor {
  position: absolute;
  inset: -50%;
  background-image: repeating-linear-gradient(
    45deg,
    #111 0, #111 20px,
    #000 20px, #000 40px
  );
  z-index: 1;
  animation: bg-scroll 20s linear infinite;
}

@keyframes bg-scroll {
  0% { transform: translate(0, 0); }
  100% { transform: translate(-10%, -10%); }
}

.summary-title-huge {
  flex: 1;
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-left: 50px;
}

.title-bg {
  position: absolute;
  font-size: 300px;
  font-family: "Impact", sans-serif;
  color: #222;
  transform: rotate(-90deg) translateX(-20%);
  white-space: nowrap;
  letter-spacing: 20px;
  z-index: -1;
}

.title-fg {
  font-family: "Impact", sans-serif;
  font-size: 140px;
  color: var(--manga-yellow);
  line-height: 0.85;
  text-shadow: 
    -5px -5px 0 #fff,
    5px -5px 0 #fff,
    -5px 5px 0 #fff,
    5px 5px 0 #fff,
    20px 20px 0 var(--manga-red);
  transform: rotate(-10deg);
}

.radical-list-container {
  flex: 2;
  z-index: 10;
  padding: 80px 100px 80px 0;
  display: flex;
  flex-direction: column;
  gap: 30px;
  overflow-y: auto;
  overflow-x: hidden;
}

.radical-list-container::-webkit-scrollbar { display: none; }

.radical-item {
  position: relative;
  background: #fff;
  border: 8px solid #000;
  padding: 20px 40px;
  display: flex;
  align-items: center;
  gap: 30px;
  box-shadow: 15px 15px 0 var(--manga-red);
  transition: transform 0.2s, box-shadow 0.2s;
}

.radical-item:hover {
  transform: scale(1.05) !important;
  box-shadow: 25px 25px 0 var(--manga-blue);
  z-index: 20;
}

.tilt-left { transform: rotate(-3deg) translateX(-40px); }
.tilt-right { transform: rotate(2deg) translateX(40px); }

.r-tape {
  position: absolute;
  top: -15px;
  left: 50%;
  width: 120px;
  height: 40px;
  background: rgba(255, 255, 255, 0.7);
  border: 2px solid #ddd;
  transform: translateX(-50%) rotate(-5deg);
  box-shadow: 2px 2px 5px rgba(0,0,0,0.1);
  backdrop-filter: blur(2px);
}

.r-idx {
  font-family: "Impact", sans-serif;
  font-size: 80px;
  color: #000;
  -webkit-text-stroke: 2px #fff;
  line-height: 1;
}

.r-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.r-name {
  font-size: 48px;
  font-weight: 900;
  letter-spacing: 5px;
  text-transform: uppercase;
}

.r-poem {
  font-size: 24px;
  font-weight: bold;
  color: #666;
  font-style: italic;
  margin-top: -5px;
}

.r-stamp {
  font-family: "Impact", sans-serif;
  font-size: 60px;
  padding: 10px 20px;
  border: 8px solid currentColor;
  border-radius: 10px;
  transform: rotate(15deg);
}

.radical-item.PASS .r-stamp {
  color: #00E5FF;
  text-shadow: 4px 4px 0 #000;
  box-shadow: inset 0 0 20px rgba(0, 229, 255, 0.5);
}

.radical-item.FAIL .r-stamp {
  color: #FF2A55;
  text-shadow: 4px 4px 0 #000;
  box-shadow: inset 0 0 20px rgba(255, 42, 85, 0.5);
  transform: rotate(-15deg);
}

.btn-radical-return {
  position: absolute;
  bottom: 50px;
  left: 50px;
  z-index: 100;
  background: var(--manga-red);
  color: #fff;
  border: 8px solid #000;
  padding: 20px 60px;
  font-size: 40px;
  font-weight: 900;
  font-style: italic;
  cursor: pointer;
  box-shadow: 15px 15px 0 var(--manga-yellow);
  transform: skewX(-15deg);
  transition: all 0.2s;
}

.btn-radical-return:hover {
  background: var(--manga-yellow);
  color: #000;
  box-shadow: 20px 20px 0 var(--manga-red);
  transform: skewX(-15deg) translate(-5px, -5px);
}

.btn-radical-return span {
  display: block;
  transform: skewX(15deg); /* 文字摆正 */
}

.modal-overlay {
  position: absolute; top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0, 0, 0, 0.05); backdrop-filter: blur(10px);
  display: flex; justify-content: center; align-items: center; z-index: 9999;
}
</style>