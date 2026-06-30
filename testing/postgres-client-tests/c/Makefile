LIBPQ_CFLAGS  := $(shell pkg-config --cflags libpq)
LIBPQ_LDFLAGS := $(shell pkg-config --libs libpq)
APR_CFLAGS    := $(shell pkg-config --cflags apr-1 apr-util-1)
APR_LDFLAGS   := $(shell pkg-config --libs apr-1 apr-util-1)
CFLAGS="$(apr-1-config --includes) $(apu-1-config --includes)"

all: postgres-c-connector-test libaprutil-test libdbi-test

postgres-c-connector-test: postgres-c-connector-test.c
	$(CC) $(LIBPQ_CFLAGS) -o $@ $^ $(LIBPQ_LDFLAGS)

libaprutil-test: libaprutil-test.c
	$(CC) $(APR_CFLAGS) -o $@ $^ $(APR_LDFLAGS)

libdbi-test: libdbi-test.c
	$(CC) -o $@ $^ -ldbi

.PHONY: clean
clean:
	rm -f postgres-c-connector-test libaprutil-test libdbi-test
