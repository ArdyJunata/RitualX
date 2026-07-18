// frontend/src/modules/routines/types.ts

export type PeriodType = 'daily' | 'weekly' | 'monthly'

export interface Routine {
  id: string
  user_id: string
  title: string
  period_type: PeriodType
  target_count: number
  icon: string
  color: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateRoutineRequest {
  title: string
  period_type: PeriodType
  target_count: number
  icon: string
  color: string
}

export interface RoutineLog {
  id: string
  routine_id: string
  user_id: string
  logged_at: string
  count: number
  note: string
  created_at: string
}

export interface DailyRoutine extends Routine {
  todayLog: RoutineLog | null
  todayCount: number
  isDone: boolean
}
