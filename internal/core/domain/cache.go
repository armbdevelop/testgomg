package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
)

func IDCacheKey(id int64) string {
	return "mortgage:id:" + strconv.FormatInt(id, 10)
}

func CacheKey(p MortgageProfile) (string, error) {
	// В БД -0 и 0 равны; они должны брать одну транзакционную блокировку.
	if p.DownPaymentAmount == 0 {
		p.DownPaymentAmount = 0
	}

	if p.MatCapitalAmount != nil && *p.MatCapitalAmount == 0 {
		zero := 0.0
		p.MatCapitalAmount = &zero
	}

	data, err := json.Marshal(struct {
		PropertyPrice      float64      `json:"propertyPrice"`
		PropertyType       PropertyType `json:"propertyType"`
		DownPaymentAmount  float64      `json:"downPaymentAmount"`
		MatCapitalAmount   *float64     `json:"matCapitalAmount"`
		MatCapitalIncluded bool         `json:"matCapitalIncluded"`
		MortgageTermYears  int          `json:"mortgageTermYears"`
		InterestRate       float64      `json:"interestRate"`
	}{p.PropertyPrice, p.PropertyType, p.DownPaymentAmount, p.MatCapitalAmount,
		p.MatCapitalIncluded, p.MortgageTermYears, p.InterestRate})
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}

func UserCacheKey(p MortgageProfile) (string, error) {
	paramsHash, err := CacheKey(p)
	if err != nil {
		return "", err
	}

	return "mortgage:u:" + p.UserID + ":" + paramsHash, nil
}
