package api

import (
	"encoding/json"

	usecases "github.com/els/backend/internal/application/core/use_cases"
	"github.com/els/backend/internal/domain/core"
)

func toIngestEventsCommand(in *IngestInput) (usecases.IngestEventsCommand, error) {
	events := make([]core.RawEvent, 0, len(in.Body.Events))
	for _, e := range in.Body.Events {
		events = append(events, toRawEvent(e))
	}
	return usecases.IngestEventsCommand{Events: events}, nil
}

func toIngestEventsOutput(r usecases.IngestEventsResult) IngestOutput {
	return IngestOutput{Accepted: r.Accepted}
}

func toMarkUnclearCommand(in *MarkUnclearInput) usecases.MarkUnclearCommand {
	return usecases.MarkUnclearCommand{Event: toRawEvent(in.Body)}
}

func toListEventsQuery(in *ListInput) (usecases.ListEventsQuery, error) {
	status, err := core.ParseStatus(in.Status)
	if err != nil {
		return usecases.ListEventsQuery{}, err
	}
	return usecases.ListEventsQuery{Status: status}, nil
}

func toListEventsOutput(r usecases.ListEventsResult) ListOutput {
	switch r.Status {
	case core.StatusProcessed:
		return ListOutput{Events: eventViews(r.Processed)}
	case core.StatusAll:
		return ListOutput{Events: append(eventViews(r.Processed), rawViews(r.Raws)...)}
	default:
		return ListOutput{Events: rawViews(r.Raws)}
	}
}

func toRawEvent(e EventEnvelope) core.RawEvent {
	r := core.RawEvent{
		ClientID: e.ClientID,
		Skill:    e.Skill,
		Text:     e.Text,
		Target:   e.Target,
		Outcome:  e.Outcome,
		Context:  e.Context,
		Source:   marshal(e.Source),
		Meta:     marshal(e.Meta),
	}
	if e.OccurredAt != nil {
		r.OccurredAt = *e.OccurredAt
	}
	return r
}

func rawViews(raws []core.RawEvent) []EventView {
	out := make([]EventView, 0, len(raws))
	for _, r := range raws {
		status := r.Status
		if status == string(core.StatusProcessing) {
			status = string(core.StatusPending)
		}
		v := EventView{
			ID: r.ID, ClientID: r.ClientID, Status: status, Skill: r.Skill, Text: r.Text, Target: r.Target,
			Outcome: r.Outcome, Context: r.Context, Source: jsonToMap(r.Source), Meta: jsonToMap(r.Meta),
			OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt,
		}
		if r.Error != "" {
			v.Error = map[string]any{"reason": r.Error}
		}
		out = append(out, v)
	}
	return out
}

func eventViews(events []core.Event) []EventView {
	out := make([]EventView, 0, len(events))
	for _, e := range events {
		out = append(out, EventView{
			ID: e.ID, RawEventID: e.RawEventID, ClientID: e.ClientID, Status: string(core.StatusProcessed), Skill: e.Type,
			Context: e.Context, Action: e.Action, Lemma: e.Lemma, POS: e.POS, GrammarKey: e.GrammarKey,
			Outcome: e.Outcome, Error: errorView(e.Error), Source: jsonToMap(e.Source), Meta: jsonToMap(e.Meta),
			OccurredAt: e.OccurredAt, CreatedAt: e.CreatedAt,
		})
	}
	return out
}

func toCatalogOutput(r usecases.ListCatalogResult) CatalogOutput {
	words := make([]WordView, 0, len(r.Words))
	for _, w := range r.Words {
		words = append(words, WordView{
			ID: w.ID, Key: w.Key, Lemma: w.Lemma, POS: w.POS, Type: w.Type,
			CEFR: w.CEFR, Frequency: w.Frequency, Enriched: w.Enriched, CreatedAt: w.CreatedAt,
		})
	}
	rules := make([]GrammarRuleView, 0, len(r.Rules))
	for _, g := range r.Rules {
		rules = append(rules, GrammarRuleView{
			ID: g.ID, Key: g.Key, ParentKey: g.ParentKey, Title: g.Title,
			CEFRLevel: g.CEFRLevel, Enriched: g.Enriched, CreatedAt: g.CreatedAt,
		})
	}
	return CatalogOutput{Words: words, Rules: rules}
}

func toDictionariesOutput(d map[string][]core.DictEntry) DictionariesOutput {
	out := make(map[string][]DictEntryView, len(d))
	for k, entries := range d {
		views := make([]DictEntryView, 0, len(entries))
		for _, e := range entries {
			views = append(views, DictEntryView{Value: e.Value, Label: e.Label, Icon: e.Icon})
		}
		out[k] = views
	}
	return DictionariesOutput{Dictionaries: out}
}

func jsonToMap(b json.RawMessage) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(b, &m) != nil || len(m) == 0 {
		return nil
	}
	return m
}

func marshal(m map[string]any) json.RawMessage {
	if len(m) == 0 {
		return nil
	}
	b, _ := json.Marshal(m)
	return b
}

func errorView(e *core.EventError) map[string]any {
	if e == nil {
		return nil
	}
	m := map[string]any{}
	if e.Name != "" {
		m["name"] = e.Name
	}
	if e.Sentence != "" {
		m["sentence"] = e.Sentence
	}
	if e.Fragment != "" {
		m["fragment"] = e.Fragment
	}
	if e.Correction != "" {
		m["correction"] = e.Correction
	}
	if e.Description != "" {
		m["description"] = e.Description
	}
	return m
}
