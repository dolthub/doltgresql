CFLAGS := $(shell pkg-config --cflags libpq)
LDFLAGS := $(shell pkg-config --libs libpq)

all: postgres-c-connector-test

postgres-c-connector-test: postgres-c-connector-test.c
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

.PHONY: clean
clean:
	rm -f postgres-c-connector-test
