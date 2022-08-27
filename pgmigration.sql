do $$
    declare
        t record;
    begin
        for t IN select column_name, table_name
                 from information_schema.columns
                 where data_type='bigint' and table_name NOT LIKE 'pg_%'
        loop
            execute 'alter table ' || t.table_name || ' alter column ' || t.column_name || ' type INTEGER';
        end loop;
end$$;

ALTER TABLE documents ALTER COLUMN timestamp TYPE BIGINT;

