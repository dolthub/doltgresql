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

#include <stdarg.h>
#include <stdio.h>
#include <stdbool.h>

#if defined(_WIN32) || defined(_WIN64)
#define DLLEXPORT __declspec(dllexport)
#else
#define DLLEXPORT __attribute__((visibility("default")))
#endif

static char last_error[512];

DLLEXPORT bool errstart(int elevel, const char* domain) {
	last_error[0] = '\0';
	return 1;
}

DLLEXPORT bool errstart_cold(int elevel, const char *domain) {
	return errstart(elevel, domain);
}

DLLEXPORT int errmsg(const char *fmt, ...) {
	va_list ap;
	va_start(ap, fmt);
	vsnprintf(last_error, sizeof(last_error), fmt, ap);
	va_end(ap);
	return 0;
}

DLLEXPORT int errmsg_internal(const char *fmt, ...) {
	va_list ap;
	va_start(ap, fmt);
	vsnprintf(last_error, sizeof(last_error), fmt, ap);
	va_end(ap);
	return 0;
}

DLLEXPORT int errfinish(int dummy, ...) {
	if (last_error[0]) {
		fprintf(stderr, "Postgres ERROR: %s\n", last_error);
	}
	return 0;
}
