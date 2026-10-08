LDFLAGS := -lodbc

all: psqlodbc-test

psqlodbc-test: psqlodbc-test.c
	$(CC) -o $@ $^ $(LDFLAGS)

.PHONY: clean
clean:
	rm -f psqlodbc-test
