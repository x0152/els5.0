package onboarding_test

import (
	"testing"

	"github.com/els/backend/internal/domain/onboarding"
)

func TestMergeKeepsHighWaterMark(t *testing.T) {
	// arrange
	stored := map[string]int{onboarding.MetricVocabWords: 5, onboarding.MetricDiaryEntries: 2}
	live := map[string]int{onboarding.MetricVocabWords: 3, onboarding.MetricDiaryEntries: 4}

	// act
	merged := onboarding.Merge(stored, live)

	// assert
	if merged[onboarding.MetricVocabWords] != 5 {
		t.Errorf("vocab: got %d, want 5", merged[onboarding.MetricVocabWords])
	}
	if merged[onboarding.MetricDiaryEntries] != 4 {
		t.Errorf("diary: got %d, want 4", merged[onboarding.MetricDiaryEntries])
	}
	if merged[onboarding.MetricQuestsCompleted] != 0 {
		t.Errorf("quests: got %d, want 0", merged[onboarding.MetricQuestsCompleted])
	}
}

func TestIncreased(t *testing.T) {
	// arrange
	stored := map[string]int{onboarding.MetricVocabWords: 5}
	merged := map[string]int{onboarding.MetricVocabWords: 5, onboarding.MetricDiaryEntries: 1, onboarding.MetricChatMessages: 0}

	// act
	inc := onboarding.Increased(stored, merged)

	// assert
	if len(inc) != 1 || inc[onboarding.MetricDiaryEntries] != 1 {
		t.Errorf("got %v, want only diary=1", inc)
	}
}

func TestStatusesMarksDoneAtThreshold(t *testing.T) {
	// arrange
	watermarks := map[string]int{onboarding.MetricQuestsCompleted: 5}

	// act
	statuses := onboarding.Statuses(watermarks, map[string]bool{"quests_completed_5": true})

	// assert
	byID := map[string]onboarding.Status{}
	for _, s := range statuses {
		byID[s.ID] = s
	}
	if !byID["quests_completed_5"].Done {
		t.Error("quests_completed_5 should be done at value 5")
	}
	if byID["quests_completed_10"].Done {
		t.Error("quests_completed_10 should not be done at value 5")
	}
	if byID["quests_completed_10"].Value != 5 {
		t.Errorf("value: got %d, want 5", byID["quests_completed_10"].Value)
	}
	if !byID["quests_completed_5"].Acked || byID["quests_completed_10"].Acked {
		t.Error("only quests_completed_5 should be acked")
	}
}
