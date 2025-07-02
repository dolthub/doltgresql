// Copyright 2025 Dolthub, Inc.
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

#ifndef PG_EXT_EXPORTS_H
#define PG_EXT_EXPORTS_H

#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>

// This doesn't compile unless it has a value, but Postgres defines this as an empty value intentionally
#define FLEXIBLE_ARRAY_MEMBER 8

typedef uintptr_t Datum;
typedef struct FunctionCallInfoBaseData* FunctionCallInfo;
typedef Datum (*PGFunction) (FunctionCallInfo fcinfo);

typedef struct NullableDatum {
	Datum value;
	bool  isnull;
} NullableDatum;

typedef struct FmgrInfo {
	void*         fn_addr;
	uint32_t      fn_oid;
	short         fn_nargs;
	bool          fn_strict;
	bool          fn_retset;
	unsigned char fn_stats;
	void*         fn_extra;
	void*         fn_mcxt;
	void*         fn_expr;
} FmgrInfo;

typedef struct FunctionCallInfoBaseData {
	FmgrInfo*     flinfo;
	void*         context;
	void*         resultinfo;
	uint32_t      fncollation;
	bool          isnull;
	short         nargs;
	NullableDatum args[FLEXIBLE_ARRAY_MEMBER];
} FunctionCallInfoBaseData;

enum {
	SZ_FMGRINFO = sizeof(FmgrInfo),
	SZ_FCINFO   = sizeof(FunctionCallInfoBaseData)
};

typedef const char pgext_const_char;
typedef unsigned char pgext_unsigned_char;
typedef const uint8_t pgext_const_uint8;

#endif //PG_EXT_EXPORTS_H
