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

func (p *pgRepo) GetOrCreateCalculation(
	ctx context.Context,
	profile domain.MortgageProfile,
) (calc domain.MortgageCalculation, err error) {
	paramsHash, err := domain.CacheKey(profile)
	if err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Hash: %w", err)
	}

	tx, err := p.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Begin: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("pgRepo.GetOrCreateCalculation.Rollback: %v", rbErr)
		}
	}()

	// Блокировка живёт до конца транзакции, включая поиск существующей записи.
	if _, err = tx.ExecContext(ctx, queryLockMortgageCalculation, profile.UserID+":"+paramsHash); err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Lock(userID: %s): %w", profile.UserID, err)
	}

	var row mortgageCalculationRow

	err = tx.GetContext(ctx, &row, querySelectCalculationByParams,
		profile.UserID, profile.PropertyPrice, profile.PropertyType, profile.DownPaymentAmount,
		profile.MatCapitalAmount, profile.MatCapitalIncluded, profile.MortgageTermYears, profile.InterestRate)
	if err == nil {
		calc, err = row.toDomain()
		if err != nil {
			return calc, err
		}

		if err = tx.Commit(); err != nil {
			return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Commit: %w", err)
		}

		return calc, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Select(userID: %s): %w", profile.UserID, err)
	}

	if _, err = tx.ExecContext(ctx, queryUpsertUser, profile.UserID); err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.UpsertUser: %w", err)
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
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.InsertProfile: %w", err)
	}

	if err = tx.GetContext(ctx, &calc.ID, queryInsertMortgageCalculation, profile.UserID, profileID); err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.InsertCalculation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return calc, fmt.Errorf("pgRepo.GetOrCreateCalculation.Commit: %w", err)
	}

	calc.UserID = profile.UserID
	calc.MortgageProfileID = profileID
	calc.Status = domain.StatusPending

	return calc, nil
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

func (p *pgRepo) GetCalculation(ctx context.Context, id int64,
	userID string) (calc domain.MortgageCalculation, err error) {
	var row mortgageCalculationRow

	if err := p.db.GetContext(ctx, &row, querySelectMortgageCalculation, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return calc, domain.ErrNotFound
		}

		return domain.MortgageCalculation{}, fmt.Errorf("pgRepo.GetCalculation.Select: %w", err)
	}

	return row.toDomain()
}

func (row mortgageCalculationRow) toDomain() (calc domain.MortgageCalculation, err error) {
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
		Status:                  domain.StatusPending,
	}

	if len(row.PaymentSchedule) > 0 {
		if err = json.Unmarshal(row.PaymentSchedule, &calc.PaymentSchedule); err != nil {
			return calc, fmt.Errorf("pgRepo.GetCalculation.Schedule: %w", err)
		}
	}

	if calc.PaymentSchedule != nil {
		calc.Status = domain.StatusDone
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
