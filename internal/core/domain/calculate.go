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
	motherCapital := 0.0
	if profile.MatCapitalIncluded && profile.MatCapitalAmount != nil {
		motherCapital = *profile.MatCapitalAmount
	}

	loan := profile.PropertyPrice - profile.DownPaymentAmount - motherCapital // Сумма кредита
	n := profile.MortgageTermYears * monthsInYear                             // месяцев
	i := profile.InterestRate / monthsInYear / kopecksInRuble                 // месячная ставка

	payment, overpayment := annuity(loan, i, n) // аннуитетный платеж и переплата

	total := payment * float64(n) // общая сумма выплат

	taxDeduction := min(profile.PropertyPrice, deductionPropertyCap)*deductionRate +
		min(overpayment, deductionInterestCap)*deductionRate
	remainder := loan

	calc.PaymentSchedule = make(PaymentSchedule)

	start := time.Now().AddDate(0, 1, 0) // первый платёж в следующем месяце

	for k := range n {
		if remainder <= 0 {
			break
		}

		proc := remainder * i
		body := payment - proc
		remainder -= body
		date := start.AddDate(0, k, 0)

		year := date.Format("2006") // "2026"
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

	// Переплата в сценарии без маткапитала (кредит был бы на него больше).
	_, overpaymentNoMC := annuity(loan+motherCapital, i, n)

	// Экономия от маткапитала это разница переплат двух сценариев.
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
	// Аннуитет платёж одинаковый каждый месяц, из формулы 4 в ТЗ.
	payment = loan * monthlyRate * pow / (pow - 1)
	overpayment = payment*float64(months) - loan

	return payment, overpayment
}

func round2(x float64) float64 {
	return math.Round(x*kopecksInRuble) / kopecksInRuble
}
