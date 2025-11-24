package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RecommendationType represents the type of cost optimization recommendation
type RecommendationType string

const (
	RecommendationTypeDownsize   RecommendationType = "downsize"
	RecommendationTypeRightsize  RecommendationType = "rightsize"
	RecommendationTypeUnused     RecommendationType = "unused"
	RecommendationTypeOverprovisioned RecommendationType = "overprovisioned"
)

// RecommendationSeverity represents how critical a recommendation is
type RecommendationSeverity string

const (
	RecommendationSeverityLow    RecommendationSeverity = "low"
	RecommendationSeverityMedium RecommendationSeverity = "medium"
	RecommendationSeverityHigh   RecommendationSeverity = "high"
)

// CostRecommendation represents a cost optimization recommendation
type CostRecommendation struct {
	ID                uuid.UUID              `json:"id"`
	NodeID            uuid.UUID              `json:"node_id"`
	NodeName          string                 `json:"node_name"`
	NodeType          string                 `json:"node_type"`
	Type              RecommendationType     `json:"type"`
	Severity          RecommendationSeverity `json:"severity"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	CurrentCost       decimal.Decimal        `json:"current_cost"`
	PotentialSavings  decimal.Decimal        `json:"potential_savings"`
	Currency          string                 `json:"currency"`
	Metric            string                 `json:"metric"`
	CurrentValue      decimal.Decimal        `json:"current_value"`
	PeakValue         decimal.Decimal        `json:"peak_value"`
	AverageValue      decimal.Decimal        `json:"average_value"`
	UtilizationPercent decimal.Decimal       `json:"utilization_percent"`
	RecommendedAction string                 `json:"recommended_action"`
	AnalysisPeriod    string                 `json:"analysis_period"`
	StartDate         time.Time              `json:"start_date"`
	EndDate           time.Time              `json:"end_date"`
	CreatedAt         time.Time              `json:"created_at"`
}

// UsageMetrics represents usage metrics for a node
type UsageMetrics struct {
	NodeID       uuid.UUID       `json:"node_id"`
	Metric       string          `json:"metric"`
	CurrentValue decimal.Decimal `json:"current_value"`
	PeakValue    decimal.Decimal `json:"peak_value"`
	AverageValue decimal.Decimal `json:"average_value"`
	MinValue     decimal.Decimal `json:"min_value"`
	Unit         string          `json:"unit"`
	StartDate    time.Time       `json:"start_date"`
	EndDate      time.Time       `json:"end_date"`
}

