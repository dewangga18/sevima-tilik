package domain

import (
	"time"
)

type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	GradeLevel   int       `json:"grade_level,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Skill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Domain      string   `json:"domain"`
	GradeLevel  int      `json:"grade_level"`
	Description string   `json:"description"`
	Prereqs     []string `json:"prereqs,omitempty"`
}

type Question struct {
	ID            string   `json:"id"`
	SkillID       string   `json:"skill_id"`
	Difficulty    int      `json:"difficulty"`
	Prompt        string   `json:"prompt"`
	Options       []string `json:"options"`
	AnswerKey     string   `json:"-"` // Never serialized to client
	Explanation   string   `json:"explanation,omitempty"`
	Misconception string   `json:"misconception,omitempty"`
}

type AssessmentStatus string

const (
	AssessmentInProgress AssessmentStatus = "in_progress"
	AssessmentCompleted  AssessmentStatus = "completed"
)

type Assessment struct {
	ID          string           `json:"id"`
	StudentID   string           `json:"student_id"`
	GradeLevel  int              `json:"grade_level"`
	Status      AssessmentStatus `json:"status"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Items       []AssessmentItem `json:"items,omitempty"`
	Results     []SkillResult    `json:"results,omitempty"`
}

type AssessmentItem struct {
	ID             string    `json:"id"`
	AssessmentID   string    `json:"assessment_id"`
	QuestionID     string    `json:"question_id"`
	OrderIndex     int       `json:"order_index"`
	Question       *Question `json:"question,omitempty"`
	StudentAnswer  string    `json:"student_answer,omitempty"`
	IsCorrect      *bool     `json:"is_correct,omitempty"`
	AnsweredAt     *time.Time `json:"answered_at,omitempty"`
}

type SkillStatus string

const (
	StatusUnassessed    SkillStatus = "unassessed"
	StatusNeedsPractice SkillStatus = "needs_practice"
	StatusMastered      SkillStatus = "mastered"
)

type SkillResult struct {
	SkillID       string      `json:"skill_id"`
	SkillName     string      `json:"skill_name"`
	Status        SkillStatus `json:"status"`
	TotalAnswered int         `json:"total_answered"`
	TotalCorrect  int         `json:"total_correct"`
	EvidenceCount int         `json:"evidence_count"`
	Confidence    string      `json:"confidence"` // "low", "medium", "high"
	IsRootGap     bool        `json:"is_root_gap"`
}
