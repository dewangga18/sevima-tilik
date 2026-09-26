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
  probe_for_skill_id?: string
  id: string
  assessment_id: string
  question_id: string
  order_index: number
  question?: Question
  student_answer?: string
  is_correct?: boolean
  answered_at?: string
}

export type SkillStatus = 'unassessed' | 'needs_practice' | 'mastered' | 'strong_evidence' | 'inconclusive'

export interface SkillResult {
  related_target_skill_id?: string
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
  rule_version?: string
  target_skill_id?: string
  revision?: number
  max_questions?: number
  stop_reason?: 'evidence_complete' | 'insufficient_evidence' | 'question_limit'
  learning_path?: { skill_id: string; skill_name: string; reason: string }[]
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
  correct_count?: number
  assessed_skill_count: number
}

export interface AssessmentHistory {
  completed_count: number
  items: AssessmentSummary[]
}

export interface Lesson {
  id: string
  skill_id: string
  title: string
  estimated_minutes: number
  steps: { heading: string; body: string; example: string }[]
  hint: string
}
export type LearningStage = 'lesson' | 'practice' | 'reassessment' | 'completed' | 'exhausted'
export interface LearningItem {
  id: string
  question_id: string
  stage: 'practice' | 'reassessment'
  order_index: number
  question?: Question
  student_answer?: string
  is_correct?: boolean
  answered_at?: string
}
export interface LearningSession {
  id: string
  source_assessment_id: string
  skill_id: string
  skill_name: string
  rule_version: string
  stage: LearningStage
  revision: number
  started_at: string
  completed_at?: string
  before_score?: number
  score?: number
  outcome?: SkillStatus
  review_skill_id?: string
  stop_reason?: string
  lesson?: Lesson
  items: LearningItem[]
}
export interface LearningSummary {
  id: string
  skill_id: string
  skill_name: string
  stage: LearningStage
  started_at: string
  completed_at?: string
  score?: number
}
export interface LearningProgress {
  source_assessment_id?: string
  skills: { skill_id: string; skill_name: string; status: SkillStatus; score?: number; evidence_count: number; correct_count: number; source: 'diagnostic' | 'reassessment' | 'unassessed'; updated_at?: string }[]
  learning_path: { skill_id: string; skill_name: string; reason: string }[]
  active_sessions: LearningSummary[]
  recent_sessions: LearningSummary[]
  completed_count: number
}

export const learningStageLabels: Record<LearningStage, string> = { lesson: 'Membaca materi', practice: 'Latihan', reassessment: 'Cek pemahaman ulang', completed: 'Selesai', exhausted: 'Menunggu soal baru' }
