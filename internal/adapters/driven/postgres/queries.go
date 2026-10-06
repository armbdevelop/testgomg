package postgres

const (
	queryLockMortgageCalculation = `
		select pg_advisory_xact_lock(hashtextextended($1, 0));
	`

	queryUpsertUser = `
	insert into users (tg_id) values ($1) on conflict (tg_id) do nothing;
	`

	queryInsertMortgageProfile = `
	insert into mortgage_profile                                              
         (user_id, 
          property_price, 
          property_type, 
          down_payment_amount,         
          mat_capital_amount, 
          mat_capital_included, 
          mortgage_term_years,       
   		  interest_rate)                                                              
     values ($1,$2,$3,$4,$5,$6,$7,$8)                                          
     returning id;   
	`

	queryInsertMortgageCalculation = `
     insert into mortgage_calculation (user_id, mortgage_profile_id)           
     values ($1, $2)                                                           
     returning id;      
	`

	queryUpdateMortgageCalculation = `
	update mortgage_calculation set                                           
         monthly_payment=$2, 
         total_payment=$3, 
         total_overpayment_amount=$4,    
         possible_tax_deduction=$5, 
         savings_due_mother_capital=$6,             
         recommended_income=$7, 
         payment_schedule=$8, 
         updated_at=now()          
     where id=$1;     
	`

	querySelectCalculationByParams = `
		select c.id, c.user_id, c.mortgage_profile_id, c.monthly_payment,
			c.total_payment, c.total_overpayment_amount, c.possible_tax_deduction,
			c.savings_due_mother_capital, c.recommended_income, c.payment_schedule
		from mortgage_calculation c
		join mortgage_profile p on p.id = c.mortgage_profile_id
		where c.user_id = $1 and p.user_id = $1
			and p.property_price = $2 and p.property_type = $3
			and p.down_payment_amount = $4
			and p.mat_capital_amount is not distinct from $5
			and p.mat_capital_included = $6
			and p.mortgage_term_years = $7 and p.interest_rate = $8
		order by c.id
		limit 1;
	`

	querySelectMortgageCalculation = `
	select 
	    id, 
	    user_id, 
	    mortgage_profile_id, 
	    monthly_payment, 
	    total_payment,  
    	total_overpayment_amount, 
    	possible_tax_deduction,                  
   		savings_due_mother_capital, 
   		recommended_income, 
   		payment_schedule                               
    from mortgage_calculation 
    where id=$1 and user_id=$2;
	`
)
