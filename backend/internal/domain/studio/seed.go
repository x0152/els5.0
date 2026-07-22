package studio

type SeedItem struct {
	Text              string
	Transcription     string
	Translation       string
	Explanation       string
	ExplanationNative string
	Example           string
}

type SeedArea struct {
	Title string
	Icon  string
	Items []SeedItem
}

var Seed = []SeedArea{
	{
		Title: "Small talk",
		Icon:  "coffee",
		Items: []SeedItem{
			{
				Text:              "How's it going?",
				Transcription:     "haʊz ɪt ˈɡəʊɪŋ",
				Translation:       "Как дела?",
				Explanation:       "A casual way to ask how someone is doing when you greet them.",
				ExplanationNative: "Неформальный способ спросить, как у человека дела, при встрече.",
				Example:           "Hey Tom, how's it going? I haven't seen you all week.",
			},
			{
				Text:              "I can't complain.",
				Transcription:     "aɪ kɑːnt kəmˈpleɪn",
				Translation:       "Не жалуюсь.",
				Explanation:       "A relaxed answer meaning things are going fine, nothing bad to report.",
				ExplanationNative: "Спокойный ответ: всё в порядке, жаловаться не на что.",
				Example:           "Work is busy but I can't complain — things are going well.",
			},
			{
				Text:              "It's been ages!",
				Transcription:     "ɪts biːn ˈeɪdʒɪz",
				Translation:       "Сто лет не виделись!",
				Explanation:       "Say this when you meet someone you haven't seen for a very long time.",
				ExplanationNative: "Так говорят при встрече с человеком, которого очень давно не видели.",
				Example:           "Anna! It's been ages! When did we last meet?",
			},
			{
				Text:              "What have you been up to?",
				Transcription:     "wɒt həv ju biːn ʌp tuː",
				Translation:       "Чем ты занимался в последнее время?",
				Explanation:       "A friendly question about what someone has been doing lately.",
				ExplanationNative: "Дружеский вопрос о том, чем человек занимался в последнее время.",
				Example:           "So, what have you been up to since you moved?",
			},
			{
				Text:              "Let's grab a coffee sometime.",
				Transcription:     "lɛts ɡræb ə ˈkɒfi ˈsʌmtaɪm",
				Translation:       "Давай как-нибудь выпьем кофе.",
				Explanation:       "A casual invitation to meet up, without fixing an exact time.",
				ExplanationNative: "Непринуждённое предложение встретиться, без точного времени.",
				Example:           "It was great running into you — let's grab a coffee sometime.",
			},
		},
	},
	{
		Title: "Work",
		Icon:  "briefcase",
		Items: []SeedItem{
			{
				Text:              "I'd like to schedule a follow-up call.",
				Transcription:     "aɪd laɪk tə ˈʃɛdjuːl ə ˈfɒləʊʌp kɔːl",
				Translation:       "Я бы хотел назначить повторный звонок.",
				Explanation:       "A polite way to suggest another call to continue the discussion.",
				ExplanationNative: "Вежливый способ предложить ещё один звонок, чтобы продолжить обсуждение.",
				Example:           "Great talking to you — I'd like to schedule a follow-up call next week.",
			},
			{
				Text:              "Let's touch base early next week.",
				Transcription:     "lɛts tʌtʃ beɪs ˈɜːli nɛkst wiːk",
				Translation:       "Давайте свяжемся в начале следующей недели.",
				Explanation:       "'Touch base' means to contact someone briefly to share updates.",
				ExplanationNative: "'Touch base' — коротко связаться, чтобы обменяться новостями.",
				Example:           "I'll send the draft on Friday, and let's touch base early next week.",
			},
			{
				Text:              "Could you elaborate on that?",
				Transcription:     "kʊd ju ɪˈlæbəreɪt ɒn ðæt",
				Translation:       "Не могли бы вы рассказать об этом подробнее?",
				Explanation:       "A polite request to explain something in more detail.",
				ExplanationNative: "Вежливая просьба объяснить что-то подробнее.",
				Example:           "That's an interesting point — could you elaborate on that?",
			},
			{
				Text:              "I'll keep you posted.",
				Transcription:     "aɪl kiːp ju ˈpəʊstɪd",
				Translation:       "Буду держать вас в курсе.",
				Explanation:       "A promise to share news and updates as soon as you have them.",
				ExplanationNative: "Обещание сообщать новости, как только они появятся.",
				Example:           "We're still waiting for the results, but I'll keep you posted.",
			},
			{
				Text:              "to hit the ground running",
				Transcription:     "tə hɪt ðə ɡraʊnd ˈrʌnɪŋ",
				Translation:       "сразу активно включиться в работу",
				Explanation:       "An idiom: to start something new with full energy, without a slow warm-up.",
				ExplanationNative: "Идиома: начать новое дело сразу в полную силу, без раскачки.",
				Example:           "She knows the industry well, so she'll hit the ground running.",
			},
		},
	},
	{
		Title: "Travel",
		Icon:  "plane",
		Items: []SeedItem{
			{
				Text:              "Could you tell me how to get to the city centre?",
				Transcription:     "kʊd ju tɛl mi haʊ tə ɡɛt tə ðə ˈsɪti ˈsɛntə",
				Translation:       "Не подскажете, как добраться до центра города?",
				Explanation:       "A polite way to ask a stranger for directions.",
				ExplanationNative: "Вежливый способ спросить дорогу у незнакомого человека.",
				Example:           "Excuse me, could you tell me how to get to the city centre from here?",
			},
			{
				Text:              "I'd like to check in, please.",
				Transcription:     "aɪd laɪk tə tʃɛk ɪn pliːz",
				Translation:       "Я бы хотел заселиться, пожалуйста.",
				Explanation:       "Say this at a hotel reception when you arrive to get your room.",
				ExplanationNative: "Так говорят на стойке отеля по прибытии, чтобы получить номер.",
				Example:           "Good evening! I'd like to check in, please — the booking is under Smith.",
			},
			{
				Text:              "Is breakfast included?",
				Transcription:     "ɪz ˈbrɛkfəst ɪnˈkluːdɪd",
				Translation:       "Завтрак включён?",
				Explanation:       "A question to check whether breakfast is part of the room price.",
				ExplanationNative: "Вопрос, входит ли завтрак в стоимость номера.",
				Example:           "The room looks great — is breakfast included in the price?",
			},
			{
				Text:              "Could I get the bill, please?",
				Transcription:     "kʊd aɪ ɡɛt ðə bɪl pliːz",
				Translation:       "Можно счёт, пожалуйста?",
				Explanation:       "A polite way to ask for the bill in a restaurant or café.",
				ExplanationNative: "Вежливый способ попросить счёт в ресторане или кафе.",
				Example:           "Everything was delicious. Could I get the bill, please?",
			},
			{
				Text:              "What time does the last train leave?",
				Transcription:     "wɒt taɪm dʌz ðə lɑːst treɪn liːv",
				Translation:       "Во сколько уходит последний поезд?",
				Explanation:       "Ask this to find out the departure time of the final train of the day.",
				ExplanationNative: "Вопрос о времени отправления последнего поезда за день.",
				Example:           "We should check what time the last train leaves so we don't get stuck.",
			},
		},
	},
}
