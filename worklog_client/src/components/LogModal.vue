<script setup lang="ts">
import type { ILog } from '@/interfaces'
import { DateTime } from 'luxon'
import { ref, watch } from 'vue'

interface IProps {
  fetchLogs: () => Promise<void>
  selectedLogIds: string[]
  logs: ILog[]
  isEditLog: boolean
}

interface IEmits {
  (e: 'closeModal'): void
}

type TaskTypes = 'task' | 'bug' | 'story'
type TaskProgress = 'backlog' | 'pending' | 'progress' | 'pr' | 'staging'

const { fetchLogs, selectedLogIds, logs, isEditLog } = defineProps<IProps>()
const modalRef = ref<HTMLDialogElement | null>(null)
const taskName = ref<string>('')
const taskType = ref<TaskTypes>('task')
const taskStatus = ref<TaskProgress>('pending')
const taskPriority = ref<number>(1)
const startedAt = ref<string>(DateTime.now().toFormat("yyyy-MM-dd'T'HH:mm"))
const completedAt = ref<string>('')
const notes = ref<string>('')
const loading = ref<boolean>(false)
const logId = ref<string | null>(null)
const emit = defineEmits<IEmits>()
defineExpose({ modalRef })

function resetState() {
  taskName.value = ''
  taskType.value = 'task'
  taskStatus.value = 'pending'
  taskPriority.value = 1
  startedAt.value = DateTime.now().toFormat("yyyy-MM-dd'T'HH:mm")
  completedAt.value = ''
  notes.value = ''
  loading.value = false
  logId.value = null
}

function handleClose() {
  resetState()
  emit('closeModal')
}

async function handleSubmit() {
  try {
    loading.value = true
    const payload = {
      taskName: taskName.value,
      taskType: taskType.value,
      taskStatus: taskStatus.value,
      priority: taskPriority.value,
      ...(logId.value ? { logId: logId.value } : {}),
      ...(notes.value.trim().length > 0 ? { notes: notes.value } : {}),
      ...(startedAt.value.length > 0
        ? { startedAt: DateTime.fromFormat(startedAt.value, "yyyy-MM-dd'T'HH:mm").toISO() }
        : {}),
      ...(completedAt.value.length > 0
        ? { completedAt: DateTime.fromFormat(completedAt.value, "yyyy-MM-dd'T'HH:mm").toISO() }
        : {}),
    }

    const method = logId.value ? 'PUT' : 'POST'
    const res = await fetch('http://localhost:5001/log', {
      method,
      body: JSON.stringify(payload),
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
    })

    if (res.status == 400) {
      const errorMessage = await res.text()
      window.alert(errorMessage)
      return
    }

    await fetchLogs()
  } catch (e) {
    console.log(e)
  } finally {
    loading.value = false
    handleClose()
  }
}

// pass is edit also
watch(
  [() => selectedLogIds, () => logs, () => isEditLog],
  ([newSelectedLogIds, newLogs, isEditLog]) => {
    if (isEditLog && newSelectedLogIds.length > 0) {
      const selectedLogs = newLogs.filter((log) => newSelectedLogIds.includes(log.logId))
      if (selectedLogs.length > 0) {
        const selectedLog = selectedLogs[0]
        logId.value = selectedLog.logId
        taskName.value = selectedLog.taskName
        taskType.value = selectedLog.taskType
        taskStatus.value = selectedLog.taskStatus
        taskPriority.value = selectedLog.priority
        if (selectedLog.startedAt) {
          startedAt.value = DateTime.fromISO(selectedLog.startedAt).toFormat("yyyy-MM-dd'T'HH:mm")
        }

        if (selectedLog.completedAt) {
          completedAt.value = DateTime.fromISO(selectedLog.completedAt).toFormat(
            "yyyy-MM-dd'T'HH:mm",
          )
        }

        if (selectedLog.notes) {
          notes.value = selectedLog.notes
        }
      }
    }
  },
)
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
          <label for="completedAt">End On</label>
          <input type="datetime-local" name="completedAt" id="completedAt" v-model="completedAt" />
        </div>
      </div>

      <div class="input-group notes">
        <label for="notes">Notes</label>
        <textarea name="notes" id="notes" rows="3" v-model="notes"></textarea>
      </div>

      <div class="form-action">
        <button type="submit" class="primary-button">
          {{ loading ? 'LOADING...' : logId ? 'EDIT' : 'SAVE' }}
        </button>
      </div>
    </form>
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

.notes textarea {
  resize: none;
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
