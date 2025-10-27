XMLLINT := $(shell command -v xmllint 2> /dev/null)

.PHONY: all clean

all: glade/smppizdez_min.glade
	go build
glade/smppizdez_min.glade: glade/smppizdez.glade
ifndef XMLLINT
	cp glade/smppizdez.glade glade/smppizdez_min.glade
else
	$(XMLLINT) --noblanks glade/smppizdez.glade > glade/smppizdez_min.glade
endif
clean:
	rm glade/smppizdez_min.glade
	go clean
