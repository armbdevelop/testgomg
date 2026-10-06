package http

import (
	"strconv"

	"github.com/armbdevelop/testgomg/internal/core/domain"
)

type calculationResponse struct {
	ID                      int64                  `json:"id"`
	UserID                  string                 `json:"userId"`
	MortgageProfileID       int64                  `json:"mortgageProfileId"`
	MonthlyPayment          string                 `json:"monthlyPayment"`
	TotalPayment            string                 `json:"totalPayment"`
	TotalOverpaymentAmount  string                 `json:"totalOverpaymentAmount"`
	PossibleTaxDeduction    string                 `json:"possibleTaxDeduction"`
	SavingsDueMotherCapital string                 `json:"savingsDueMotherCapital"`
	RecommendedIncome       string                 `json:"recommendedIncome"`
	Status                  string                 `json:"status"`
	PaymentSchedule         domain.PaymentSchedule `json:"mortgagePaymentSchedule"`
}

func newCalculationResponse(calc domain.MortgageCalculation) calculationResponse {
	money := func(x float64) string { return strconv.FormatFloat(x, 'f', 2, 64) }

	return calculationResponse{
		ID:                      calc.ID,
		UserID:                  calc.UserID,
		MortgageProfileID:       calc.MortgageProfileID,
		MonthlyPayment:          money(calc.MonthlyPayment),
		TotalPayment:            money(calc.TotalPayment),
		TotalOverpaymentAmount:  money(calc.TotalOverpaymentAmount),
		PossibleTaxDeduction:    money(calc.PossibleTaxDeduction),
		SavingsDueMotherCapital: money(calc.SavingsDueMotherCapital),
		RecommendedIncome:       money(calc.RecommendedIncome),
		Status:                  calc.Status,
		PaymentSchedule:         calc.PaymentSchedule,
	}
}
