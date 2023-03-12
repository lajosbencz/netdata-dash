.PHONY: all clean dash

# LDFLAGS := "-w -s"
LDFLAGS := ""
SRCDIR := ./cmd
DISTDIR := ./dist
SOURCES := $(wildcard $(SRCDIR)/*)
OBJECTS := $(SOURCES:$(SRCDIR)/%=$(DISTDIR)/%)

all: $(OBJECTS)

$(DISTDIR)/%: $(SRCDIR)/%
	go build -ldflags $(LDFLAGS) -o $@ "./$<"

clean:
	rm -f $(DISTDIR)/*
