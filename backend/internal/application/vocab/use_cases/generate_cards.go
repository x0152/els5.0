package usecases

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/illustration"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

const (
	cardsDeckSize     = 10
	cardsOptionCount  = 4
	cardsMinWords     = 4
	cardsPoolLimit    = 500
	cardsLearningPart = 6
	cardsNewPart      = 3
)

type GenerateCardsUseCase struct {
	units   vocab.Repository
	storage media.Storage
	urls    media.PublicURL
	bucket  string
}

func NewGenerateCardsUseCase(units vocab.Repository, storage media.Storage, urls media.PublicURL, bucket string) *GenerateCardsUseCase {
	return &GenerateCardsUseCase{units: units, storage: storage, urls: urls, bucket: bucket}
}

func (uc *GenerateCardsUseCase) Execute(ctx context.Context, actor *iam.Actor, imagesOnly bool) ([]vocab.Card, error) {
	// 1. Load the whole pool: it supplies both card words and distractors.
	pool, _, err := uc.units.List(ctx, actor.AccountID().String(), vocab.ListFilter{Limit: cardsPoolLimit})
	if err != nil {
		return nil, err
	}
	if len(pool) < cardsMinWords {
		return nil, shared.Validation(fmt.Errorf("words: need at least %d words to practice cards", cardsMinWords))
	}

	// 2. Order candidates by status quota: mostly learning, some new, a bit of learned review.
	candidates := orderCandidates(pool)

	// 3. Fill the deck; with imagesOnly skip words without a generated illustration.
	cards := make([]vocab.Card, 0, cardsDeckSize)
	for _, u := range candidates {
		url := uc.imageURL(ctx, u.Text)
		if imagesOnly && url == "" {
			continue
		}
		card := vocab.Card{Unit: u, Mode: vocab.ModeFor(u), Direction: vocab.CardDirectionWord, ImageURL: url}
		if card.Mode == vocab.CardModeChoice {
			if translations := translationOptions(u, pool); translations != nil && !imagesOnly && actor.Account().ShowTranslations() && rand.Intn(3) == 0 {
				card.Direction = vocab.CardDirectionTranslation
				card.Options = translations
				card.ImageURL = ""
			} else {
				card.Options = options(u, pool)
			}
		}
		cards = append(cards, card)
		if len(cards) == cardsDeckSize {
			break
		}
	}
	if len(cards) == 0 {
		return nil, shared.Validation(fmt.Errorf("words: no words with generated images yet"))
	}

	// 4. Mix statuses within the deck.
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	return cards, nil
}

func orderCandidates(pool []vocab.Unit) []vocab.Unit {
	groups := map[vocab.Status][]vocab.Unit{}
	for _, u := range pool {
		groups[u.Status] = append(groups[u.Status], u)
	}
	for _, g := range groups {
		rand.Shuffle(len(g), func(i, j int) { g[i], g[j] = g[j], g[i] })
	}
	out := make([]vocab.Unit, 0, len(pool))
	take := func(status vocab.Status, n int) {
		g := groups[status]
		if n > len(g) {
			n = len(g)
		}
		out = append(out, g[:n]...)
		groups[status] = g[n:]
	}
	take(vocab.StatusLearning, cardsLearningPart)
	take(vocab.StatusNew, cardsNewPart)
	take(vocab.StatusLearned, cardsDeckSize-cardsLearningPart-cardsNewPart)
	for _, status := range []vocab.Status{vocab.StatusLearning, vocab.StatusNew, vocab.StatusLearned} {
		out = append(out, groups[status]...)
	}
	return out
}

func options(u vocab.Unit, pool []vocab.Unit) []string {
	sameKind := make([]string, 0, len(pool))
	others := make([]string, 0, len(pool))
	for _, p := range pool {
		if p.ID == u.ID {
			continue
		}
		if p.Kind == u.Kind {
			sameKind = append(sameKind, p.Text)
		} else {
			others = append(others, p.Text)
		}
	}
	rand.Shuffle(len(sameKind), func(i, j int) { sameKind[i], sameKind[j] = sameKind[j], sameKind[i] })
	rand.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	opts := append(sameKind, others...)
	if len(opts) > cardsOptionCount-1 {
		opts = opts[:cardsOptionCount-1]
	}
	opts = append(opts, u.Text)
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	return opts
}

func translationOptions(u vocab.Unit, pool []vocab.Unit) []string {
	if u.Translation == "" {
		return nil
	}
	seen := map[string]bool{u.Translation: true}
	distractors := make([]string, 0, len(pool))
	for _, p := range pool {
		if p.ID == u.ID || p.Translation == "" || seen[p.Translation] {
			continue
		}
		seen[p.Translation] = true
		distractors = append(distractors, p.Translation)
	}
	if len(distractors) < cardsOptionCount-1 {
		return nil
	}
	rand.Shuffle(len(distractors), func(i, j int) { distractors[i], distractors[j] = distractors[j], distractors[i] })
	opts := append(distractors[:cardsOptionCount-1], u.Translation)
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	return opts
}

func (uc *GenerateCardsUseCase) imageURL(ctx context.Context, text string) string {
	id := illustration.Key(vocab.ImagePrompt(text), "square")
	path, err := media.NewPath(uc.bucket + "/" + illustration.Filename(id))
	if err != nil {
		return ""
	}
	reader, _, err := uc.storage.Get(ctx, path)
	if err != nil {
		return ""
	}
	_ = reader.Close()
	return uc.urls.Build(path)
}
