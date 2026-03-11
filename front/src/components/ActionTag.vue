<template>
  <div class="manga-tag" :class="{ 'is-add': isAdd }" @click="$emit('click')">
    {{ text }}
    <div v-if="!isAdd" class="delete-mark" @click.stop="$emit('delete')">x</div>
  </div>
</template>

<script setup>
defineProps({
  text: {
    type: String,
    required: true
  },
  isAdd: {
    type: Boolean,
    default: false // 如果是 true，就会变成虚线的“添加按钮”
  }
});
defineEmits(['delete', 'click']);
</script>

<style scoped>
.manga-tag {
  background: var(--manga-white);
  border: 3px solid var(--manga-black);
  padding: 6px 14px;
  font-size: 18px;
  font-weight: bold;
  position: relative;
  cursor: pointer;
  box-shadow: 3px 3px 0 var(--manga-black);
  transition: all 0.2s ease;
  display: inline-block;
}

.manga-tag:hover {
  transform: translate(-2px, -2px) scale(1.05);
  box-shadow: 5px 5px 0 var(--manga-black);
  background: var(--manga-yellow);
}

/* 添加按钮的特殊样式（虚线） */
.manga-tag.is-add {
  border-style: dashed;
  background: transparent;
  box-shadow: none;
}
.manga-tag.is-add:hover {
  background: rgba(255, 215, 0, 0.3); /* 半透明黄 */
  transform: scale(1.05);
}

/* 爆炸风删除按钮 */
.delete-mark {
  position: absolute;
  top: -12px;
  right: -12px;
  background: var(--manga-red);
  color: white;
  width: 24px;
  height: 24px;
  border: 2px solid var(--manga-black);
  font-size: 14px;
  text-align: center;
  line-height: 20px;
  clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%);
  display: none;
  z-index: 10;
}

.manga-tag:hover .delete-mark {
  display: block;
  animation: popIn 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes popIn {
  0% { transform: scale(0) rotate(-45deg); }
  100% { transform: scale(1) rotate(0deg); }
}
</style>