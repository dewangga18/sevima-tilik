package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sevima/tilik-api/internal/domain"
)

type CurriculumRepository struct {
	db *DB
}

func NewCurriculumRepository(db *DB) *CurriculumRepository {
	return &CurriculumRepository{db: db}
}

func (r *CurriculumRepository) GetSkills(ctx context.Context, gradeLevel int) ([]domain.Skill, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.name, s.domain, s.grade_level, s.description
		FROM skills s
		ORDER BY s.grade_level ASC, s.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query skills: %w", err)
	}
	defer rows.Close()

	skills := make([]domain.Skill, 0)
	for rows.Next() {
		var s domain.Skill
		if err := rows.Scan(&s.ID, &s.Name, &s.Domain, &s.GradeLevel, &s.Description); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read skills: %w", err)
	}

	// Fetch prerequisites
	for i := range skills {
		pRows, err := r.db.QueryContext(ctx, `
			SELECT prereq_id FROM skill_prerequisites WHERE skill_id = $1 ORDER BY prereq_id ASC
		`, skills[i].ID)
		if err != nil {
			return nil, err
		}
		for pRows.Next() {
			var pid string
			if err := pRows.Scan(&pid); err != nil {
				pRows.Close()
				return nil, fmt.Errorf("read skill prerequisite: %w", err)
			}
			skills[i].Prereqs = append(skills[i].Prereqs, pid)
		}
		if err := pRows.Err(); err != nil {
			pRows.Close()
			return nil, fmt.Errorf("read prerequisites: %w", err)
		}
		pRows.Close()
	}

	return skills, nil
}

func (r *CurriculumRepository) GetInitialDiagnosticQuestions(ctx context.Context, gradeLevel int) ([]domain.Question, error) {
	// For Phase 1 initial diagnostic: pick 1 question per skill in grade 4 slice
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ON (q.skill_id) q.id, q.skill_id, q.difficulty, q.prompt, q.options, q.answer_key, q.explanation, q.misconception
		FROM questions q
		JOIN skills s ON s.id = q.skill_id
		ORDER BY q.skill_id, q.difficulty ASC, q.id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query diagnostic questions: %w", err)
	}
	defer rows.Close()

	var questions []domain.Question
	for rows.Next() {
		var q domain.Question
		var optsJSON []byte
		if err := rows.Scan(&q.ID, &q.SkillID, &q.Difficulty, &q.Prompt, &optsJSON, &q.AnswerKey, &q.Explanation, &q.Misconception); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(optsJSON, &q.Options); err != nil {
			return nil, fmt.Errorf("unmarshal question options: %w", err)
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read diagnostic questions: %w", err)
	}
	return questions, nil
}

func (r *CurriculumRepository) GetQuestionByID(ctx context.Context, id string) (*domain.Question, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT q.id, q.skill_id, q.difficulty, q.prompt, q.options, q.answer_key, q.explanation, q.misconception
		FROM questions q
		WHERE q.id = $1
	`, id)

	var q domain.Question
	var optsJSON []byte
	if err := row.Scan(&q.ID, &q.SkillID, &q.Difficulty, &q.Prompt, &optsJSON, &q.AnswerKey, &q.Explanation, &q.Misconception); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(optsJSON, &q.Options); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *CurriculumRepository) GetQuestionBank(ctx context.Context) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,skill_id,difficulty,prompt,options,answer_key,explanation,misconception FROM questions ORDER BY skill_id,difficulty,id`)
	if err != nil {
		return nil, fmt.Errorf("query question bank: %w", err)
	}
	defer rows.Close()
	questions := make([]domain.Question, 0)
	for rows.Next() {
		var q domain.Question
		var options []byte
		if err := rows.Scan(&q.ID, &q.SkillID, &q.Difficulty, &q.Prompt, &options, &q.AnswerKey, &q.Explanation, &q.Misconception); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(options, &q.Options); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return questions, nil
}
