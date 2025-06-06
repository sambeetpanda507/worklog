<script setup lang="ts">
import ConfirmModal from '@/components/ConfirmModal.vue'
import LogModal from '@/components/LogModal.vue'
import TaskSummary from '@/components/TaskSummary.vue'
import { DateTime } from 'luxon'
import { onBeforeUnmount, onMounted, ref, watch, computed } from 'vue'
import TypeSummary from './components/TypeSummary.vue'
import type { ILog } from './interfaces'
import TaskCompletionSummary from './components/TaskCompletionSummary.vue'

type ConfirmModalExposed = {
  modalRef: HTMLDialogElement | null
}

type LogModalExposed = {
  modalRef: HTMLDialogElement | null
}

type FetchParams = {
  sortBy?: string
  sortOrder?: string
  page: number
  limit: number
}

const logs = ref<ILog[]>([])
const loading = ref<boolean>(false)
const searchValue = ref<string>('')
const selectedLogIds = ref<string[]>([])
const showActionMenu = ref<boolean>(false)
const menuRef = ref<HTMLDivElement | null>(null)
const showDeleteConfirmation = ref<boolean>(false)
const confirmDialogRef = ref<ConfirmModalExposed | null>(null)
const logModalRef = ref<LogModalExposed | null>(null)
const isEditLog = ref<boolean>(false)
const sortBy = ref<string>('created_at')
const sortOrder = ref<string>('desc')
const columns = ref<{ label: string; value: string }[]>([
  { label: 'Task Name', value: 'task_name' },
  { label: 'Task Type', value: 'task_type' },
  { label: 'Task Status', value: 'task_status' },
  { label: 'Priority', value: 'priority' },
  { label: 'Notes', value: 'notes' },
  { label: 'Started At', value: 'started_at' },
  { label: 'Completed At', value: 'completed_at' },
  { label: 'Created At', value: 'created_at' },
  { label: 'Updated At', value: 'updated_at' },
])
const page = ref<number>(0)
const limit = ref<number>(10)

const totalPages = computed(() => {
  if (logs.value.length == 0) {
    return 0
  } else {
    return logs.value[0].totalPages
  }
})

async function fetchLogs({ sortBy, sortOrder, page, limit }: FetchParams): Promise<void> {
  try {
    loading.value = true
    const baseURL: string = `${import.meta.env.VITE_SERVER_BASE_URL}/logs`
    const searchParams = new URLSearchParams()
    searchParams.set('page', page.toString())
    searchParams.set('limit', limit.toString())
    if (searchValue.value.trim().length > 0) {
      searchParams.set('s', searchValue.value)
    }

    if (sortBy && sortOrder) {
      searchParams.set('sortBy', sortBy)
      searchParams.set('sortOrder', sortOrder)
    }

    const url: string = `${baseURL}?${searchParams.toString()}`
    const response = await fetch(url)
    if (response.status != 200) {
      throw new Error("Something wen't wrong while fetching logs.")
    }

    const data = await response.json()
    const logData: ILog[] = data.logs
    logs.value = logData
  } catch (e: unknown) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function handleClearSearch() {
  searchValue.value = ''
  await fetchLogs({ page: page.value, limit: limit.value })
}

function handleSelectAll(e: Event) {
  const checkbox = e.target as HTMLInputElement
  if (checkbox.checked) {
    selectedLogIds.value = logs.value.map((log) => log.logId)
  } else {
    selectedLogIds.value = []
  }
}

function toggleAction() {
  showActionMenu.value = !showActionMenu.value
}

function handleClickOutside(e: MouseEvent) {
  if (showActionMenu.value && menuRef.value && !menuRef.value.contains(e.target as Node)) {
    showActionMenu.value = false
  }
}

async function handleConfirmDeleteLogs() {
  try {
    loading.value = true
    const logIds: string[] = selectedLogIds.value
    const baseURL: string = `${import.meta.env.VITE_SERVER_BASE_URL}/logs`
    const searchParms: URLSearchParams = new URLSearchParams({ logIds: JSON.stringify(logIds) })
    const url: string = `${baseURL}?${searchParms.toString()}`
    await fetch(url, { method: 'DELETE' })
    selectedLogIds.value = []
    await fetchLogs({ page: page.value, limit: limit.value })
  } catch (e: unknown) {
    console.error(e)
  } finally {
    loading.value = false
    if (confirmDialogRef.value && confirmDialogRef.value.modalRef) {
      confirmDialogRef.value.modalRef.close()
    }
  }
}

function closeActionMenu() {
  showActionMenu.value = false
}

function handleDeleteLogs() {
  if (selectedLogIds.value.length > 0) {
    showDeleteConfirmation.value = true
    const modalRef = confirmDialogRef.value?.modalRef
    modalRef?.showModal()
  } else {
    window.alert('No log selected')
  }

  // close the menu
  closeActionMenu()
}

function handleCreateNew() {
  logModalRef.value?.modalRef?.showModal()
  closeActionMenu()
}

function handleCloseLogModal() {
  isEditLog.value = false
  selectedLogIds.value = []
  logModalRef.value?.modalRef?.close()
}

function handleUpdateLog() {
  closeActionMenu()

  // check if row is selected or not
  if (!selectedLogIds.value.length) {
    window.alert('No log selected')
    return
  }

  // show validation for multiple row selection
  if (selectedLogIds.value.length > 1) {
    window.alert('Only 1 log can be updated at a time')
    return
  }

  // open the LogModal and pre-populate the fields
  isEditLog.value = true
  logModalRef.value?.modalRef?.showModal()

  // update api and send payload
}

function handleSort(col: string) {
  if (col == sortBy.value) {
    if (sortOrder.value == 'asc') {
      sortOrder.value = 'desc'
    } else {
      sortOrder.value = 'asc'
    }
  }

  sortBy.value = col
}

function handlePageIncrement() {
  if (page.value < totalPages.value - 1) {
    page.value += 1
  }
}

function handlePageDecrement() {
  if (page.value > 0) {
    page.value -= 1
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  fetchLogs({
    sortBy: sortBy.value,
    sortOrder: sortOrder.value,
    page: page.value,
    limit: limit.value,
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

watch([sortBy, sortOrder, page, limit], ([newSortBy, newSortOrder, newPageCount, newLimit]) => {
  fetchLogs({ sortBy: newSortBy, sortOrder: newSortOrder, page: newPageCount, limit: newLimit })
})
</script>

<template>
  <main class="container">
    <h1 class="page-title">
      Task Management Dashboard <span class="material-symbols-outlined title-icon"> list_alt </span>
    </h1>

    <ConfirmModal
      ref="confirmDialogRef"
      :confirm-button-text="'CONFIRM'"
      :confirm-text="'Are you sure you want to delete logs(s)?'"
      :handle-confirm="handleConfirmDeleteLogs"
      :loading="loading"
    />

    <!-- LOG MODAL -->
    <LogModal
      ref="logModalRef"
      @close-modal="handleCloseLogModal"
      :fetch-logs="() => fetchLogs({ page, limit })"
      :logs="logs"
      :selected-log-ids="selectedLogIds"
      :is-edit-log="isEditLog"
    />

    <!-- TASK COMPLETION SYMMARY -->
    <div class="task-completion-summary">
      <TaskCompletionSummary />
    </div>

    <!-- CARDS -->
    <div class="summary-cards">
      <!-- STATUS DONUT CHART -->
      <TaskSummary />

      <!-- STATUS DONUT CHART -->
      <TypeSummary />
    </div>

    <section class="data-grid">
      <!-- search bar -->
      <div class="search-box-container">
        <form
          class="search-box"
          @submit.prevent="() => fetchLogs({ sortBy, sortOrder, page, limit })"
        >
          <input type="text" v-model="searchValue" placeholder="Search logs or notes..." required />
          <button type="submit" class="primary-button search-button">SEARCH</button>
          <button
            v-if="searchValue.trim().length > 0"
            type="button"
            @click="handleClearSearch"
            class="clear-button danger-button"
          >
            Clear
          </button>
        </form>
      </div>

      <!-- action buttion -->
      <div class="table-buttons">
        <div class="action-container" ref="menuRef">
          <button class="primary-button" @click="toggleAction">ACTION</button>
          <ul :class="['menu', showActionMenu ? 'active' : 'hidden']">
            <li class="menu-item" @click="handleCreateNew">Create New</li>
            <li class="menu-item" @click="handleDeleteLogs">Delete</li>
            <li class="menu-item" @click="handleUpdateLog">Update</li>
          </ul>
        </div>
      </div>

      <div class="table-container">
        <div class="logs-container">
          <table class="logs">
            <thead>
              <tr>
                <th>
                  <input type="checkbox" id="allLogs" class="checkbox" @change="handleSelectAll" />
                </th>
                <th v-for="column in columns" :key="column.value">
                  <span class="task-header-title">
                    <span class="col-name">{{ column.label }}</span>
                    <span
                      :class="[
                        'material-symbols-outlined',
                        'arrow-icon',
                        sortBy == column.value && 'active',
                      ]"
                      @click.prevent="() => handleSort(column.value)"
                    >
                      {{ sortOrder == 'asc' ? 'arrow_upward' : 'arrow_downward' }}
                    </span>
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <!-- Iterate over logs and render logId -->
              <tr v-for="log of logs" :key="log.logId">
                <td>
                  <div class="checkbox-container">
                    <input
                      type="checkbox"
                      id="log"
                      class="checkbox"
                      :value="log.logId"
                      v-model="selectedLogIds"
                    />
                  </div>
                </td>
                <td>{{ log.taskName }}</td>
                <td>{{ log.taskType }}</td>
                <td>{{ log.taskStatus }}</td>
                <td>{{ log.priority }}</td>
                <td class="notes" :title="log.notes">{{ log.notes }}</td>
                <td v-if="log.startedAt" class="date-cell">
                  <span class="date">{{
                    DateTime.fromISO(log.startedAt).toFormat('dd-LLL-yyyy')
                  }}</span>
                  <span class="seperator">|</span>
                  <span class="date">{{
                    DateTime.fromISO(log.startedAt).toFormat('hh:mm a')
                  }}</span>
                </td>
                <td v-else>N/A</td>
                <td v-if="log.completedAt" class="date-cell">
                  {{ DateTime.fromISO(log.completedAt).toFormat('dd-LLL-yyyy') }}
                </td>
                <td v-else>N/A</td>
                <td class="date-cell">
                  {{ DateTime.fromISO(log.createdAt).toFormat('dd-LLL-yyyy') }}
                </td>
                <td class="date-cell">
                  {{ DateTime.fromISO(log.updatedAt).toFormat('dd-LLL-yyyy') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Table pagination container -->
        <div class="pagination-container">
          <span class="material-symbols-outlined decrement-icon" @click="handlePageDecrement">
            chevron_left
          </span>
          <span class="page-number">{{ page + 1 }}</span>
          <span class="highphen">-</span>
          <span class="total-pages">{{ totalPages }}</span>
          <span class="material-symbols-outlined increment-icon" @click="handlePageIncrement">
            chevron_right
          </span>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.checkbox-container {
  display: flex;
  align-items: center;
  justify-self: center;
}

.page-title {
  text-align: center;
  margin-bottom: 5rem;
  display: flex;
  align-items: center;
  width: 100%;
  justify-content: center;
  gap: 1rem;
  font-size: 1.5rem;
}

.title-icon {
  font-size: 2rem;
  display: none;
}

@media (min-width: 40rem) {
  .page-title {
    font-size: 2rem;
  }

  .title-icon {
    display: block;
  }
}

.task-completion-summary {
  margin: 2rem;
}

.summary-cards {
  margin: 2rem;
  display: grid;
  gap: 1rem;
}

@media (min-width: 40rem) {
  .summary-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

.data-grid {
  margin: 2rem;
}

.search-box-container {
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 0.5rem;
  background-color: #fff;
  margin-bottom: 1rem;
  border-radius: 0.25rem;
}

.search-box {
  border: 2px solid #cbd5e1;
  display: flex;
  border-radius: 0.25rem;
  padding: 0.5rem 1rem;
  font-size: 0.95rem;
  transition: all 0.2s ease-in-out;
}

.search-box > input {
  width: 100%;
  outline: none;
  border: none;
}

.search-box:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.2);
}

.clear-button {
  margin-left: 0.5rem;
}

.logs-container {
  overflow: auto;
}

.logs {
  border-collapse: collapse;
  text-align: left;
}

.logs thead {
  background-color: var(--primary);
  color: white;
  text-align: left;
}

th,
td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e0e0e0;
  white-space: nowrap;
}

th {
  font-weight: 500;
}

.task-header-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: bold;
}

.col-name {
  font-weight: 500;
}

.task-header-title:hover > .arrow-icon {
  opacity: 1;
}

table.logs tbody tr:nth-child(even) {
  background-color: #f9f9f9;
}

table.logs tbody tr:hover {
  background-color: #f1f5f9;
  /* light slate on hover */
  transition: background-color 0.2s ease-in-out;
}

table.logs th {
  font-weight: 600;
  font-size: 0.95rem;
}

table.logs td {
  font-size: 0.9rem;
  color: #333;
}

.date-cell {
  padding-right: 2rem;
}

.table-buttons {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 1rem;
}

.action-container {
  position: relative;
}

.menu {
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  border-radius: 0.25rem;
  list-style-type: none;
  padding-block: 0.5rem;
  position: absolute;
  z-index: 1;
  background-color: #fff;
  top: 100%;
  right: 50%;
}

.menu.active {
  display: block;
}

.menu.hidden {
  display: none;
}

.menu-item {
  padding: 0.25rem 0.5rem;
  cursor: pointer;
  color: rgb(76, 76, 76);
  white-space: nowrap;
}

.menu-item:hover {
  background-color: rgb(247, 247, 247);
}

.notes {
  width: 100%;
  max-width: 20rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.seperator {
  padding-inline: 0.25rem;
  color: rgb(159, 159, 159);
  font-weight: 300;
}

.arrow-icon {
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
  opacity: 0;
}

.arrow-icon.active {
  opacity: 1;
}

.pagination-container {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  width: 100%;
  gap: 0.25rem;
  padding: 1rem;
  position: sticky;
  left: 0;
  bottom: 0;
}

.decrement-icon,
.increment-icon {
  font-size: 1.5rem;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
}

.page-number,
.total-pages {
  font-size: 0.875rem;
  display: block;
  font-weight: 600;
}

.table-container {
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  border-radius: 0.25rem;
  overflow: hidden;
}
</style>
