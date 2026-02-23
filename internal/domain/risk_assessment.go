package domain

type RiskAssessment struct {
	IsBlocked       bool
	ConfidenceScore int `json:"confidence_score"`
	Reason          string
	AIPushMessage   string
}
