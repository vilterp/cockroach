// Code generated by help.awk. DO NOT EDIT.
// GENERATED FILE DO NOT EDIT

package parser

var helpMessages = map[string]HelpMessageBody{
	//line sql.y: 982
	`ALTER`: {
		//line sql.y: 983
		Category: hGroup,
		//line sql.y: 984
		Text: `ALTER TABLE, ALTER INDEX, ALTER VIEW, ALTER DATABASE
`,
	},
	//line sql.y: 993
	`ALTER TABLE`: {
		ShortDescription: `change the definition of a table`,
		//line sql.y: 994
		Category: hDDL,
		//line sql.y: 995
		Text: `
ALTER TABLE [IF EXISTS] <tablename> <command> [, ...]

Commands:
  ALTER TABLE ... ADD [COLUMN] [IF NOT EXISTS] <colname> <type> [<qualifiers...>]
  ALTER TABLE ... ADD <constraint>
  ALTER TABLE ... DROP [COLUMN] [IF EXISTS] <colname> [RESTRICT | CASCADE]
  ALTER TABLE ... DROP CONSTRAINT [IF EXISTS] <constraintname> [RESTRICT | CASCADE]
  ALTER TABLE ... ALTER [COLUMN] <colname> {SET DEFAULT <expr> | DROP DEFAULT}
  ALTER TABLE ... ALTER [COLUMN] <colname> DROP NOT NULL
  ALTER TABLE ... RENAME TO <newname>
  ALTER TABLE ... RENAME [COLUMN] <colname> TO <newname>
  ALTER TABLE ... VALIDATE CONSTRAINT <constraintname>
  ALTER TABLE ... SPLIT AT <selectclause>
  ALTER TABLE ... SCATTER [ FROM ( <exprs...> ) TO ( <exprs...> ) ]

Column qualifiers:
  [CONSTRAINT <constraintname>] {NULL | NOT NULL | UNIQUE | PRIMARY KEY | CHECK (<expr>) | DEFAULT <expr>}
  FAMILY <familyname>, CREATE [IF NOT EXISTS] FAMILY [<familyname>]
  REFERENCES <tablename> [( <colnames...> )]
  COLLATE <collationname>

`,
		//line sql.y: 1017
		SeeAlso: `WEBDOCS/alter-table.html
`,
	},
	//line sql.y: 1029
	`ALTER VIEW`: {
		ShortDescription: `change the definition of a view`,
		//line sql.y: 1030
		Category: hDDL,
		//line sql.y: 1031
		Text: `
ALTER VIEW [IF EXISTS] <name> RENAME TO <newname>
`,
		//line sql.y: 1033
		SeeAlso: `WEBDOCS/alter-view.html
`,
	},
	//line sql.y: 1040
	`ALTER DATABASE`: {
		ShortDescription: `change the definition of a database`,
		//line sql.y: 1041
		Category: hDDL,
		//line sql.y: 1042
		Text: `
ALTER DATABASE <name> RENAME TO <newname>
`,
		//line sql.y: 1044
		SeeAlso: `WEBDOCS/alter-database.html
`,
	},
	//line sql.y: 1055
	`ALTER INDEX`: {
		ShortDescription: `change the definition of an index`,
		//line sql.y: 1056
		Category: hDDL,
		//line sql.y: 1057
		Text: `
ALTER INDEX [IF EXISTS] <idxname> <command>

Commands:
  ALTER INDEX ... RENAME TO <newname>
  ALTER INDEX ... SPLIT AT <selectclause>
  ALTER INDEX ... SCATTER [ FROM ( <exprs...> ) TO ( <exprs...> ) ]

`,
		//line sql.y: 1065
		SeeAlso: `WEBDOCS/alter-index.html
`,
	},
	//line sql.y: 1305
	`BACKUP`: {
		ShortDescription: `back up data to external storage`,
		//line sql.y: 1306
		Category: hCCL,
		//line sql.y: 1307
		Text: `
BACKUP <targets...> TO <location...>
       [ AS OF SYSTEM TIME <expr> ]
       [ INCREMENTAL FROM <location...> ]
       [ WITH <option> [= <value>] [, ...] ]

Targets:
   TABLE <pattern> [, ...]
   DATABASE <databasename> [, ...]

Location:
   "[scheme]://[host]/[path to backup]?[parameters]"

Options:
   INTO_DB
   SKIP_MISSING_FOREIGN_KEYS

`,
		//line sql.y: 1324
		SeeAlso: `RESTORE, WEBDOCS/backup.html
`,
	},
	//line sql.y: 1332
	`RESTORE`: {
		ShortDescription: `restore data from external storage`,
		//line sql.y: 1333
		Category: hCCL,
		//line sql.y: 1334
		Text: `
RESTORE <targets...> FROM <location...>
        [ AS OF SYSTEM TIME <expr> ]
        [ WITH <option> [= <value>] [, ...] ]

Targets:
   TABLE <pattern> [, ...]
   DATABASE <databasename> [, ...]

Locations:
   "[scheme]://[host]/[path to backup]?[parameters]"

Options:
   INTO_DB
   SKIP_MISSING_FOREIGN_KEYS

`,
		//line sql.y: 1350
		SeeAlso: `BACKUP, WEBDOCS/restore.html
`,
	},
	//line sql.y: 1364
	`IMPORT`: {
		ShortDescription: `load data from file in a distributed manner`,
		//line sql.y: 1365
		Category: hCCL,
		//line sql.y: 1366
		Text: `
IMPORT TABLE <tablename>
       { ( <elements> ) | CREATE USING <schemafile> }
       <format>
       DATA ( <datafile> [, ...] )
       [ WITH <option> [= <value>] [, ...] ]

Formats:
   CSV

Options:
   distributed = '...'
   sstsize = '...'
   temp = '...'
   comma = '...'          [CSV-specific]
   comment = '...'        [CSV-specific]
   nullif = '...'         [CSV-specific]

`,
		//line sql.y: 1384
		SeeAlso: `CREATE TABLE
`,
	},
	//line sql.y: 1479
	`CANCEL`: {
		//line sql.y: 1480
		Category: hGroup,
		//line sql.y: 1481
		Text: `CANCEL JOB, CANCEL QUERY
`,
	},
	//line sql.y: 1487
	`CANCEL JOB`: {
		ShortDescription: `cancel a background job`,
		//line sql.y: 1488
		Category: hMisc,
		//line sql.y: 1489
		Text: `CANCEL JOB <jobid>
`,
		//line sql.y: 1490
		SeeAlso: `SHOW JOBS, PAUSE JOBS, RESUME JOB
`,
	},
	//line sql.y: 1498
	`CANCEL QUERY`: {
		ShortDescription: `cancel a running query`,
		//line sql.y: 1499
		Category: hMisc,
		//line sql.y: 1500
		Text: `CANCEL QUERY <queryid>
`,
		//line sql.y: 1501
		SeeAlso: `SHOW QUERIES
`,
	},
	//line sql.y: 1509
	`CREATE`: {
		//line sql.y: 1510
		Category: hGroup,
		//line sql.y: 1511
		Text: `
CREATE DATABASE, CREATE TABLE, CREATE INDEX, CREATE TABLE AS,
CREATE USER, CREATE VIEW, CREATE SEQUENCE
`,
	},
	//line sql.y: 1526
	`DELETE`: {
		ShortDescription: `delete rows from a table`,
		//line sql.y: 1527
		Category: hDML,
		//line sql.y: 1528
		Text: `DELETE FROM <tablename> [WHERE <expr>]
              [LIMIT <expr>]
              [RETURNING <exprs...>]
`,
		//line sql.y: 1531
		SeeAlso: `WEBDOCS/delete.html
`,
	},
	//line sql.y: 1544
	`DISCARD`: {
		ShortDescription: `reset the session to its initial state`,
		//line sql.y: 1545
		Category: hCfg,
		//line sql.y: 1546
		Text: `DISCARD ALL
`,
	},
	//line sql.y: 1558
	`DROP`: {
		//line sql.y: 1559
		Category: hGroup,
		//line sql.y: 1560
		Text: `DROP DATABASE, DROP INDEX, DROP TABLE, DROP VIEW, DROP USER
`,
	},
	//line sql.y: 1569
	`DROP VIEW`: {
		ShortDescription: `remove a view`,
		//line sql.y: 1570
		Category: hDDL,
		//line sql.y: 1571
		Text: `DROP VIEW [IF EXISTS] <tablename> [, ...] [CASCADE | RESTRICT]
`,
		//line sql.y: 1572
		SeeAlso: `WEBDOCS/drop-index.html
`,
	},
	//line sql.y: 1584
	`DROP TABLE`: {
		ShortDescription: `remove a table`,
		//line sql.y: 1585
		Category: hDDL,
		//line sql.y: 1586
		Text: `DROP TABLE [IF EXISTS] <tablename> [, ...] [CASCADE | RESTRICT]
`,
		//line sql.y: 1587
		SeeAlso: `WEBDOCS/drop-table.html
`,
	},
	//line sql.y: 1599
	`DROP INDEX`: {
		ShortDescription: `remove an index`,
		//line sql.y: 1600
		Category: hDDL,
		//line sql.y: 1601
		Text: `DROP INDEX [IF EXISTS] <idxname> [, ...] [CASCADE | RESTRICT]
`,
		//line sql.y: 1602
		SeeAlso: `WEBDOCS/drop-index.html
`,
	},
	//line sql.y: 1622
	`DROP DATABASE`: {
		ShortDescription: `remove a database`,
		//line sql.y: 1623
		Category: hDDL,
		//line sql.y: 1624
		Text: `DROP DATABASE [IF EXISTS] <databasename> [CASCADE | RESTRICT]
`,
		//line sql.y: 1625
		SeeAlso: `WEBDOCS/drop-database.html
`,
	},
	//line sql.y: 1645
	`DROP USER`: {
		ShortDescription: `remove a user`,
		//line sql.y: 1646
		Category: hPriv,
		//line sql.y: 1647
		Text: `DROP USER [IF EXISTS] <user> [, ...]
`,
		//line sql.y: 1648
		SeeAlso: `CREATE USER, SHOW USERS
`,
	},
	//line sql.y: 1690
	`EXPLAIN`: {
		ShortDescription: `show the logical plan of a query`,
		//line sql.y: 1691
		Category: hMisc,
		//line sql.y: 1692
		Text: `
EXPLAIN <statement>
EXPLAIN [( [PLAN ,] <planoptions...> )] <statement>

Explainable statements:
    SELECT, CREATE, DROP, ALTER, INSERT, UPSERT, UPDATE, DELETE,
    SHOW, EXPLAIN, EXECUTE

Plan options:
    TYPES, EXPRS, METADATA, QUALIFY, INDENT, VERBOSE, DIST_SQL

`,
		//line sql.y: 1703
		SeeAlso: `WEBDOCS/explain.html
`,
	},
	//line sql.y: 1761
	`PREPARE`: {
		ShortDescription: `prepare a statement for later execution`,
		//line sql.y: 1762
		Category: hMisc,
		//line sql.y: 1763
		Text: `PREPARE <name> [ ( <types...> ) ] AS <query>
`,
		//line sql.y: 1764
		SeeAlso: `EXECUTE, DEALLOCATE, DISCARD
`,
	},
	//line sql.y: 1786
	`EXECUTE`: {
		ShortDescription: `execute a statement prepared previously`,
		//line sql.y: 1787
		Category: hMisc,
		//line sql.y: 1788
		Text: `EXECUTE <name> [ ( <exprs...> ) ]
`,
		//line sql.y: 1789
		SeeAlso: `PREPARE, DEALLOCATE, DISCARD
`,
	},
	//line sql.y: 1812
	`DEALLOCATE`: {
		ShortDescription: `remove a prepared statement`,
		//line sql.y: 1813
		Category: hMisc,
		//line sql.y: 1814
		Text: `DEALLOCATE [PREPARE] { <name> | ALL }
`,
		//line sql.y: 1815
		SeeAlso: `PREPARE, EXECUTE, DISCARD
`,
	},
	//line sql.y: 1835
	`GRANT`: {
		ShortDescription: `define access privileges`,
		//line sql.y: 1836
		Category: hPriv,
		//line sql.y: 1837
		Text: `
GRANT {ALL | <privileges...> } ON <targets...> TO <grantees...>

Privileges:
  CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE

Targets:
  DATABASE <databasename> [, ...]
  [TABLE] [<databasename> .] { <tablename> | * } [, ...]

`,
		//line sql.y: 1847
		SeeAlso: `REVOKE, WEBDOCS/grant.html
`,
	},
	//line sql.y: 1855
	`REVOKE`: {
		ShortDescription: `remove access privileges`,
		//line sql.y: 1856
		Category: hPriv,
		//line sql.y: 1857
		Text: `
REVOKE {ALL | <privileges...> } ON <targets...> FROM <grantees...>

Privileges:
  CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE

Targets:
  DATABASE <databasename> [, <databasename>]...
  [TABLE] [<databasename> .] { <tablename> | * } [, ...]

`,
		//line sql.y: 1867
		SeeAlso: `GRANT, WEBDOCS/revoke.html
`,
	},
	//line sql.y: 1954
	`RESET`: {
		ShortDescription: `reset a session variable to its default value`,
		//line sql.y: 1955
		Category: hCfg,
		//line sql.y: 1956
		Text: `RESET [SESSION] <var>
`,
		//line sql.y: 1957
		SeeAlso: `RESET CLUSTER SETTING, WEBDOCS/set-vars.html
`,
	},
	//line sql.y: 1969
	`RESET CLUSTER SETTING`: {
		ShortDescription: `reset a cluster setting to its default value`,
		//line sql.y: 1970
		Category: hCfg,
		//line sql.y: 1971
		Text: `RESET CLUSTER SETTING <var>
`,
		//line sql.y: 1972
		SeeAlso: `SET CLUSTER SETTING, RESET
`,
	},
	//line sql.y: 2002
	`SCRUB TABLE`: {
		ShortDescription: `run a scrub check on a table`,
		//line sql.y: 2003
		Category: hMisc,
		//line sql.y: 2004
		Text: `
SCRUB TABLE <tablename> [WITH <option> [, ...]]

Options:
  SCRUB TABLE ... WITH OPTIONS INDEX ALL
  SCRUB TABLE ... WITH OPTIONS INDEX (<index>...)

`,
	},
	//line sql.y: 2041
	`SET CLUSTER SETTING`: {
		ShortDescription: `change a cluster setting`,
		//line sql.y: 2042
		Category: hCfg,
		//line sql.y: 2043
		Text: `SET CLUSTER SETTING <var> { TO | = } <value>
`,
		//line sql.y: 2044
		SeeAlso: `SHOW CLUSTER SETTING, RESET CLUSTER SETTING, SET SESSION,
WEBDOCS/cluster-settings.html
`,
	},
	//line sql.y: 2065
	`SET SESSION`: {
		ShortDescription: `change a session variable`,
		//line sql.y: 2066
		Category: hCfg,
		//line sql.y: 2067
		Text: `
SET [SESSION] <var> { TO | = } <values...>
SET [SESSION] TIME ZONE <tz>
SET [SESSION] CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL { SNAPSHOT | SERIALIZABLE }

`,
		//line sql.y: 2072
		SeeAlso: `SHOW SESSION, RESET, DISCARD, SHOW, SET CLUSTER SETTING, SET TRANSACTION,
WEBDOCS/set-vars.html
`,
	},
	//line sql.y: 2089
	`SET TRANSACTION`: {
		ShortDescription: `configure the transaction settings`,
		//line sql.y: 2090
		Category: hTxn,
		//line sql.y: 2091
		Text: `
SET [SESSION] TRANSACTION <txnparameters...>

Transaction parameters:
   ISOLATION LEVEL { SNAPSHOT | SERIALIZABLE }
   PRIORITY { LOW | NORMAL | HIGH }

`,
		//line sql.y: 2098
		SeeAlso: `SHOW TRANSACTION, SET SESSION,
WEBDOCS/set-transaction.html
`,
	},
	//line sql.y: 2237
	`SHOW`: {
		//line sql.y: 2238
		Category: hGroup,
		//line sql.y: 2239
		Text: `
SHOW SESSION, SHOW CLUSTER SETTING, SHOW DATABASES, SHOW TABLES, SHOW COLUMNS, SHOW INDEXES,
SHOW CONSTRAINTS, SHOW CREATE TABLE, SHOW CREATE VIEW, SHOW USERS, SHOW TRANSACTION, SHOW BACKUP,
SHOW JOBS, SHOW QUERIES, SHOW SESSIONS, SHOW TRACE
`,
	},
	//line sql.y: 2265
	`SHOW SESSION`: {
		ShortDescription: `display session variables`,
		//line sql.y: 2266
		Category: hCfg,
		//line sql.y: 2267
		Text: `SHOW [SESSION] { <var> | ALL }
`,
		//line sql.y: 2268
		SeeAlso: `WEBDOCS/show-vars.html
`,
	},
	//line sql.y: 2289
	`SHOW BACKUP`: {
		ShortDescription: `list backup contents`,
		//line sql.y: 2290
		Category: hCCL,
		//line sql.y: 2291
		Text: `SHOW BACKUP <location>
`,
		//line sql.y: 2292
		SeeAlso: `WEBDOCS/show-backup.html
`,
	},
	//line sql.y: 2300
	`SHOW CLUSTER SETTING`: {
		ShortDescription: `display cluster settings`,
		//line sql.y: 2301
		Category: hCfg,
		//line sql.y: 2302
		Text: `
SHOW CLUSTER SETTING <var>
SHOW ALL CLUSTER SETTINGS
`,
		//line sql.y: 2305
		SeeAlso: `WEBDOCS/cluster-settings.html
`,
	},
	//line sql.y: 2322
	`SHOW COLUMNS`: {
		ShortDescription: `list columns in relation`,
		//line sql.y: 2323
		Category: hDDL,
		//line sql.y: 2324
		Text: `SHOW COLUMNS FROM <tablename>
`,
		//line sql.y: 2325
		SeeAlso: `WEBDOCS/show-columns.html
`,
	},
	//line sql.y: 2333
	`SHOW DATABASES`: {
		ShortDescription: `list databases`,
		//line sql.y: 2334
		Category: hDDL,
		//line sql.y: 2335
		Text: `SHOW DATABASES
`,
		//line sql.y: 2336
		SeeAlso: `WEBDOCS/show-databases.html
`,
	},
	//line sql.y: 2344
	`SHOW GRANTS`: {
		ShortDescription: `list grants`,
		//line sql.y: 2345
		Category: hPriv,
		//line sql.y: 2346
		Text: `SHOW GRANTS [ON <targets...>] [FOR <users...>]
`,
		//line sql.y: 2347
		SeeAlso: `WEBDOCS/show-grants.html
`,
	},
	//line sql.y: 2355
	`SHOW INDEXES`: {
		ShortDescription: `list indexes`,
		//line sql.y: 2356
		Category: hDDL,
		//line sql.y: 2357
		Text: `SHOW INDEXES FROM <tablename>
`,
		//line sql.y: 2358
		SeeAlso: `WEBDOCS/show-index.html
`,
	},
	//line sql.y: 2376
	`SHOW CONSTRAINTS`: {
		ShortDescription: `list constraints`,
		//line sql.y: 2377
		Category: hDDL,
		//line sql.y: 2378
		Text: `SHOW CONSTRAINTS FROM <tablename>
`,
		//line sql.y: 2379
		SeeAlso: `WEBDOCS/show-constraints.html
`,
	},
	//line sql.y: 2392
	`SHOW QUERIES`: {
		ShortDescription: `list running queries`,
		//line sql.y: 2393
		Category: hMisc,
		//line sql.y: 2394
		Text: `SHOW [CLUSTER | LOCAL] QUERIES
`,
		//line sql.y: 2395
		SeeAlso: `CANCEL QUERY
`,
	},
	//line sql.y: 2411
	`SHOW JOBS`: {
		ShortDescription: `list background jobs`,
		//line sql.y: 2412
		Category: hMisc,
		//line sql.y: 2413
		Text: `SHOW JOBS
`,
		//line sql.y: 2414
		SeeAlso: `CANCEL JOB, PAUSE JOB, RESUME JOB
`,
	},
	//line sql.y: 2422
	`SHOW TRACE`: {
		ShortDescription: `display an execution trace`,
		//line sql.y: 2423
		Category: hMisc,
		//line sql.y: 2424
		Text: `
SHOW [KV] TRACE FOR SESSION
SHOW [KV] TRACE FOR <statement>
`,
		//line sql.y: 2427
		SeeAlso: `EXPLAIN
`,
	},
	//line sql.y: 2448
	`SHOW SESSIONS`: {
		ShortDescription: `list open client sessions`,
		//line sql.y: 2449
		Category: hMisc,
		//line sql.y: 2450
		Text: `SHOW [CLUSTER | LOCAL] SESSIONS
`,
	},
	//line sql.y: 2466
	`SHOW TABLES`: {
		ShortDescription: `list tables`,
		//line sql.y: 2467
		Category: hDDL,
		//line sql.y: 2468
		Text: `SHOW TABLES [FROM <databasename>]
`,
		//line sql.y: 2469
		SeeAlso: `WEBDOCS/show-tables.html
`,
	},
	//line sql.y: 2481
	`SHOW TRANSACTION`: {
		ShortDescription: `display current transaction properties`,
		//line sql.y: 2482
		Category: hCfg,
		//line sql.y: 2483
		Text: `SHOW TRANSACTION {ISOLATION LEVEL | PRIORITY | STATUS}
`,
		//line sql.y: 2484
		SeeAlso: `WEBDOCS/show-transaction.html
`,
	},
	//line sql.y: 2503
	`SHOW CREATE TABLE`: {
		ShortDescription: `display the CREATE TABLE statement for a table`,
		//line sql.y: 2504
		Category: hDDL,
		//line sql.y: 2505
		Text: `SHOW CREATE TABLE <tablename>
`,
		//line sql.y: 2506
		SeeAlso: `WEBDOCS/show-create-table.html
`,
	},
	//line sql.y: 2514
	`SHOW CREATE VIEW`: {
		ShortDescription: `display the CREATE VIEW statement for a view`,
		//line sql.y: 2515
		Category: hDDL,
		//line sql.y: 2516
		Text: `SHOW CREATE VIEW <viewname>
`,
		//line sql.y: 2517
		SeeAlso: `WEBDOCS/show-create-view.html
`,
	},
	//line sql.y: 2525
	`SHOW USERS`: {
		ShortDescription: `list defined users`,
		//line sql.y: 2526
		Category: hPriv,
		//line sql.y: 2527
		Text: `SHOW USERS
`,
		//line sql.y: 2528
		SeeAlso: `CREATE USER, DROP USER, WEBDOCS/show-users.html
`,
	},
	//line sql.y: 2601
	`PAUSE JOB`: {
		ShortDescription: `pause a background job`,
		//line sql.y: 2602
		Category: hMisc,
		//line sql.y: 2603
		Text: `PAUSE JOB <jobid>
`,
		//line sql.y: 2604
		SeeAlso: `SHOW JOBS, CANCEL JOB, RESUME JOB
`,
	},
	//line sql.y: 2612
	`CREATE TABLE`: {
		ShortDescription: `create a new table`,
		//line sql.y: 2613
		Category: hDDL,
		//line sql.y: 2614
		Text: `
CREATE TABLE [IF NOT EXISTS] <tablename> ( <elements...> ) [<interleave>]
CREATE TABLE [IF NOT EXISTS] <tablename> [( <colnames...> )] AS <source>

Table elements:
   <name> <type> [<qualifiers...>]
   [UNIQUE] INDEX [<name>] ( <colname> [ASC | DESC] [, ...] )
                           [STORING ( <colnames...> )] [<interleave>]
   FAMILY [<name>] ( <colnames...> )
   [CONSTRAINT <name>] <constraint>

Table constraints:
   PRIMARY KEY ( <colnames...> )
   FOREIGN KEY ( <colnames...> ) REFERENCES <tablename> [( <colnames...> )] [ON DELETE {NO ACTION | RESTRICT}] [ON UPDATE {NO ACTION | RESTRICT}]
   UNIQUE ( <colnames... ) [STORING ( <colnames...> )] [<interleave>]
   CHECK ( <expr> )

Column qualifiers:
  [CONSTRAINT <constraintname>] {NULL | NOT NULL | UNIQUE | PRIMARY KEY | CHECK (<expr>) | DEFAULT <expr>}
  FAMILY <familyname>, CREATE [IF NOT EXISTS] FAMILY [<familyname>]
  REFERENCES <tablename> [( <colnames...> )] [ON DELETE {NO ACTION | RESTRICT}] [ON UPDATE {NO ACTION | RESTRICT}]
  COLLATE <collationname>

Interleave clause:
   INTERLEAVE IN PARENT <tablename> ( <colnames...> ) [CASCADE | RESTRICT]

`,
		//line sql.y: 2640
		SeeAlso: `SHOW TABLES, CREATE VIEW, SHOW CREATE TABLE,
WEBDOCS/create-table.html
WEBDOCS/create-table-as.html
`,
	},
	//line sql.y: 3124
	`CREATE SEQUENCE`: {
		ShortDescription: `create a new sequence`,
		//line sql.y: 3125
		Category: hDDL,
		//line sql.y: 3126
		Text: `
CREATE [UNIQUE] INDEX [IF NOT EXISTS] [<idxname>]
       ON <tablename> ( <colname> [ASC | DESC] [, ...] )
       [STORING ( <colnames...> )] [<interleave>]

CREATE SEQUENCE <seqname>
  [INCREMENT <increment>]
  [MINVALUE <minvalue> | NO MINVALUE]
  [MAXVALUE <maxvalue> | NO MAXVALUE]
  [START <start>]
  [CACHE <cache>]
  [[NO] CYCLE]
  [OWNED BY { <table_name.column_name> | NONE }]

`,
		//line sql.y: 3140
		SeeAlso: `CREATE TABLE
WEBDOCS/create-sequence.html
`,
	},
	//line sql.y: 3181
	`TRUNCATE`: {
		ShortDescription: `empty one or more tables`,
		//line sql.y: 3182
		Category: hDML,
		//line sql.y: 3183
		Text: `TRUNCATE [TABLE] <tablename> [, ...] [CASCADE | RESTRICT]
`,
		//line sql.y: 3184
		SeeAlso: `WEBDOCS/truncate.html
`,
	},
	//line sql.y: 3192
	`CREATE USER`: {
		ShortDescription: `define a new user`,
		//line sql.y: 3193
		Category: hPriv,
		//line sql.y: 3194
		Text: `CREATE USER <name> [ [WITH] PASSWORD <passwd> ]
`,
		//line sql.y: 3195
		SeeAlso: `DROP USER, SHOW USERS, WEBDOCS/create-user.html
`,
	},
	//line sql.y: 3213
	`CREATE VIEW`: {
		ShortDescription: `create a new view`,
		//line sql.y: 3214
		Category: hDDL,
		//line sql.y: 3215
		Text: `CREATE VIEW <viewname> [( <colnames...> )] AS <source>
`,
		//line sql.y: 3216
		SeeAlso: `CREATE TABLE, SHOW CREATE VIEW, WEBDOCS/create-view.html
`,
	},
	//line sql.y: 3230
	`CREATE INDEX`: {
		ShortDescription: `create a new index`,
		//line sql.y: 3231
		Category: hDDL,
		//line sql.y: 3232
		Text: `
CREATE [UNIQUE] INDEX [IF NOT EXISTS] [<idxname>]
       ON <tablename> ( <colname> [ASC | DESC] [, ...] )
       [STORING ( <colnames...> )] [<interleave>]

Interleave clause:
   INTERLEAVE IN PARENT <tablename> ( <colnames...> ) [CASCADE | RESTRICT]

`,
		//line sql.y: 3240
		SeeAlso: `CREATE TABLE, SHOW INDEXES, SHOW CREATE INDEX,
WEBDOCS/create-index.html
`,
	},
	//line sql.y: 3379
	`RELEASE`: {
		ShortDescription: `complete a retryable block`,
		//line sql.y: 3380
		Category: hTxn,
		//line sql.y: 3381
		Text: `RELEASE [SAVEPOINT] cockroach_restart
`,
		//line sql.y: 3382
		SeeAlso: `SAVEPOINT, WEBDOCS/savepoint.html
`,
	},
	//line sql.y: 3390
	`RESUME JOB`: {
		ShortDescription: `resume a background job`,
		//line sql.y: 3391
		Category: hMisc,
		//line sql.y: 3392
		Text: `RESUME JOB <jobid>
`,
		//line sql.y: 3393
		SeeAlso: `SHOW JOBS, CANCEL JOB, PAUSE JOB
`,
	},
	//line sql.y: 3401
	`SAVEPOINT`: {
		ShortDescription: `start a retryable block`,
		//line sql.y: 3402
		Category: hTxn,
		//line sql.y: 3403
		Text: `SAVEPOINT cockroach_restart
`,
		//line sql.y: 3404
		SeeAlso: `RELEASE, WEBDOCS/savepoint.html
`,
	},
	//line sql.y: 3418
	`BEGIN`: {
		ShortDescription: `start a transaction`,
		//line sql.y: 3419
		Category: hTxn,
		//line sql.y: 3420
		Text: `
BEGIN [TRANSACTION] [ <txnparameter> [[,] ...] ]
START TRANSACTION [ <txnparameter> [[,] ...] ]

Transaction parameters:
   ISOLATION LEVEL { SNAPSHOT | SERIALIZABLE }
   PRIORITY { LOW | NORMAL | HIGH }

`,
		//line sql.y: 3428
		SeeAlso: `COMMIT, ROLLBACK, WEBDOCS/begin-transaction.html
`,
	},
	//line sql.y: 3441
	`COMMIT`: {
		ShortDescription: `commit the current transaction`,
		//line sql.y: 3442
		Category: hTxn,
		//line sql.y: 3443
		Text: `
COMMIT [TRANSACTION]
END [TRANSACTION]
`,
		//line sql.y: 3446
		SeeAlso: `BEGIN, ROLLBACK, WEBDOCS/commit-transaction.html
`,
	},
	//line sql.y: 3459
	`ROLLBACK`: {
		ShortDescription: `abort the current transaction`,
		//line sql.y: 3460
		Category: hTxn,
		//line sql.y: 3461
		Text: `ROLLBACK [TRANSACTION] [TO [SAVEPOINT] cockroach_restart]
`,
		//line sql.y: 3462
		SeeAlso: `BEGIN, COMMIT, SAVEPOINT, WEBDOCS/rollback-transaction.html
`,
	},
	//line sql.y: 3575
	`CREATE DATABASE`: {
		ShortDescription: `create a new database`,
		//line sql.y: 3576
		Category: hDDL,
		//line sql.y: 3577
		Text: `CREATE DATABASE [IF NOT EXISTS] <name>
`,
		//line sql.y: 3578
		SeeAlso: `WEBDOCS/create-database.html
`,
	},
	//line sql.y: 3647
	`INSERT`: {
		ShortDescription: `create new rows in a table`,
		//line sql.y: 3648
		Category: hDML,
		//line sql.y: 3649
		Text: `
INSERT INTO <tablename> [[AS] <name>] [( <colnames...> )]
       <selectclause>
       [ON CONFLICT [( <colnames...> )] {DO UPDATE SET ... [WHERE <expr>] | DO NOTHING}]
       [RETURNING <exprs...>]
`,
		//line sql.y: 3654
		SeeAlso: `UPSERT, UPDATE, DELETE, WEBDOCS/insert.html
`,
	},
	//line sql.y: 3671
	`UPSERT`: {
		ShortDescription: `create or replace rows in a table`,
		//line sql.y: 3672
		Category: hDML,
		//line sql.y: 3673
		Text: `
UPSERT INTO <tablename> [AS <name>] [( <colnames...> )]
       <selectclause>
       [RETURNING <exprs...>]
`,
		//line sql.y: 3677
		SeeAlso: `INSERT, UPDATE, DELETE, WEBDOCS/upsert.html
`,
	},
	//line sql.y: 3753
	`UPDATE`: {
		ShortDescription: `update rows of a table`,
		//line sql.y: 3754
		Category: hDML,
		//line sql.y: 3755
		Text: `UPDATE <tablename> [[AS] <name>] SET ... [WHERE <expr>] [RETURNING <exprs...>]
`,
		//line sql.y: 3756
		SeeAlso: `INSERT, UPSERT, DELETE, WEBDOCS/update.html
`,
	},
	//line sql.y: 3932
	`<SELECTCLAUSE>`: {
		ShortDescription: `access tabular data`,
		//line sql.y: 3933
		Category: hDML,
		//line sql.y: 3934
		Text: `
Select clause:
  TABLE <tablename>
  VALUES ( <exprs...> ) [ , ... ]
  SELECT ... [ { INTERSECT | UNION | EXCEPT } [ ALL | DISTINCT ] <selectclause> ]
`,
	},
	//line sql.y: 3945
	`SELECT`: {
		ShortDescription: `retrieve rows from a data source and compute a result`,
		//line sql.y: 3946
		Category: hDML,
		//line sql.y: 3947
		Text: `
SELECT [DISTINCT]
       { <expr> [[AS] <name>] | [ [<dbname>.] <tablename>. ] * } [, ...]
       [ FROM <source> ]
       [ WHERE <expr> ]
       [ GROUP BY <expr> [ , ... ] ]
       [ HAVING <expr> ]
       [ WINDOW <name> AS ( <definition> ) ]
       [ { UNION | INTERSECT | EXCEPT } [ ALL | DISTINCT ] <selectclause> ]
       [ ORDER BY <expr> [ ASC | DESC ] [, ...] ]
       [ LIMIT { <expr> | ALL } ]
       [ OFFSET <expr> [ ROW | ROWS ] ]
       [ FOR UPDATE ]
`,
		//line sql.y: 3960
		SeeAlso: `WEBDOCS/select.html
`,
	},
	//line sql.y: 4020
	`TABLE`: {
		ShortDescription: `select an entire table`,
		//line sql.y: 4021
		Category: hDML,
		//line sql.y: 4022
		Text: `TABLE <tablename>
`,
		//line sql.y: 4023
		SeeAlso: `SELECT, VALUES, WEBDOCS/table-expressions.html
`,
	},
	//line sql.y: 4286
	`VALUES`: {
		ShortDescription: `select a given set of values`,
		//line sql.y: 4287
		Category: hDML,
		//line sql.y: 4288
		Text: `VALUES ( <exprs...> ) [, ...]
`,
		//line sql.y: 4289
		SeeAlso: `SELECT, TABLE, WEBDOCS/table-expressions.html
`,
	},
	//line sql.y: 4394
	`<SOURCE>`: {
		ShortDescription: `define a data source for SELECT`,
		//line sql.y: 4395
		Category: hDML,
		//line sql.y: 4396
		Text: `
Data sources:
  <tablename> [ @ { <idxname> | <indexhint> } ]
  <tablefunc> ( <exprs...> )
  ( { <selectclause> | <source> } )
  <source> [AS] <alias> [( <colnames...> )]
  <source> { [INNER] | { LEFT | RIGHT | FULL } [OUTER] } JOIN <source> ON <expr>
  <source> { [INNER] | { LEFT | RIGHT | FULL } [OUTER] } JOIN <source> USING ( <colnames...> )
  <source> NATURAL { [INNER] | { LEFT | RIGHT | FULL } [OUTER] } JOIN <source>
  <source> CROSS JOIN <source>
  <source> WITH ORDINALITY
  '[' EXPLAIN ... ']'
  '[' SHOW ... ']'

Index hints:
  '{' FORCE_INDEX = <idxname> [, ...] '}'
  '{' NO_INDEX_JOIN [, ...] '}'

`,
		//line sql.y: 4414
		SeeAlso: `WEBDOCS/table-expressions.html
`,
	},
}
