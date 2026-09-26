package service

import (
	"fmt"
	"sort"

	"github.com/sevima/tilik-api/internal/domain"
)

// prerequisiteOrder validates the DAG and orders prerequisites before dependents.
// Sorting ties makes the result reproducible independently of database row order.
func prerequisiteOrder(skills []domain.Skill) ([]string, error) {
	remaining := make(map[string]int, len(skills))
	dependents := make(map[string][]string, len(skills))
	for _, skill := range skills {
		if skill.ID == "" {
			return nil, fmt.Errorf("skill ID is empty")
		}
		if _, exists := remaining[skill.ID]; exists {
			return nil, fmt.Errorf("duplicate skill %s", skill.ID)
		}
		remaining[skill.ID] = 0
	}
	for _, skill := range skills {
		seen := make(map[string]bool, len(skill.Prereqs))
		for _, prerequisite := range skill.Prereqs {
			if _, exists := remaining[prerequisite]; !exists {
				return nil, fmt.Errorf("skill %s references missing prerequisite %s", skill.ID, prerequisite)
			}
			if seen[prerequisite] {
				return nil, fmt.Errorf("skill %s has duplicate prerequisite %s", skill.ID, prerequisite)
			}
			seen[prerequisite] = true
			remaining[skill.ID]++
			dependents[prerequisite] = append(dependents[prerequisite], skill.ID)
		}
	}
	ready := make([]string, 0, len(skills))
	for id, count := range remaining {
		if count == 0 {
			ready = append(ready, id)
		}
	}
	order := make([]string, 0, len(skills))
	for len(ready) > 0 {
		sort.Strings(ready)
		id := ready[0]
		ready = ready[1:]
		order = append(order, id)
		for _, dependent := range dependents[id] {
			remaining[dependent]--
			if remaining[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
	}
	if len(order) != len(skills) {
		return nil, fmt.Errorf("prerequisite graph contains a cycle")
	}
	return order, nil
}
