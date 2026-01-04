DO $$
BEGIN
CREATE ROLE rds_iam;
EXCEPTION WHEN duplicate_object THEN RAISE NOTICE '%, skipping', SQLERRM USING ERRCODE = SQLSTATE;
END
$$;

CREATE USER lambda_user;
GRANT rds_iam TO lambda_user;