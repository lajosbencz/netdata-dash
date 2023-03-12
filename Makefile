.PHONY: all clean dash

LDFLAGS := "-w -s"
SRCDIR := ./cmd
DISTDIR := ./dist
SOURCES := $(wildcard $(SRCDIR)/*)
OBJECTS := $(SOURCES:$(SRCDIR)/%=$(DISTDIR)/%)

all: $(OBJECTS)

$(DISTDIR)/%: $(SRCDIR)/%
	CGO_ENABLED=0 GODEBUG=http2client=0 go build -ldflags $(LDFLAGS) -o $@ "./$<"

clean:
	rm -f $(DISTDIR)/*
