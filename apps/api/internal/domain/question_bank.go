package domain

import "encoding/json"

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
