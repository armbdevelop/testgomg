package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/armbdevelop/testgomg/internal/core/domain"
	"github.com/armbdevelop/testgomg/internal/core/ports"
	"github.com/jmoiron/sqlx"
)

type pgRepo struct {
	db *sqlx.DB
}

func NewPGRepository(db *sqlx.DB) ports.PGRepository {
	return &pgRepo{db: db}
}

func (p *pgRepo) CreateCalculation(ctx context.Context, profile domain.MortgageProfile,
	_ domain.MortgageCalculation) (id int64, err error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("pgRepo.CreateCalculation.Begin: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("pgRepo.CreateCalculation.Rollback: %v", rbErr)
		}
	}()

	if _, err = tx.ExecContext(ctx, queryUpsertUser, profile.UserID); err != nil {
		return 0, fmt.Errorf("pgRepo.CreateCalculation.UpsertUser: %w", err)
	}

	var profileID int64
	if err = tx.GetContext(ctx, &profileID, queryInsertMortgageProfile,
		profile.UserID,
		profile.PropertyPrice,
		profile.PropertyType,
		profile.DownPaymentAmount,
		profile.MatCapitalAmount,
		profile.MatCapitalIncluded,
		profile.MortgageTermYears,
		profile.InterestRate); err != nil {
		return 0, fmt.Errorf("pgRepo.CreateCalculation.InsertProfile: %w", err)
	}

	if err = tx.GetContext(ctx, &id, queryInsertMortgageCalculation, profile.UserID, profileID); err != nil {
		return 0, fmt.Errorf("pgRepo.CreateCalculation.InsertCalculation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("pgRepo.CreateCalculation.Commit: %w", err)
	}

	return id, nil
}

type mortgageCalculationRow struct {
	ID                      int64           `db:"id"`
	UserID                  string          `db:"user_id"`
	MortgageProfileID       int64           `db:"mortgage_profile_id"`
	MonthlyPayment          sql.NullFloat64 `db:"monthly_payment"`
	TotalPayment            sql.NullFloat64 `db:"total_payment"`
	TotalOverpaymentAmount  sql.NullFloat64 `db:"total_overpayment_amount"`
	PossibleTaxDeduction    sql.NullFloat64 `db:"possible_tax_deduction"`
	SavingsDueMotherCapital sql.NullFloat64 `db:"savings_due_mother_capital"`
	RecommendedIncome       sql.NullFloat64 `db:"recommended_income"`
	PaymentSchedule         []byte          `db:"payment_schedule"`
}

func (p *pgRepo) GetCalculation(ctx context.Context, id int64) (calc domain.MortgageCalculation, err error) {
	var row mortgageCalculationRow

	if err := p.db.GetContext(ctx, &row, querySelectMortgageCalculation, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calc, domain.ErrNotFound
		}

		return domain.MortgageCalculation{}, fmt.Errorf("pgRepo.GetCalculation.Select: %w", err)
	}

	calc = domain.MortgageCalculation{
		ID:                      row.ID,
		UserID:                  row.UserID,
		MortgageProfileID:       row.MortgageProfileID,
		MonthlyPayment:          row.MonthlyPayment.Float64,
		TotalPayment:            row.TotalPayment.Float64,
		TotalOverpaymentAmount:  row.TotalOverpaymentAmount.Float64,
		PossibleTaxDeduction:    row.PossibleTaxDeduction.Float64,
		SavingsDueMotherCapital: row.SavingsDueMotherCapital.Float64,
		RecommendedIncome:       row.RecommendedIncome.Float64,
	}

	if len(row.PaymentSchedule) > 0 {
		if err = json.Unmarshal(row.PaymentSchedule, &calc.PaymentSchedule); err != nil {
			return calc, fmt.Errorf("pgRepo.GetCalculation.Schedule: %w", err)
		}
	}

	return calc, nil
}

func (p *pgRepo) UpdateCalculation(ctx context.Context, calc domain.MortgageCalculation) (err error) {
	scheduleJSON, err := json.Marshal(calc.PaymentSchedule)
	if err != nil {
		return fmt.Errorf("pgRepo.UpdateCalculation.Marshal: %w", err)
	}

	res, err := p.db.ExecContext(ctx, queryUpdateMortgageCalculation,
		calc.ID,
		calc.MonthlyPayment,
		calc.TotalPayment,
		calc.TotalOverpaymentAmount,
		calc.PossibleTaxDeduction,
		calc.SavingsDueMotherCapital,
		calc.RecommendedIncome,
		string(scheduleJSON),
	)
	if err != nil {
		return fmt.Errorf("pgRepo.UpdateCalculation.Update: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("pgRepo.UpdateCalculation.RowsAffected: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}
