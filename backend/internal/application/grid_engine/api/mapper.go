package api

import (
	usecases "github.com/els/backend/internal/application/grid_engine/use_cases"
	"github.com/els/backend/internal/domain/grid"
)

func toDescribeGridOutput(r usecases.DescribeResult) DescribeGridOutput {
	return DescribeGridOutput{
		SchemaVersion: r.SchemaVersion,
		Columns:       toColumnsDTO(r.Columns),
		Sources:       toSourcesDTO(r.Sources),
		Rows:          toRowsDTO(r.Rows),
		Total:         r.Total,
		Limit:         r.Limit,
		Offset:        r.Offset,
		RefsHydrated:  toRefsHydratedDTO(r.RefsHydrated),
		GeneratedAt:   r.GeneratedAt,
	}
}

func toApplyGridOutput(r usecases.ApplyResult) ApplyGridOutput {
	out := ApplyGridOutput{
		SchemaVersion: r.SchemaVersion,
		Applied:       make([]OpResultDTO, 0, len(r.Applied)),
		Failed:        make([]OpErrorDTO, 0, len(r.Failed)),
	}
	for _, a := range r.Applied {
		out.Applied = append(out.Applied, OpResultDTO{
			Index:       a.Index,
			Kind:        string(a.Kind),
			TempID:      a.TempID,
			ID:          a.ID,
			BaseVersion: a.BaseVersion,
		})
	}
	for _, f := range r.Failed {
		out.Failed = append(out.Failed, OpErrorDTO{
			Index:   f.Index,
			TempID:  f.TempID,
			ID:      f.ID,
			Code:    f.Code,
			Field:   f.Field,
			Message: f.Message,
		})
	}
	return out
}

func toLookupGridOutput(r usecases.LookupResult, inputs []LookupQueryDTO) LookupGridOutput {
	out := LookupGridOutput{Queries: make([]LookupQueryResultDTO, 0, len(r.Queries))}
	for i, q := range r.Queries {
		dto := LookupQueryResultDTO{Source: string(q.Source), NextCursor: q.NextCursor}
		if len(q.Items) > 0 {
			dto.Items = make([]LookupItemDTO, 0, len(q.Items))
			for _, it := range q.Items {
				dto.Items = append(dto.Items, LookupItemDTO{Key: it.Key, Label: it.Label})
			}
		}
		if len(q.Resolutions) > 0 || len(q.Unresolved) > 0 {
			var requested []string
			if i < len(inputs) {
				requested = inputs[i].Values
			}
			dto.Resolutions = toResolutionsDTO(q, requested)
		}
		out.Queries = append(out.Queries, dto)
	}
	return out
}

func toResolutionsDTO(q usecases.LookupQueryResult, requested []string) []LookupResolutionDTO {
	index := make(map[string]LookupResolutionDTO, len(q.Resolutions)+len(q.Unresolved))
	for _, r := range q.Resolutions {
		index[r.Input] = LookupResolutionDTO{
			Input:     r.Input,
			Key:       r.Key,
			Label:     r.Label,
			MatchedBy: string(r.MatchedBy),
			Resolved:  true,
		}
	}
	for _, v := range q.Unresolved {
		index[v] = LookupResolutionDTO{Input: v, Resolved: false}
	}

	if len(requested) == 0 {
		out := make([]LookupResolutionDTO, 0, len(index))
		for _, v := range index {
			out = append(out, v)
		}
		return out
	}

	out := make([]LookupResolutionDTO, 0, len(requested))
	for _, v := range requested {
		if r, ok := index[v]; ok {
			out = append(out, r)
			continue
		}
		out = append(out, LookupResolutionDTO{Input: v, Resolved: false})
	}
	return out
}

func fromLookupQueryDTO(in []LookupQueryDTO) []usecases.LookupQueryRequest {
	out := make([]usecases.LookupQueryRequest, 0, len(in))
	for _, q := range in {
		out = append(out, usecases.LookupQueryRequest{
			Source: grid.SourceID(q.Source),
			Values: q.Values,
			Q:      q.Q,
			Limit:  q.Limit,
			Cursor: q.Cursor,
		})
	}
	return out
}

func fromOpsDTO(in []OpDTO) []grid.Op {
	out := make([]grid.Op, 0, len(in))
	for _, d := range in {
		data := make(map[grid.ColumnID]any, len(d.Data))
		for k, v := range d.Data {
			data[grid.ColumnID(k)] = v
		}
		out = append(out, grid.Op{
			Kind:        grid.OpKind(d.Kind),
			TempID:      d.TempID,
			ID:          d.ID,
			BaseVersion: d.BaseVersion,
			Data:        data,
		})
	}
	return out
}

func toColumnsDTO(cols []grid.Column) []ColumnDTO {
	out := make([]ColumnDTO, 0, len(cols))
	for _, c := range cols {
		out = append(out, ColumnDTO{
			ID:          string(c.ID),
			Title:       c.Title,
			Type:        string(c.Type),
			Required:    c.Required,
			Readonly:    c.Readonly,
			Enum:        toEnumDTO(c.Enum),
			Ref:         toRefDTO(c.Ref),
			Constraints: toConstraintsDTO(c.Constraints),
		})
	}
	return out
}

func toEnumDTO(e []grid.EnumOption) []EnumOptionDTO {
	if len(e) == 0 {
		return nil
	}
	out := make([]EnumOptionDTO, len(e))
	for i, o := range e {
		out[i] = EnumOptionDTO{Value: o.Value, Label: o.Label}
	}
	return out
}

func toRefDTO(r *grid.RefSpec) *RefSpecDTO {
	if r == nil {
		return nil
	}
	return &RefSpecDTO{
		Source:     string(r.Source),
		KeyField:   r.KeyField,
		LabelField: r.LabelField,
		Multi:      r.Multi,
	}
}

func toConstraintsDTO(c *grid.Constraints) *ConstraintsDTO {
	if c == nil {
		return nil
	}
	return &ConstraintsDTO{
		MinLength: c.MinLength,
		MaxLength: c.MaxLength,
		Min:       c.Min,
		Max:       c.Max,
		Pattern:   c.Pattern,
		Unique:    c.Unique,
	}
}

func toRowsDTO(rows []grid.Row) []RowDTO {
	out := make([]RowDTO, 0, len(rows))
	for _, r := range rows {
		cells := make(map[string]any, len(r.Cells))
		for k, v := range r.Cells {
			cells[string(k)] = v
		}
		out = append(out, RowDTO{
			ID:          r.ID,
			BaseVersion: r.BaseVersion,
			Cells:       cells,
		})
	}
	return out
}

func toRefsHydratedDTO(in map[grid.SourceID]map[string]string) RefsHydratedDTO {
	out := make(RefsHydratedDTO, len(in))
	for k, v := range in {
		out[string(k)] = v
	}
	return out
}

func toSourcesDTO(in []grid.SourceID) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = string(s)
	}
	return out
}
