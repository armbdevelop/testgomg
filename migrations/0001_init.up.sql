create table users (
    tg_id      varchar(255) primary key,
    username   varchar(255),
    first_name varchar(255),
    last_name  varchar(255),
    lang_code  varchar(10),
    invited_by varchar(255),
    is_active  boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table mortgage_profile (
    id                   bigserial primary key,
    user_id              varchar(255) not null references users (tg_id),
    property_price       double precision not null,
    property_type        varchar(50) not null,
    down_payment_amount  double precision not null,
    mat_capital_amount   double precision,
    mat_capital_included boolean not null default false,
    mortgage_term_years  integer not null,
    interest_rate        double precision not null,
    created_at           timestamptz not null default now(),
    updated_at           timestamptz not null default now()
);

create table mortgage_calculation (
    id                          bigserial primary key,
    user_id                     varchar(255) not null references users (tg_id),
    mortgage_profile_id         bigint not null references mortgage_profile (id),
    monthly_payment             double precision,
    total_payment               double precision,
    total_overpayment_amount    double precision,
    possible_tax_deduction      double precision,
    savings_due_mother_capital  double precision,
    recommended_income          double precision,
    payment_schedule            jsonb,
    created_at                  timestamptz not null default now(),
    updated_at                  timestamptz not null default now()
);
