include $(GOROOT)/src/Make.inc

TARG=golex
DEPS=../dsview

GOFILES=\
	lex.go \
	graph.go

include $(GOROOT)/src/Make.cmd
