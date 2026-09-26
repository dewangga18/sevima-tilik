export type Role = 'student' | 'teacher' | 'admin'

export interface User {
  id: string
  email: string
  name: string
  role: Role
  grade_level?: number
  created_at: string
}

export interface Skill {
  id: string
  name: string
  domain: string
  grade_level: number
  description: string
  prereqs?: string[]
}

export interface Question {
  id: string
  skill_id: string
  difficulty: number
  prompt: string
  options: string[]
  explanation?: string
  misconception?: string
}

export interface AssessmentItem {
  id: string
  assessment_id: string
  question_id: string
  order_index: number
  question?: Question
  student_answer?: string
  is_correct?: boolean
  answered_at?: string
}

export type SkillStatus = 'unassessed' | 'needs_practice' | 'mastered'

export interface SkillResult {
  skill_id: string
  skill_name: string
  status: SkillStatus
  total_answered: number
  total_correct: number
  evidence_count: number
  confidence: 'low' | 'medium' | 'high'
  is_root_gap: boolean
}

export type AssessmentStatus = 'in_progress' | 'completed'

export interface Assessment {
  id: string
  student_id: string
  grade_level: number
  status: AssessmentStatus
  started_at: string
  completed_at?: string
  items?: AssessmentItem[]
  results?: SkillResult[]
}

export interface AssessmentSummary {
  id: string
  grade_level: number
  status: AssessmentStatus
  started_at: string
  completed_at?: string
  question_count: number
  answered_count: number
  correct_count: number
  assessed_skill_count: number
}

export interface AssessmentHistory {
  completed_count: number
  items: AssessmentSummary[]
}
