package domain

import "time"

type LessonStep struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
	Example string `json:"example"`
}
type Lesson struct {
	ID               string       `json:"id"`
	SkillID          string       `json:"skill_id"`
	Title            string       `json:"title"`
	EstimatedMinutes int          `json:"estimated_minutes"`
	Steps            []LessonStep `json:"steps"`
	Hint             string       `json:"hint"`
}
type LearningItem struct {
	ID            string     `json:"id"`
	QuestionID    string     `json:"question_id"`
	Stage         string     `json:"stage"`
	OrderIndex    int        `json:"order_index"`
	Question      *Question  `json:"question,omitempty"`
	StudentAnswer string     `json:"student_answer,omitempty"`
	IsCorrect     *bool      `json:"is_correct,omitempty"`
	AnsweredAt    *time.Time `json:"answered_at,omitempty"`
}
type LearningSession struct {
	ID                 string         `json:"id"`
	StudentID          string         `json:"student_id"`
	SourceAssessmentID string         `json:"source_assessment_id"`
	SkillID            string         `json:"skill_id"`
	SkillName          string         `json:"skill_name"`
	RequestID          string         `json:"-"`
	RuleVersion        string         `json:"rule_version"`
	Stage              string         `json:"stage"`
	Revision           int            `json:"revision"`
	StartedAt          time.Time      `json:"started_at"`
	LessonCompletedAt  *time.Time     `json:"lesson_completed_at,omitempty"`
	CompletedAt        *time.Time     `json:"completed_at,omitempty"`
	BeforeScore        *int           `json:"before_score,omitempty"`
	Score              *int           `json:"score,omitempty"`
	Outcome            SkillStatus    `json:"outcome,omitempty"`
	StopReason         string         `json:"stop_reason,omitempty"`
	ReviewSkillID      string         `json:"review_skill_id,omitempty"`
	Lesson             *Lesson        `json:"lesson,omitempty"`
	Items              []LearningItem `json:"items"`
}
type SkillProgress struct {
	SkillID       string      `json:"skill_id"`
	SkillName     string      `json:"skill_name"`
	Status        SkillStatus `json:"status"`
	Score         *int        `json:"score,omitempty"`
	EvidenceCount int         `json:"evidence_count"`
	CorrectCount  int         `json:"correct_count"`
	Source        string      `json:"source"`
	UpdatedAt     *time.Time  `json:"updated_at,omitempty"`
}
type LearningSummary struct {
	ID          string     `json:"id"`
	SkillID     string     `json:"skill_id"`
	SkillName   string     `json:"skill_name"`
	Stage       string     `json:"stage"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Score       *int       `json:"score,omitempty"`
}
type LearningProgress struct {
	SourceAssessmentID string             `json:"source_assessment_id,omitempty"`
	Skills             []SkillProgress    `json:"skills"`
	LearningPath       []LearningPathItem `json:"learning_path"`
	ActiveSessions     []LearningSummary  `json:"active_sessions"`
	RecentSessions     []LearningSummary  `json:"recent_sessions"`
	CompletedCount     int                `json:"completed_count"`
}
