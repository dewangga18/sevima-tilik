package domain

import (
	"encoding/json"
	"time"
)

type QuestionCandidate struct {
	ID                 string                   `json:"id"`
	Purpose            string                   `json:"purpose"`
	Grade              int                      `json:"grade"`
	Phase              string                   `json:"phase"`
	Domain             string                   `json:"domain"`
	Topic              string                   `json:"topic"`
	Skill              string                   `json:"skill"`
	Difficulty         int                      `json:"difficulty"`
	CognitiveLevel     string                   `json:"cognitive_level"`
	QuestionType       string                   `json:"question_type"`
	Context            string                   `json:"context"`
	Prompt             string                   `json:"question"`
	Options            []CandidateOption        `json:"options"`
	CorrectAnswer      string                   `json:"correct_answer"`
	Explanation        string                   `json:"explanation"`
	Prerequisites      []string                 `json:"prerequisite_skills"`
	Misconceptions     []CandidateMisconception `json:"possible_misconceptions"`
	MetadataInferred   bool                     `json:"metadata_inferred"`
	SourceURL          string                   `json:"source_url"`
	SourceTitle        string                   `json:"source_title"`
	SourceType         string                   `json:"source_type"`
	LicenseNote        string                   `json:"license_note"`
	RawSourceReference string                   `json:"raw_source_reference"`
}
type CandidateOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type CandidateMisconception struct {
	WrongAnswer string `json:"wrong_answer"`
	Reason      string `json:"reason"`
}
type CandidateReview struct {
	Hash           string `json:"hash"`
	Status         string `json:"status"`
	Note           string `json:"note"`
	ExpectedAnswer string `json:"expected_answer,omitempty"`
}
type BankEntry struct {
	Candidate  QuestionCandidate
	Hash       string
	SourceFile string
	Payload    json.RawMessage
	Status     string
	Note       string
	Question   Question
}
type BankImportItem struct {
	ID     string `json:"id"`
	Hash   string `json:"hash"`
	Status string `json:"status"`
	Reason string `json:"reason"`
	Action string `json:"action"`
}
type BankImportReport struct {
	Applied bool             `json:"applied"`
	Items   []BankImportItem `json:"items"`
}

// QuestionCandidateRecord is a stored candidate plus its review state. The
// embedded candidate carries correct_answer, so this shape must only ever be
// served to admins who are deciding whether to activate it.
type QuestionCandidateRecord struct {
	ID           string            `json:"id"`
	Status       string            `json:"status"`
	ReviewNote   string            `json:"review_note"`
	SourceFile   string            `json:"source_file"`
	ContentHash  string            `json:"content_hash"`
	RegisteredAt time.Time         `json:"registered_at"`
	Candidate    QuestionCandidate `json:"candidate"`
	// Payload keeps the exact stored bytes so re-registration can never
	// rewrite historical content with a re-serialized equivalent.
	Payload json.RawMessage `json:"-"`
}

// CandidateReviewView adds why a candidate can or cannot be activated, so the
// review UI can explain the rule before an admin clicks rather than after a
// rejected request.
type CandidateReviewView struct {
	QuestionCandidateRecord
	Eligible      bool   `json:"eligible"`
	BlockedReason string `json:"blocked_reason,omitempty"`
}
