<template>
  <div class="manga-panel">
    <!-- 1. 固定头部 -->
    <div class="panel-header">
      <span>⚙️ SYSTEM CONFIG / 作战统筹</span>
      <span class="close-btn" @click="$emit('close')">[X] 退出</span>
    </div>

    <!-- 2. 核心滚动区 -->
    <div class="panel-scroll-area">
      <div class="panel-body">
        <!-- 左侧分镜：学生 -->
        <div class="comic-frame">
          <div class="frame-title">A. 目标锁定 (Targets)</div>
          <div class="tag-list">
            <ActionTag 
              v-for="(student, index) in students" 
              :key="'s-'+index" 
              :text="student" 
              @delete="removeStudent(index)" 
            />
            <ActionTag text="+ 增加目标" :isAdd="true" @click="openStudentDialog" />
          </div>
        </div>

        <!-- 右侧分镜：篇目 -->
        <div class="comic-frame">
          <div class="frame-title" style="background: var(--manga-blue);">B. 弹药装填 (Arsenal)</div>
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
import { ref, onMounted, reactive } from 'vue';
import ActionTag from './ActionTag.vue';
import MangaDialog from './MangaDialog.vue';

interface Poem { title: string; author: string; }
const emit = defineEmits<{ (e: 'close'): void; (e: 'saved'): void; }>();

const students = ref<string[]>([]);
const poems = ref<Poem[]>([]);
const isSaving = ref(false);

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

const fetchConfig = async () => {
  try {
    const res = await fetch('/api/config');
    const json = await res.json();
    if (json.code === 200) {
      students.value = json.data.students || [];
      poems.value = json.data.poems || [];
    }
  } catch (e) { console.error(e); }
};

const saveConfig = async () => {
  isSaving.value = true;
  try {
    const res = await fetch('/api/config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ students: students.value, poems: poems.value })
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

onMounted(() => fetchConfig());

const openStudentDialog = () => {
  summonDialog('input', '🎯 录入新目标', '输入学生姓名...', (n) => n && students.value.push(n));
};
const removeStudent = (i: number) => students.value.splice(i, 1);

const openPoemDialog = () => {
  summonDialog('input', '📜 填装篇目', '格式: 《篇目》-作者 (例: 《短歌行》-曹操)', (p) => {
    if (!p) return;
    const parts = p.split(/[-—]/);
    if (parts.length >= 2) {
      const title = parts[0].replace(/[《》]/g, '').trim();
      const author = parts[1].trim();
      poems.value.push({ title, author });
    } else {
      // 如果没按格式填，至少把输入的当标题
      poems.value.push({ title: p.trim(), author: '佚名' });
    }
  });
};
const removePoem = (i: number) => poems.value.splice(i, 1);
</script>

<style scoped>
/* 样式保持不变，但移除了 book-group 相关 */
.manga-panel {
  background: var(--manga-white);
  border: 5px solid var(--manga-black);
  box-shadow: 12px 12px 0px var(--manga-black);
  width: 950px;
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
  z-index: 10;
  clip-path: polygon(0 0, 100% 0, 99% 100%, 1% 100%);
}

.close-btn { cursor: pointer; transition: color 0.2s; }
.close-btn:hover { color: var(--manga-red); }

.panel-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
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
  padding: 40px 15px 15px;
  position: relative;
  background: #fff;
  min-height: 200px;
}

.frame-title {
  position: absolute;
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
