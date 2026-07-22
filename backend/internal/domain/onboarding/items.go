package onboarding

const (
	MetricWorkoutsCompleted = "workouts_completed"
	MetricVocabWords        = "vocab_words"
	MetricChatMessages      = "chat_messages"
	MetricDiaryEntries      = "diary_entries"
	MetricQuestsCompleted   = "quests_completed"
	MetricFilmsWatched      = "films_watched"
	MetricArticlesRead      = "articles_read"
	MetricBookChapters      = "book_chapters_completed"
)

var Metrics = []string{
	MetricWorkoutsCompleted,
	MetricVocabWords,
	MetricChatMessages,
	MetricDiaryEntries,
	MetricQuestsCompleted,
	MetricFilmsWatched,
	MetricArticlesRead,
	MetricBookChapters,
}

type Kind string

const (
	KindChecklist   Kind = "checklist"
	KindAchievement Kind = "achievement"
)

type Item struct {
	ID        string
	Kind      Kind
	Metric    string
	Threshold int
}

var Items = []Item{
	{ID: "first_film", Kind: KindChecklist, Metric: MetricFilmsWatched, Threshold: 1},
	{ID: "first_quest", Kind: KindChecklist, Metric: MetricQuestsCompleted, Threshold: 1},
	{ID: "first_article", Kind: KindChecklist, Metric: MetricArticlesRead, Threshold: 1},
	{ID: "first_workout", Kind: KindChecklist, Metric: MetricWorkoutsCompleted, Threshold: 1},
	{ID: "first_chat", Kind: KindChecklist, Metric: MetricChatMessages, Threshold: 1},
	{ID: "first_words", Kind: KindChecklist, Metric: MetricVocabWords, Threshold: 5},
	{ID: "first_chapter", Kind: KindChecklist, Metric: MetricBookChapters, Threshold: 1},

	{ID: "quests_completed_1", Kind: KindAchievement, Metric: MetricQuestsCompleted, Threshold: 1},
	{ID: "quests_completed_5", Kind: KindAchievement, Metric: MetricQuestsCompleted, Threshold: 5},
	{ID: "quests_completed_10", Kind: KindAchievement, Metric: MetricQuestsCompleted, Threshold: 10},
	{ID: "quests_completed_25", Kind: KindAchievement, Metric: MetricQuestsCompleted, Threshold: 25},
	{ID: "workouts_completed_5", Kind: KindAchievement, Metric: MetricWorkoutsCompleted, Threshold: 5},
	{ID: "workouts_completed_15", Kind: KindAchievement, Metric: MetricWorkoutsCompleted, Threshold: 15},
	{ID: "workouts_completed_30", Kind: KindAchievement, Metric: MetricWorkoutsCompleted, Threshold: 30},
	{ID: "vocab_words_25", Kind: KindAchievement, Metric: MetricVocabWords, Threshold: 25},
	{ID: "vocab_words_100", Kind: KindAchievement, Metric: MetricVocabWords, Threshold: 100},
	{ID: "vocab_words_500", Kind: KindAchievement, Metric: MetricVocabWords, Threshold: 500},
	{ID: "diary_entries_7", Kind: KindAchievement, Metric: MetricDiaryEntries, Threshold: 7},
	{ID: "diary_entries_30", Kind: KindAchievement, Metric: MetricDiaryEntries, Threshold: 30},
	{ID: "chat_messages_25", Kind: KindAchievement, Metric: MetricChatMessages, Threshold: 25},
}

type Status struct {
	Item
	Value int
	Done  bool
	Acked bool
}

func ValidItemID(id string) bool {
	for _, it := range Items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// Merge keeps every metric at its high-water mark: values only grow,
// so deleting words or quests never rolls progress back.
func Merge(stored, live map[string]int) map[string]int {
	out := make(map[string]int, len(Metrics))
	for _, m := range Metrics {
		v := stored[m]
		if live[m] > v {
			v = live[m]
		}
		out[m] = v
	}
	return out
}

func Increased(stored, merged map[string]int) map[string]int {
	out := map[string]int{}
	for m, v := range merged {
		if v > stored[m] {
			out[m] = v
		}
	}
	return out
}

func Statuses(watermarks map[string]int, acked map[string]bool) []Status {
	out := make([]Status, 0, len(Items))
	for _, it := range Items {
		v := watermarks[it.Metric]
		out = append(out, Status{Item: it, Value: v, Done: v >= it.Threshold, Acked: acked[it.ID]})
	}
	return out
}
