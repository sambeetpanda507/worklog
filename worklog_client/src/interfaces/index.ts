import type { Priority, TaskStatus, TaskType } from '@/types'

export interface ILog {
  logId: string
  taskName: string
  taskType: TaskType
  taskStatus: TaskStatus
  priority: Priority
  notes?: string
  startedAt?: string
  completedAt?: string
  createdAt: string
  updatedAt: string
}

export interface ITaskStatusSummary {
  taskStatus: string
  statusCount: number
  percentage: number
}

export interface ITaskTypeSummary {
  taskType: string
  statusCount: number
  percentage: number
}
