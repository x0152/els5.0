package settings

import (
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/shared"
)

type Feature string

const (
	FeatureMain     Feature = "main"
	FeatureAnalysis Feature = "analysis"
	FeatureVision   Feature = "vision"
	FeatureImage    Feature = "image"
)

func Features() []Feature {
	return []Feature{FeatureMain, FeatureAnalysis, FeatureVision, FeatureImage}
}

func ParseFeature(s string) (Feature, error) {
	f := Feature(strings.ToLower(strings.TrimSpace(s)))
	switch f {
	case FeatureMain, FeatureAnalysis, FeatureVision, FeatureImage:
		return f, nil
	default:
		return "", shared.Validation(fmt.Errorf("feature: unknown %q", s))
	}
}

type Kind string

const (
	KindOpenAI  Kind = "openai"
	KindComfyUI Kind = "comfyui"
)

func ParseKind(s string) (Kind, error) {
	k := Kind(strings.ToLower(strings.TrimSpace(s)))
	switch k {
	case "":
		return KindOpenAI, nil
	case KindOpenAI, KindComfyUI:
		return k, nil
	default:
		return "", shared.Validation(fmt.Errorf("kind: unknown %q", s))
	}
}

type AIProvider struct {
	Feature Feature
	Kind    Kind
	BaseURL string
	APIKey  string
	Model   string
	Params  map[string]string
}

func NewAIProvider(feature Feature, baseURL, apiKey, model string) AIProvider {
	return AIProvider{
		Feature: feature,
		Kind:    KindOpenAI,
		BaseURL: strings.TrimSpace(baseURL),
		APIKey:  strings.TrimSpace(apiKey),
		Model:   strings.TrimSpace(model),
	}
}

func (p AIProvider) HasKey() bool { return strings.TrimSpace(p.APIKey) != "" }

func (p AIProvider) IsEmpty() bool {
	return p.BaseURL == "" && p.APIKey == "" && p.Model == ""
}
