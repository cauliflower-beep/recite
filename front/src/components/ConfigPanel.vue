<template>
  <div class="manga-panel">
    <!-- 1. 固定头部 -->
    <div class="panel-header">
      <span>⚙️ SYSTEM CONFIG / 作战统筹</span>
      <span class="close-btn" @click="$emit('close')">[X] 退出</span>
    </div>

    <!-- 2. 核心区：两栏布局 -->
    <div class="panel-body">
      <!-- A. 篇目分镜 (Arsenal) -->
      <div class="comic-frame">
        <div class="frame-title" style="background: var(--manga-blue);">A. 弹药装填 (Arsenal)</div>
        
        <!-- 导航箭头 -->
        <button 
          class="nav-btn left" 
          :disabled="poemScroll.isAtStart" 
          @click="scrollList('poem', -200)"
        > ◀ </button>
        <button 
          class="nav-btn right" 
          :disabled="poemScroll.isAtEnd" 
          @click="scrollList('poem', 200)"
        > ▶ </button>

        <div class="scroll-viewport" ref="poemRef" @scroll="updateScrollState('poem')">
          <div class="tag-list">
            <ActionTag 
              v-for="(poem, index) in poems" 
              :key="'p-'+index" 
              :text="`《${poem.title}》- ${poem.author}`" 
              @delete="removePoem(index)" 
            />
            <ActionTag text="+ 增加篇目" :isAdd="true" @click="openPoemDialog" />
          </div>
        </div>
      </div>

      <!-- B. 学生分镜 (Targets) -->
      <div class="comic-frame">
        <div class="frame-title" style="display: flex; flex-direction: column; gap: 10px;">
          <span>B. 目标锁定 (Targets)</span>
          <!-- 班级 Tag 列表 -->
          <div class="class-tags" style="display: flex; gap: 10px; flex-wrap: wrap; margin-top: 5px; pointer-events: auto;">
            <div 
              v-for="(cls, index) in classes" 
              :key="cls.className" 
              class="class-tag" 
              :class="{ active: currentClassName === cls.className }"
              @click.stop="currentClassName = cls.className; handleClassChange()"
            >
              {{ cls.className }}
              <span class="delete-class-btn" @click.stop="removeClass(index)" v-if="classes.length > 1">×</span>
            </div>
            <div class="class-tag add-class-btn" @click="openClassDialog">+ 添加班级</div>
          </div>
        </div>
        
        <!-- 导航箭头 -->
        <button 
          class="nav-btn left" 
          :disabled="studentScroll.isAtStart" 
          @click="scrollList('student', -200)"
        > ◀ </button>
        <button 
          class="nav-btn right" 
          :disabled="studentScroll.isAtEnd" 
          @click="scrollList('student', 200)"
        > ▶ </button>

        <div class="scroll-viewport" ref="studentRef" @scroll="updateScrollState('student')">
          <div class="tag-list">
            <ActionTag 
              v-for="(student, index) in currentStudents" 
              :key="'s-'+index" 
              :text="student" 
              @delete="removeStudent(index)" 
            />
            <ActionTag text="+ 增加目标" :isAdd="true" @click="openStudentDialog" />
          </div>
        </div>
      </div>
    </div>

    <!-- 3. 固定底部 -->
    <div class="panel-footer">
      <button class="btn-save" @click="saveConfig" :disabled="isSaving">
        {{ isSaving ? '⏳ 正在写入神经元...' : '💥 确认覆写 (SAVE)' }}
      </button>
    </div>

    <!-- 漫画风弹窗 -->
    <MangaDialog
      v-model:visible="dialog.show"
      :type="dialog.type"
      :title="dialog.title"
      :placeholder="dialog.placeholder"
      @confirm="handleDialogConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, nextTick, computed } from 'vue';
import ActionTag from './ActionTag.vue';
import MangaDialog from './MangaDialog.vue';

interface Poem { title: string; author: string; }
interface ClassStudents { className: string; members: string[]; }

const emit = defineEmits<{ (e: 'close'): void; (e: 'saved'): void; }>();

const classes = ref<ClassStudents[]>([]);
const currentClassName = ref<string>('');
const poems = ref<Poem[]>([]);
const isSaving = ref(false);

const currentStudents = computed(() => {
  const cls = classes.value.find(c => c.className === currentClassName.value);
  return cls ? cls.members : [];
});

const handleClassChange = () => {
  nextTick(() => updateScrollState('student'));
};

// 滚动控制逻辑
const studentRef = ref<HTMLElement | null>(null);
const poemRef = ref<HTMLElement | null>(null);

const studentScroll = reactive({ isAtStart: true, isAtEnd: false });
const poemScroll = reactive({ isAtStart: true, isAtEnd: false });

const updateScrollState = (type: 'student' | 'poem') => {
  const el = type === 'student' ? studentRef.value : poemRef.value;
  const state = type === 'student' ? studentScroll : poemScroll;
  if (!el) return;

  state.isAtStart = el.scrollLeft <= 10;
  state.isAtEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 10;
};

const scrollList = (type: 'student' | 'poem', offset: number) => {
  const el = type === 'student' ? studentRef.value : poemRef.value;
  if (el) {
    el.scrollBy({ left: offset, behavior: 'smooth' });
  }
};

const dialog = reactive({
  show: false,
  type: 'input' as 'input' | 'confirm' | 'info',
  title: '',
  placeholder: '',
});

let pendingAction: ((val?: string) => void) | null = null;

const summonDialog = (type: 'input'|'confirm'|'info', title: string, placeholder: string, action: (val?: string)=>void) => {
  dialog.type = type;
  dialog.title = title;
  dialog.placeholder = placeholder;
  pendingAction = action;
  dialog.show = true;
};

const handleDialogConfirm = (payload?: string) => {
  if (pendingAction) {
    pendingAction(payload);
    pendingAction = null;
    nextTick(() => {
      updateScrollState('student');
      updateScrollState('poem');
    });
  }
};

const fetchConfig = async () => {
  try {
    const res = await fetch('/api/config');
    const json = await res.json();
    if (json.code === 200) {
      classes.value = json.data.classes || [];
      if (classes.value.length > 0) {
        currentClassName.value = classes.value[0]?.className || '';
      }
      poems.value = json.data.poems || [];
      nextTick(() => {
        updateScrollState('student');
        updateScrollState('poem');
      });
    }
  } catch (e) { console.error(e); }
};

const saveConfig = async () => {
  isSaving.value = true;
  try {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ classes: classes.value, poems: poems.value })
    });
    const json = await res.json();
    if (json.code === 200) {
      summonDialog('info', '🛰️ 同步成功', '作战数据库覆写完成！', () => {
        emit('saved');
        emit('close');
      });
    }
  } catch (e) { 
    summonDialog('info', '💥 错误', '连接基地服务器失败！', ()=>{});
  } finally { isSaving.value = false; }
};

onMounted(() => {
  fetchConfig();
});

const openStudentDialog = () => {
  summonDialog('input', '🎯 录入新目标', '输入学生姓名...', (n) => {
    if (n) {
      const cls = classes.value.find(c => c.className === currentClassName.value);
      if (cls) cls.members.push(n);
    }
  });
};
const removeStudent = (i: number) => {
  const cls = classes.value.find(c => c.className === currentClassName.value);
  if (cls) {
    cls.members.splice(i, 1);
    nextTick(() => updateScrollState('student'));
  }
};

const openClassDialog = () => {
  summonDialog('input', '🏫 新建班级', '输入班级名称 (如: 12班)...', (n) => {
    if (n && !classes.value.some(c => c.className === n)) {
      classes.value.push({ className: n, members: [] });
      currentClassName.value = n; // 自动选中新班级
      nextTick(() => updateScrollState('student'));
    } else if (classes.value.some(c => c.className === n)) {
      summonDialog('info', '⚠️ 提示', '班级名称已存在！', ()=>{});
    }
  });
};

const removeClass = (index: number) => {
  const cls = classes.value[index];
  if (!cls) return;
  summonDialog('confirm', '⚠️ 警告', `确定要删除【${cls.className}】及其所有学生数据吗？`, () => {
    const deletedClass = classes.value[index]?.className || '';
    classes.value.splice(index, 1);
    // 如果删除的是当前选中的班级，自动选中第一个
    if (currentClassName.value === deletedClass && classes.value.length > 0) {
      currentClassName.value = classes.value[0]?.className || '';
    }
    nextTick(() => updateScrollState('student'));
  });
};

const openPoemDialog = () => {
  summonDialog('input', '📜 填装篇目', '格式: 《篇目》-作者 (例: 《短歌行》-曹操)', (p) => {
    if (!p) return;
    const parts = p.split(/[-—]/);
    if (parts.length >= 2 && parts[0] !== undefined && parts[1] !== undefined) {
      const title = parts[0].replace(/[《》]/g, '').trim();
      const author = parts[1].trim();
      poems.value.push({ title, author });
    } else {
      poems.value.push({ title: p.trim(), author: '佚名' });
    }
    nextTick(() => updateScrollState('poem'));
  });
};
const removePoem = (i: number) => {
  poems.value.splice(i, 1);
  nextTick(() => updateScrollState('poem'));
};
</script>

<style scoped>
.manga-panel {
  background: var(--manga-white);
  border: 5px solid var(--manga-black);
  box-shadow: 12px 12px 0px var(--manga-black);
  width: 1000px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  transform: rotate(-0.5deg);
  position: relative;
  animation: panelDrop 0.5s cubic-bezier(0.25, 1, 0.5, 1);
}

@keyframes panelDrop {
  0% { transform: translateY(-50px) rotate(5deg); opacity: 0; }
  100% { transform: translateY(0) rotate(-0.5deg); opacity: 1; }
}

.panel-header {
  flex-shrink: 0;
  background: var(--manga-black);
  color: var(--manga-yellow);
  padding: 15px 20px;
  font-size: 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  z-index: 100;
  clip-path: polygon(0 0, 100% 0, 99% 100%, 1% 100%);
}

.close-btn { cursor: pointer; transition: color 0.2s; }
.close-btn:hover { color: var(--manga-red); }

.panel-body { 
  display: flex; 
  gap: 20px; 
  padding: 30px;
  height: 520px; 
}

.comic-frame {
  flex: 1;
  border: 4px solid var(--manga-black);
  position: relative;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden; 
  box-shadow: 5px 5px 0 rgba(0,0,0,0.1);
}

.scroll-viewport {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 120px 20px 30px 40px; /* 增加上边距，避免被变高的 frame-title 遮挡 */
  scrollbar-width: none; 
}
.scroll-viewport::-webkit-scrollbar { display: none; }

/* 增加末尾缓冲，防止最后一列贴边 */
.tag-list::after {
  content: "";
  display: block;
  width: 60px; /* 战术留白空间 */
  height: 100%;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 45px;
  height: 45px;
  background: var(--manga-yellow);
  border: 3px solid var(--manga-black);
  color: var(--manga-black);
  font-size: 24px;
  font-weight: 900;
  cursor: pointer;
  z-index: 50;
  box-shadow: 4px 4px 0 var(--manga-black);
  transition: all 0.1s;
  display: flex;
  justify-content: center;
  align-items: center;
}
.nav-btn:active { transform: translateY(-50%) translate(2px, 2px); box-shadow: 2px 2px 0 var(--manga-black); }
.nav-btn:disabled { background: #ccc; cursor: not-allowed; opacity: 0.3; box-shadow: none; }
.nav-btn.left { left: 5px; }
.nav-btn.right { right: 5px; }

.frame-title {
  position: absolute;
  top: 15px; left: 15px;
  background: var(--manga-yellow);
  border: 3px solid var(--manga-black);
  padding: 5px 15px;
  font-size: 18px;
  font-weight: 900;
  box-shadow: 4px 4px 0 var(--manga-black);
  z-index: 60;
  pointer-events: auto;
}

.class-tags {
  margin-top: 10px;
}

.class-tag {
  background: #333;
  color: #fff;
  padding: 5px 12px;
  border: 2px solid #000;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s;
  box-shadow: 2px 2px 0 #000;
  pointer-events: auto;
}

.class-tag:hover {
  transform: translateY(-2px);
  box-shadow: 4px 4px 0 #000;
}

.class-tag.active {
  background: var(--manga-yellow);
  color: #000;
  font-weight: bold;
  border-color: #000;
}

.delete-class-btn {
  font-weight: bold;
  cursor: pointer;
  padding: 0 4px;
  border-radius: 50%;
}

.delete-class-btn:hover {
  background: rgba(255, 0, 0, 0.2);
  color: red;
}

.add-class-btn {
  background: #fff;
  color: #000;
  border-style: dashed;
}

.add-class-btn:hover {
  background: #f0f0f0;
}

.tag-list { 
  display: flex; 
  flex-flow: column wrap; 
  gap: 12px; 
  height: 100%; 
  align-content: flex-start;
}

.panel-footer {
  flex-shrink: 0;
  padding: 15px 30px;
  text-align: right;
  border-top: 5px solid var(--manga-black);
  background: var(--manga-white);
}

.btn-save {
  background: var(--manga-blue);
  border: 4px solid var(--manga-black);
  box-shadow: 6px 6px 0 var(--manga-black);
  padding: 10px 40px;
  font-size: 24px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.1s;
}
.btn-save:active:not(:disabled) { transform: translate(3px, 3px); box-shadow: 3px 3px 0 var(--manga-black); }
.btn-save:disabled { background: #ccc; cursor: not-allowed; opacity: 0.7; }
</style>