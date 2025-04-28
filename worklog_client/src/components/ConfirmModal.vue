<script setup lang="ts">
import { ref } from 'vue'

interface IProps {
  confirmText: string
  confirmButtonText: string
  handleConfirm: () => void
  loading: boolean
}

const { confirmButtonText, confirmText, loading } = defineProps<IProps>()
const modalRef = ref<HTMLDialogElement | null>(null)

defineExpose({ modalRef })
function handleClose() {
  if (loading) {
    return
  }

  modalRef.value?.close()
}
</script>

<template>
  <dialog ref="modalRef" class="dialog">
    <div class="modal-header">
      <span :class="['close-button', loading && 'disabled-icon-button']" @click="handleClose"
        >&times;</span
      >
    </div>
    <div class="modal-body">
      <p>{{ confirmText }}</p>
    </div>
    <div class="modal-action">
      <button
        @click="handleConfirm"
        :disabled="loading === true"
        :class="['confirm-button', 'primary-button', loading && 'disabled-button']"
      >
        {{ loading ? 'Loading' : confirmButtonText }}
      </button>
    </div>
  </dialog>
</template>

<style scoped>
.dialog {
  outline: none;
  border: 0;
  border-radius: 0.25rem;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 0.1),
    0 2px 4px -2px rgb(0 0 0 / 0.1);
}

.modal-header {
  border-bottom: 1px solid rgb(216, 216, 216);
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding-inline: 0.5rem;
}

.close-button {
  font-size: 1.5rem;
  cursor: pointer;
  display: block;
  color: rgb(120, 120, 120);
}

.close-button:hover {
  color: rgb(62, 62, 62);
}

.modal-body {
  padding: 2rem 1rem;
  font-weight: 500;
  letter-spacing: 0.05rem;
}

.modal-action {
  border-top: 1px solid rgb(216, 216, 216);
  padding: 0.5rem;
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.confirm-button {
  font-size: small;
}
</style>
