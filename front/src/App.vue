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

interface ClassStudents {
  className: string;
  members: string[];
}

type GameState = 'IDLE' | 'ROLLING' | 'BATTLE' | 'SUMMARY';

// --- 基础配置与响应式状态 ---
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
  "这种程度，不过是热身罢了！",
  "这就是...羁绊（知识）的力量！",
  "记忆宫殿已完全解锁，满分通关！",
  "区区课文，拿下！",
  "天选之子，背得漂亮！"
];

const failSlogans = [
  "卡壳了？回家再练一百年吧！",
  "就这？就这？",
  "大脑内存已清空...",
  "吟唱中断，魔法反噬！",
  "连这种常识都忘了？太逊了！",
  "被自己的结界困住了吗？",
  "你的记忆力被封印了吗？",
  "CPU温度过高，请求重启！",
  "这可是送分题啊，我的朋友！",
  "快醒醒，现在不是在梦境里！",
  "技能冷却中...请回炉重造！",
  "宋老师的死亡凝视，你感受到了吗？",
  "难道...你中了遗忘魔咒？",
  "不会吧不会吧，真有人背不出来？"
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

type SoundType = 'roll' | 'lock' | 'pass' | 'fail' | 'smash' | 'ui';

let webAudioCtx: AudioContext | null = null;
let rollTimer: number | null = null;
let lastSelectionSfxAt = 0;

const ensureAudioContext = () => {
  if (webAudioCtx) return webAudioCtx;
  const Ctx = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
  if (!Ctx) return null;
  webAudioCtx = new Ctx();
  return webAudioCtx;
};

const unlockAudio = async () => {
  const ctx = ensureAudioContext();
  if (!ctx) return;
  if (ctx.state === 'suspended') {
    await ctx.resume();
  }
};

const playTone = (frequency: number, durationMs: number, type: OscillatorType, volume: number, delayMs = 0) => {
  const ctx = ensureAudioContext();
  if (!ctx) return;
  const startAt = ctx.currentTime + delayMs / 1000;
  const durationSec = durationMs / 1000;

  const oscillator = ctx.createOscillator();
  const gainNode = ctx.createGain();
  oscillator.type = type;
  oscillator.frequency.setValueAtTime(frequency, startAt);
  gainNode.gain.setValueAtTime(0.0001, startAt);
  gainNode.gain.exponentialRampToValueAtTime(Math.max(volume, 0.0001), startAt + 0.01);
  gainNode.gain.exponentialRampToValueAtTime(0.0001, startAt + durationSec);

  oscillator.connect(gainNode);
  gainNode.connect(ctx.destination);
  oscillator.start(startAt);
  oscillator.stop(startAt + durationSec);
};

const startRollSound = () => {
  if (rollTimer !== null) return;
  rollTimer = window.setInterval(() => {
    playTone(1000, 45, 'square', 0.05);
  }, 70);
};

const stopRollSound = () => {
  if (rollTimer === null) return;
  window.clearInterval(rollTimer);
  rollTimer = null;
};

const handleSelectionCountInput = () => {
  if (gameState.value !== 'IDLE') return;
  const now = performance.now();
  if (now - lastSelectionSfxAt < 40) return;
  lastSelectionSfxAt = now;
  playSound('ui');
};

const playSound = async (type: SoundType) => {
  await unlockAudio();
  switch (type) {
    case 'roll':
      startRollSound();
      break;
    case 'lock':
      playTone(1500, 70, 'triangle', 0.08);
      playTone(900, 50, 'triangle', 0.05, 25);
      break;
    case 'pass':
      playTone(659, 70, 'triangle', 0.05);
      playTone(880, 120, 'triangle', 0.08, 70);
      playTone(1175, 180, 'sine', 0.1, 190);
      break;
    case 'fail':
      playTone(392, 120, 'sawtooth', 0.07);
      playTone(311, 130, 'sawtooth', 0.07, 100);
      playTone(262, 180, 'sawtooth', 0.08, 200);
      break;
    case 'smash':
      playTone(180, 120, 'square', 0.09);
      playTone(120, 180, 'square', 0.07, 45);
      break;
    case 'ui':
      playTone(1200, 45, 'triangle', 0.05);
      break;
  }
};

const classes = ref<ClassStudents[]>([]);
const currentClassName = ref<string>('');
const poems = ref<{ title: string; author: string }[]>([]);

const currentStudents = computed(() => {
  const cls = classes.value.find(c => c.className === currentClassName.value);
  return cls ? cls.members : [];
});

// 当前选中的任务（用于大屏显示）
const activeMission = computed(() => missionQueue.value[activeIdx.value] || null);

// 任务进度计算
const missionProgress = computed(() => {
  if (missionQueue.value.length === 0) return 0;
  if (gameState.value === 'SUMMARY') return 100;
  return (activeIdx.value / missionQueue.value.length) * 100;
});

// --- 数据加载 ---
const loadData = async () => {
  try {
    const res = await fetch('/api/config');
    const json = await res.json();
    if (json.code === 200) {
      classes.value = json.data.classes || [];
      if (classes.value.length > 0) {
        currentClassName.value = classes.value[0]?.className ?? '';
      }
      poems.value = json.data.poems || [];
    }
  } catch (e) {
    console.error("数据加载失败");
  }
};

// --- 核心业务逻辑：批量滚动抽选 ---
const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const startBatchRoll = async () => {
  if (currentStudents.value.length < selectionCount.value) return alert("当前班级学生人数不足！");
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
      const availableStudents = currentStudents.value.filter(s => !usedStudents.has(s));
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
    stopRollSound();
    
    // 锁定最终结果并加入去重集合，播放锁定音效
    usedStudents.add(currentMission.student);
    usedPoems.add(currentMission.poem);
    playSound('lock');
    
    await sleep(800); // 定格停顿1秒，给被抽到的同学反应时间
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
  document.addEventListener('pointerdown', unlockAudio, { once: true });
  loadData();
});

onUnmounted(() => {
  stopRollSound();
  document.removeEventListener('pointerdown', unlockAudio);
  if (webAudioCtx && webAudioCtx.state !== 'closed') {
    webAudioCtx.close();
  }
});
</script>

<template>
  <div class="app-container">
    <!-- 极致燃向：漫画集中线背景层 -->
    <div class="manga-focus-lines"></div>
    
    <!-- 极致燃向：无限滚动的警告封条 -->
    <div class="hazard-tape tape-top">
      <div class="tape-text">CAUTION /// BATTLE ZONE /// NO ESCAPE /// CAUTION /// BATTLE ZONE /// NO ESCAPE /// CAUTION /// BATTLE ZONE /// NO ESCAPE ///</div>
    </div>
    <div class="hazard-tape tape-bottom">
      <div class="tape-text">STAND BY /// TARGET LOCKED /// ENGAGE /// STAND BY /// TARGET LOCKED /// ENGAGE /// STAND BY /// TARGET LOCKED /// ENGAGE ///</div>
    </div>

    <!-- 极致燃向：动态漂浮的漫画拟声词/图腾 -->
    <div class="manga-sfx-container">
      <div class="manga-sfx sfx-1">ドドドド</div>
      <div class="manga-sfx sfx-2">ゴゴゴゴ</div>
      <div class="manga-sfx sfx-3">BAM!</div>
    </div>
    
    <div class="main-layout">
      <!-- 顶栏：配置与人数设置 -->
      <div class="top-bar">
        <div class="radical-class-selector" v-if="gameState === 'IDLE'">
          <div class="rcs-track">
            <div 
              v-for="cls in classes" 
              :key="cls.className" 
              class="rcs-item"
              :class="{ 'is-active': currentClassName === cls.className }"
              @click="currentClassName = cls.className; playSound('ui')"
            >
              <span class="rcs-text">{{ cls.className }}</span>
              <div class="rcs-bg"></div>
            </div>
          </div>
        </div>
        <div class="count-selector" v-if="gameState === 'IDLE'">
            <span class="label">作战人数:</span>
            <input type="range" min="1" max="10" v-model.number="selectionCount" @input="handleSelectionCountInput">
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

        <!-- B. 任务队列分镜 (移除白色背景框，改为悬浮堆叠卡片流) -->
        <div class="mission-log-radical" v-if="gameState !== 'SUMMARY'">
          <!-- 悬浮的巨型标题 -->
          <div class="radical-log-title">
            <span class="bg-text">QUEUE</span>
            <span class="fg-text">作战序列</span>
          </div>

          <div class="log-list-radical">
            <div 
              v-for="(mission, index) in missionQueue" 
              :key="index" 
              class="radical-log-card"
              :class="{ 
                'is-active': index === activeIdx, 
                'is-past': index < activeIdx,
                [mission.status]: true 
              }"
              :style="{ '--card-idx': index }"
            >
              <!-- 左侧大序号 -->
              <div class="r-card-idx">{{ String(index + 1).padStart(2, '0') }}</div>
              
              <!-- 核心内容区 -->
              <div class="r-card-body">
                <div class="r-card-name">{{ mission.student }}</div>
                <div class="r-card-mission">
                  <span class="r-poem">《{{ mission.poem }}》</span>
                  <span class="r-author">{{ mission.author }}</span>
                </div>
              </div>
              
              <!-- 爆裂状态印章 -->
              <div class="r-card-stamp" v-if="mission.status !== 'PENDING'">
                {{ mission.status }}
              </div>
              
              <!-- 装饰性胶带/涂鸦 -->
              <div class="r-card-deco"></div>
            </div>
            
            <div v-if="missionQueue.length === 0" class="radical-log-empty">
              WAITING FOR DEPLOYMENT...
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
          <span>撤退！！！</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  width: 100vw; height: 100vh;
  overflow: hidden;
  background-color: #000; /* 深邃的黑底，衬托高对比度 */
  position: relative;
  display: flex; 
  flex-direction: column;
}

/* 1. 动态网点底图：红黑高对比度 */
.app-container::before {
  content: '';
  position: absolute;
  top: -50%; left: -50%; right: -50%; bottom: -50%;
  background-image: radial-gradient(var(--manga-red) 15%, transparent 16%), 
                    radial-gradient(var(--manga-red) 15%, transparent 16%);
  background-size: 30px 30px;
  background-position: 0 0, 15px 15px;
  opacity: 0.15;
  z-index: 0;
  pointer-events: none;
  transform: rotate(-15deg);
  animation: moveDotsFast 15s linear infinite;
}

@keyframes moveDotsFast {
  0% { transform: rotate(-15deg) translate(0, 0); }
  100% { transform: rotate(-15deg) translate(-150px, -150px); }
}

/* 2. 漫画集中线特效 (CSS 手搓放射线) */
.manga-focus-lines {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  background: repeating-conic-gradient(
    from 0deg,
    transparent 0deg 10deg,
    rgba(255, 255, 255, 0.05) 10deg 12deg,
    transparent 12deg 25deg,
    rgba(255, 255, 255, 0.08) 25deg 26deg
  );
  z-index: 0;
  pointer-events: none;
  animation: focusPulse 0.1s infinite alternate;
}

@keyframes focusPulse {
  0% { opacity: 0.8; transform: scale(1); }
  100% { opacity: 1; transform: scale(1.02); }
}

/* 3. 无限滚动的警告封条 */
.hazard-tape {
  position: absolute;
  width: 120vw;
  height: 40px;
  background: var(--manga-yellow);
  color: #000;
  font-weight: 900;
  font-size: 24px;
  line-height: 40px;
  overflow: hidden;
  z-index: 1;
  pointer-events: none;
  box-shadow: 0 5px 15px rgba(0,0,0,0.5);
  border-top: 4px solid #000;
  border-bottom: 4px solid #000;
  white-space: nowrap;
}

.tape-top {
  top: 10%;
  left: -10vw;
  transform: rotate(-5deg);
}

.tape-bottom {
  bottom: 10%;
  left: -10vw;
  transform: rotate(5deg);
  background: var(--manga-red);
  color: #fff;
}

.tape-text {
  display: inline-block;
  animation: scrollTape 10s linear infinite;
  letter-spacing: 4px;
}

.tape-bottom .tape-text {
  animation: scrollTape 12s linear infinite reverse;
}

@keyframes scrollTape {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

/* 4. 二次元拟声词/特效文字 */
.manga-sfx-container {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  pointer-events: none;
  z-index: 2;
  overflow: hidden;
}

.manga-sfx {
  position: absolute;
  font-family: impact, sans-serif;
  font-weight: 900;
  font-style: italic;
  color: transparent;
  -webkit-text-stroke: 2px rgba(255, 255, 255, 0.15);
  filter: drop-shadow(4px 4px 0 rgba(0,0,0,0.5));
}

.sfx-1 { font-size: 15vw; top: 15%; left: 5%; transform: rotate(-15deg); animation: pulseSfx 2s infinite alternate; }
.sfx-2 { font-size: 12vw; bottom: 20%; right: 5%; transform: rotate(10deg); animation: pulseSfx 3s infinite alternate-reverse; }
.sfx-3 { font-size: 20vw; top: 40%; left: 30%; -webkit-text-stroke: 4px rgba(255, 0, 0, 0.05); transform: rotate(-5deg) scale(1.5); z-index: 0; }

@keyframes pulseSfx {
  0% { transform: rotate(-15deg) scale(1); opacity: 0.5; }
  100% { transform: rotate(-15deg) scale(1.1); opacity: 1; }
}

.main-layout {
  flex: 1;
  display: flex; flex-direction: column;
  background: transparent;
  width: 100%;
  height: 100%;
  position: relative;
  z-index: 1;
}

.top-bar { 
  height: 100px; display: flex; justify-content: flex-end; align-items: center; 
  padding: 0 40px; z-index: 10; gap: 30px;
}

/* 激进版班级选择器 */
.radical-class-selector {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
  position: relative;
}

.rcs-track {
  display: flex;
  gap: 10px;
  background: #000;
  padding: 8px 12px;
  border: 4px solid #000;
  transform: skewX(-10deg);
  box-shadow: 6px 6px 0 rgba(0,0,0,0.8), -4px -4px 0 var(--manga-yellow);
  position: relative;
  overflow: hidden;
}

/* 给轨道加个漫画网点背景 */
.rcs-track::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image: radial-gradient(#fff 10%, transparent 11%);
  background-size: 4px 4px;
  opacity: 0.1;
  pointer-events: none;
}

.rcs-item {
  position: relative;
  padding: 8px 20px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.2s cubic-bezier(0.25, 1.5, 0.5, 1);
  border: 3px solid transparent;
}

.rcs-text {
  position: relative;
  z-index: 2;
  font-size: 1.2rem;
  font-weight: 900;
  color: #fff;
  letter-spacing: 1px;
  text-shadow: 2px 2px 0 #000;
  transition: all 0.2s;
}

.rcs-bg {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: #333;
  z-index: 1;
  transform: scaleY(0);
  transform-origin: bottom;
  transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1);
}

/* 悬浮态 */
.rcs-item:hover .rcs-bg {
  transform: scaleY(1);
  background: #555;
}

/* 选中态（燃炸效果） */
.rcs-item.is-active {
  border-color: #000;
  transform: scale(1.1) translateY(-2px);
  z-index: 3;
}

.rcs-item.is-active .rcs-bg {
  transform: scaleY(1);
  background: var(--manga-yellow);
}

.rcs-item.is-active .rcs-text {
  color: #000;
  text-shadow: none;
}

/* 选中时的装饰锯齿 */
.rcs-item.is-active::after {
  content: '';
  position: absolute;
  bottom: -3px; left: 0; right: 0;
  height: 6px;
  background: var(--manga-red);
  z-index: 4;
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
  flex: 1; display: flex; gap: 2vw; padding: 2vh 4vw; z-index: 5;
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
  padding: 4px 20px; font-weight: 900; font-size: 18px; z-index: 10;
}

.log-tag { left: auto; right: 30px; background: var(--manga-red); color: #fff; }

/* 左侧主显示区 */
.display-main {
  flex: 1 1 auto; 
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  position: relative;
  min-width: 0;
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
    width: 100%; 
    flex: 1;
    display: flex; flex-direction: column; align-items: center; justify-content: center;
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

/* === 极致燃向：右侧任务队列 (Mission Log) === */
.mission-log-radical {
  width: 35vw;
  max-width: 500px;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 10;
  perspective: 1000px; /* 增加 3D 景深 */
}

/* 巨型错位标题 */
.radical-log-title {
  position: relative;
  margin-bottom: 30px;
  height: 80px;
}

.radical-log-title .bg-text {
  position: absolute;
  top: -10px; left: -10px;
  font-size: 80px;
  font-weight: 900;
  font-style: italic;
  color: transparent;
  -webkit-text-stroke: 2px rgba(255, 255, 255, 0.1);
  letter-spacing: 5px;
  z-index: 0;
}

.radical-log-title .fg-text {
  position: absolute;
  bottom: 0; left: 20px;
  font-size: 32px;
  font-weight: 900;
  color: var(--manga-yellow);
  background: #000;
  padding: 5px 15px;
  transform: skewX(-15deg);
  border: 3px solid #fff;
  box-shadow: 4px 4px 0 var(--manga-red);
  z-index: 1;
}

/* 列表容器 */
.log-list-radical {
  flex: 1;
  overflow-y: auto;
  padding: 10px 20px 50px 10px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  /* 隐藏滚动条但保留功能 */
  scrollbar-width: none;
}
.log-list-radical::-webkit-scrollbar { display: none; }

/* 待处刑卡片本体 */
.radical-log-card {
  position: relative;
  background: #fff;
  border: 4px solid #000;
  display: flex;
  align-items: stretch;
  transform: skewX(-5deg) rotate(calc(var(--card-idx) * -1deg + 0.5deg)); /* 错落排版 */
  box-shadow: 8px 8px 0 rgba(0,0,0,0.8);
  transition: all 0.3s cubic-bezier(0.25, 1.5, 0.5, 1);
  overflow: hidden;
  opacity: 0.9;
}

/* 状态：已过去 (置灰、缩小) */
.radical-log-card.is-past {
  opacity: 0.5;
  filter: grayscale(100%);
  transform: scale(0.95) translateX(-10px) skewX(-5deg);
  box-shadow: 4px 4px 0 rgba(0,0,0,0.5);
}

/* 状态：正在处刑 (极致放大、高亮、爆破感) */
.radical-log-card.is-active {
  opacity: 1;
  transform: scale(1.1) translateX(-15px) skewX(-5deg) rotate(0deg);
  box-shadow: 
    15px 15px 0 #000,
    -5px -5px 0 var(--manga-yellow),
    10px -5px 0 var(--manga-red);
  border-color: var(--manga-red);
  border-width: 6px;
  z-index: 100;
  animation: cardShake 0.5s infinite alternate;
}

@keyframes cardShake {
  0% { transform: scale(1.1) translateX(-15px) skewX(-5deg) rotate(0deg) translateY(0); }
  100% { transform: scale(1.1) translateX(-15px) skewX(-5deg) rotate(0deg) translateY(-3px); }
}

/* 卡片左侧：大序号 */
.r-card-idx {
  background: #000;
  color: var(--manga-yellow);
  font-size: 32px;
  font-weight: 900;
  font-style: italic;
  padding: 10px 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-right: 4px solid #000;
}
.is-active .r-card-idx {
  background: var(--manga-red);
  color: #fff;
}

/* 卡片核心内容 */
.r-card-body {
  flex: 1;
  padding: 10px 15px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  background-image: radial-gradient(rgba(0,0,0,0.05) 20%, transparent 20%);
  background-size: 4px 4px;
}

.r-card-name {
  font-size: 24px;
  font-weight: 900;
  color: #000;
  margin-bottom: 5px;
  letter-spacing: 2px;
}

.r-card-mission {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.r-poem {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}
.r-author {
  font-size: 12px;
  color: #666;
  background: #eee;
  padding: 2px 6px;
  border: 1px solid #000;
}

/* 爆裂状态印章 (PASS / FAIL) */
.r-card-stamp {
  position: absolute;
  right: -10px;
  top: 50%;
  transform: translateY(-50%) rotate(-15deg);
  font-size: 40px;
  font-weight: 900;
  font-style: italic;
  text-transform: uppercase;
  padding: 0 10px;
  border: 6px solid currentColor;
  mix-blend-mode: multiply;
  opacity: 0.8;
  pointer-events: none;
}

.radical-log-card.PASS .r-card-stamp {
  color: #00AA00;
  transform: translateY(-50%) rotate(-15deg) scale(1.2);
}

.radical-log-card.FAIL .r-card-stamp {
  color: #DD0000;
  transform: translateY(-50%) rotate(10deg) scale(1.5);
  border-style: dashed;
}

/* 装饰性黄色胶带 */
.r-card-deco {
  position: absolute;
  bottom: -10px; right: 20px;
  width: 60px; height: 15px;
  background: var(--manga-yellow);
  transform: rotate(-25deg);
  mix-blend-mode: multiply;
  z-index: 5;
  opacity: 0.8;
}

/* 空状态 */
.radical-log-empty {
  color: rgba(255, 255, 255, 0.3);
  font-size: 24px;
  font-weight: 900;
  font-style: italic;
  text-align: center;
  margin-top: 50px;
  transform: skewX(-10deg);
}

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

@media (max-width: 1024px) {
  .battle-arena {
    flex-direction: column;
    overflow-y: auto;
  }
  
  .display-main {
    flex: 0 0 auto;
    min-height: 50vh;
  }

  .mission-log {
    flex: 1 1 auto;
    width: 100%;
  }

  .arena-divider {
    flex-direction: row;
    width: 100%;
    height: 40px;
  }

  .divider-line {
    height: 6px;
    width: 100%;
    background: repeating-linear-gradient(
      90deg,
      #000,
      #000 15px,
      transparent 15px,
      transparent 30px
    );
  }

  .divider-line::after {
    top: 10px;
    left: 0;
    width: 100%;
    height: 2px;
  }

  .divider-text {
    writing-mode: horizontal-tb;
    padding: 5px 20px;
    letter-spacing: 5px;
  }
}
</style>
