export interface TourInfo {
  title: string
  description: string
  features: string[]
}

export const TOURS: Record<string, TourInfo> = {
  profile: {
    title: 'Your profile',
    description:
      'Your personal settings — the whole platform adapts lessons, translations and feedback to them.',
    features: [
      'Set your name, English level and a few words about yourself',
      'Pick your native language — translations and AI explanations use it',
      'Toggle translations on or off across the whole platform',
      'Choose pronunciation strictness (Easy / Normal / Strict) for Speaking and vocab checks',
      'Upload a profile photo',
    ],
  },
  chat: {
    title: 'Assistant',
    description:
      'Your personal AI English tutor, available from any page as a side panel or full screen.',
    features: [
      'Ask anything: grammar, word meanings, translations, examples',
      'Replies stream in live — stop generation or regenerate the last answer',
      'Tap suggestion chips to continue the conversation',
      'Switch the AI model in the header',
      'Reset the context or clear the history completely',
    ],
  },
  workout: {
    title: 'Workout',
    description:
      'One guided lesson per day built from a film scene and your recent mistakes — all skills in one session.',
    features: [
      'Steps: warm-up, watch a clip, questions, speaking, dictation, reading, writing, grammar and vocab',
      'Speak lines aloud and get a pronunciation score, or type what you hear',
      'Writing and grammar steps are checked by AI with hints',
      'Tap unknown words in the reading step to look them up',
      'Keep your streak: day streak, monthly calendar and best-streak stats',
      'Replay finished steps or take “one more” lesson after completing today’s',
    ],
  },
  quest: {
    title: 'Quests',
    description:
      'Interactive story adventures where you progress by writing English in dialogues.',
    features: [
      'Create a mission: pick a genre, describe a story or go Random',
      'Talk to AI characters by typing your replies',
      'Grammar gate: fix your mistakes before the story continues',
      'See “how a native would say it” for your drafts',
      'Track goals and characters in the sidebar, restart any time',
      'Reach an ending — from perfect to abandoned',
    ],
  },
  diary: {
    title: 'Diary',
    description:
      'Write a few English sentences every day and get instant AI feedback.',
    features: [
      'A fresh prompt every day — write 3–5 sentences',
      'AI checks grammar before sending; reveal fixes and try again',
      'After sending you get a friendly reply plus language notes',
      'Warm-up exercises built from your past corrections',
      'See a diff of your draft vs the corrected version',
      'Keep a streak and browse your history',
    ],
  },
  vocab: {
    title: 'My Vocabulary',
    description:
      'Your personal word collection — every word you look up anywhere on the platform lands here.',
    features: [
      'Add words, phrases, phrasal verbs and idioms — AI normalizes them',
      'Word details: pronunciation audio, CEFR level, frequency, AI-generated image',
      'Train with spaced-repetition cards — choose or type the answer',
      'Check your pronunciation right after answering a card',
      'Practice mode: AI exercises on the words you are learning',
      'Filter by status: New / Learning / Learned',
    ],
  },
  speaking: {
    title: 'Speaking',
    description:
      'Pronunciation trainer: read sentences aloud and get feedback on every sound.',
    features: [
      'Record a sentence and get an overall score plus a per-phoneme breakdown',
      'Tap any phoneme to see how to articulate it',
      '“Explain with AI” — tips tailored to your native language',
      'Practice your weakest sounds with generated sentences',
      'Listen to the example, replay your own recording',
      'Sound guide: all vowels, diphthongs and consonants with diagrams',
    ],
  },
  writing: {
    title: 'Writing',
    description:
      'Phrase trainer: write replies in real-life situations and polish them until they sound natural.',
    features: [
      'Paste a dialogue context or let AI suggest a situation',
      'Three levels: No mistakes → Natural → Like a native',
      'AI marks the issues but never writes the answer for you',
      'Retry as many times as you need, then raise the level',
      'Cmd/Ctrl+Enter to check instantly',
    ],
  },
  reading: {
    title: 'Reading',
    description:
      'Read texts generated for your level and turn unknown words into vocabulary.',
    features: [
      'Generate a text: topic, difficulty and length of your choice',
      'Weave in the words you are currently learning',
      'Tap words you don’t know — they are saved for training',
      'Listen to the text with natural TTS',
      'Finish to see your reading speed and the list of new words',
    ],
  },
  listening: {
    title: 'Listening',
    description:
      'Dictation trainer: listen to short clips and type exactly what you hear.',
    features: [
      'Generate 3–10 clips on any topic and level, optionally with your vocab words',
      'Type what you hear — accuracy is checked word by word',
      'Replay, slow down to 0.7×, or take a first-letter hint',
      'Pick a voice or keep it random',
      'End summary shows the words you missed',
    ],
  },
  grammarbook: {
    title: 'Grammarbook',
    description:
      'An interactive grammar book: theory on the left page, exercises on the right.',
    features: [
      'Units with clear explanations and illustrations',
      'Fill-in exercises checked instantly, progress is saved',
      'Practice session mode — exercises one by one',
      'Generate extra exercise variants for any unit',
      'Generate a whole new unit from a topic you need',
      '“Continue” jumps to your next incomplete unit',
    ],
  },
  wordbook: {
    title: 'Wordbook',
    description:
      'Topic-based vocabulary book: learn word groups in context, like Vocabulary in Use.',
    features: [
      'Lessons with theory, examples and illustrations',
      'Interactive exercises with instant checking',
      'Practice session mode and extra generated variants',
      'Generate a new lesson from any topic',
      'Progress per lesson with a “Continue” shortcut',
    ],
  },
  phrasebook: {
    title: 'Phrasebook',
    description:
      'Collocations and set phrases — learn how English words naturally go together.',
    features: [
      'Lessons on collocations with theory and examples',
      'Gap-fill exercises checked instantly',
      'Practice session mode and generated exercise variants',
      'Generate a new lesson from any topic',
      'Progress tracking with a “Continue” shortcut',
    ],
  },
  essentialbook: {
    title: 'Essentialbook',
    description:
      '504 essential words: high-frequency vocabulary unit by unit.',
    features: [
      'Lessons with word lists, theory and examples',
      'Exercises for every unit with instant checking',
      'Practice session mode and extra variants',
      'Word counts per lesson to plan your pace',
      'Progress tracking with a “Continue” shortcut',
    ],
  },
  reader: {
    title: 'Reader',
    description:
      'Your own library: upload books and articles and read them with instant word lookup.',
    features: [
      'Upload books (EPUB, FB2, DOCX, TXT and more) or articles from a URL',
      'Organize into collections, edit covers and metadata',
      'Tap any word for a lookup — it is saved to your vocabulary',
      'Draggable reading line and auto-saved position',
      'What you read feeds the learning system automatically',
    ],
  },
  studio: {
    title: 'Studio',
    description:
      'Your own phrases and words, trained across every skill on one screen.',
    features: [
      'Organize phrases into areas — e.g. “Job interview” or “Travel”',
      'Add any phrase — AI prepares transcription, translation and an example',
      'Listening: play the phrase and type what you hear',
      'Speaking: record it and get a sound-by-sound score',
      'Use it: reply to an AI mini-situation with the phrase',
      'Regenerate examples and situations any time, progress per phrase',
    ],
  },
  films: {
    title: 'Films',
    description:
      'Watch films and series with interactive subtitles that feed your vocabulary.',
    features: [
      'Multiple audio and subtitle tracks — your choice is remembered',
      'Click any word in subtitles to see meaning, hear it and save it',
      'Subtitle panel: jump to any line, analyze it or ask the assistant',
      'Mark unclear lines to practice them later',
      'Resume where you left off; series remember the last episode',
      'Picture-in-picture mini-player to keep watching while you browse',
    ],
  },
}
