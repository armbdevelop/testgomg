package domain

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// Контрольный пример: 10 000 000 ₽, взнос 2 000 000 ₽, 8% годовых, 20 лет
// → платёж ≈ 66 916 ₽ (сверено с аннуитетными калькуляторами).
func TestCalculateReferenceCase(t *testing.T) {
	t.Parallel()

	calc := Calculate(MortgageProfile{
		PropertyPrice:     10_000_000,
		PropertyType:      ApartmentInNewBuilding,
		DownPaymentAmount: 2_000_000,
		MortgageTermYears: 20,
		InterestRate:      8,
	})

	require.Less(t, math.Abs(calc.MonthlyPayment-66_916), 1.0)
	require.Greater(t, calc.TotalPayment, 10_000_000.0)
	// Платёж и сумма округляются независимо - расхождение до рубля на каждом шаге.
	require.Less(t, math.Abs(calc.TotalPayment-calc.MonthlyPayment*240), 3.0)
	require.Less(t, math.Abs(calc.TotalOverpaymentAmount-(calc.TotalPayment-8_000_000)), 1.0)

	// Кредит гасится полностью: минимальный баланс в графике (последний платёж) ≈ 0.
	minBalance := math.Inf(1)

	for _, months := range calc.PaymentSchedule {
		for _, payment := range months {
			minBalance = min(minBalance, payment.MortgageBalance)
		}
	}

	require.Less(t, math.Abs(minBalance), 1.0)
}

func TestCalculateMatCapitalReducesPayment(t *testing.T) {
	t.Parallel()

	matCapital := 500_000.0
	profile := MortgageProfile{
		PropertyPrice:     10_000_000,
		PropertyType:      House,
		DownPaymentAmount: 2_000_000,
		MortgageTermYears: 20,
		InterestRate:      8,
	}

	without := Calculate(profile)

	profile.MatCapitalAmount = &matCapital
	profile.MatCapitalIncluded = true
	with := Calculate(profile)

	require.Less(t, with.MonthlyPayment, without.MonthlyPayment)
	require.Positive(t, with.SavingsDueMotherCapital)
}

func TestPropertyTypeValid(t *testing.T) {
	t.Parallel()

	require.True(t, ApartmentInNewBuilding.Valid())
	require.True(t, Other.Valid())
	require.False(t, PropertyType("castle").Valid())
	require.False(t, PropertyType("").Valid())
}
