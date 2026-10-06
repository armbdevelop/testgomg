package postgres

const (
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
    where id=$1;         
	`
)
