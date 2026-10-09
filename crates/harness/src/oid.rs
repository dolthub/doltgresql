// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//! The OID of every type that Postgres 15 defines, named after pg_type.typname. Array types, whose names start
//! with an underscore, are named with an _ARRAY suffix instead.

/// The OID of bool.
pub const BOOL: u32 = 16;
/// The OID of bytea.
pub const BYTEA: u32 = 17;
/// The OID of char.
pub const CHAR: u32 = 18;
/// The OID of name.
pub const NAME: u32 = 19;
/// The OID of int8.
pub const INT8: u32 = 20;
/// The OID of int2.
pub const INT2: u32 = 21;
/// The OID of int2vector.
pub const INT2VECTOR: u32 = 22;
/// The OID of int4.
pub const INT4: u32 = 23;
/// The OID of regproc.
pub const REGPROC: u32 = 24;
/// The OID of text.
pub const TEXT: u32 = 25;
/// The OID of oid.
pub const OID: u32 = 26;
/// The OID of tid.
pub const TID: u32 = 27;
/// The OID of xid.
pub const XID: u32 = 28;
/// The OID of cid.
pub const CID: u32 = 29;
/// The OID of oidvector.
pub const OIDVECTOR: u32 = 30;
/// The OID of pg_ddl_command.
pub const PG_DDL_COMMAND: u32 = 32;
/// The OID of pg_type.
pub const PG_TYPE: u32 = 71;
/// The OID of pg_attribute.
pub const PG_ATTRIBUTE: u32 = 75;
/// The OID of pg_proc.
pub const PG_PROC: u32 = 81;
/// The OID of pg_class.
pub const PG_CLASS: u32 = 83;
/// The OID of json.
pub const JSON: u32 = 114;
/// The OID of xml.
pub const XML: u32 = 142;
/// The OID of _xml.
pub const XML_ARRAY: u32 = 143;
/// The OID of pg_node_tree.
pub const PG_NODE_TREE: u32 = 194;
/// The OID of _json.
pub const JSON_ARRAY: u32 = 199;
/// The OID of _pg_type.
pub const PG_TYPE_ARRAY: u32 = 210;
/// The OID of table_am_handler.
pub const TABLE_AM_HANDLER: u32 = 269;
/// The OID of _pg_attribute.
pub const PG_ATTRIBUTE_ARRAY: u32 = 270;
/// The OID of _xid8.
pub const XID8_ARRAY: u32 = 271;
/// The OID of _pg_proc.
pub const PG_PROC_ARRAY: u32 = 272;
/// The OID of _pg_class.
pub const PG_CLASS_ARRAY: u32 = 273;
/// The OID of index_am_handler.
pub const INDEX_AM_HANDLER: u32 = 325;
/// The OID of point.
pub const POINT: u32 = 600;
/// The OID of lseg.
pub const LSEG: u32 = 601;
/// The OID of path.
pub const PATH: u32 = 602;
/// The OID of box.
pub const BOX: u32 = 603;
/// The OID of polygon.
pub const POLYGON: u32 = 604;
/// The OID of line.
pub const LINE: u32 = 628;
/// The OID of _line.
pub const LINE_ARRAY: u32 = 629;
/// The OID of cidr.
pub const CIDR: u32 = 650;
/// The OID of _cidr.
pub const CIDR_ARRAY: u32 = 651;
/// The OID of float4.
pub const FLOAT4: u32 = 700;
/// The OID of float8.
pub const FLOAT8: u32 = 701;
/// The OID of unknown.
pub const UNKNOWN: u32 = 705;
/// The OID of circle.
pub const CIRCLE: u32 = 718;
/// The OID of _circle.
pub const CIRCLE_ARRAY: u32 = 719;
/// The OID of macaddr8.
pub const MACADDR8: u32 = 774;
/// The OID of _macaddr8.
pub const MACADDR8_ARRAY: u32 = 775;
/// The OID of money.
pub const MONEY: u32 = 790;
/// The OID of _money.
pub const MONEY_ARRAY: u32 = 791;
/// The OID of macaddr.
pub const MACADDR: u32 = 829;
/// The OID of inet.
pub const INET: u32 = 869;
/// The OID of _bool.
pub const BOOL_ARRAY: u32 = 1000;
/// The OID of _bytea.
pub const BYTEA_ARRAY: u32 = 1001;
/// The OID of _char.
pub const CHAR_ARRAY: u32 = 1002;
/// The OID of _name.
pub const NAME_ARRAY: u32 = 1003;
/// The OID of _int2.
pub const INT2_ARRAY: u32 = 1005;
/// The OID of _int2vector.
pub const INT2VECTOR_ARRAY: u32 = 1006;
/// The OID of _int4.
pub const INT4_ARRAY: u32 = 1007;
/// The OID of _regproc.
pub const REGPROC_ARRAY: u32 = 1008;
/// The OID of _text.
pub const TEXT_ARRAY: u32 = 1009;
/// The OID of _tid.
pub const TID_ARRAY: u32 = 1010;
/// The OID of _xid.
pub const XID_ARRAY: u32 = 1011;
/// The OID of _cid.
pub const CID_ARRAY: u32 = 1012;
/// The OID of _oidvector.
pub const OIDVECTOR_ARRAY: u32 = 1013;
/// The OID of _bpchar.
pub const BPCHAR_ARRAY: u32 = 1014;
/// The OID of _varchar.
pub const VARCHAR_ARRAY: u32 = 1015;
/// The OID of _int8.
pub const INT8_ARRAY: u32 = 1016;
/// The OID of _point.
pub const POINT_ARRAY: u32 = 1017;
/// The OID of _lseg.
pub const LSEG_ARRAY: u32 = 1018;
/// The OID of _path.
pub const PATH_ARRAY: u32 = 1019;
/// The OID of _box.
pub const BOX_ARRAY: u32 = 1020;
/// The OID of _float4.
pub const FLOAT4_ARRAY: u32 = 1021;
/// The OID of _float8.
pub const FLOAT8_ARRAY: u32 = 1022;
/// The OID of _polygon.
pub const POLYGON_ARRAY: u32 = 1027;
/// The OID of _oid.
pub const OID_ARRAY: u32 = 1028;
/// The OID of aclitem.
pub const ACLITEM: u32 = 1033;
/// The OID of _aclitem.
pub const ACLITEM_ARRAY: u32 = 1034;
/// The OID of _macaddr.
pub const MACADDR_ARRAY: u32 = 1040;
/// The OID of _inet.
pub const INET_ARRAY: u32 = 1041;
/// The OID of bpchar.
pub const BPCHAR: u32 = 1042;
/// The OID of varchar.
pub const VARCHAR: u32 = 1043;
/// The OID of date.
pub const DATE: u32 = 1082;
/// The OID of time.
pub const TIME: u32 = 1083;
/// The OID of timestamp.
pub const TIMESTAMP: u32 = 1114;
/// The OID of _timestamp.
pub const TIMESTAMP_ARRAY: u32 = 1115;
/// The OID of _date.
pub const DATE_ARRAY: u32 = 1182;
/// The OID of _time.
pub const TIME_ARRAY: u32 = 1183;
/// The OID of timestamptz.
pub const TIMESTAMPTZ: u32 = 1184;
/// The OID of _timestamptz.
pub const TIMESTAMPTZ_ARRAY: u32 = 1185;
/// The OID of interval.
pub const INTERVAL: u32 = 1186;
/// The OID of _interval.
pub const INTERVAL_ARRAY: u32 = 1187;
/// The OID of _numeric.
pub const NUMERIC_ARRAY: u32 = 1231;
/// The OID of pg_database.
pub const PG_DATABASE: u32 = 1248;
/// The OID of _cstring.
pub const CSTRING_ARRAY: u32 = 1263;
/// The OID of timetz.
pub const TIMETZ: u32 = 1266;
/// The OID of _timetz.
pub const TIMETZ_ARRAY: u32 = 1270;
/// The OID of bit.
pub const BIT: u32 = 1560;
/// The OID of _bit.
pub const BIT_ARRAY: u32 = 1561;
/// The OID of varbit.
pub const VARBIT: u32 = 1562;
/// The OID of _varbit.
pub const VARBIT_ARRAY: u32 = 1563;
/// The OID of numeric.
pub const NUMERIC: u32 = 1700;
/// The OID of refcursor.
pub const REFCURSOR: u32 = 1790;
/// The OID of _refcursor.
pub const REFCURSOR_ARRAY: u32 = 2201;
/// The OID of regprocedure.
pub const REGPROCEDURE: u32 = 2202;
/// The OID of regoper.
pub const REGOPER: u32 = 2203;
/// The OID of regoperator.
pub const REGOPERATOR: u32 = 2204;
/// The OID of regclass.
pub const REGCLASS: u32 = 2205;
/// The OID of regtype.
pub const REGTYPE: u32 = 2206;
/// The OID of _regprocedure.
pub const REGPROCEDURE_ARRAY: u32 = 2207;
/// The OID of _regoper.
pub const REGOPER_ARRAY: u32 = 2208;
/// The OID of _regoperator.
pub const REGOPERATOR_ARRAY: u32 = 2209;
/// The OID of _regclass.
pub const REGCLASS_ARRAY: u32 = 2210;
/// The OID of _regtype.
pub const REGTYPE_ARRAY: u32 = 2211;
/// The OID of record.
pub const RECORD: u32 = 2249;
/// The OID of cstring.
pub const CSTRING: u32 = 2275;
/// The OID of any.
pub const ANY: u32 = 2276;
/// The OID of anyarray.
pub const ANYARRAY: u32 = 2277;
/// The OID of void.
pub const VOID: u32 = 2278;
/// The OID of trigger.
pub const TRIGGER: u32 = 2279;
/// The OID of language_handler.
pub const LANGUAGE_HANDLER: u32 = 2280;
/// The OID of internal.
pub const INTERNAL: u32 = 2281;
/// The OID of anyelement.
pub const ANYELEMENT: u32 = 2283;
/// The OID of _record.
pub const RECORD_ARRAY: u32 = 2287;
/// The OID of anynonarray.
pub const ANYNONARRAY: u32 = 2776;
/// The OID of pg_authid.
pub const PG_AUTHID: u32 = 2842;
/// The OID of pg_auth_members.
pub const PG_AUTH_MEMBERS: u32 = 2843;
/// The OID of _txid_snapshot.
pub const TXID_SNAPSHOT_ARRAY: u32 = 2949;
/// The OID of uuid.
pub const UUID: u32 = 2950;
/// The OID of _uuid.
pub const UUID_ARRAY: u32 = 2951;
/// The OID of txid_snapshot.
pub const TXID_SNAPSHOT: u32 = 2970;
/// The OID of fdw_handler.
pub const FDW_HANDLER: u32 = 3115;
/// The OID of pg_lsn.
pub const PG_LSN: u32 = 3220;
/// The OID of _pg_lsn.
pub const PG_LSN_ARRAY: u32 = 3221;
/// The OID of tsm_handler.
pub const TSM_HANDLER: u32 = 3310;
/// The OID of pg_ndistinct.
pub const PG_NDISTINCT: u32 = 3361;
/// The OID of pg_dependencies.
pub const PG_DEPENDENCIES: u32 = 3402;
/// The OID of anyenum.
pub const ANYENUM: u32 = 3500;
/// The OID of tsvector.
pub const TSVECTOR: u32 = 3614;
/// The OID of tsquery.
pub const TSQUERY: u32 = 3615;
/// The OID of gtsvector.
pub const GTSVECTOR: u32 = 3642;
/// The OID of _tsvector.
pub const TSVECTOR_ARRAY: u32 = 3643;
/// The OID of _gtsvector.
pub const GTSVECTOR_ARRAY: u32 = 3644;
/// The OID of _tsquery.
pub const TSQUERY_ARRAY: u32 = 3645;
/// The OID of regconfig.
pub const REGCONFIG: u32 = 3734;
/// The OID of _regconfig.
pub const REGCONFIG_ARRAY: u32 = 3735;
/// The OID of regdictionary.
pub const REGDICTIONARY: u32 = 3769;
/// The OID of _regdictionary.
pub const REGDICTIONARY_ARRAY: u32 = 3770;
/// The OID of jsonb.
pub const JSONB: u32 = 3802;
/// The OID of _jsonb.
pub const JSONB_ARRAY: u32 = 3807;
/// The OID of anyrange.
pub const ANYRANGE: u32 = 3831;
/// The OID of event_trigger.
pub const EVENT_TRIGGER: u32 = 3838;
/// The OID of int4range.
pub const INT4RANGE: u32 = 3904;
/// The OID of _int4range.
pub const INT4RANGE_ARRAY: u32 = 3905;
/// The OID of numrange.
pub const NUMRANGE: u32 = 3906;
/// The OID of _numrange.
pub const NUMRANGE_ARRAY: u32 = 3907;
/// The OID of tsrange.
pub const TSRANGE: u32 = 3908;
/// The OID of _tsrange.
pub const TSRANGE_ARRAY: u32 = 3909;
/// The OID of tstzrange.
pub const TSTZRANGE: u32 = 3910;
/// The OID of _tstzrange.
pub const TSTZRANGE_ARRAY: u32 = 3911;
/// The OID of daterange.
pub const DATERANGE: u32 = 3912;
/// The OID of _daterange.
pub const DATERANGE_ARRAY: u32 = 3913;
/// The OID of int8range.
pub const INT8RANGE: u32 = 3926;
/// The OID of _int8range.
pub const INT8RANGE_ARRAY: u32 = 3927;
/// The OID of pg_shseclabel.
pub const PG_SHSECLABEL: u32 = 4066;
/// The OID of jsonpath.
pub const JSONPATH: u32 = 4072;
/// The OID of _jsonpath.
pub const JSONPATH_ARRAY: u32 = 4073;
/// The OID of regnamespace.
pub const REGNAMESPACE: u32 = 4089;
/// The OID of _regnamespace.
pub const REGNAMESPACE_ARRAY: u32 = 4090;
/// The OID of regrole.
pub const REGROLE: u32 = 4096;
/// The OID of _regrole.
pub const REGROLE_ARRAY: u32 = 4097;
/// The OID of regcollation.
pub const REGCOLLATION: u32 = 4191;
/// The OID of _regcollation.
pub const REGCOLLATION_ARRAY: u32 = 4192;
/// The OID of int4multirange.
pub const INT4MULTIRANGE: u32 = 4451;
/// The OID of nummultirange.
pub const NUMMULTIRANGE: u32 = 4532;
/// The OID of tsmultirange.
pub const TSMULTIRANGE: u32 = 4533;
/// The OID of tstzmultirange.
pub const TSTZMULTIRANGE: u32 = 4534;
/// The OID of datemultirange.
pub const DATEMULTIRANGE: u32 = 4535;
/// The OID of int8multirange.
pub const INT8MULTIRANGE: u32 = 4536;
/// The OID of anymultirange.
pub const ANYMULTIRANGE: u32 = 4537;
/// The OID of anycompatiblemultirange.
pub const ANYCOMPATIBLEMULTIRANGE: u32 = 4538;
/// The OID of pg_brin_bloom_summary.
pub const PG_BRIN_BLOOM_SUMMARY: u32 = 4600;
/// The OID of pg_brin_minmax_multi_summary.
pub const PG_BRIN_MINMAX_MULTI_SUMMARY: u32 = 4601;
/// The OID of pg_mcv_list.
pub const PG_MCV_LIST: u32 = 5017;
/// The OID of pg_snapshot.
pub const PG_SNAPSHOT: u32 = 5038;
/// The OID of _pg_snapshot.
pub const PG_SNAPSHOT_ARRAY: u32 = 5039;
/// The OID of xid8.
pub const XID8: u32 = 5069;
/// The OID of anycompatible.
pub const ANYCOMPATIBLE: u32 = 5077;
/// The OID of anycompatiblearray.
pub const ANYCOMPATIBLEARRAY: u32 = 5078;
/// The OID of anycompatiblenonarray.
pub const ANYCOMPATIBLENONARRAY: u32 = 5079;
/// The OID of anycompatiblerange.
pub const ANYCOMPATIBLERANGE: u32 = 5080;
/// The OID of pg_subscription.
pub const PG_SUBSCRIPTION: u32 = 6101;
/// The OID of _int4multirange.
pub const INT4MULTIRANGE_ARRAY: u32 = 6150;
/// The OID of _nummultirange.
pub const NUMMULTIRANGE_ARRAY: u32 = 6151;
/// The OID of _tsmultirange.
pub const TSMULTIRANGE_ARRAY: u32 = 6152;
/// The OID of _tstzmultirange.
pub const TSTZMULTIRANGE_ARRAY: u32 = 6153;
/// The OID of _datemultirange.
pub const DATEMULTIRANGE_ARRAY: u32 = 6155;
/// The OID of _int8multirange.
pub const INT8MULTIRANGE_ARRAY: u32 = 6157;
/// The OID of _pg_attrdef.
pub const PG_ATTRDEF_ARRAY: u32 = 10000;
/// The OID of pg_attrdef.
pub const PG_ATTRDEF: u32 = 10001;
/// The OID of _pg_constraint.
pub const PG_CONSTRAINT_ARRAY: u32 = 10002;
/// The OID of pg_constraint.
pub const PG_CONSTRAINT: u32 = 10003;
/// The OID of _pg_inherits.
pub const PG_INHERITS_ARRAY: u32 = 10004;
/// The OID of pg_inherits.
pub const PG_INHERITS: u32 = 10005;
/// The OID of _pg_index.
pub const PG_INDEX_ARRAY: u32 = 10006;
/// The OID of pg_index.
pub const PG_INDEX: u32 = 10007;
/// The OID of _pg_operator.
pub const PG_OPERATOR_ARRAY: u32 = 10008;
/// The OID of pg_operator.
pub const PG_OPERATOR: u32 = 10009;
/// The OID of _pg_opfamily.
pub const PG_OPFAMILY_ARRAY: u32 = 10010;
/// The OID of pg_opfamily.
pub const PG_OPFAMILY: u32 = 10011;
/// The OID of _pg_opclass.
pub const PG_OPCLASS_ARRAY: u32 = 10012;
/// The OID of pg_opclass.
pub const PG_OPCLASS: u32 = 10013;
/// The OID of _pg_am.
pub const PG_AM_ARRAY: u32 = 10014;
/// The OID of pg_am.
pub const PG_AM: u32 = 10015;
/// The OID of _pg_amop.
pub const PG_AMOP_ARRAY: u32 = 10016;
/// The OID of pg_amop.
pub const PG_AMOP: u32 = 10017;
/// The OID of _pg_amproc.
pub const PG_AMPROC_ARRAY: u32 = 10018;
/// The OID of pg_amproc.
pub const PG_AMPROC: u32 = 10019;
/// The OID of _pg_language.
pub const PG_LANGUAGE_ARRAY: u32 = 10020;
/// The OID of pg_language.
pub const PG_LANGUAGE: u32 = 10021;
/// The OID of _pg_largeobject_metadata.
pub const PG_LARGEOBJECT_METADATA_ARRAY: u32 = 10022;
/// The OID of pg_largeobject_metadata.
pub const PG_LARGEOBJECT_METADATA: u32 = 10023;
/// The OID of _pg_largeobject.
pub const PG_LARGEOBJECT_ARRAY: u32 = 10024;
/// The OID of pg_largeobject.
pub const PG_LARGEOBJECT: u32 = 10025;
/// The OID of _pg_aggregate.
pub const PG_AGGREGATE_ARRAY: u32 = 10026;
/// The OID of pg_aggregate.
pub const PG_AGGREGATE: u32 = 10027;
/// The OID of _pg_statistic.
pub const PG_STATISTIC_ARRAY: u32 = 10028;
/// The OID of pg_statistic.
pub const PG_STATISTIC: u32 = 10029;
/// The OID of _pg_statistic_ext.
pub const PG_STATISTIC_EXT_ARRAY: u32 = 10030;
/// The OID of pg_statistic_ext.
pub const PG_STATISTIC_EXT: u32 = 10031;
/// The OID of _pg_statistic_ext_data.
pub const PG_STATISTIC_EXT_DATA_ARRAY: u32 = 10032;
/// The OID of pg_statistic_ext_data.
pub const PG_STATISTIC_EXT_DATA: u32 = 10033;
/// The OID of _pg_rewrite.
pub const PG_REWRITE_ARRAY: u32 = 10034;
/// The OID of pg_rewrite.
pub const PG_REWRITE: u32 = 10035;
/// The OID of _pg_trigger.
pub const PG_TRIGGER_ARRAY: u32 = 10036;
/// The OID of pg_trigger.
pub const PG_TRIGGER: u32 = 10037;
/// The OID of _pg_event_trigger.
pub const PG_EVENT_TRIGGER_ARRAY: u32 = 10038;
/// The OID of pg_event_trigger.
pub const PG_EVENT_TRIGGER: u32 = 10039;
/// The OID of _pg_description.
pub const PG_DESCRIPTION_ARRAY: u32 = 10040;
/// The OID of pg_description.
pub const PG_DESCRIPTION: u32 = 10041;
/// The OID of _pg_cast.
pub const PG_CAST_ARRAY: u32 = 10042;
/// The OID of pg_cast.
pub const PG_CAST: u32 = 10043;
/// The OID of _pg_enum.
pub const PG_ENUM_ARRAY: u32 = 10044;
/// The OID of pg_enum.
pub const PG_ENUM: u32 = 10045;
/// The OID of _pg_namespace.
pub const PG_NAMESPACE_ARRAY: u32 = 10046;
/// The OID of pg_namespace.
pub const PG_NAMESPACE: u32 = 10047;
/// The OID of _pg_conversion.
pub const PG_CONVERSION_ARRAY: u32 = 10048;
/// The OID of pg_conversion.
pub const PG_CONVERSION: u32 = 10049;
/// The OID of _pg_depend.
pub const PG_DEPEND_ARRAY: u32 = 10050;
/// The OID of pg_depend.
pub const PG_DEPEND: u32 = 10051;
/// The OID of _pg_database.
pub const PG_DATABASE_ARRAY: u32 = 10052;
/// The OID of _pg_db_role_setting.
pub const PG_DB_ROLE_SETTING_ARRAY: u32 = 10053;
/// The OID of pg_db_role_setting.
pub const PG_DB_ROLE_SETTING: u32 = 10054;
/// The OID of _pg_tablespace.
pub const PG_TABLESPACE_ARRAY: u32 = 10055;
/// The OID of pg_tablespace.
pub const PG_TABLESPACE: u32 = 10056;
/// The OID of _pg_authid.
pub const PG_AUTHID_ARRAY: u32 = 10057;
/// The OID of _pg_auth_members.
pub const PG_AUTH_MEMBERS_ARRAY: u32 = 10058;
/// The OID of _pg_shdepend.
pub const PG_SHDEPEND_ARRAY: u32 = 10059;
/// The OID of pg_shdepend.
pub const PG_SHDEPEND: u32 = 10060;
/// The OID of _pg_shdescription.
pub const PG_SHDESCRIPTION_ARRAY: u32 = 10061;
/// The OID of pg_shdescription.
pub const PG_SHDESCRIPTION: u32 = 10062;
/// The OID of _pg_ts_config.
pub const PG_TS_CONFIG_ARRAY: u32 = 10063;
/// The OID of pg_ts_config.
pub const PG_TS_CONFIG: u32 = 10064;
/// The OID of _pg_ts_config_map.
pub const PG_TS_CONFIG_MAP_ARRAY: u32 = 10065;
/// The OID of pg_ts_config_map.
pub const PG_TS_CONFIG_MAP: u32 = 10066;
/// The OID of _pg_ts_dict.
pub const PG_TS_DICT_ARRAY: u32 = 10067;
/// The OID of pg_ts_dict.
pub const PG_TS_DICT: u32 = 10068;
/// The OID of _pg_ts_parser.
pub const PG_TS_PARSER_ARRAY: u32 = 10069;
/// The OID of pg_ts_parser.
pub const PG_TS_PARSER: u32 = 10070;
/// The OID of _pg_ts_template.
pub const PG_TS_TEMPLATE_ARRAY: u32 = 10071;
/// The OID of pg_ts_template.
pub const PG_TS_TEMPLATE: u32 = 10072;
/// The OID of _pg_extension.
pub const PG_EXTENSION_ARRAY: u32 = 10073;
/// The OID of pg_extension.
pub const PG_EXTENSION: u32 = 10074;
/// The OID of _pg_foreign_data_wrapper.
pub const PG_FOREIGN_DATA_WRAPPER_ARRAY: u32 = 10075;
/// The OID of pg_foreign_data_wrapper.
pub const PG_FOREIGN_DATA_WRAPPER: u32 = 10076;
/// The OID of _pg_foreign_server.
pub const PG_FOREIGN_SERVER_ARRAY: u32 = 10077;
/// The OID of pg_foreign_server.
pub const PG_FOREIGN_SERVER: u32 = 10078;
/// The OID of _pg_user_mapping.
pub const PG_USER_MAPPING_ARRAY: u32 = 10079;
/// The OID of pg_user_mapping.
pub const PG_USER_MAPPING: u32 = 10080;
/// The OID of _pg_foreign_table.
pub const PG_FOREIGN_TABLE_ARRAY: u32 = 10081;
/// The OID of pg_foreign_table.
pub const PG_FOREIGN_TABLE: u32 = 10082;
/// The OID of _pg_policy.
pub const PG_POLICY_ARRAY: u32 = 10083;
/// The OID of pg_policy.
pub const PG_POLICY: u32 = 10084;
/// The OID of _pg_replication_origin.
pub const PG_REPLICATION_ORIGIN_ARRAY: u32 = 10085;
/// The OID of pg_replication_origin.
pub const PG_REPLICATION_ORIGIN: u32 = 10086;
/// The OID of _pg_default_acl.
pub const PG_DEFAULT_ACL_ARRAY: u32 = 10087;
/// The OID of pg_default_acl.
pub const PG_DEFAULT_ACL: u32 = 10088;
/// The OID of _pg_init_privs.
pub const PG_INIT_PRIVS_ARRAY: u32 = 10089;
/// The OID of pg_init_privs.
pub const PG_INIT_PRIVS: u32 = 10090;
/// The OID of _pg_seclabel.
pub const PG_SECLABEL_ARRAY: u32 = 10091;
/// The OID of pg_seclabel.
pub const PG_SECLABEL: u32 = 10092;
/// The OID of _pg_shseclabel.
pub const PG_SHSECLABEL_ARRAY: u32 = 10093;
/// The OID of _pg_collation.
pub const PG_COLLATION_ARRAY: u32 = 10094;
/// The OID of pg_collation.
pub const PG_COLLATION: u32 = 10095;
/// The OID of _pg_parameter_acl.
pub const PG_PARAMETER_ACL_ARRAY: u32 = 10096;
/// The OID of pg_parameter_acl.
pub const PG_PARAMETER_ACL: u32 = 10097;
/// The OID of _pg_partitioned_table.
pub const PG_PARTITIONED_TABLE_ARRAY: u32 = 10098;
/// The OID of pg_partitioned_table.
pub const PG_PARTITIONED_TABLE: u32 = 10099;
/// The OID of _pg_range.
pub const PG_RANGE_ARRAY: u32 = 10100;
/// The OID of pg_range.
pub const PG_RANGE: u32 = 10101;
/// The OID of _pg_transform.
pub const PG_TRANSFORM_ARRAY: u32 = 10102;
/// The OID of pg_transform.
pub const PG_TRANSFORM: u32 = 10103;
/// The OID of _pg_sequence.
pub const PG_SEQUENCE_ARRAY: u32 = 10104;
/// The OID of pg_sequence.
pub const PG_SEQUENCE: u32 = 10105;
/// The OID of _pg_publication.
pub const PG_PUBLICATION_ARRAY: u32 = 10106;
/// The OID of pg_publication.
pub const PG_PUBLICATION: u32 = 10107;
/// The OID of _pg_publication_namespace.
pub const PG_PUBLICATION_NAMESPACE_ARRAY: u32 = 10108;
/// The OID of pg_publication_namespace.
pub const PG_PUBLICATION_NAMESPACE: u32 = 10109;
/// The OID of _pg_publication_rel.
pub const PG_PUBLICATION_REL_ARRAY: u32 = 10110;
/// The OID of pg_publication_rel.
pub const PG_PUBLICATION_REL: u32 = 10111;
/// The OID of _pg_subscription.
pub const PG_SUBSCRIPTION_ARRAY: u32 = 10112;
/// The OID of _pg_subscription_rel.
pub const PG_SUBSCRIPTION_REL_ARRAY: u32 = 10113;
/// The OID of pg_subscription_rel.
pub const PG_SUBSCRIPTION_REL: u32 = 10114;
/// The OID of _pg_roles.
pub const PG_ROLES_ARRAY: u32 = 12001;
/// The OID of pg_roles.
pub const PG_ROLES: u32 = 12002;
/// The OID of _pg_shadow.
pub const PG_SHADOW_ARRAY: u32 = 12006;
/// The OID of pg_shadow.
pub const PG_SHADOW: u32 = 12007;
/// The OID of _pg_group.
pub const PG_GROUP_ARRAY: u32 = 12011;
/// The OID of pg_group.
pub const PG_GROUP: u32 = 12012;
/// The OID of _pg_user.
pub const PG_USER_ARRAY: u32 = 12015;
/// The OID of pg_user.
pub const PG_USER: u32 = 12016;
/// The OID of _pg_policies.
pub const PG_POLICIES_ARRAY: u32 = 12019;
/// The OID of pg_policies.
pub const PG_POLICIES: u32 = 12020;
/// The OID of _pg_rules.
pub const PG_RULES_ARRAY: u32 = 12024;
/// The OID of pg_rules.
pub const PG_RULES: u32 = 12025;
/// The OID of _pg_views.
pub const PG_VIEWS_ARRAY: u32 = 12029;
/// The OID of pg_views.
pub const PG_VIEWS: u32 = 12030;
/// The OID of _pg_tables.
pub const PG_TABLES_ARRAY: u32 = 12034;
/// The OID of pg_tables.
pub const PG_TABLES: u32 = 12035;
/// The OID of _pg_matviews.
pub const PG_MATVIEWS_ARRAY: u32 = 12039;
/// The OID of pg_matviews.
pub const PG_MATVIEWS: u32 = 12040;
/// The OID of _pg_indexes.
pub const PG_INDEXES_ARRAY: u32 = 12044;
/// The OID of pg_indexes.
pub const PG_INDEXES: u32 = 12045;
/// The OID of _pg_sequences.
pub const PG_SEQUENCES_ARRAY: u32 = 12049;
/// The OID of pg_sequences.
pub const PG_SEQUENCES: u32 = 12050;
/// The OID of _pg_stats.
pub const PG_STATS_ARRAY: u32 = 12054;
/// The OID of pg_stats.
pub const PG_STATS: u32 = 12055;
/// The OID of _pg_stats_ext.
pub const PG_STATS_EXT_ARRAY: u32 = 12059;
/// The OID of pg_stats_ext.
pub const PG_STATS_EXT: u32 = 12060;
/// The OID of _pg_stats_ext_exprs.
pub const PG_STATS_EXT_EXPRS_ARRAY: u32 = 12064;
/// The OID of pg_stats_ext_exprs.
pub const PG_STATS_EXT_EXPRS: u32 = 12065;
/// The OID of _pg_publication_tables.
pub const PG_PUBLICATION_TABLES_ARRAY: u32 = 12069;
/// The OID of pg_publication_tables.
pub const PG_PUBLICATION_TABLES: u32 = 12070;
/// The OID of _pg_locks.
pub const PG_LOCKS_ARRAY: u32 = 12074;
/// The OID of pg_locks.
pub const PG_LOCKS: u32 = 12075;
/// The OID of _pg_cursors.
pub const PG_CURSORS_ARRAY: u32 = 12078;
/// The OID of pg_cursors.
pub const PG_CURSORS: u32 = 12079;
/// The OID of _pg_available_extensions.
pub const PG_AVAILABLE_EXTENSIONS_ARRAY: u32 = 12082;
/// The OID of pg_available_extensions.
pub const PG_AVAILABLE_EXTENSIONS: u32 = 12083;
/// The OID of _pg_available_extension_versions.
pub const PG_AVAILABLE_EXTENSION_VERSIONS_ARRAY: u32 = 12086;
/// The OID of pg_available_extension_versions.
pub const PG_AVAILABLE_EXTENSION_VERSIONS: u32 = 12087;
/// The OID of _pg_prepared_xacts.
pub const PG_PREPARED_XACTS_ARRAY: u32 = 12091;
/// The OID of pg_prepared_xacts.
pub const PG_PREPARED_XACTS: u32 = 12092;
/// The OID of _pg_prepared_statements.
pub const PG_PREPARED_STATEMENTS_ARRAY: u32 = 12096;
/// The OID of pg_prepared_statements.
pub const PG_PREPARED_STATEMENTS: u32 = 12097;
/// The OID of _pg_seclabels.
pub const PG_SECLABELS_ARRAY: u32 = 12100;
/// The OID of pg_seclabels.
pub const PG_SECLABELS: u32 = 12101;
/// The OID of _pg_settings.
pub const PG_SETTINGS_ARRAY: u32 = 12105;
/// The OID of pg_settings.
pub const PG_SETTINGS: u32 = 12106;
/// The OID of _pg_file_settings.
pub const PG_FILE_SETTINGS_ARRAY: u32 = 12111;
/// The OID of pg_file_settings.
pub const PG_FILE_SETTINGS: u32 = 12112;
/// The OID of _pg_hba_file_rules.
pub const PG_HBA_FILE_RULES_ARRAY: u32 = 12115;
/// The OID of pg_hba_file_rules.
pub const PG_HBA_FILE_RULES: u32 = 12116;
/// The OID of _pg_ident_file_mappings.
pub const PG_IDENT_FILE_MAPPINGS_ARRAY: u32 = 12119;
/// The OID of pg_ident_file_mappings.
pub const PG_IDENT_FILE_MAPPINGS: u32 = 12120;
/// The OID of _pg_timezone_abbrevs.
pub const PG_TIMEZONE_ABBREVS_ARRAY: u32 = 12123;
/// The OID of pg_timezone_abbrevs.
pub const PG_TIMEZONE_ABBREVS: u32 = 12124;
/// The OID of _pg_timezone_names.
pub const PG_TIMEZONE_NAMES_ARRAY: u32 = 12127;
/// The OID of pg_timezone_names.
pub const PG_TIMEZONE_NAMES: u32 = 12128;
/// The OID of _pg_config.
pub const PG_CONFIG_ARRAY: u32 = 12131;
/// The OID of pg_config.
pub const PG_CONFIG: u32 = 12132;
/// The OID of _pg_shmem_allocations.
pub const PG_SHMEM_ALLOCATIONS_ARRAY: u32 = 12135;
/// The OID of pg_shmem_allocations.
pub const PG_SHMEM_ALLOCATIONS: u32 = 12136;
/// The OID of _pg_backend_memory_contexts.
pub const PG_BACKEND_MEMORY_CONTEXTS_ARRAY: u32 = 12139;
/// The OID of pg_backend_memory_contexts.
pub const PG_BACKEND_MEMORY_CONTEXTS: u32 = 12140;
/// The OID of _pg_stat_all_tables.
pub const PG_STAT_ALL_TABLES_ARRAY: u32 = 12143;
/// The OID of pg_stat_all_tables.
pub const PG_STAT_ALL_TABLES: u32 = 12144;
/// The OID of _pg_stat_xact_all_tables.
pub const PG_STAT_XACT_ALL_TABLES_ARRAY: u32 = 12148;
/// The OID of pg_stat_xact_all_tables.
pub const PG_STAT_XACT_ALL_TABLES: u32 = 12149;
/// The OID of _pg_stat_sys_tables.
pub const PG_STAT_SYS_TABLES_ARRAY: u32 = 12153;
/// The OID of pg_stat_sys_tables.
pub const PG_STAT_SYS_TABLES: u32 = 12154;
/// The OID of _pg_stat_xact_sys_tables.
pub const PG_STAT_XACT_SYS_TABLES_ARRAY: u32 = 12158;
/// The OID of pg_stat_xact_sys_tables.
pub const PG_STAT_XACT_SYS_TABLES: u32 = 12159;
/// The OID of _pg_stat_user_tables.
pub const PG_STAT_USER_TABLES_ARRAY: u32 = 12162;
/// The OID of pg_stat_user_tables.
pub const PG_STAT_USER_TABLES: u32 = 12163;
/// The OID of _pg_stat_xact_user_tables.
pub const PG_STAT_XACT_USER_TABLES_ARRAY: u32 = 12167;
/// The OID of pg_stat_xact_user_tables.
pub const PG_STAT_XACT_USER_TABLES: u32 = 12168;
/// The OID of _pg_statio_all_tables.
pub const PG_STATIO_ALL_TABLES_ARRAY: u32 = 12171;
/// The OID of pg_statio_all_tables.
pub const PG_STATIO_ALL_TABLES: u32 = 12172;
/// The OID of _pg_statio_sys_tables.
pub const PG_STATIO_SYS_TABLES_ARRAY: u32 = 12176;
/// The OID of pg_statio_sys_tables.
pub const PG_STATIO_SYS_TABLES: u32 = 12177;
/// The OID of _pg_statio_user_tables.
pub const PG_STATIO_USER_TABLES_ARRAY: u32 = 12180;
/// The OID of pg_statio_user_tables.
pub const PG_STATIO_USER_TABLES: u32 = 12181;
/// The OID of _pg_stat_all_indexes.
pub const PG_STAT_ALL_INDEXES_ARRAY: u32 = 12184;
/// The OID of pg_stat_all_indexes.
pub const PG_STAT_ALL_INDEXES: u32 = 12185;
/// The OID of _pg_stat_sys_indexes.
pub const PG_STAT_SYS_INDEXES_ARRAY: u32 = 12189;
/// The OID of pg_stat_sys_indexes.
pub const PG_STAT_SYS_INDEXES: u32 = 12190;
/// The OID of _pg_stat_user_indexes.
pub const PG_STAT_USER_INDEXES_ARRAY: u32 = 12193;
/// The OID of pg_stat_user_indexes.
pub const PG_STAT_USER_INDEXES: u32 = 12194;
/// The OID of _pg_statio_all_indexes.
pub const PG_STATIO_ALL_INDEXES_ARRAY: u32 = 12197;
/// The OID of pg_statio_all_indexes.
pub const PG_STATIO_ALL_INDEXES: u32 = 12198;
/// The OID of _pg_statio_sys_indexes.
pub const PG_STATIO_SYS_INDEXES_ARRAY: u32 = 12202;
/// The OID of pg_statio_sys_indexes.
pub const PG_STATIO_SYS_INDEXES: u32 = 12203;
/// The OID of _pg_statio_user_indexes.
pub const PG_STATIO_USER_INDEXES_ARRAY: u32 = 12206;
/// The OID of pg_statio_user_indexes.
pub const PG_STATIO_USER_INDEXES: u32 = 12207;
/// The OID of _pg_statio_all_sequences.
pub const PG_STATIO_ALL_SEQUENCES_ARRAY: u32 = 12210;
/// The OID of pg_statio_all_sequences.
pub const PG_STATIO_ALL_SEQUENCES: u32 = 12211;
/// The OID of _pg_statio_sys_sequences.
pub const PG_STATIO_SYS_SEQUENCES_ARRAY: u32 = 12215;
/// The OID of pg_statio_sys_sequences.
pub const PG_STATIO_SYS_SEQUENCES: u32 = 12216;
/// The OID of _pg_statio_user_sequences.
pub const PG_STATIO_USER_SEQUENCES_ARRAY: u32 = 12219;
/// The OID of pg_statio_user_sequences.
pub const PG_STATIO_USER_SEQUENCES: u32 = 12220;
/// The OID of _pg_stat_activity.
pub const PG_STAT_ACTIVITY_ARRAY: u32 = 12223;
/// The OID of pg_stat_activity.
pub const PG_STAT_ACTIVITY: u32 = 12224;
/// The OID of _pg_stat_replication.
pub const PG_STAT_REPLICATION_ARRAY: u32 = 12228;
/// The OID of pg_stat_replication.
pub const PG_STAT_REPLICATION: u32 = 12229;
/// The OID of _pg_stat_slru.
pub const PG_STAT_SLRU_ARRAY: u32 = 12233;
/// The OID of pg_stat_slru.
pub const PG_STAT_SLRU: u32 = 12234;
/// The OID of _pg_stat_wal_receiver.
pub const PG_STAT_WAL_RECEIVER_ARRAY: u32 = 12237;
/// The OID of pg_stat_wal_receiver.
pub const PG_STAT_WAL_RECEIVER: u32 = 12238;
/// The OID of _pg_stat_recovery_prefetch.
pub const PG_STAT_RECOVERY_PREFETCH_ARRAY: u32 = 12241;
/// The OID of pg_stat_recovery_prefetch.
pub const PG_STAT_RECOVERY_PREFETCH: u32 = 12242;
/// The OID of _pg_stat_subscription.
pub const PG_STAT_SUBSCRIPTION_ARRAY: u32 = 12245;
/// The OID of pg_stat_subscription.
pub const PG_STAT_SUBSCRIPTION: u32 = 12246;
/// The OID of _pg_stat_ssl.
pub const PG_STAT_SSL_ARRAY: u32 = 12250;
/// The OID of pg_stat_ssl.
pub const PG_STAT_SSL: u32 = 12251;
/// The OID of _pg_stat_gssapi.
pub const PG_STAT_GSSAPI_ARRAY: u32 = 12254;
/// The OID of pg_stat_gssapi.
pub const PG_STAT_GSSAPI: u32 = 12255;
/// The OID of _pg_replication_slots.
pub const PG_REPLICATION_SLOTS_ARRAY: u32 = 12258;
/// The OID of pg_replication_slots.
pub const PG_REPLICATION_SLOTS: u32 = 12259;
/// The OID of _pg_stat_replication_slots.
pub const PG_STAT_REPLICATION_SLOTS_ARRAY: u32 = 12263;
/// The OID of pg_stat_replication_slots.
pub const PG_STAT_REPLICATION_SLOTS: u32 = 12264;
/// The OID of _pg_stat_database.
pub const PG_STAT_DATABASE_ARRAY: u32 = 12267;
/// The OID of pg_stat_database.
pub const PG_STAT_DATABASE: u32 = 12268;
/// The OID of _pg_stat_database_conflicts.
pub const PG_STAT_DATABASE_CONFLICTS_ARRAY: u32 = 12272;
/// The OID of pg_stat_database_conflicts.
pub const PG_STAT_DATABASE_CONFLICTS: u32 = 12273;
/// The OID of _pg_stat_user_functions.
pub const PG_STAT_USER_FUNCTIONS_ARRAY: u32 = 12276;
/// The OID of pg_stat_user_functions.
pub const PG_STAT_USER_FUNCTIONS: u32 = 12277;
/// The OID of _pg_stat_xact_user_functions.
pub const PG_STAT_XACT_USER_FUNCTIONS_ARRAY: u32 = 12281;
/// The OID of pg_stat_xact_user_functions.
pub const PG_STAT_XACT_USER_FUNCTIONS: u32 = 12282;
/// The OID of _pg_stat_archiver.
pub const PG_STAT_ARCHIVER_ARRAY: u32 = 12286;
/// The OID of pg_stat_archiver.
pub const PG_STAT_ARCHIVER: u32 = 12287;
/// The OID of _pg_stat_bgwriter.
pub const PG_STAT_BGWRITER_ARRAY: u32 = 12290;
/// The OID of pg_stat_bgwriter.
pub const PG_STAT_BGWRITER: u32 = 12291;
/// The OID of _pg_stat_wal.
pub const PG_STAT_WAL_ARRAY: u32 = 12294;
/// The OID of pg_stat_wal.
pub const PG_STAT_WAL: u32 = 12295;
/// The OID of _pg_stat_progress_analyze.
pub const PG_STAT_PROGRESS_ANALYZE_ARRAY: u32 = 12298;
/// The OID of pg_stat_progress_analyze.
pub const PG_STAT_PROGRESS_ANALYZE: u32 = 12299;
/// The OID of _pg_stat_progress_vacuum.
pub const PG_STAT_PROGRESS_VACUUM_ARRAY: u32 = 12303;
/// The OID of pg_stat_progress_vacuum.
pub const PG_STAT_PROGRESS_VACUUM: u32 = 12304;
/// The OID of _pg_stat_progress_cluster.
pub const PG_STAT_PROGRESS_CLUSTER_ARRAY: u32 = 12308;
/// The OID of pg_stat_progress_cluster.
pub const PG_STAT_PROGRESS_CLUSTER: u32 = 12309;
/// The OID of _pg_stat_progress_create_index.
pub const PG_STAT_PROGRESS_CREATE_INDEX_ARRAY: u32 = 12313;
/// The OID of pg_stat_progress_create_index.
pub const PG_STAT_PROGRESS_CREATE_INDEX: u32 = 12314;
/// The OID of _pg_stat_progress_basebackup.
pub const PG_STAT_PROGRESS_BASEBACKUP_ARRAY: u32 = 12318;
/// The OID of pg_stat_progress_basebackup.
pub const PG_STAT_PROGRESS_BASEBACKUP: u32 = 12319;
/// The OID of _pg_stat_progress_copy.
pub const PG_STAT_PROGRESS_COPY_ARRAY: u32 = 12323;
/// The OID of pg_stat_progress_copy.
pub const PG_STAT_PROGRESS_COPY: u32 = 12324;
/// The OID of _pg_user_mappings.
pub const PG_USER_MAPPINGS_ARRAY: u32 = 12328;
/// The OID of pg_user_mappings.
pub const PG_USER_MAPPINGS: u32 = 12329;
/// The OID of _pg_replication_origin_status.
pub const PG_REPLICATION_ORIGIN_STATUS_ARRAY: u32 = 12333;
/// The OID of pg_replication_origin_status.
pub const PG_REPLICATION_ORIGIN_STATUS: u32 = 12334;
/// The OID of _pg_stat_subscription_stats.
pub const PG_STAT_SUBSCRIPTION_STATS_ARRAY: u32 = 12337;
/// The OID of pg_stat_subscription_stats.
pub const PG_STAT_SUBSCRIPTION_STATS: u32 = 12338;
/// The OID of _cardinal_number.
pub const CARDINAL_NUMBER_ARRAY: u32 = 13560;
/// The OID of cardinal_number.
pub const CARDINAL_NUMBER: u32 = 13561;
/// The OID of _character_data.
pub const CHARACTER_DATA_ARRAY: u32 = 13563;
/// The OID of character_data.
pub const CHARACTER_DATA: u32 = 13564;
/// The OID of _sql_identifier.
pub const SQL_IDENTIFIER_ARRAY: u32 = 13565;
/// The OID of sql_identifier.
pub const SQL_IDENTIFIER: u32 = 13566;
/// The OID of _information_schema_catalog_name.
pub const INFORMATION_SCHEMA_CATALOG_NAME_ARRAY: u32 = 13568;
/// The OID of information_schema_catalog_name.
pub const INFORMATION_SCHEMA_CATALOG_NAME: u32 = 13569;
/// The OID of _time_stamp.
pub const TIME_STAMP_ARRAY: u32 = 13571;
/// The OID of time_stamp.
pub const TIME_STAMP: u32 = 13572;
/// The OID of _yes_or_no.
pub const YES_OR_NO_ARRAY: u32 = 13573;
/// The OID of yes_or_no.
pub const YES_OR_NO: u32 = 13574;
/// The OID of _applicable_roles.
pub const APPLICABLE_ROLES_ARRAY: u32 = 13577;
/// The OID of applicable_roles.
pub const APPLICABLE_ROLES: u32 = 13578;
/// The OID of _administrable_role_authorizations.
pub const ADMINISTRABLE_ROLE_AUTHORIZATIONS_ARRAY: u32 = 13582;
/// The OID of administrable_role_authorizations.
pub const ADMINISTRABLE_ROLE_AUTHORIZATIONS: u32 = 13583;
/// The OID of _attributes.
pub const ATTRIBUTES_ARRAY: u32 = 13586;
/// The OID of attributes.
pub const ATTRIBUTES: u32 = 13587;
/// The OID of _character_sets.
pub const CHARACTER_SETS_ARRAY: u32 = 13591;
/// The OID of character_sets.
pub const CHARACTER_SETS: u32 = 13592;
/// The OID of _check_constraint_routine_usage.
pub const CHECK_CONSTRAINT_ROUTINE_USAGE_ARRAY: u32 = 13596;
/// The OID of check_constraint_routine_usage.
pub const CHECK_CONSTRAINT_ROUTINE_USAGE: u32 = 13597;
/// The OID of _check_constraints.
pub const CHECK_CONSTRAINTS_ARRAY: u32 = 13601;
/// The OID of check_constraints.
pub const CHECK_CONSTRAINTS: u32 = 13602;
/// The OID of _collations.
pub const COLLATIONS_ARRAY: u32 = 13606;
/// The OID of collations.
pub const COLLATIONS: u32 = 13607;
/// The OID of _collation_character_set_applicability.
pub const COLLATION_CHARACTER_SET_APPLICABILITY_ARRAY: u32 = 13611;
/// The OID of collation_character_set_applicability.
pub const COLLATION_CHARACTER_SET_APPLICABILITY: u32 = 13612;
/// The OID of _column_column_usage.
pub const COLUMN_COLUMN_USAGE_ARRAY: u32 = 13616;
/// The OID of column_column_usage.
pub const COLUMN_COLUMN_USAGE: u32 = 13617;
/// The OID of _column_domain_usage.
pub const COLUMN_DOMAIN_USAGE_ARRAY: u32 = 13621;
/// The OID of column_domain_usage.
pub const COLUMN_DOMAIN_USAGE: u32 = 13622;
/// The OID of _column_privileges.
pub const COLUMN_PRIVILEGES_ARRAY: u32 = 13626;
/// The OID of column_privileges.
pub const COLUMN_PRIVILEGES: u32 = 13627;
/// The OID of _column_udt_usage.
pub const COLUMN_UDT_USAGE_ARRAY: u32 = 13631;
/// The OID of column_udt_usage.
pub const COLUMN_UDT_USAGE: u32 = 13632;
/// The OID of _columns.
pub const COLUMNS_ARRAY: u32 = 13636;
/// The OID of columns.
pub const COLUMNS: u32 = 13637;
/// The OID of _constraint_column_usage.
pub const CONSTRAINT_COLUMN_USAGE_ARRAY: u32 = 13641;
/// The OID of constraint_column_usage.
pub const CONSTRAINT_COLUMN_USAGE: u32 = 13642;
/// The OID of _constraint_table_usage.
pub const CONSTRAINT_TABLE_USAGE_ARRAY: u32 = 13646;
/// The OID of constraint_table_usage.
pub const CONSTRAINT_TABLE_USAGE: u32 = 13647;
/// The OID of _domain_constraints.
pub const DOMAIN_CONSTRAINTS_ARRAY: u32 = 13651;
/// The OID of domain_constraints.
pub const DOMAIN_CONSTRAINTS: u32 = 13652;
/// The OID of _domain_udt_usage.
pub const DOMAIN_UDT_USAGE_ARRAY: u32 = 13656;
/// The OID of domain_udt_usage.
pub const DOMAIN_UDT_USAGE: u32 = 13657;
/// The OID of _domains.
pub const DOMAINS_ARRAY: u32 = 13661;
/// The OID of domains.
pub const DOMAINS: u32 = 13662;
/// The OID of _enabled_roles.
pub const ENABLED_ROLES_ARRAY: u32 = 13666;
/// The OID of enabled_roles.
pub const ENABLED_ROLES: u32 = 13667;
/// The OID of _key_column_usage.
pub const KEY_COLUMN_USAGE_ARRAY: u32 = 13670;
/// The OID of key_column_usage.
pub const KEY_COLUMN_USAGE: u32 = 13671;
/// The OID of _parameters.
pub const PARAMETERS_ARRAY: u32 = 13675;
/// The OID of parameters.
pub const PARAMETERS: u32 = 13676;
/// The OID of _referential_constraints.
pub const REFERENTIAL_CONSTRAINTS_ARRAY: u32 = 13680;
/// The OID of referential_constraints.
pub const REFERENTIAL_CONSTRAINTS: u32 = 13681;
/// The OID of _role_column_grants.
pub const ROLE_COLUMN_GRANTS_ARRAY: u32 = 13685;
/// The OID of role_column_grants.
pub const ROLE_COLUMN_GRANTS: u32 = 13686;
/// The OID of _routine_column_usage.
pub const ROUTINE_COLUMN_USAGE_ARRAY: u32 = 13689;
/// The OID of routine_column_usage.
pub const ROUTINE_COLUMN_USAGE: u32 = 13690;
/// The OID of _routine_privileges.
pub const ROUTINE_PRIVILEGES_ARRAY: u32 = 13694;
/// The OID of routine_privileges.
pub const ROUTINE_PRIVILEGES: u32 = 13695;
/// The OID of _role_routine_grants.
pub const ROLE_ROUTINE_GRANTS_ARRAY: u32 = 13699;
/// The OID of role_routine_grants.
pub const ROLE_ROUTINE_GRANTS: u32 = 13700;
/// The OID of _routine_routine_usage.
pub const ROUTINE_ROUTINE_USAGE_ARRAY: u32 = 13703;
/// The OID of routine_routine_usage.
pub const ROUTINE_ROUTINE_USAGE: u32 = 13704;
/// The OID of _routine_sequence_usage.
pub const ROUTINE_SEQUENCE_USAGE_ARRAY: u32 = 13708;
/// The OID of routine_sequence_usage.
pub const ROUTINE_SEQUENCE_USAGE: u32 = 13709;
/// The OID of _routine_table_usage.
pub const ROUTINE_TABLE_USAGE_ARRAY: u32 = 13713;
/// The OID of routine_table_usage.
pub const ROUTINE_TABLE_USAGE: u32 = 13714;
/// The OID of _routines.
pub const ROUTINES_ARRAY: u32 = 13718;
/// The OID of routines.
pub const ROUTINES: u32 = 13719;
/// The OID of _schemata.
pub const SCHEMATA_ARRAY: u32 = 13723;
/// The OID of schemata.
pub const SCHEMATA: u32 = 13724;
/// The OID of _sequences.
pub const SEQUENCES_ARRAY: u32 = 13727;
/// The OID of sequences.
pub const SEQUENCES: u32 = 13728;
/// The OID of _sql_features.
pub const SQL_FEATURES_ARRAY: u32 = 13732;
/// The OID of sql_features.
pub const SQL_FEATURES: u32 = 13733;
/// The OID of _sql_implementation_info.
pub const SQL_IMPLEMENTATION_INFO_ARRAY: u32 = 13737;
/// The OID of sql_implementation_info.
pub const SQL_IMPLEMENTATION_INFO: u32 = 13738;
/// The OID of _sql_parts.
pub const SQL_PARTS_ARRAY: u32 = 13742;
/// The OID of sql_parts.
pub const SQL_PARTS: u32 = 13743;
/// The OID of _sql_sizing.
pub const SQL_SIZING_ARRAY: u32 = 13747;
/// The OID of sql_sizing.
pub const SQL_SIZING: u32 = 13748;
/// The OID of _table_constraints.
pub const TABLE_CONSTRAINTS_ARRAY: u32 = 13752;
/// The OID of table_constraints.
pub const TABLE_CONSTRAINTS: u32 = 13753;
/// The OID of _table_privileges.
pub const TABLE_PRIVILEGES_ARRAY: u32 = 13757;
/// The OID of table_privileges.
pub const TABLE_PRIVILEGES: u32 = 13758;
/// The OID of _role_table_grants.
pub const ROLE_TABLE_GRANTS_ARRAY: u32 = 13762;
/// The OID of role_table_grants.
pub const ROLE_TABLE_GRANTS: u32 = 13763;
/// The OID of _tables.
pub const TABLES_ARRAY: u32 = 13766;
/// The OID of tables.
pub const TABLES: u32 = 13767;
/// The OID of _transforms.
pub const TRANSFORMS_ARRAY: u32 = 13771;
/// The OID of transforms.
pub const TRANSFORMS: u32 = 13772;
/// The OID of _triggered_update_columns.
pub const TRIGGERED_UPDATE_COLUMNS_ARRAY: u32 = 13776;
/// The OID of triggered_update_columns.
pub const TRIGGERED_UPDATE_COLUMNS: u32 = 13777;
/// The OID of _triggers.
pub const TRIGGERS_ARRAY: u32 = 13781;
/// The OID of triggers.
pub const TRIGGERS: u32 = 13782;
/// The OID of _udt_privileges.
pub const UDT_PRIVILEGES_ARRAY: u32 = 13786;
/// The OID of udt_privileges.
pub const UDT_PRIVILEGES: u32 = 13787;
/// The OID of _role_udt_grants.
pub const ROLE_UDT_GRANTS_ARRAY: u32 = 13791;
/// The OID of role_udt_grants.
pub const ROLE_UDT_GRANTS: u32 = 13792;
/// The OID of _usage_privileges.
pub const USAGE_PRIVILEGES_ARRAY: u32 = 13795;
/// The OID of usage_privileges.
pub const USAGE_PRIVILEGES: u32 = 13796;
/// The OID of _role_usage_grants.
pub const ROLE_USAGE_GRANTS_ARRAY: u32 = 13800;
/// The OID of role_usage_grants.
pub const ROLE_USAGE_GRANTS: u32 = 13801;
/// The OID of _user_defined_types.
pub const USER_DEFINED_TYPES_ARRAY: u32 = 13804;
/// The OID of user_defined_types.
pub const USER_DEFINED_TYPES: u32 = 13805;
/// The OID of _view_column_usage.
pub const VIEW_COLUMN_USAGE_ARRAY: u32 = 13809;
/// The OID of view_column_usage.
pub const VIEW_COLUMN_USAGE: u32 = 13810;
/// The OID of _view_routine_usage.
pub const VIEW_ROUTINE_USAGE_ARRAY: u32 = 13814;
/// The OID of view_routine_usage.
pub const VIEW_ROUTINE_USAGE: u32 = 13815;
/// The OID of _view_table_usage.
pub const VIEW_TABLE_USAGE_ARRAY: u32 = 13819;
/// The OID of view_table_usage.
pub const VIEW_TABLE_USAGE: u32 = 13820;
/// The OID of _views.
pub const VIEWS_ARRAY: u32 = 13824;
/// The OID of views.
pub const VIEWS: u32 = 13825;
/// The OID of _data_type_privileges.
pub const DATA_TYPE_PRIVILEGES_ARRAY: u32 = 13829;
/// The OID of data_type_privileges.
pub const DATA_TYPE_PRIVILEGES: u32 = 13830;
/// The OID of _element_types.
pub const ELEMENT_TYPES_ARRAY: u32 = 13834;
/// The OID of element_types.
pub const ELEMENT_TYPES: u32 = 13835;
/// The OID of __pg_foreign_table_columns.
pub const _PG_FOREIGN_TABLE_COLUMNS_ARRAY: u32 = 13839;
/// The OID of _pg_foreign_table_columns.
pub const PG_FOREIGN_TABLE_COLUMNS_ARRAY: u32 = 13840;
/// The OID of _column_options.
pub const COLUMN_OPTIONS_ARRAY: u32 = 13844;
/// The OID of column_options.
pub const COLUMN_OPTIONS: u32 = 13845;
/// The OID of __pg_foreign_data_wrappers.
pub const _PG_FOREIGN_DATA_WRAPPERS_ARRAY: u32 = 13848;
/// The OID of _pg_foreign_data_wrappers.
pub const PG_FOREIGN_DATA_WRAPPERS_ARRAY: u32 = 13849;
/// The OID of _foreign_data_wrapper_options.
pub const FOREIGN_DATA_WRAPPER_OPTIONS_ARRAY: u32 = 13852;
/// The OID of foreign_data_wrapper_options.
pub const FOREIGN_DATA_WRAPPER_OPTIONS: u32 = 13853;
/// The OID of _foreign_data_wrappers.
pub const FOREIGN_DATA_WRAPPERS_ARRAY: u32 = 13856;
/// The OID of foreign_data_wrappers.
pub const FOREIGN_DATA_WRAPPERS: u32 = 13857;
/// The OID of __pg_foreign_servers.
pub const _PG_FOREIGN_SERVERS_ARRAY: u32 = 13860;
/// The OID of _pg_foreign_servers.
pub const PG_FOREIGN_SERVERS_ARRAY: u32 = 13861;
/// The OID of _foreign_server_options.
pub const FOREIGN_SERVER_OPTIONS_ARRAY: u32 = 13865;
/// The OID of foreign_server_options.
pub const FOREIGN_SERVER_OPTIONS: u32 = 13866;
/// The OID of _foreign_servers.
pub const FOREIGN_SERVERS_ARRAY: u32 = 13869;
/// The OID of foreign_servers.
pub const FOREIGN_SERVERS: u32 = 13870;
/// The OID of __pg_foreign_tables.
pub const _PG_FOREIGN_TABLES_ARRAY: u32 = 13873;
/// The OID of _pg_foreign_tables.
pub const PG_FOREIGN_TABLES_ARRAY: u32 = 13874;
/// The OID of _foreign_table_options.
pub const FOREIGN_TABLE_OPTIONS_ARRAY: u32 = 13878;
/// The OID of foreign_table_options.
pub const FOREIGN_TABLE_OPTIONS: u32 = 13879;
/// The OID of _foreign_tables.
pub const FOREIGN_TABLES_ARRAY: u32 = 13882;
/// The OID of foreign_tables.
pub const FOREIGN_TABLES: u32 = 13883;
/// The OID of __pg_user_mappings.
pub const _PG_USER_MAPPINGS_ARRAY: u32 = 13886;
/// The OID of _pg_user_mappings.
pub const PG_USER_MAPPINGS_ARRAY_13887: u32 = 13887;
/// The OID of _user_mapping_options.
pub const USER_MAPPING_OPTIONS_ARRAY: u32 = 13891;
/// The OID of user_mapping_options.
pub const USER_MAPPING_OPTIONS: u32 = 13892;
/// The OID of _user_mappings.
pub const USER_MAPPINGS_ARRAY: u32 = 13896;
/// The OID of user_mappings.
pub const USER_MAPPINGS: u32 = 13897;

/// name returns the constant name of a built-in type OID.
pub fn name(oid: u32) -> Option<&'static str> {
    Some(match oid {
        16 => "BOOL",
        17 => "BYTEA",
        18 => "CHAR",
        19 => "NAME",
        20 => "INT8",
        21 => "INT2",
        22 => "INT2VECTOR",
        23 => "INT4",
        24 => "REGPROC",
        25 => "TEXT",
        26 => "OID",
        27 => "TID",
        28 => "XID",
        29 => "CID",
        30 => "OIDVECTOR",
        32 => "PG_DDL_COMMAND",
        71 => "PG_TYPE",
        75 => "PG_ATTRIBUTE",
        81 => "PG_PROC",
        83 => "PG_CLASS",
        114 => "JSON",
        142 => "XML",
        143 => "XML_ARRAY",
        194 => "PG_NODE_TREE",
        199 => "JSON_ARRAY",
        210 => "PG_TYPE_ARRAY",
        269 => "TABLE_AM_HANDLER",
        270 => "PG_ATTRIBUTE_ARRAY",
        271 => "XID8_ARRAY",
        272 => "PG_PROC_ARRAY",
        273 => "PG_CLASS_ARRAY",
        325 => "INDEX_AM_HANDLER",
        600 => "POINT",
        601 => "LSEG",
        602 => "PATH",
        603 => "BOX",
        604 => "POLYGON",
        628 => "LINE",
        629 => "LINE_ARRAY",
        650 => "CIDR",
        651 => "CIDR_ARRAY",
        700 => "FLOAT4",
        701 => "FLOAT8",
        705 => "UNKNOWN",
        718 => "CIRCLE",
        719 => "CIRCLE_ARRAY",
        774 => "MACADDR8",
        775 => "MACADDR8_ARRAY",
        790 => "MONEY",
        791 => "MONEY_ARRAY",
        829 => "MACADDR",
        869 => "INET",
        1000 => "BOOL_ARRAY",
        1001 => "BYTEA_ARRAY",
        1002 => "CHAR_ARRAY",
        1003 => "NAME_ARRAY",
        1005 => "INT2_ARRAY",
        1006 => "INT2VECTOR_ARRAY",
        1007 => "INT4_ARRAY",
        1008 => "REGPROC_ARRAY",
        1009 => "TEXT_ARRAY",
        1010 => "TID_ARRAY",
        1011 => "XID_ARRAY",
        1012 => "CID_ARRAY",
        1013 => "OIDVECTOR_ARRAY",
        1014 => "BPCHAR_ARRAY",
        1015 => "VARCHAR_ARRAY",
        1016 => "INT8_ARRAY",
        1017 => "POINT_ARRAY",
        1018 => "LSEG_ARRAY",
        1019 => "PATH_ARRAY",
        1020 => "BOX_ARRAY",
        1021 => "FLOAT4_ARRAY",
        1022 => "FLOAT8_ARRAY",
        1027 => "POLYGON_ARRAY",
        1028 => "OID_ARRAY",
        1033 => "ACLITEM",
        1034 => "ACLITEM_ARRAY",
        1040 => "MACADDR_ARRAY",
        1041 => "INET_ARRAY",
        1042 => "BPCHAR",
        1043 => "VARCHAR",
        1082 => "DATE",
        1083 => "TIME",
        1114 => "TIMESTAMP",
        1115 => "TIMESTAMP_ARRAY",
        1182 => "DATE_ARRAY",
        1183 => "TIME_ARRAY",
        1184 => "TIMESTAMPTZ",
        1185 => "TIMESTAMPTZ_ARRAY",
        1186 => "INTERVAL",
        1187 => "INTERVAL_ARRAY",
        1231 => "NUMERIC_ARRAY",
        1248 => "PG_DATABASE",
        1263 => "CSTRING_ARRAY",
        1266 => "TIMETZ",
        1270 => "TIMETZ_ARRAY",
        1560 => "BIT",
        1561 => "BIT_ARRAY",
        1562 => "VARBIT",
        1563 => "VARBIT_ARRAY",
        1700 => "NUMERIC",
        1790 => "REFCURSOR",
        2201 => "REFCURSOR_ARRAY",
        2202 => "REGPROCEDURE",
        2203 => "REGOPER",
        2204 => "REGOPERATOR",
        2205 => "REGCLASS",
        2206 => "REGTYPE",
        2207 => "REGPROCEDURE_ARRAY",
        2208 => "REGOPER_ARRAY",
        2209 => "REGOPERATOR_ARRAY",
        2210 => "REGCLASS_ARRAY",
        2211 => "REGTYPE_ARRAY",
        2249 => "RECORD",
        2275 => "CSTRING",
        2276 => "ANY",
        2277 => "ANYARRAY",
        2278 => "VOID",
        2279 => "TRIGGER",
        2280 => "LANGUAGE_HANDLER",
        2281 => "INTERNAL",
        2283 => "ANYELEMENT",
        2287 => "RECORD_ARRAY",
        2776 => "ANYNONARRAY",
        2842 => "PG_AUTHID",
        2843 => "PG_AUTH_MEMBERS",
        2949 => "TXID_SNAPSHOT_ARRAY",
        2950 => "UUID",
        2951 => "UUID_ARRAY",
        2970 => "TXID_SNAPSHOT",
        3115 => "FDW_HANDLER",
        3220 => "PG_LSN",
        3221 => "PG_LSN_ARRAY",
        3310 => "TSM_HANDLER",
        3361 => "PG_NDISTINCT",
        3402 => "PG_DEPENDENCIES",
        3500 => "ANYENUM",
        3614 => "TSVECTOR",
        3615 => "TSQUERY",
        3642 => "GTSVECTOR",
        3643 => "TSVECTOR_ARRAY",
        3644 => "GTSVECTOR_ARRAY",
        3645 => "TSQUERY_ARRAY",
        3734 => "REGCONFIG",
        3735 => "REGCONFIG_ARRAY",
        3769 => "REGDICTIONARY",
        3770 => "REGDICTIONARY_ARRAY",
        3802 => "JSONB",
        3807 => "JSONB_ARRAY",
        3831 => "ANYRANGE",
        3838 => "EVENT_TRIGGER",
        3904 => "INT4RANGE",
        3905 => "INT4RANGE_ARRAY",
        3906 => "NUMRANGE",
        3907 => "NUMRANGE_ARRAY",
        3908 => "TSRANGE",
        3909 => "TSRANGE_ARRAY",
        3910 => "TSTZRANGE",
        3911 => "TSTZRANGE_ARRAY",
        3912 => "DATERANGE",
        3913 => "DATERANGE_ARRAY",
        3926 => "INT8RANGE",
        3927 => "INT8RANGE_ARRAY",
        4066 => "PG_SHSECLABEL",
        4072 => "JSONPATH",
        4073 => "JSONPATH_ARRAY",
        4089 => "REGNAMESPACE",
        4090 => "REGNAMESPACE_ARRAY",
        4096 => "REGROLE",
        4097 => "REGROLE_ARRAY",
        4191 => "REGCOLLATION",
        4192 => "REGCOLLATION_ARRAY",
        4451 => "INT4MULTIRANGE",
        4532 => "NUMMULTIRANGE",
        4533 => "TSMULTIRANGE",
        4534 => "TSTZMULTIRANGE",
        4535 => "DATEMULTIRANGE",
        4536 => "INT8MULTIRANGE",
        4537 => "ANYMULTIRANGE",
        4538 => "ANYCOMPATIBLEMULTIRANGE",
        4600 => "PG_BRIN_BLOOM_SUMMARY",
        4601 => "PG_BRIN_MINMAX_MULTI_SUMMARY",
        5017 => "PG_MCV_LIST",
        5038 => "PG_SNAPSHOT",
        5039 => "PG_SNAPSHOT_ARRAY",
        5069 => "XID8",
        5077 => "ANYCOMPATIBLE",
        5078 => "ANYCOMPATIBLEARRAY",
        5079 => "ANYCOMPATIBLENONARRAY",
        5080 => "ANYCOMPATIBLERANGE",
        6101 => "PG_SUBSCRIPTION",
        6150 => "INT4MULTIRANGE_ARRAY",
        6151 => "NUMMULTIRANGE_ARRAY",
        6152 => "TSMULTIRANGE_ARRAY",
        6153 => "TSTZMULTIRANGE_ARRAY",
        6155 => "DATEMULTIRANGE_ARRAY",
        6157 => "INT8MULTIRANGE_ARRAY",
        10000 => "PG_ATTRDEF_ARRAY",
        10001 => "PG_ATTRDEF",
        10002 => "PG_CONSTRAINT_ARRAY",
        10003 => "PG_CONSTRAINT",
        10004 => "PG_INHERITS_ARRAY",
        10005 => "PG_INHERITS",
        10006 => "PG_INDEX_ARRAY",
        10007 => "PG_INDEX",
        10008 => "PG_OPERATOR_ARRAY",
        10009 => "PG_OPERATOR",
        10010 => "PG_OPFAMILY_ARRAY",
        10011 => "PG_OPFAMILY",
        10012 => "PG_OPCLASS_ARRAY",
        10013 => "PG_OPCLASS",
        10014 => "PG_AM_ARRAY",
        10015 => "PG_AM",
        10016 => "PG_AMOP_ARRAY",
        10017 => "PG_AMOP",
        10018 => "PG_AMPROC_ARRAY",
        10019 => "PG_AMPROC",
        10020 => "PG_LANGUAGE_ARRAY",
        10021 => "PG_LANGUAGE",
        10022 => "PG_LARGEOBJECT_METADATA_ARRAY",
        10023 => "PG_LARGEOBJECT_METADATA",
        10024 => "PG_LARGEOBJECT_ARRAY",
        10025 => "PG_LARGEOBJECT",
        10026 => "PG_AGGREGATE_ARRAY",
        10027 => "PG_AGGREGATE",
        10028 => "PG_STATISTIC_ARRAY",
        10029 => "PG_STATISTIC",
        10030 => "PG_STATISTIC_EXT_ARRAY",
        10031 => "PG_STATISTIC_EXT",
        10032 => "PG_STATISTIC_EXT_DATA_ARRAY",
        10033 => "PG_STATISTIC_EXT_DATA",
        10034 => "PG_REWRITE_ARRAY",
        10035 => "PG_REWRITE",
        10036 => "PG_TRIGGER_ARRAY",
        10037 => "PG_TRIGGER",
        10038 => "PG_EVENT_TRIGGER_ARRAY",
        10039 => "PG_EVENT_TRIGGER",
        10040 => "PG_DESCRIPTION_ARRAY",
        10041 => "PG_DESCRIPTION",
        10042 => "PG_CAST_ARRAY",
        10043 => "PG_CAST",
        10044 => "PG_ENUM_ARRAY",
        10045 => "PG_ENUM",
        10046 => "PG_NAMESPACE_ARRAY",
        10047 => "PG_NAMESPACE",
        10048 => "PG_CONVERSION_ARRAY",
        10049 => "PG_CONVERSION",
        10050 => "PG_DEPEND_ARRAY",
        10051 => "PG_DEPEND",
        10052 => "PG_DATABASE_ARRAY",
        10053 => "PG_DB_ROLE_SETTING_ARRAY",
        10054 => "PG_DB_ROLE_SETTING",
        10055 => "PG_TABLESPACE_ARRAY",
        10056 => "PG_TABLESPACE",
        10057 => "PG_AUTHID_ARRAY",
        10058 => "PG_AUTH_MEMBERS_ARRAY",
        10059 => "PG_SHDEPEND_ARRAY",
        10060 => "PG_SHDEPEND",
        10061 => "PG_SHDESCRIPTION_ARRAY",
        10062 => "PG_SHDESCRIPTION",
        10063 => "PG_TS_CONFIG_ARRAY",
        10064 => "PG_TS_CONFIG",
        10065 => "PG_TS_CONFIG_MAP_ARRAY",
        10066 => "PG_TS_CONFIG_MAP",
        10067 => "PG_TS_DICT_ARRAY",
        10068 => "PG_TS_DICT",
        10069 => "PG_TS_PARSER_ARRAY",
        10070 => "PG_TS_PARSER",
        10071 => "PG_TS_TEMPLATE_ARRAY",
        10072 => "PG_TS_TEMPLATE",
        10073 => "PG_EXTENSION_ARRAY",
        10074 => "PG_EXTENSION",
        10075 => "PG_FOREIGN_DATA_WRAPPER_ARRAY",
        10076 => "PG_FOREIGN_DATA_WRAPPER",
        10077 => "PG_FOREIGN_SERVER_ARRAY",
        10078 => "PG_FOREIGN_SERVER",
        10079 => "PG_USER_MAPPING_ARRAY",
        10080 => "PG_USER_MAPPING",
        10081 => "PG_FOREIGN_TABLE_ARRAY",
        10082 => "PG_FOREIGN_TABLE",
        10083 => "PG_POLICY_ARRAY",
        10084 => "PG_POLICY",
        10085 => "PG_REPLICATION_ORIGIN_ARRAY",
        10086 => "PG_REPLICATION_ORIGIN",
        10087 => "PG_DEFAULT_ACL_ARRAY",
        10088 => "PG_DEFAULT_ACL",
        10089 => "PG_INIT_PRIVS_ARRAY",
        10090 => "PG_INIT_PRIVS",
        10091 => "PG_SECLABEL_ARRAY",
        10092 => "PG_SECLABEL",
        10093 => "PG_SHSECLABEL_ARRAY",
        10094 => "PG_COLLATION_ARRAY",
        10095 => "PG_COLLATION",
        10096 => "PG_PARAMETER_ACL_ARRAY",
        10097 => "PG_PARAMETER_ACL",
        10098 => "PG_PARTITIONED_TABLE_ARRAY",
        10099 => "PG_PARTITIONED_TABLE",
        10100 => "PG_RANGE_ARRAY",
        10101 => "PG_RANGE",
        10102 => "PG_TRANSFORM_ARRAY",
        10103 => "PG_TRANSFORM",
        10104 => "PG_SEQUENCE_ARRAY",
        10105 => "PG_SEQUENCE",
        10106 => "PG_PUBLICATION_ARRAY",
        10107 => "PG_PUBLICATION",
        10108 => "PG_PUBLICATION_NAMESPACE_ARRAY",
        10109 => "PG_PUBLICATION_NAMESPACE",
        10110 => "PG_PUBLICATION_REL_ARRAY",
        10111 => "PG_PUBLICATION_REL",
        10112 => "PG_SUBSCRIPTION_ARRAY",
        10113 => "PG_SUBSCRIPTION_REL_ARRAY",
        10114 => "PG_SUBSCRIPTION_REL",
        12001 => "PG_ROLES_ARRAY",
        12002 => "PG_ROLES",
        12006 => "PG_SHADOW_ARRAY",
        12007 => "PG_SHADOW",
        12011 => "PG_GROUP_ARRAY",
        12012 => "PG_GROUP",
        12015 => "PG_USER_ARRAY",
        12016 => "PG_USER",
        12019 => "PG_POLICIES_ARRAY",
        12020 => "PG_POLICIES",
        12024 => "PG_RULES_ARRAY",
        12025 => "PG_RULES",
        12029 => "PG_VIEWS_ARRAY",
        12030 => "PG_VIEWS",
        12034 => "PG_TABLES_ARRAY",
        12035 => "PG_TABLES",
        12039 => "PG_MATVIEWS_ARRAY",
        12040 => "PG_MATVIEWS",
        12044 => "PG_INDEXES_ARRAY",
        12045 => "PG_INDEXES",
        12049 => "PG_SEQUENCES_ARRAY",
        12050 => "PG_SEQUENCES",
        12054 => "PG_STATS_ARRAY",
        12055 => "PG_STATS",
        12059 => "PG_STATS_EXT_ARRAY",
        12060 => "PG_STATS_EXT",
        12064 => "PG_STATS_EXT_EXPRS_ARRAY",
        12065 => "PG_STATS_EXT_EXPRS",
        12069 => "PG_PUBLICATION_TABLES_ARRAY",
        12070 => "PG_PUBLICATION_TABLES",
        12074 => "PG_LOCKS_ARRAY",
        12075 => "PG_LOCKS",
        12078 => "PG_CURSORS_ARRAY",
        12079 => "PG_CURSORS",
        12082 => "PG_AVAILABLE_EXTENSIONS_ARRAY",
        12083 => "PG_AVAILABLE_EXTENSIONS",
        12086 => "PG_AVAILABLE_EXTENSION_VERSIONS_ARRAY",
        12087 => "PG_AVAILABLE_EXTENSION_VERSIONS",
        12091 => "PG_PREPARED_XACTS_ARRAY",
        12092 => "PG_PREPARED_XACTS",
        12096 => "PG_PREPARED_STATEMENTS_ARRAY",
        12097 => "PG_PREPARED_STATEMENTS",
        12100 => "PG_SECLABELS_ARRAY",
        12101 => "PG_SECLABELS",
        12105 => "PG_SETTINGS_ARRAY",
        12106 => "PG_SETTINGS",
        12111 => "PG_FILE_SETTINGS_ARRAY",
        12112 => "PG_FILE_SETTINGS",
        12115 => "PG_HBA_FILE_RULES_ARRAY",
        12116 => "PG_HBA_FILE_RULES",
        12119 => "PG_IDENT_FILE_MAPPINGS_ARRAY",
        12120 => "PG_IDENT_FILE_MAPPINGS",
        12123 => "PG_TIMEZONE_ABBREVS_ARRAY",
        12124 => "PG_TIMEZONE_ABBREVS",
        12127 => "PG_TIMEZONE_NAMES_ARRAY",
        12128 => "PG_TIMEZONE_NAMES",
        12131 => "PG_CONFIG_ARRAY",
        12132 => "PG_CONFIG",
        12135 => "PG_SHMEM_ALLOCATIONS_ARRAY",
        12136 => "PG_SHMEM_ALLOCATIONS",
        12139 => "PG_BACKEND_MEMORY_CONTEXTS_ARRAY",
        12140 => "PG_BACKEND_MEMORY_CONTEXTS",
        12143 => "PG_STAT_ALL_TABLES_ARRAY",
        12144 => "PG_STAT_ALL_TABLES",
        12148 => "PG_STAT_XACT_ALL_TABLES_ARRAY",
        12149 => "PG_STAT_XACT_ALL_TABLES",
        12153 => "PG_STAT_SYS_TABLES_ARRAY",
        12154 => "PG_STAT_SYS_TABLES",
        12158 => "PG_STAT_XACT_SYS_TABLES_ARRAY",
        12159 => "PG_STAT_XACT_SYS_TABLES",
        12162 => "PG_STAT_USER_TABLES_ARRAY",
        12163 => "PG_STAT_USER_TABLES",
        12167 => "PG_STAT_XACT_USER_TABLES_ARRAY",
        12168 => "PG_STAT_XACT_USER_TABLES",
        12171 => "PG_STATIO_ALL_TABLES_ARRAY",
        12172 => "PG_STATIO_ALL_TABLES",
        12176 => "PG_STATIO_SYS_TABLES_ARRAY",
        12177 => "PG_STATIO_SYS_TABLES",
        12180 => "PG_STATIO_USER_TABLES_ARRAY",
        12181 => "PG_STATIO_USER_TABLES",
        12184 => "PG_STAT_ALL_INDEXES_ARRAY",
        12185 => "PG_STAT_ALL_INDEXES",
        12189 => "PG_STAT_SYS_INDEXES_ARRAY",
        12190 => "PG_STAT_SYS_INDEXES",
        12193 => "PG_STAT_USER_INDEXES_ARRAY",
        12194 => "PG_STAT_USER_INDEXES",
        12197 => "PG_STATIO_ALL_INDEXES_ARRAY",
        12198 => "PG_STATIO_ALL_INDEXES",
        12202 => "PG_STATIO_SYS_INDEXES_ARRAY",
        12203 => "PG_STATIO_SYS_INDEXES",
        12206 => "PG_STATIO_USER_INDEXES_ARRAY",
        12207 => "PG_STATIO_USER_INDEXES",
        12210 => "PG_STATIO_ALL_SEQUENCES_ARRAY",
        12211 => "PG_STATIO_ALL_SEQUENCES",
        12215 => "PG_STATIO_SYS_SEQUENCES_ARRAY",
        12216 => "PG_STATIO_SYS_SEQUENCES",
        12219 => "PG_STATIO_USER_SEQUENCES_ARRAY",
        12220 => "PG_STATIO_USER_SEQUENCES",
        12223 => "PG_STAT_ACTIVITY_ARRAY",
        12224 => "PG_STAT_ACTIVITY",
        12228 => "PG_STAT_REPLICATION_ARRAY",
        12229 => "PG_STAT_REPLICATION",
        12233 => "PG_STAT_SLRU_ARRAY",
        12234 => "PG_STAT_SLRU",
        12237 => "PG_STAT_WAL_RECEIVER_ARRAY",
        12238 => "PG_STAT_WAL_RECEIVER",
        12241 => "PG_STAT_RECOVERY_PREFETCH_ARRAY",
        12242 => "PG_STAT_RECOVERY_PREFETCH",
        12245 => "PG_STAT_SUBSCRIPTION_ARRAY",
        12246 => "PG_STAT_SUBSCRIPTION",
        12250 => "PG_STAT_SSL_ARRAY",
        12251 => "PG_STAT_SSL",
        12254 => "PG_STAT_GSSAPI_ARRAY",
        12255 => "PG_STAT_GSSAPI",
        12258 => "PG_REPLICATION_SLOTS_ARRAY",
        12259 => "PG_REPLICATION_SLOTS",
        12263 => "PG_STAT_REPLICATION_SLOTS_ARRAY",
        12264 => "PG_STAT_REPLICATION_SLOTS",
        12267 => "PG_STAT_DATABASE_ARRAY",
        12268 => "PG_STAT_DATABASE",
        12272 => "PG_STAT_DATABASE_CONFLICTS_ARRAY",
        12273 => "PG_STAT_DATABASE_CONFLICTS",
        12276 => "PG_STAT_USER_FUNCTIONS_ARRAY",
        12277 => "PG_STAT_USER_FUNCTIONS",
        12281 => "PG_STAT_XACT_USER_FUNCTIONS_ARRAY",
        12282 => "PG_STAT_XACT_USER_FUNCTIONS",
        12286 => "PG_STAT_ARCHIVER_ARRAY",
        12287 => "PG_STAT_ARCHIVER",
        12290 => "PG_STAT_BGWRITER_ARRAY",
        12291 => "PG_STAT_BGWRITER",
        12294 => "PG_STAT_WAL_ARRAY",
        12295 => "PG_STAT_WAL",
        12298 => "PG_STAT_PROGRESS_ANALYZE_ARRAY",
        12299 => "PG_STAT_PROGRESS_ANALYZE",
        12303 => "PG_STAT_PROGRESS_VACUUM_ARRAY",
        12304 => "PG_STAT_PROGRESS_VACUUM",
        12308 => "PG_STAT_PROGRESS_CLUSTER_ARRAY",
        12309 => "PG_STAT_PROGRESS_CLUSTER",
        12313 => "PG_STAT_PROGRESS_CREATE_INDEX_ARRAY",
        12314 => "PG_STAT_PROGRESS_CREATE_INDEX",
        12318 => "PG_STAT_PROGRESS_BASEBACKUP_ARRAY",
        12319 => "PG_STAT_PROGRESS_BASEBACKUP",
        12323 => "PG_STAT_PROGRESS_COPY_ARRAY",
        12324 => "PG_STAT_PROGRESS_COPY",
        12328 => "PG_USER_MAPPINGS_ARRAY",
        12329 => "PG_USER_MAPPINGS",
        12333 => "PG_REPLICATION_ORIGIN_STATUS_ARRAY",
        12334 => "PG_REPLICATION_ORIGIN_STATUS",
        12337 => "PG_STAT_SUBSCRIPTION_STATS_ARRAY",
        12338 => "PG_STAT_SUBSCRIPTION_STATS",
        13560 => "CARDINAL_NUMBER_ARRAY",
        13561 => "CARDINAL_NUMBER",
        13563 => "CHARACTER_DATA_ARRAY",
        13564 => "CHARACTER_DATA",
        13565 => "SQL_IDENTIFIER_ARRAY",
        13566 => "SQL_IDENTIFIER",
        13568 => "INFORMATION_SCHEMA_CATALOG_NAME_ARRAY",
        13569 => "INFORMATION_SCHEMA_CATALOG_NAME",
        13571 => "TIME_STAMP_ARRAY",
        13572 => "TIME_STAMP",
        13573 => "YES_OR_NO_ARRAY",
        13574 => "YES_OR_NO",
        13577 => "APPLICABLE_ROLES_ARRAY",
        13578 => "APPLICABLE_ROLES",
        13582 => "ADMINISTRABLE_ROLE_AUTHORIZATIONS_ARRAY",
        13583 => "ADMINISTRABLE_ROLE_AUTHORIZATIONS",
        13586 => "ATTRIBUTES_ARRAY",
        13587 => "ATTRIBUTES",
        13591 => "CHARACTER_SETS_ARRAY",
        13592 => "CHARACTER_SETS",
        13596 => "CHECK_CONSTRAINT_ROUTINE_USAGE_ARRAY",
        13597 => "CHECK_CONSTRAINT_ROUTINE_USAGE",
        13601 => "CHECK_CONSTRAINTS_ARRAY",
        13602 => "CHECK_CONSTRAINTS",
        13606 => "COLLATIONS_ARRAY",
        13607 => "COLLATIONS",
        13611 => "COLLATION_CHARACTER_SET_APPLICABILITY_ARRAY",
        13612 => "COLLATION_CHARACTER_SET_APPLICABILITY",
        13616 => "COLUMN_COLUMN_USAGE_ARRAY",
        13617 => "COLUMN_COLUMN_USAGE",
        13621 => "COLUMN_DOMAIN_USAGE_ARRAY",
        13622 => "COLUMN_DOMAIN_USAGE",
        13626 => "COLUMN_PRIVILEGES_ARRAY",
        13627 => "COLUMN_PRIVILEGES",
        13631 => "COLUMN_UDT_USAGE_ARRAY",
        13632 => "COLUMN_UDT_USAGE",
        13636 => "COLUMNS_ARRAY",
        13637 => "COLUMNS",
        13641 => "CONSTRAINT_COLUMN_USAGE_ARRAY",
        13642 => "CONSTRAINT_COLUMN_USAGE",
        13646 => "CONSTRAINT_TABLE_USAGE_ARRAY",
        13647 => "CONSTRAINT_TABLE_USAGE",
        13651 => "DOMAIN_CONSTRAINTS_ARRAY",
        13652 => "DOMAIN_CONSTRAINTS",
        13656 => "DOMAIN_UDT_USAGE_ARRAY",
        13657 => "DOMAIN_UDT_USAGE",
        13661 => "DOMAINS_ARRAY",
        13662 => "DOMAINS",
        13666 => "ENABLED_ROLES_ARRAY",
        13667 => "ENABLED_ROLES",
        13670 => "KEY_COLUMN_USAGE_ARRAY",
        13671 => "KEY_COLUMN_USAGE",
        13675 => "PARAMETERS_ARRAY",
        13676 => "PARAMETERS",
        13680 => "REFERENTIAL_CONSTRAINTS_ARRAY",
        13681 => "REFERENTIAL_CONSTRAINTS",
        13685 => "ROLE_COLUMN_GRANTS_ARRAY",
        13686 => "ROLE_COLUMN_GRANTS",
        13689 => "ROUTINE_COLUMN_USAGE_ARRAY",
        13690 => "ROUTINE_COLUMN_USAGE",
        13694 => "ROUTINE_PRIVILEGES_ARRAY",
        13695 => "ROUTINE_PRIVILEGES",
        13699 => "ROLE_ROUTINE_GRANTS_ARRAY",
        13700 => "ROLE_ROUTINE_GRANTS",
        13703 => "ROUTINE_ROUTINE_USAGE_ARRAY",
        13704 => "ROUTINE_ROUTINE_USAGE",
        13708 => "ROUTINE_SEQUENCE_USAGE_ARRAY",
        13709 => "ROUTINE_SEQUENCE_USAGE",
        13713 => "ROUTINE_TABLE_USAGE_ARRAY",
        13714 => "ROUTINE_TABLE_USAGE",
        13718 => "ROUTINES_ARRAY",
        13719 => "ROUTINES",
        13723 => "SCHEMATA_ARRAY",
        13724 => "SCHEMATA",
        13727 => "SEQUENCES_ARRAY",
        13728 => "SEQUENCES",
        13732 => "SQL_FEATURES_ARRAY",
        13733 => "SQL_FEATURES",
        13737 => "SQL_IMPLEMENTATION_INFO_ARRAY",
        13738 => "SQL_IMPLEMENTATION_INFO",
        13742 => "SQL_PARTS_ARRAY",
        13743 => "SQL_PARTS",
        13747 => "SQL_SIZING_ARRAY",
        13748 => "SQL_SIZING",
        13752 => "TABLE_CONSTRAINTS_ARRAY",
        13753 => "TABLE_CONSTRAINTS",
        13757 => "TABLE_PRIVILEGES_ARRAY",
        13758 => "TABLE_PRIVILEGES",
        13762 => "ROLE_TABLE_GRANTS_ARRAY",
        13763 => "ROLE_TABLE_GRANTS",
        13766 => "TABLES_ARRAY",
        13767 => "TABLES",
        13771 => "TRANSFORMS_ARRAY",
        13772 => "TRANSFORMS",
        13776 => "TRIGGERED_UPDATE_COLUMNS_ARRAY",
        13777 => "TRIGGERED_UPDATE_COLUMNS",
        13781 => "TRIGGERS_ARRAY",
        13782 => "TRIGGERS",
        13786 => "UDT_PRIVILEGES_ARRAY",
        13787 => "UDT_PRIVILEGES",
        13791 => "ROLE_UDT_GRANTS_ARRAY",
        13792 => "ROLE_UDT_GRANTS",
        13795 => "USAGE_PRIVILEGES_ARRAY",
        13796 => "USAGE_PRIVILEGES",
        13800 => "ROLE_USAGE_GRANTS_ARRAY",
        13801 => "ROLE_USAGE_GRANTS",
        13804 => "USER_DEFINED_TYPES_ARRAY",
        13805 => "USER_DEFINED_TYPES",
        13809 => "VIEW_COLUMN_USAGE_ARRAY",
        13810 => "VIEW_COLUMN_USAGE",
        13814 => "VIEW_ROUTINE_USAGE_ARRAY",
        13815 => "VIEW_ROUTINE_USAGE",
        13819 => "VIEW_TABLE_USAGE_ARRAY",
        13820 => "VIEW_TABLE_USAGE",
        13824 => "VIEWS_ARRAY",
        13825 => "VIEWS",
        13829 => "DATA_TYPE_PRIVILEGES_ARRAY",
        13830 => "DATA_TYPE_PRIVILEGES",
        13834 => "ELEMENT_TYPES_ARRAY",
        13835 => "ELEMENT_TYPES",
        13839 => "_PG_FOREIGN_TABLE_COLUMNS_ARRAY",
        13840 => "PG_FOREIGN_TABLE_COLUMNS_ARRAY",
        13844 => "COLUMN_OPTIONS_ARRAY",
        13845 => "COLUMN_OPTIONS",
        13848 => "_PG_FOREIGN_DATA_WRAPPERS_ARRAY",
        13849 => "PG_FOREIGN_DATA_WRAPPERS_ARRAY",
        13852 => "FOREIGN_DATA_WRAPPER_OPTIONS_ARRAY",
        13853 => "FOREIGN_DATA_WRAPPER_OPTIONS",
        13856 => "FOREIGN_DATA_WRAPPERS_ARRAY",
        13857 => "FOREIGN_DATA_WRAPPERS",
        13860 => "_PG_FOREIGN_SERVERS_ARRAY",
        13861 => "PG_FOREIGN_SERVERS_ARRAY",
        13865 => "FOREIGN_SERVER_OPTIONS_ARRAY",
        13866 => "FOREIGN_SERVER_OPTIONS",
        13869 => "FOREIGN_SERVERS_ARRAY",
        13870 => "FOREIGN_SERVERS",
        13873 => "_PG_FOREIGN_TABLES_ARRAY",
        13874 => "PG_FOREIGN_TABLES_ARRAY",
        13878 => "FOREIGN_TABLE_OPTIONS_ARRAY",
        13879 => "FOREIGN_TABLE_OPTIONS",
        13882 => "FOREIGN_TABLES_ARRAY",
        13883 => "FOREIGN_TABLES",
        13886 => "_PG_USER_MAPPINGS_ARRAY",
        13887 => "PG_USER_MAPPINGS_ARRAY_13887",
        13891 => "USER_MAPPING_OPTIONS_ARRAY",
        13892 => "USER_MAPPING_OPTIONS",
        13896 => "USER_MAPPINGS_ARRAY",
        13897 => "USER_MAPPINGS",
        _ => return None,
    })
}
