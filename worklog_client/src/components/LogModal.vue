<script setup lang="ts">
import { DateTime } from 'luxon'
import { ref, defineEmits } from 'vue'

interface IProps {}

interface IEmits {
  (e: 'closeModal'): void
}

type TaskTypes = 'task' | 'bug' | 'story'
type TaskProgress = 'backlog' | 'pending' | 'progress' | 'pr' | 'staging'

const {} = defineProps<IProps>()
const modalRef = ref<HTMLDialogElement | null>(null)
const emit = defineEmits<IEmits>()
const taskName = ref<string>('')
const taskType = ref<TaskTypes | null>(null)
const taskStatus = ref<TaskProgress | null>(null)
const taskPriority = ref<number | null>(null)
const startedAt = ref<string>(DateTime.now().toFormat("yyyy-MM-dd'T'HH:mm"))
const completedAt = ref<string>('')
defineExpose({ modalRef })

function handleClose() {
  emit('closeModal')
}

function handleSubmit() {
  console.log('submit')
}
</script>

<template>
  <dialog ref="modalRef" class="dialog">
    <div class="modal-header">
      <p>Create new log</p>
      <span class="close-button" @click="handleClose">&times;</span>
    </div>

    <form @submit.prevent="handleSubmit" class="log-form">
      <div class="group-wrapper">
        <!-- task name -->
        <div class="input-group">
          <label for="taskName">Task Name</label>
          <input
            type="text"
            placeholder="Enter task name"
            id="taskName"
            v-model="taskName"
            required
          />
        </div>

        <!-- task type -->
        <div class="input-group">
          <label for="taskType">Task Type</label>
          <select
            name="taskType"
            id="taskType"
            placeholder="Select task type"
            v-model="taskType"
            required
          >
            <option :value="null" disabled>--- select ---</option>
            <option value="task">Task</option>
            <option value="story">Story</option>
            <option value="bug">Bug</option>
          </select>
        </div>
      </div>

      <div class="group-wrapper">
        <!-- task status -->
        <div class="input-group">
          <label for="taskStatus">Task Status</label>
          <select
            name="taskStatus"
            id="taskStatus"
            placeholder="Select task progress"
            v-model="taskStatus"
            required
          >
            <option :value="null" disabled>--- select ---</option>
            <option value="backlog">Backlog</option>
            <option value="pending">Pending</option>
            <option value="progress">Progress</option>
            <option value="pr">PR</option>
            <option value="staging">Staging</option>
          </select>
        </div>

        <!-- task priority -->
        <div class="input-group">
          <label for="priority">Task Priority</label>
          <select
            name="priority"
            id="priority"
            placeholder="Select task priotiry"
            v-model="taskPriority"
            required
          >
            <option :value="null" disabled>--- select ---</option>
            <option :value="1">Low</option>
            <option :value="5">Medium</option>
            <option :value="7">High</option>
            <option :value="10">Highest</option>
          </select>
        </div>
      </div>

      <div class="group-wrapper">
        <!-- start date -->
        <div class="input-group">
          <label for="startedAt">Start Date</label>
          <input
            type="datetime-local"
            name="startedAt"
            id="startedAt"
            v-model="startedAt"
            required
          />
        </div>

        <!-- completed at date -->
        <div class="input-group">
          {{ completedAt }}
          <label for="completedAt">End On</label>
          <input type="datetime-local" name="completedAt" id="completedAt" v-model="completedAt" />
        </div>
      </div>

      <div class="input-group notes">
        <label for="notes">Notes</label>
        <textarea name="notes" id="notes" rows="3"></textarea>
      </div>

      <div class="form-action">
        <button type="submit" class="primary-button">SAVE</button>
      </div>
    </form>
  </dialog>
</template>

<style>
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
  justify-content: space-between;
  align-items: center;
  padding-inline: 0.75rem;
  position: sticky;
  top: 0;
  background-color: white;
}

.modal-header > p {
  font-weight: 500;
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

.log-form {
  font-weight: 500;
  letter-spacing: 0.05rem;
  padding-top: 1rem;
}

.group-wrapper {
  display: flex;
  flex-direction: column;
  padding-inline: 1rem;
}

@media (min-width: 40rem) {
  .group-wrapper {
    justify-content: space-between;
    flex-direction: row;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .group-wrapper > .input-group {
    flex-basis: 50%;
  }
}

.input-group {
  display: flex;
  flex-direction: column;
  margin-bottom: 0.5rem;
}

@media (min-width: 40rem) {
  .input-group {
    margin-bottom: 0;
  }
}

.input-group > label {
  font-size: 1rem;
  font-weight: 500;
  padding-bottom: 0.25rem;
}

.input-group > input,
.input-group > select,
.input-group > textarea {
  padding: 0.875rem;
  border-radius: 0.25rem;
  padding: 0.5rem 1rem;
  font-size: 0.95rem;
  transition: all 0.2s ease-in-out;
  border-color: var(--primary);
  width: 100%;
}

.input-group input:focus,
.input-group select:focus,
.input-group > textarea {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.2);
}

.notes {
  padding-inline: 1rem;
}

.form-action {
  border-top: 1px solid rgb(216, 216, 216);
  padding-inline: 0.75rem;
  padding-block: 0.5rem;
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
  position: sticky;
  bottom: 0;
  background-color: white;
}
</style>
