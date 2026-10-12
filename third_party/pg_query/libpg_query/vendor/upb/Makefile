# Maintenance Makefile for the vendored upb runtime (see README).
#
#   make -C vendor/upb update TAG=v36.2
#
# Re-vendors upb from a protobuf git tag and AMALGAMATES the runtime subset into
# a single upb.c + upb/generated_code_support.h (upb's amalgamate.py, shipped in
# the upb tree). It also keeps port/def.inc + undef.inc and third_party/utf8_range,
# applies the local patches in patches/*.patch to upb.c, records the tag in VERSION, and
# regenerates protobuf/pg_query.upb*.{c,h}. The vendored runtime and the generated
# code MUST come from the same version (the generated minitables are ABI-locked to
# the runtime), so this refuses to run unless protoc matches TAG. Requires git +
# python3 + protoc + protoc-gen-upb + protoc-gen-upb_minitable on PATH (on macOS:
# `brew install protobuf`).

SHELL := /bin/bash
TAG ?=
REPO := https://github.com/protocolbuffers/protobuf.git
ROOT := $(abspath $(CURDIR)/../..)
TMP  := $(CURDIR)/.upb-update-tmp
# Runtime .c subset to amalgamate -- keep in sync with the list in README.
# Reflection/JSON/text/lex/io/util and the fast-path parser are omitted.
SUBSET := upb/base upb/hash upb/mem upb/message upb/mini_descriptor upb/mini_table upb/port upb/wire

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo 'Update the vendored upb runtime + regenerate the message code:'
	@echo '    make -C vendor/upb update TAG=v36.2'
	@echo 'git, python3, protoc and protoc-gen-upb must be on PATH and match TAG.'

.PHONY: update
update:
	@test -n "$(TAG)" || { echo "usage: make update TAG=v36.2 (a protobuf git tag)" >&2; exit 2; }
	@for t in git python3 protoc protoc-gen-upb protoc-gen-upb_minitable; do \
		command -v $$t >/dev/null || { echo "ERROR: '$$t' not found on PATH" >&2; exit 1; }; \
	done
	@want=$$(echo "$(TAG)" | sed 's/^v//'); pver=$$(protoc --version | awk '{print $$2}'); \
		if [ "$$pver" != "$$want" ] && [ "$(ALLOW_VERSION_MISMATCH)" != "1" ]; then \
			echo "ERROR: protoc is $$pver but TAG=$(TAG); install protobuf $$want (the generated" >&2; \
			echo "       code must match the vendored runtime), or set ALLOW_VERSION_MISMATCH=1." >&2; \
			exit 1; \
		fi
	@echo "==> Fetching protobuf $(TAG)"
	@rm -rf "$(TMP)"
	@git clone --quiet --filter=blob:none --no-checkout --depth 1 --branch "$(TAG)" "$(REPO)" "$(TMP)"
	@git -C "$(TMP)" sparse-checkout init --cone >/dev/null
	@git -C "$(TMP)" sparse-checkout set upb third_party/utf8_range >/dev/null
	@git -C "$(TMP)" checkout --quiet
	@echo "==> Amalgamating runtime into upb.c + upb/generated_code_support.h"
	@rm -rf upb third_party
	@mkdir -p upb/port third_party/utf8_range
	@cp "$(TMP)"/upb/generated_code_support.h "$(TMP)"/gcs_root.c
	@cd "$(TMP)" && amalg=$$(find upb -name amalgamate.py | head -1); \
		test -n "$$amalg" || { echo "ERROR: amalgamate.py not found in the upb tree" >&2; exit 1; }; \
		python3 "$$amalg" \
			"$(CURDIR)/upb.c" "$(CURDIR)/upb/generated_code_support.h" \
			gcs_root.c \
			$$(find $(SUBSET) -name '*.c' ! -name '*test*' ! -path '*/decode_fast/*') \
			$$(find upb \( -name '*.h' -o -name '*.inc' \) ! -name '*test*' ! -path '*/cmake/*' ! -path '*/conformance/*')
	@# amalgamate emits `#include "generated_code_support.h"` (basename); make it
	@# resolve via -Ivendor/upb like the generated code does.
	@sed 's|#include "generated_code_support.h"|#include "upb/generated_code_support.h"|' upb.c > upb.c.tmp && mv upb.c.tmp upb.c
	@cp "$(TMP)"/upb/port/def.inc "$(TMP)"/upb/port/undef.inc upb/port/
	@cp "$(TMP)"/third_party/utf8_range/utf8_range.c "$(TMP)"/third_party/utf8_range/utf8_range.h third_party/utf8_range/
	@cp "$(TMP)"/third_party/utf8_range/*.inc third_party/utf8_range/ 2>/dev/null || true
	@cp "$(TMP)"/third_party/utf8_range/utf8_validity.h third_party/utf8_range/ 2>/dev/null || true
	@cp "$(TMP)"/LICENSE LICENSE
	@cp "$(TMP)"/third_party/utf8_range/LICENSE third_party/utf8_range/LICENSE 2>/dev/null || true
	@echo "==> Applying local patches to upb.c"
	@shopt -s nullglob; for p in patches/*.patch; do \
		echo "    $$p"; \
		patch -p1 -N -s -F 0 < "$$p" || { \
			rm -f upb.c.rej upb.c.orig; \
			echo "ERROR: $$p no longer applies to upb.c; update the patch (see README)." >&2; \
			exit 1; \
		}; \
	done
	@echo "$(TAG)" > VERSION
	@rm -rf "$(TMP)"
	@echo "==> Regenerating protobuf/pg_query.upb*.{c,h} with protoc $$(protoc --version | awk '{print $$2}')"
	@cd "$(ROOT)" && protoc --upb_out=. --upb_minitable_out=. -I. protobuf/pg_query.proto
	@rm -f "$(ROOT)/protobuf/pg_query.upb.c" # always empty (accessors are inline in the header)
	@echo "Done ($(TAG)). Now run 'make && make test' at the repo root and review the diff."
