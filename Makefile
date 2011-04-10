include $(GOROOT)/src/Make.inc

TARG=golex
DEPS=../dsview

GOFILES=\
	lex.go

include $(GOROOT)/src/Make.cmd