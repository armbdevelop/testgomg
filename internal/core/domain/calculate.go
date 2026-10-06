package domain

import (
	"math"
	"time"
)

const (
	monthsInYear         = 12
	deductionPropertyCap = 2_000_000
	deductionInterestCap = 3_000_000
	deductionRate        = 0.13
	incomePaymentRatio   = 0.4
	kopecksInRuble       = 100
)

func Calculate(profile MortgageProfile) (calc MortgageCalculation) {
	calc.UserID = profile.UserID

	motherCapital := 0.0
	if profile.MatCapitalIncluded && profile.MatCapitalAmount != nil {
		motherCapital = *profile.MatCapitalAmount
	}

	loan := profile.PropertyPrice - profile.DownPaymentAmount - motherCapital
	n := profile.MortgageTermYears * monthsInYear
	i := profile.InterestRate / monthsInYear / kopecksInRuble

	payment, overpayment := annuity(loan, i, n)

	total := payment * float64(n)

	taxDeduction := min(profile.PropertyPrice, deductionPropertyCap)*deductionRate +
		min(overpayment, deductionInterestCap)*deductionRate
	remainder := loan

	calc.PaymentSchedule = make(PaymentSchedule)

	now := time.Now()
	start := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)

	for k := range n {
		if remainder <= 0 {
			break
		}

		proc := remainder * i
		body := payment - proc
		remainder -= body
		date := start.AddDate(0, k, 0)

		year := date.Format("2006")
		month := date.Format("01")

		if calc.PaymentSchedule[year] == nil {
			calc.PaymentSchedule[year] = make(map[string]MortgagePayment)
		}

		calc.PaymentSchedule[year][month] = MortgagePayment{
			TotalPayment:                round2(payment),
			RepaymentOfMortgageBody:     round2(body),
			RepaymentOfMortgageInterest: round2(proc),
			MortgageBalance:             round2(remainder),
		}
	}

	_, overpaymentNoMC := annuity(loan+motherCapital, i, n)

	calc.SavingsDueMotherCapital = round2(overpaymentNoMC - overpayment)
	calc.MonthlyPayment = round2(payment)
	calc.TotalPayment = round2(total)
	calc.TotalOverpaymentAmount = round2(overpayment)
	calc.PossibleTaxDeduction = round2(taxDeduction)
	calc.RecommendedIncome = round2(payment / incomePaymentRatio)

	return calc
}

func annuity(loan, monthlyRate float64, months int) (payment, overpayment float64) {
	pow := math.Pow(1+monthlyRate, float64(months))
	payment = loan * monthlyRate * pow / (pow - 1)
	overpayment = payment*float64(months) - loan

	return payment, overpayment
}

func round2(x float64) float64 {
	return math.Round(x*kopecksInRuble) / kopecksInRuble
}
