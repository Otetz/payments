CREATE INDEX payments_account_direction_index ON payments (account, direction);

CREATE OR REPLACE VIEW accounts_view AS
SELECT A.id,
       A.balance
           + (SELECT COALESCE(SUM(P.amount), 0)
              FROM payments AS P
              WHERE P.account = A.id
              AND P.direction='incoming')
           - (SELECT COALESCE(SUM(P.amount), 0)
              FROM payments AS P
              WHERE P.account = A.id
              AND P.direction='outgoing')
       AS balance,
       A.currency,
       A.deleted
FROM accounts AS A;
