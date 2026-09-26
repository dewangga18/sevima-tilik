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
	IsActive     bool      `json:"is_active"`
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

type LearningPathItem struct {
	SkillID   string `json:"skill_id"`
	SkillName string `json:"skill_name"`
	Reason    string `json:"reason"`
}

type Assessment struct {
	RuleVersion   string             `json:"rule_version"`
	TargetSkillID string             `json:"target_skill_id,omitempty"`
	Revision      int                `json:"revision"`
	MaxQuestions  int                `json:"max_questions,omitempty"`
	StopReason    string             `json:"stop_reason,omitempty"`
	LearningPath  []LearningPathItem `json:"learning_path,omitempty"`
	ID            string             `json:"id"`
	StudentID     string             `json:"student_id"`
	GradeLevel    int                `json:"grade_level"`
	Status        AssessmentStatus   `json:"status"`
	StartedAt     time.Time          `json:"started_at"`
	CompletedAt   *time.Time         `json:"completed_at,omitempty"`
	Items         []AssessmentItem   `json:"items,omitempty"`
	Results       []SkillResult      `json:"results,omitempty"`
}

type AssessmentItem struct {
	ProbeForSkillID string     `json:"probe_for_skill_id,omitempty"`
	ID              string     `json:"id"`
	AssessmentID    string     `json:"assessment_id"`
	QuestionID      string     `json:"question_id"`
	OrderIndex      int        `json:"order_index"`
	Question        *Question  `json:"question,omitempty"`
	StudentAnswer   string     `json:"student_answer,omitempty"`
	IsCorrect       *bool      `json:"is_correct,omitempty"`
	AnsweredAt      *time.Time `json:"answered_at,omitempty"`
}

type AssessmentSummary struct {
	ID                 string           `json:"id"`
	GradeLevel         int              `json:"grade_level"`
	Status             AssessmentStatus `json:"status"`
	StartedAt          time.Time        `json:"started_at"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty"`
	QuestionCount      int              `json:"question_count"`
	AnsweredCount      int              `json:"answered_count"`
	CorrectCount       *int             `json:"correct_count,omitempty"`
	AssessedSkillCount int              `json:"assessed_skill_count"`
}

type AssessmentHistory struct {
	CompletedCount int                 `json:"completed_count"`
	Items          []AssessmentSummary `json:"items"`
}

type SkillStatus string

const (
	StatusStrongEvidence SkillStatus = "strong_evidence"
	StatusInconclusive   SkillStatus = "inconclusive"
	StatusUnassessed     SkillStatus = "unassessed"
	StatusNeedsPractice  SkillStatus = "needs_practice"
	StatusMastered       SkillStatus = "mastered"
)

type SkillResult struct {
	RelatedTargetSkillID string      `json:"related_target_skill_id,omitempty"`
	SkillID              string      `json:"skill_id"`
	SkillName            string      `json:"skill_name"`
	Status               SkillStatus `json:"status"`
	TotalAnswered        int         `json:"total_answered"`
	TotalCorrect         int         `json:"total_correct"`
	EvidenceCount        int         `json:"evidence_count"`
	Confidence           string      `json:"confidence"` // "low", "medium", "high"
	IsRootGap            bool        `json:"is_root_gap"`
}

type TeacherClass struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	GradeLevel    int    `json:"grade_level"`
	StudentCount  int    `json:"student_count"`
	AssessedCount int    `json:"assessed_count"`
}

type StudentOverview struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	LatestStatus    string     `json:"latest_status"` // "completed", "in_progress", "unassessed"
	AssessedSkills  int        `json:"assessed_skills"`
	VisibleGapCount int        `json:"visible_gap_count"`
	RootGapCount    int        `json:"root_gap_count"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type ProgressEntry struct {
	SkillID        string    `json:"skill_id"`
	SkillName      string    `json:"skill_name"`
	Score          int       `json:"score"`
	Status         string    `json:"status"`
	EvidenceCount  int       `json:"evidence_count"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type StudentInsight struct {
	StudentID      string             `json:"student_id"`
	StudentName    string             `json:"student_name"`
	ClassroomID    string             `json:"classroom_id"`
	ClassroomName  string             `json:"classroom_name"`
	AssessmentID   string             `json:"assessment_id,omitempty"`
	AssessedAt     *time.Time         `json:"assessed_at,omitempty"`
	TargetSkillID  string             `json:"target_skill_id,omitempty"`
	Results        []SkillResult      `json:"results"`
	Recommendations []LearningPathItem `json:"recommendations"`
	Progress       []ProgressEntry    `json:"progress"`
}

type QuestionBankSkillRow struct {
	SkillID       string `json:"skill_id"`
	SkillName     string `json:"skill_name"`
	QuestionCount int    `json:"question_count"`
}

type QuestionBankStatusRow struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type QuestionBankSummary struct {
	Active    []QuestionBankSkillRow  `json:"active"`
	Candidate []QuestionBankStatusRow `json:"candidate"`
}
