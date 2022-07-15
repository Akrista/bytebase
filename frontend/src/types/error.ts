export enum GeneralErrorCode {
  OK = 0,
  INTERNAL = 1,
  NOT_AUTHORIZED = 2,
  INVALID = 3,
  NOT_FOUND = 4,
  CONFLICT = 5,
  NOT_IMPLEMENTED = 6,
}

export enum DBErrorCode {
  CONNECTION_ERROR = 101,
  SYNTAX_ERROR = 102,
  EXECUTION_ERROR = 103,
}

export enum MigrationErrorCode {
  MIGRATION_SCHEMA_MISSING = 201,
  MIGRAITON_ALREADY_APPLIED = 202,
  MGIRATION_OUT_OF_ORDER = 203,
  MIGRATION_BASELINE_MISSING = 204,
}

export enum SQLReviewPolicyErrorCode {
  EMPTY_POLICY = 401,
  STATEMENT_NO_WHERE = 10101,
  STATEMENT_NO_SELECT_ALL = 10102,
  STATEMENT_LEADING_WILDCARD_LIKE = 10103,
  TABLE_NAMING_DISMATCH = 10201,
  COLUMN_NAMING_DISMATCH = 10202,
  INDEX_NAMING_DISMATCH = 10203,
  UK_NAMING_DISMATCH = 10204,
  FK_NAMING_DISMATCH = 10205,
  NO_REQUIRED_COLUMN = 10301,
  COLUMN_CANBE_NULL = 10302,
  NOT_INNODB_ENGINE = 10401,
  NO_PK_IN_TABLE = 10501,
}

export enum CompatibilityErrorCode {
  COMPATIBILITY_DROP_DATABASE = 10001,
  COMPATIBILITY_RENAME_TABLE = 10002,
  COMPATIBILITY_DROP_TABLE = 10003,
  COMPATIBILITY_RENAME_COLUMN = 10004,
  COMPATIBILITY_DROP_COLUMN = 10005,
  COMPATIBILITY_ADD_PRIMARY_KEY = 10006,
  COMPATIBILITY_ADD_UNIQUE_KEY = 10007,
  COMPATIBILITY_ADD_FOREIGN_KEY = 10008,
  COMPATIBILITY_ADD_CHECK = 10009,
  COMPATIBILITY_ALTER_CHECK = 10010,
  COMPATIBILITY_ALTER_COLUMN = 10011,
}

export type ErrorCode =
  | GeneralErrorCode
  | DBErrorCode
  | MigrationErrorCode
  | CompatibilityErrorCode
  | SQLReviewPolicyErrorCode;

export type ErrorTag = "General" | "Compatibility";

export type ErrorMeta = {
  code: ErrorCode;
  hash: string;
};
