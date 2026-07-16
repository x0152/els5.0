package speech

type PhonemeInfo struct {
	Symbol      string
	Kind        string
	Examples    string
	Description string
	Pitfall     string
}

var PhonemeGuide = []PhonemeInfo{
	{"i", "vowel", "see, tree, be", "Long tense vowel. Lips slightly spread, tongue high and forward, muscles tense.", "Keep it long and tense; the short relaxed variant turns 'sheep' into 'ship'."},
	{"ɪ", "vowel", "ship, big, sit", "Short lax vowel between Russian 'и' and 'ы'. Tongue slightly lower and further back than for /i/.", "Russian speakers often say a tense 'и', which sounds like /i/ and merges 'live' with 'leave'."},
	{"ɛ", "vowel", "bed, head, said", "Short vowel similar to Russian 'э', tongue mid-front, mouth slightly open.", "Do not open the mouth too wide, or it drifts toward /æ/."},
	{"æ", "vowel", "cat, bad, hand", "Open front vowel. Mouth open noticeably wider than for 'э', jaw drops, corners of lips pulled aside.", "Russian has no /æ/; replacing it with 'э' makes 'bad' sound like 'bed'."},
	{"ɑ", "vowel", "father, hot, car", "Open back vowel. Mouth wide open, tongue low and pulled back, like showing your throat to a doctor.", "Deeper and further back than Russian 'а'."},
	{"ɔ", "vowel", "law, thought, door", "Open-mid back rounded vowel. Lips slightly rounded, tongue back.", "Russian 'о' is more rounded and closed; English /ɔ/ is more open."},
	{"ʊ", "vowel", "good, put, book", "Short lax vowel. Lips loosely rounded, tongue back but relaxed.", "Do not tense the lips as in Russian 'у'; 'pull' must differ from 'pool'."},
	{"u", "vowel", "food, blue, two", "Long tense vowel. Lips firmly rounded, tongue high and back.", "Keep it long; a short version collapses 'fool' into 'full'."},
	{"ʌ", "vowel", "cup, love, but", "Short mid-open vowel, like a quick relaxed 'а'.", "Close to unstressed Russian 'а' in 'вода́' — short and neutral, not 'у' despite the spelling."},
	{"ə", "vowel", "about, sofa, banana", "Schwa: fully relaxed neutral vowel in unstressed syllables. Mouth barely open.", "Russian speakers over-articulate unstressed vowels; schwa should be almost inaudible."},
	{"ɚ", "vowel", "teacher, doctor, better", "Schwa colored by an /r/: tongue tip curls slightly back while saying a neutral vowel.", "Do not roll the r; the tongue never touches the roof of the mouth."},
	{"ɝ", "vowel", "bird, work, learn", "Stressed r-colored vowel. Lips neutral, tongue bunched with the tip curled back.", "Avoid Russian 'ё/о': 'work' is not 'ворк', there is no /o/ sound in it."},
	{"eɪ", "diphthong", "day, name, rain", "Starts at /e/ and glides to /ɪ/. One smooth movement.", "Do not flatten it into a plain 'э': 'late' must not sound like 'let'."},
	{"aɪ", "diphthong", "my, time, high", "Starts open at /a/ and glides up to /ɪ/.", "Similar to Russian 'ай' but with a wider start."},
	{"ɔɪ", "diphthong", "boy, choice, noise", "Starts at rounded /ɔ/ and glides to /ɪ/.", "Similar to Russian 'ой'; keep the first part rounded."},
	{"aʊ", "diphthong", "now, house, out", "Starts open at /a/ and glides to rounded /ʊ/.", "Similar to Russian 'ау' said quickly as one syllable."},
	{"oʊ", "diphthong", "go, home, know", "Starts at /o/ and glides to /ʊ/. Lips round progressively.", "A plain Russian 'о' sounds foreign; the glide to 'у' is what makes it English."},
	{"p", "consonant", "pen, apple, stop", "Voiceless stop with aspiration at the start of stressed syllables: a small puff of air.", "Russian 'п' has no aspiration; without the puff 'pat' can be heard as 'bat'."},
	{"b", "consonant", "big, baby, job", "Voiced stop, like Russian 'б'.", "Do not devoice at word end: 'job' must not become 'jop'."},
	{"t", "consonant", "top, letter, cat", "Voiceless stop. Tongue tip on the alveolar ridge (the bump behind upper teeth), not on the teeth. Aspirated in stressed positions.", "Russian 'т' is dental and unaspirated; move the tongue back and add the puff."},
	{"d", "consonant", "dog, ladder, bed", "Voiced alveolar stop: tongue on the ridge, not the teeth.", "Do not devoice at word end: 'bed' must not become 'bet'."},
	{"k", "consonant", "cat, school, back", "Voiceless velar stop, aspirated in stressed positions.", "Like Russian 'к' plus a puff of air at the start of words."},
	{"ɡ", "consonant", "go, bigger, bag", "Voiced velar stop, like Russian 'г'.", "Do not devoice at word end: 'bag' must not become 'back'."},
	{"f", "consonant", "fish, coffee, life", "Voiceless: upper teeth touch the lower lip.", "Same as Russian 'ф'."},
	{"v", "consonant", "very, seven, love", "Voiced: upper teeth on the lower lip with voice.", "Do not replace with /w/ (both lips) and do not devoice at the end: 'love' is not 'lof'."},
	{"θ", "consonant", "think, three, bath", "Voiceless 'th'. Tongue tip lightly between the teeth, blow air over it. No voice.", "Russian speakers substitute 'с' or 'ф'; the tongue must visibly touch the teeth."},
	{"ð", "consonant", "this, mother, breathe", "Voiced 'th'. Same tongue position as /θ/ but with voice.", "Substituting 'з' or 'д' is the classic mistake; keep the tongue between the teeth."},
	{"s", "consonant", "sun, city, miss", "Voiceless hissing sound, like Russian 'с'.", "Watch voicing: plural 's' after voiced sounds is /z/, not /s/."},
	{"z", "consonant", "zoo, easy, dogs", "Voiced buzzing sound, like Russian 'з'.", "Do not devoice at word end: 'eyes' must not sound like 'ice'."},
	{"ʃ", "consonant", "she, nation, fish", "Voiceless. Softer and more forward than Russian 'ш', lips slightly rounded.", "Russian 'ш' is too hard and retracted; aim between 'ш' and 'щ'."},
	{"ʒ", "consonant", "vision, pleasure, beige", "Voiced counterpart of /ʃ/.", "Softer than Russian 'ж'; do not devoice at the end."},
	{"tʃ", "consonant", "chair, teacher, watch", "Voiceless affricate: /t/ + /ʃ/ released together.", "Harder than Russian 'ч'; closer to 'тш' said as one sound."},
	{"dʒ", "consonant", "job, age, bridge", "Voiced affricate: /d/ + /ʒ/ released together.", "Russian has no 'дж' as one sound; do not split it and do not devoice at the end."},
	{"h", "consonant", "hat, hello, ahead", "Light breath of air, no friction in the throat.", "Russian 'х' is too harsh; /h/ is just a sigh."},
	{"m", "consonant", "man, summer, time", "Nasal, like Russian 'м'.", ""},
	{"n", "consonant", "no, dinner, sun", "Nasal. Tongue on the alveolar ridge, not the teeth.", ""},
	{"ŋ", "consonant", "sing, long, think", "Nasal made with the back of the tongue (as in 'к'), air through the nose. No /ɡ/ release.", "Russian speakers say 'нг' or plain 'н'; 'sing' has no audible 'г' at the end."},
	{"l", "consonant", "let, yellow, feel", "Tongue tip on the alveolar ridge. Dark (velarized) at word end.", "Between Russian hard 'л' and soft 'ль'; do not soften it before vowels."},
	{"ɹ", "consonant", "red, sorry, car", "Approximant: tongue tip curled back, never touching the roof of the mouth. No vibration.", "The rolled Russian 'р' is the most audible marker of a Russian accent; the tongue must not tap."},
	{"j", "consonant", "yes, you, few", "Glide, like Russian 'й'.", ""},
	{"w", "consonant", "we, one, quick", "Both lips rounded into a tight circle, then released. Teeth do not touch lips.", "Substituting /v/ makes 'west' sound like 'vest'; keep teeth away from lips."},
	{"ɾ", "consonant", "water, better, city", "Flap: American /t/ or /d/ between vowels, a single quick tap like a fast Russian 'р'.", "Hearing it instead of /t/ in 'water' is normal American pronunciation."},
	{"ʔ", "consonant", "uh-oh, button", "Glottal stop: a brief catch in the throat.", "Often appears before initial vowels; harmless."},
	{"x", "consonant", "loch (Scottish)", "Voiceless velar fricative — the Russian 'х'.", "If it appears instead of /h/, the /h/ was too harsh: make it a light sigh."},
	{"r", "consonant", "(trilled r)", "Trilled or tapped r, as in Russian 'р'.", "If this was heard, the tongue vibrated: English /ɹ/ never taps or rolls."},
}

var phonemeIndex = buildPhonemeIndex()

func buildPhonemeIndex() map[string]PhonemeInfo {
	idx := make(map[string]PhonemeInfo, len(PhonemeGuide))
	for _, p := range PhonemeGuide {
		idx[p.Symbol] = p
	}
	return idx
}

func LookupPhoneme(symbol string) (PhonemeInfo, bool) {
	info, ok := phonemeIndex[symbol]
	return info, ok
}
