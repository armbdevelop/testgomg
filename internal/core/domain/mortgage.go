package domain

type PropertyType string
type PaymentSchedule map[string]map[string]MortgagePayment

const (
	ApartmentInNewBuilding       PropertyType = "apartment_in_new_building"
	ApartmentInSecondaryBuilding PropertyType = "apartment_in_secondary_building"
	House                        PropertyType = "house"
	HouseWithLandPlot            PropertyType = "house_with_land_plot"
	LandPlot                     PropertyType = "land_plot"
	Other                        PropertyType = "other"
)

type MortgageProfile struct {
	ID                 int64        `json:"id"`
	UserID             string       `json:"userId"`
	PropertyPrice      float64      `json:"propertyPrice" validate:"required,gt=0"`
	PropertyType       PropertyType `json:"propertyType" validate:"required"`
	DownPaymentAmount  float64      `json:"downPaymentAmount" validate:"gte=0"`
	MatCapitalAmount   *float64     `json:"matCapitalAmount"`
	MatCapitalIncluded bool         `json:"matCapitalIncluded"`
	MortgageTermYears  int          `json:"mortgageTermYears" validate:"required,gt=0"`
	InterestRate       float64      `json:"interestRate" validate:"required,gt=0"` // годовая, в процентах (8 = 8%)
}

type MortgagePayment struct {
	TotalPayment                float64 `json:"totalPayment"`
	RepaymentOfMortgageBody     float64 `json:"repaymentOfMortgageBody"`
	RepaymentOfMortgageInterest float64 `json:"repaymentOfMortgageInterest"`
	MortgageBalance             float64 `json:"mortgageBalance"`
}

type MortgageCalculation struct {
	ID                      int64           `json:"id"`
	UserID                  string          `json:"userId"`
	MortgageProfileID       int64           `json:"mortgageProfileId"`
	MonthlyPayment          float64         `json:"monthlyPayment"`
	TotalPayment            float64         `json:"totalPayment"`
	TotalOverpaymentAmount  float64         `json:"totalOverpaymentAmount"`
	PossibleTaxDeduction    float64         `json:"possibleTaxDeduction"`
	SavingsDueMotherCapital float64         `json:"savingsDueMotherCapital"`
	RecommendedIncome       float64         `json:"recommendedIncome"`
	PaymentSchedule         PaymentSchedule `json:"mortgagePaymentSchedule"`
}

func (p PropertyType) Valid() bool {
	switch p {
	case ApartmentInNewBuilding,
		ApartmentInSecondaryBuilding,
		House,
		HouseWithLandPlot,
		LandPlot,
		Other:
		return true
	default:
		return false
	}
}
