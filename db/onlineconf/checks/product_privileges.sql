BEGIN READ ONLY;

SET LOCAL search_path = pg_catalog;

SELECT rolname AS role,
       has_database_privilege(oid, current_database(), 'CONNECT') AS can_connect,
       has_database_privilege(oid, current_database(), 'TEMPORARY') AS can_create_temp,
       has_database_privilege(oid, current_database(), 'CREATE') AS can_create_schema
FROM pg_roles
WHERE rolname IN ('onlineconf_migrator', 'onlineconf_runtime')
ORDER BY rolname;

SELECT r.rolname AS role, n.nspname AS schema,
       has_schema_privilege(r.oid, n.oid, 'USAGE') AS can_use,
       has_schema_privilege(r.oid, n.oid, 'CREATE') AS can_create
FROM pg_roles r CROSS JOIN pg_namespace n
WHERE r.rolname IN ('onlineconf_migrator', 'onlineconf_runtime')
  AND n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema'
ORDER BY r.rolname, n.nspname;

SELECT r.rolname AS role, n.nspname AS schema, c.relname AS object
FROM pg_roles r CROSS JOIN pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE r.rolname IN ('onlineconf_migrator', 'onlineconf_runtime')
  AND n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema'
  AND CASE
      WHEN c.relkind = 'S' THEN has_sequence_privilege(r.oid, c.oid, 'USAGE, SELECT, UPDATE')
      WHEN c.relkind IN ('r', 'p', 'v', 'm', 'f') THEN
          has_table_privilege(r.oid, c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN')
          OR has_any_column_privilege(r.oid, c.oid, 'SELECT, INSERT, UPDATE, REFERENCES')
      ELSE FALSE
  END
ORDER BY r.rolname, n.nspname, c.relname;

SELECT r.rolname AS role, n.nspname AS schema, p.oid::regprocedure AS security_definer_routine
FROM pg_roles r CROSS JOIN pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE r.rolname IN ('onlineconf_migrator', 'onlineconf_runtime')
  AND n.nspname NOT LIKE 'pg\_%' AND n.nspname <> 'information_schema'
  AND p.prosecdef AND has_function_privilege(r.oid, p.oid, 'EXECUTE')
ORDER BY r.rolname, n.nspname, p.oid::regprocedure::TEXT;

COMMIT;
