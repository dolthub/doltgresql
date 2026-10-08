CXXFLAGS := -std=c++17 $(shell pkg-config --cflags libpqxx)
LDFLAGS := $(shell pkg-config --libs libpqxx)

all: libpqxx-test

libpqxx-test: libpqxx-test.cpp
	$(CXX) $(CXXFLAGS) -o $@ $^ $(LDFLAGS)

.PHONY: clean
clean:
	rm -f libpqxx-test
