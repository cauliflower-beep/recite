<template>
  <div class="manga-panel">
    <!-- 1. 固定头部：绝对不会移动 -->
    <div class="panel-header">
      <span>⚙️ SYSTEM CONFIG / 作战统筹</span>
      <span class="close-btn" @click="$emit('close')">[X] 退出</span>
    </div>

    <!-- 2. 核心滚动区：内容再多也不怕 -->
    <div class="panel-scroll-area">
      <div class="panel-body">
        <!-- 左侧分镜：学生 -->
        <div class="comic-frame">
          <div class="frame-title">A. 目标锁定 (Targets)</div>
          <div class="tag-list">
            <ActionTag 
              v-for="(student, index) in students" 
              :key="index" 
              :text="student" 
              @delete="removeStudent(index)" 
            />
            <ActionTag text="+ 增加目标" :isAdd="true" @click="openStudentDialog" />
          </div>
        </div>

        <!-- 右侧分镜：教材篇目 -->
        <div class="comic-frame">
          <div class="frame-title" style="background: var(--manga-blue);">B. 弹药装填 (Arsenal)</div>
          <div class="book-group" v-for="(book, bIndex) in arsenal" :key="bIndex">
            <div class="book-header">
              <div class="book-title">《{{ book.title }}》</div>
              <button class="delete-book-btn" @click="openDeleteBookDialog(bIndex)">🗑️ 销毁</button>
            </div>
            <div class="tag-list" style="margin-top: 10px;">
              <ActionTag 
                v-for="(poem, pIndex) in book.poems" 
                :key="pIndex" 
                :text="poem" 
                @delete="removePoem(bIndex, pIndex)" 
              />
              <ActionTag text="+ 增加篇目" :isAdd="true" @click="openPoemDialog(bIndex)" />
            </div>
          </div>
          <!-- 增加教材 -->
          <div style="margin-top: 30px; text-align: center;">
            <ActionTag 
              text="+ 录入新教材卷宗" 
              :isAdd="true" 
              style="width: 80%; display: block; margin: 0 auto;" 
              @click="openBookDialog" 
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 3. 固定底部：保存按钮永远在手边 -->
    <div class="panel-footer">
      <button class="btn-save" @click="saveConfig" :disabled="isSaving">
        {{ isSaving ? '⏳ 正在写入神经元...' : '💥 确认覆写 (SAVE)' }}
      </button>
    </div>

    <!-- 漫画风弹窗 (录入/确认/信息) -->
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
import { ref, onMounted, reactive } from 'vue';
import ActionTag from './ActionTag.vue';
import MangaDialog from './MangaDialog.vue';

interface Book { title: string; poems: string[]; }
const emit = defineEmits<{ (e: 'close'): void; (e: 'saved'): void; }>();

const students = ref<string[]>([]);
const arsenal = ref<Book[]>([]);
const isSaving = ref(false);

// --- 弹窗逻辑控制 ---
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
  }
};

// --- API 交互 ---
const fetchConfig = async () => {
  try {
    const res = await fetch('/api/config');
    const json = await res.json();
    if (json.code === 200) {
      students.value = json.data.students || [];
      arsenal.value = json.data.arsenal || [];
    }
  } catch (e) { console.error(e); }
};

const saveConfig = async () => {
  isSaving.value = true;
  try {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ students: students.value, arsenal: arsenal.value })
    });
    const json = await res.json();
    if (json.code === 200) {
      summonDialog('info', '🛰️ 同步成功', '作战数据库覆写完成！所有情报已安全存入核心服务器。', () => {
        emit('saved');
        emit('close');
      });
    }
  } catch (e) { 
    summonDialog('info', '💥 错误', '连接基地服务器失败！', ()=>{});
  } finally { isSaving.value = false; }
};

onMounted(() => fetchConfig());

// --- 增删逻辑 ---
const openStudentDialog = () => {
  summonDialog('input', '🎯 录入新目标', '输入学生姓名...', (n) => n && students.value.push(n));
};
const removeStudent = (i: number) => students.value.splice(i, 1);

const openPoemDialog = (bi: number) => {
  const book = arsenal.value[bi];
  if (book) {
    summonDialog('input', `📜 填装篇目`, '输入古诗篇目...', (p) => p && book.poems.push(p));
  }
};
const removePoem = (bi: number, pi: number) => {
  const book = arsenal.value[bi];
  if (book && book.poems) {
    book.poems.splice(pi, 1);
  }
};

const openBookDialog = () => {
  summonDialog('input', '📚 创建新卷宗', '输入教材名称...', (t) => t && arsenal.value.push({ title: t, poems: [] }));
};
const openDeleteBookDialog = (i: number) => {
  const book = arsenal.value[i];
  if (book) {
    summonDialog('confirm', `🔥 确认抹杀?`, `即将销毁《${book.title}》`, () => arsenal.value.splice(i, 1));
  }
};
</script>

<style scoped>
/* --- 布局核心：固定头尾，中间滚动 --- */
.manga-panel {
  background: var(--manga-white);
  border: 5px solid var(--manga-black);
  box-shadow: 12px 12px 0px var(--manga-black);
  width: 950px;
  max-height: 85vh; /* 限制面板总高度 */
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
  z-index: 10;
  clip-path: polygon(0 0, 100% 0, 99% 100%, 1% 100%);
}

.close-btn { cursor: pointer; transition: color 0.2s; }
.close-btn:hover { color: var(--manga-red); }

/* --- 滚动区域美化 --- */
.panel-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  /* 粗黑漫画风滚动条 */
  scrollbar-width: thick;
  scrollbar-color: var(--manga-black) #eee;
}
.panel-scroll-area::-webkit-scrollbar { width: 12px; }
.panel-scroll-area::-webkit-scrollbar-track { background: #eee; border-left: 3px solid var(--manga-black); }
.panel-scroll-area::-webkit-scrollbar-thumb { background-color: var(--manga-black); border: 2px solid #eee; }

.panel-body { display: flex; gap: 20px; align-items: flex-start; }

.comic-frame {
  flex: 1;
  border: 3px dashed var(--manga-black);
  padding: 40px 15px 15px; /* 增加顶部 padding 给 sticky 标题留位 */
  position: relative;
  background: #fff;
  min-height: 200px;
}

.frame-title {
  position: absolute; /* 这里用 absolute 配合滚动区就足够了 */
  top: 10px; left: 10px;
  background: var(--manga-yellow);
  border: 3px solid var(--manga-black);
  padding: 5px 15px;
  font-size: 18px;
  font-weight: bold;
  box-shadow: 4px 4px 0 var(--manga-black);
  z-index: 5;
}

.tag-list { display: flex; flex-wrap: wrap; gap: 10px; }

.book-group {
  margin-top: 30px;
  border-left: 5px solid var(--manga-black);
  padding-left: 15px;
}
.book-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.book-title { font-size: 20px; background: var(--manga-black); color: white; padding: 2px 8px; transform: skewX(-10deg); }
.delete-book-btn { background: transparent; border: none; color: var(--manga-red); cursor: pointer; text-decoration: underline; font-family: inherit; }

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