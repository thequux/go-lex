include $(GOROOT)/src/Make.inc

TARG=golex

GOFILES=\
	lex.go \
	graph.go

include $(GOROOT)/src/Make.cmd
