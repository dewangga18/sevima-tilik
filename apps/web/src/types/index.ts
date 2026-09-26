export type Role = 'student' | 'teacher' | 'admin'

export interface User {
  id: string
  email: string
  name: string
  role: Role
  grade_level?: number
  is_active?: boolean
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

export interface TeacherClass {
  id: string
  name: string
  grade_level: number
  student_count: number
  assessed_count: number
}
export interface StudentOverview {
  id: string
  name: string
  latest_status: 'completed' | 'in_progress' | 'unassessed'
  assessed_skills: number
  visible_gap_count: number
  root_gap_count: number
  updated_at?: string
}
export interface SkillEvidence {
  skill_id: string
  skill_name: string
  status: SkillStatus
  total_answered: number
  total_correct: number
  evidence_count: number
  confidence: string
  is_root_gap: boolean
}
export interface StudentInsight {
  student_id: string
  student_name: string
  classroom_id: string
  classroom_name: string
  assessment_id?: string
  assessed_at?: string
  target_skill_id?: string
  results: SkillEvidence[]
  recommendations: { skill_id: string; skill_name: string; reason: string }[]
  progress: { skill_id: string; skill_name: string; score: number; status: string; evidence_count: number; updated_at: string }[]
}

// Admin management shapes. AdminClass is separate from TeacherClass because the
// admin view needs the assigned-teacher count, not the assessed-student count.
export interface AdminClass {
  id: string
  name: string
  grade_level: number
  student_count: number
  teacher_count: number
}

export interface ClassRoster {
  classroom_id: string
  student_ids: string[]
  teacher_ids: string[]
}

export type CandidateStatus = 'draft' | 'approved' | 'rejected'

export interface QuestionCandidateOption {
  id: string
  text: string
}

export interface QuestionCandidate {
  id: string
  purpose: string
  grade: number
  domain: string
  topic: string
  skill: string
  difficulty: number
  question_type: string
  question: string
  options: QuestionCandidateOption[]
  // Answer key is returned to admin review and the requesting teacher/admin generator.
  correct_answer: string
  explanation: string
  possible_misconceptions: { wrong_answer: string; reason: string }[]
  source_title: string
  license_note: string
}

export interface CandidateReview {
  id: string
  status: CandidateStatus
  review_note: string
  source_file: string
  content_hash: string
  registered_at: string
  candidate: QuestionCandidate
  eligible: boolean
  blocked_reason?: string
}

export interface QuestionBankSummary {
  active: { skill_id: string; skill_name: string; question_count: number }[]
  candidate: { status: CandidateStatus; count: number }[]
}

export interface BankReviewReport {
  activated: boolean
  report: { applied: boolean; items: { id: string; status: string; reason: string; action: string }[] }
}

export interface AISettingsView {
  configured: boolean
  storage_ready: boolean
  model: string
  updated_at?: string
}
export interface AIGenerateInput {
  request_id: string
  skill_id: string
  grade_level: 4
  difficulty: number
  purpose: 'diagnostic' | 'practice' | 'reassessment'
}
export interface AIGeneratedDraft { candidate: Omit<CandidateReview, 'eligible' | 'blocked_reason'> }
