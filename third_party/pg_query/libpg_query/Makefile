root_dir := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
word-dot = $(word $2,$(subst ., ,$1))

TARGET = pg_query
ARLIB = lib$(TARGET).a
PGDIR = $(root_dir)/tmp/postgres
PGDIRBZ2 = $(root_dir)/tmp/postgres.tar.bz2
PGDIRZIP = $(root_dir)/tmp/postgres.zip

PG_VERSION = 17.7
PG_VERSION_MAJOR = $(call word-dot,$(PG_VERSION),1)
PG_VERSION_NUM = 170007
PROTOC_VERSION = 25.1

VERSION = 6.2.5
VERSION_MAJOR = $(call word-dot,$(VERSION),1)
VERSION_MINOR = $(call word-dot,$(VERSION),2)
VERSION_PATCH = $(call word-dot,$(VERSION),3)

SO_VERSION = $(shell printf '%02d%02d' $(PG_VERSION_MAJOR) $(VERSION_MAJOR)).$(VERSION_MINOR)

ifeq ($(shell uname -s), Darwin)
	SOLIB = lib$(TARGET).dylib
	SONAME = lib$(TARGET).$(SO_VERSION).dylib
	SOLIBVER = lib$(TARGET).$(SO_VERSION).$(VERSION_PATCH).dylib
	SOFLAG = -install_name
else
	SOLIB = lib$(TARGET).so
	SONAME = $(SOLIB).$(SO_VERSION)
	SOLIBVER = $(SONAME).$(VERSION_PATCH)
	SOFLAG = -soname
endif

SRC_FILES := $(wildcard src/*.c src/postgres/*.c) vendor/protobuf-c/protobuf-c.c vendor/xxhash/xxhash.c protobuf/pg_query.pb-c.c
OBJ_FILES := $(SRC_FILES:.c=.o)

override CFLAGS += -g -I. -I./vendor -I./src/include -I./src/postgres/include -Wall -Wno-unused-function -Wno-unused-value -Wno-unused-variable -fno-strict-aliasing -fwrapv -fPIC

ifeq ($(OS),Windows_NT)
override CFLAGS += -I./src/postgres/include/port/win32
override TEST_CFLAGS += -I./src/postgres/include/port/win32
endif

override PG_CONFIGURE_FLAGS += -q --without-readline --without-zlib --without-icu

override TEST_CFLAGS += -g -I. -I./vendor -Wall
override TEST_LDFLAGS += -pthread

CFLAGS_OPT_LEVEL = -O3
ifeq ($(DEBUG),1)
	CFLAGS_OPT_LEVEL = -O0
endif
ifeq ($(VALGRIND),1)
	CFLAGS_OPT_LEVEL = -O0
endif
override CFLAGS += $(CFLAGS_OPT_LEVEL)

ifeq ($(DEBUG),1)
	# We always add -g, so this only has to enable assertion checking
	override CFLAGS += -D USE_ASSERT_CHECKING
endif
ifeq ($(VALGRIND),1)
	override CFLAGS += -DUSE_VALGRIND
	override TEST_CFLAGS += -DUSE_VALGRIND
endif

CLEANLIBS = $(ARLIB)
CLEANOBJS = $(OBJ_FILES)
CLEANFILES = $(PGDIRBZ2) $(PGDIRZIP)

AR ?= ar
ARFLAGS ?= rs
INSTALL = install
LN_S = ln -s
RM = rm -f
ECHO = echo

VALGRIND_MEMCHECK = valgrind --leak-check=full --gen-suppressions=all \
  --suppressions=test/valgrind.supp --time-stamp=yes \
  --error-markers=VALGRINDERROR-BEGIN,VALGRINDERROR-END \
  --log-file=test/valgrind.log --trace-children=yes --show-leak-kinds=all \
  --error-exitcode=1 --errors-for-leak-kinds=all --num-callers=50

CC ?= cc

# Experimental use of Protobuf C++ library, primarily used to validate JSON output matches Protobuf JSON mapping
CXX_SRC_FILES := src/pg_query_outfuncs_protobuf_cpp.cc protobuf/pg_query.pb.cc
ifeq ($(USE_PROTOBUF_CPP),1)
	override CXXFLAGS += `pkg-config --cflags protobuf` -I. -I./src/include -I./src/postgres/include -DHAVE_PTHREAD -std=c++17 -Wall -Wno-unused-function -Wno-zero-length-array -Wno-c99-extensions -fwrapv -fPIC
	ifeq ($(DEBUG),1)
		override CXXFLAGS += -O0 -g
	else
		override CXXFLAGS += -O3 -g
	endif
	override TEST_LDFLAGS += `pkg-config --libs protobuf` -lstdc++ -lm

	# Don't use regular Protobuf-C or JSON implementation (instead implement the same methods using the C++ library)
	SRC_FILES := $(filter-out src/pg_query_outfuncs_json.c src/pg_query_outfuncs_protobuf.c, $(SRC_FILES))
	OBJ_FILES := $(SRC_FILES:.c=.o) $(CXX_SRC_FILES:.cc=.o)
else
	# Make sure we always clean C++ object files
	CLEANOBJS += $(CXX_SRC_FILES:.cc=.o)
endif

all: examples test build

build: $(ARLIB)

build_shared: $(SOLIB)

clean:
	-@ $(RM) $(CLEANLIBS) $(CLEANOBJS) $(CLEANFILES) $(EXAMPLES) $(TESTS)
	-@ $(RM) -rf {test,examples}/*.dSYM
	-@ $(RM) -r $(PGDIR) $(PGDIRBZ2) $(PGDIRZIP)

.PHONY: all clean build build_shared extract_source examples test install

$(PGDIR):
	curl -o $(PGDIRBZ2) https://ftp.postgresql.org/pub/source/v$(PG_VERSION)/postgresql-$(PG_VERSION).tar.bz2
	tar -xjf $(PGDIRBZ2)
	mv $(root_dir)/postgresql-$(PG_VERSION) $(PGDIR)
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/01_parser_additional_param_ref_support.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/03_lexer_track_yyllocend.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/04_lexer_comments_as_tokens.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/05_limit_option_enum_value_default.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/06_alloc_set_delete_free_list.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/07_plpgsql_start_finish_datums.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/08_avoid_zero_length_delimiter_in_regression_tests.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/09_allow_param_junk.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/10_avoid_namespace_hashtab_impl_gen.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/11_ifndef_namedatalen.patch
	cd $(PGDIR); patch -p1 < $(root_dir)/patches/13_node_string_constant_locations.patch
	cd $(PGDIR); ./configure $(PG_CONFIGURE_FLAGS)
	cd $(PGDIR); make -C src/pl/plpgsql/src pl_gram.h plerrcodes.h pl_reserved_kwlist_d.h pl_unreserved_kwlist_d.h
	cd $(PGDIR); make -C src/port pg_config_paths.h
	cd $(PGDIR); make -C src/backend generated-headers
	cd $(PGDIR); make -C src/backend parser-recursive # Triggers copying of includes to where they belong, as well as generating gram.c/scan.c
	cd $(PGDIR); make -C src/common kwlist_d.h
	# Avoid problems with static asserts
	echo "#undef StaticAssertDecl" >> $(PGDIR)/src/include/c.h
	echo "#define StaticAssertDecl(condition, errmessage)" >> $(PGDIR)/src/include/c.h
	# Add pg_config.h overrides
	cat scripts/pg_config_overrides.h >> $(PGDIR)/src/include/pg_config.h
	# Only define strlcpy when needed
	sed -i "" '$(shell echo 's/\#include "c.h"/#include "c.h"\n#if HAVE_DECL_STRLCPY == 0/')' $(PGDIR)/src/port/strlcpy.c
	echo "#endif // HAVE_DECL_STRLCPY == 0" >> $(PGDIR)/src/port/strlcpy.c
	# Define symbols needed by elog.c that are commonly defined by win32/signal.c
	echo "#ifdef WIN32" >> $(PGDIR)/src/backend/utils/error/elog.c
	echo "volatile int pg_signal_queue;" >> $(PGDIR)/src/backend/utils/error/elog.c
	echo "int pg_signal_mask;" >> $(PGDIR)/src/backend/utils/error/elog.c
	echo "void pgwin32_dispatch_queued_signals(void) {}" >> $(PGDIR)/src/backend/utils/error/elog.c
	echo "#endif" >> $(PGDIR)/src/backend/utils/error/elog.c

extract_source: $(PGDIR)
	-@ $(RM) -rf ./src/postgres/
	mkdir ./src/postgres
	mkdir ./src/postgres/include
	LIBCLANG=/Library/Developer/CommandLineTools/usr/lib/libclang.dylib ruby ./scripts/extract_source.rb $(PGDIR)/ ./src/postgres/
	# Override OS-specific pg_config_os.h to only load Win32 logic (the primary port logic that matters for libpg_query), if needed
	echo "#if defined(_WIN32) || defined(_WIN64)" > ./src/postgres/include/pg_config_os.h
	echo "#include \"port/win32.h\"" >> ./src/postgres/include/pg_config_os.h
	# Don't mark anything as visible based on how Postgres defines it
	echo "#undef PGDLLIMPORT" >> ./src/postgres/include/pg_config_os.h
	echo "#undef PGDLLEXPORT" >> ./src/postgres/include/pg_config_os.h
	# Avoid getting incorrect sigsetjmp overrides
	echo "#ifdef __clang__" >> ./src/postgres/include/pg_config_os.h
	echo "#undef __MINGW64__" >> ./src/postgres/include/pg_config_os.h
	echo "#endif /* __clang__ */" >> ./src/postgres/include/pg_config_os.h
	echo "#endif" >> ./src/postgres/include/pg_config_os.h
	# Adjust version string to ignore differences in build environments
	sed -i "" '$(shell echo 's/\#define PG_VERSION_STR .*/#define PG_VERSION_STR "PostgreSQL $(PG_VERSION) \(libpg_query\)"/')' ./src/postgres/include/pg_config.h
	# Copy version information so its easily accessible
	sed -i "" '$(shell echo 's/\#define PG_MAJORVERSION .*/#define PG_MAJORVERSION "$(PG_VERSION_MAJOR)"/')' pg_query.h
	sed -i "" '$(shell echo 's/\#define PG_VERSION .*/#define PG_VERSION "$(PG_VERSION)"/')' pg_query.h
	sed -i "" '$(shell echo 's/\#define PG_VERSION_NUM .*/#define PG_VERSION_NUM $(PG_VERSION_NUM)/')' pg_query.h
	# Copy regress SQL files so we can use them in tests
	rm -f ./test/sql/postgres_regress/*.sql
	rm -f ./test/sql/plpgsql_regress/*.sql
	cp $(PGDIR)/src/test/regress/sql/*.sql ./test/sql/postgres_regress/
	cp $(PGDIR)/src/pl/plpgsql/src/sql/*.sql ./test/sql/plpgsql_regress/
	cp $(PGDIR)/COPYRIGHT ./src/postgres/

.c.o:
	@$(ECHO) compiling $(<)
	@$(CC) $(CPPFLAGS) $(CFLAGS) -o $@ -c $<

.cc.o:
	@$(ECHO) compiling $(<)
	@$(CXX) $(CXXFLAGS) -o $@ -c $<

$(ARLIB): $(OBJ_FILES) Makefile
	@$(AR) $(ARFLAGS) $@ $(OBJ_FILES)

$(SOLIB): $(OBJ_FILES) Makefile
	@$(CC) $(CFLAGS) -shared -Wl,$(SOFLAG),$(SONAME) $(LDFLAGS) -o $@ $(OBJ_FILES) $(LIBS)

protobuf/pg_query.pb-c.c protobuf/pg_query.pb-c.h: protobuf/pg_query.proto
ifneq ($(shell which protoc-gen-c), )
	protoc --c_out=. protobuf/pg_query.proto
else
	@echo 'Warning: protoc-gen-c not found, skipping protocol buffer regeneration'
endif

src/pg_query_protobuf.c src/pg_query_scan.c: protobuf/pg_query.pb-c.h

# Only used when USE_PROTOBUF_CPP is used (experimental for testing only)
src/pg_query_outfuncs_protobuf_cpp.cc: protobuf/pg_query.pb.cc
protobuf/pg_query.pb.cc: protobuf/pg_query.proto
ifneq ($(shell protoc --version 2>/dev/null | cut -f2 -d" "), $(PROTOC_VERSION))
	$(error "ERROR - Wrong protobuf compiler version, need $(PROTOC_VERSION)")
endif
	protoc --cpp_out=. protobuf/pg_query.proto

EXAMPLES = examples/simple examples/scan examples/normalize examples/simple_error examples/normalize_error examples/simple_plpgsql
examples: $(EXAMPLES)
	examples/simple
	examples/scan
	examples/normalize
	examples/simple_error
	examples/normalize_error
	examples/simple_plpgsql

examples/simple: examples/simple.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/simple.c $(ARLIB) $(TEST_LDFLAGS)

examples/scan: examples/scan.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/scan.c $(ARLIB) $(TEST_LDFLAGS)

examples/normalize: examples/normalize.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/normalize.c $(ARLIB) $(TEST_LDFLAGS)

examples/simple_error: examples/simple_error.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/simple_error.c $(ARLIB) $(TEST_LDFLAGS)

examples/normalize_error: examples/normalize_error.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/normalize_error.c $(ARLIB) $(TEST_LDFLAGS)

examples/simple_plpgsql: examples/simple_plpgsql.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ -g examples/simple_plpgsql.c $(ARLIB) $(TEST_LDFLAGS)

TESTS = test/complex test/concurrency test/deparse test/fingerprint test/fingerprint_opts test/is_utility_stmt test/normalize test/normalize_utility test/parse test/parse_opts test/parse_protobuf test/parse_protobuf_opts test/parse_plpgsql test/scan test/split test/summary test/summary_truncate
test: $(TESTS)
ifeq ($(VALGRIND),1)
	$(VALGRIND_MEMCHECK) test/complex || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/concurrency || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/deparse || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/fingerprint || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/fingerprint_opts || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/is_utility_stmt || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/normalize || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/normalize_utility || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/parse || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/parse_opts || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/parse_protobuf || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/parse_protobuf_opts || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/scan || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/split || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/summary || (cat test/valgrind.log && false)
	$(VALGRIND_MEMCHECK) test/summary_truncate || (cat test/valgrind.log && false)
	# Output-based tests
	$(VALGRIND_MEMCHECK) test/parse_plpgsql || (cat test/valgrind.log && false)
	diff -Naur test/plpgsql_samples.expected.json test/plpgsql_samples.actual.json
else
	test/complex
	test/concurrency
	test/deparse
	test/fingerprint
	test/fingerprint_opts
	test/is_utility_stmt
	test/normalize
	test/normalize_utility
	test/parse
	test/parse_opts
	test/parse_protobuf
	test/parse_protobuf_opts
	test/scan
	test/split
	test/summary
	test/summary_truncate
	# Output-based tests
	test/parse_plpgsql
	diff -Naur test/plpgsql_samples.expected.json test/plpgsql_samples.actual.json
endif

test/complex: test/complex.c $(ARLIB)
	# We have "-Isrc/" because this test uses pg_query_fingerprint_with_opts
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/ test/complex.c $(ARLIB) $(TEST_LDFLAGS)

test/concurrency: test/concurrency.c test/parse_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/concurrency.c $(ARLIB) $(TEST_LDFLAGS)

test/deparse: test/deparse.c test/deparse_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/deparse.c $(ARLIB) $(TEST_LDFLAGS)

test/fingerprint: test/fingerprint.c test/fingerprint_tests.c $(ARLIB)
	# We have "-Isrc/" because this test uses pg_query_fingerprint_with_opts
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/ test/fingerprint.c $(ARLIB) $(TEST_LDFLAGS)

test/fingerprint_opts: test/fingerprint_opts.c test/fingerprint_opts_tests.c $(ARLIB)
	# We have "-Isrc/" because this test uses pg_query_fingerprint_with_opts
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/ test/fingerprint_opts.c $(ARLIB) $(TEST_LDFLAGS)

test/is_utility_stmt: test/framework/main.c test/is_utility_stmt.c $(ARLIB)
	# We have "-Isrc/postgres/include" because this test uses pg_query_summary_direct
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/postgres/include test/framework/main.c test/is_utility_stmt.c $(ARLIB) $(TEST_LDFLAGS)

test/normalize: test/normalize.c test/normalize_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/normalize.c $(ARLIB) $(TEST_LDFLAGS)

test/normalize_utility: test/normalize_utility.c test/normalize_utility_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/normalize_utility.c $(ARLIB) $(TEST_LDFLAGS)

test/parse: test/parse.c test/parse_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/parse.c $(ARLIB) $(TEST_LDFLAGS)

test/summary: test/framework/main.c test/summary.c test/summary_tests.c test/summary_tests_list.c $(ARLIB)
	# We have "-Isrc/postgres/include" because this test uses pg_query_summary_direct
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/postgres/include test/framework/main.c test/summary.c $(ARLIB) $(TEST_LDFLAGS)

test/summary_truncate: test/framework/main.c test/summary_truncate.c $(ARLIB)
	# We have "-Isrc/postgres/include" because this test uses pg_query_summary_direct
	$(CC) $(TEST_CFLAGS) -o $@ -Isrc/postgres/include test/framework/main.c test/summary_truncate.c $(ARLIB) $(TEST_LDFLAGS)

test/parse_opts: test/parse_opts.c test/parse_opts_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/parse_opts.c $(ARLIB) $(TEST_LDFLAGS)

test/parse_plpgsql: test/parse_plpgsql.c test/parse_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/parse_plpgsql.c $(ARLIB) $(TEST_LDFLAGS)

test/parse_protobuf: test/parse_protobuf.c test/parse_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/parse_protobuf.c $(ARLIB) $(TEST_LDFLAGS)

test/parse_protobuf_opts: test/parse_protobuf_opts.c test/parse_opts_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/parse_protobuf_opts.c $(ARLIB) $(TEST_LDFLAGS)

test/scan: test/scan.c test/scan_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/scan.c $(ARLIB) $(TEST_LDFLAGS)

test/split: test/split.c test/split_tests.c $(ARLIB)
	$(CC) $(TEST_CFLAGS) -o $@ test/split.c $(ARLIB) $(TEST_LDFLAGS)

prefix = /usr/local
libdir = $(prefix)/lib
includedir = $(prefix)/include

install: $(ARLIB) $(SOLIB)
	$(INSTALL) -d "$(DESTDIR)"$(libdir)
	$(INSTALL) -m 644 $(ARLIB) "$(DESTDIR)"$(libdir)/$(ARLIB)
	$(INSTALL) -m 755 $(SOLIB) "$(DESTDIR)"$(libdir)/$(SOLIBVER)
	$(LN_S) $(SOLIBVER) "$(DESTDIR)"$(libdir)/$(SONAME)
	$(LN_S) $(SOLIBVER) "$(DESTDIR)"$(libdir)/$(SOLIB)
	$(INSTALL) -d "$(DESTDIR)"$(includedir)/$(TARGET)
	$(INSTALL) -m 644 pg_query.h "$(DESTDIR)"$(includedir)/pg_query.h
	$(INSTALL) -m 644 postgres_deparse.h "$(DESTDIR)"$(includedir)/postgres_deparse.h
	$(INSTALL) -m 644 protobuf/pg_query.proto "$(DESTDIR)"$(includedir)/$(TARGET)/pg_query.proto
