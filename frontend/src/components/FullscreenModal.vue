<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="handleClose">
    <div class="modal-container">
      <div class="modal-content">
        <p class="modal-text">{{ text }}</p>
        <button class="modal-button" @click="handleClose">OK</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref } from 'vue'

const isOpen = ref(false)
const text = ref('')

function open(modalText: string) {
  text.value = modalText
  isOpen.value = true
  document.body.style.overflow = 'hidden'
}

function close() {
  isOpen.value = false
  document.body.style.overflow = ''
}

function handleClose() {
  close()
}

defineExpose({
  open,
  close,
  handleClose
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-container {
  width: 90%;
  max-width: 600px;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  padding: 24px;
  animation: modal-fade-in 0.2s ease-out;
}

.modal-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.modal-text {
  margin: 0;
  font-size: 16px;
  line-height: 1.5;
  color: #333;
}

.modal-button {
  align-self: flex-end;
  padding: 8px 16px;
  background-color: #2563eb;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.modal-button:hover {
  background-color: #1d4ed8;
}

@keyframes modal-fade-in {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
