package main

import (
	"os"
	_ "strings"
	"bytes"
	"log"
)

type (
	NodeId int


	NfaNode map[byte][]NodeId

	Graph interface {
		NewNode() NodeId
		AddTransition(label byte, from, to NodeId)
	}

	Nfa struct {
		Nodes     map[NodeId]NfaNode
		Start     NodeId
		Accepting map[NodeId]string
		MaxNode   NodeId
	}

	regex interface {
		//addToGraph(g Graph, from, to NodeId)
		StringPrec(prec int) string
		//WriteTo(w io.Writer) (n int, err os.Error)
	}

	alternateRe []regex
	starRe      struct {
		regex
	}
	anyRe      struct{}
	sequenceRe []regex
	literal    byte
)

// Graph stuffs

func (re sequenceRe) StringPrec(prec int) string {
	const Prec = 10
	str := ""

	for _, sub := range re {
		str += sub.StringPrec(Prec)
	}
	if prec > Prec {
		str = "(" + str + ")"
	}
	return str
}

func (re alternateRe) StringPrec(prec int) string {
	const Prec = 9
	str := ""
	for i, sub := range re {
		if i != 0 {
			str += "|"
		}
		str += sub.StringPrec(Prec)
	}

	if prec > Prec {
		str = "(" + str + ")"
	}
	return str
}

func (re starRe) StringPrec(prec int) string {
	const Prec = 11
	str := re.regex.StringPrec(Prec)
	if prec > Prec {
		return "(" + str + ")*"
	}
	return str + "*"
}

func (anyRe) StringPrec(int) string {
	return "."
}

func (re literal) StringPrec(int) string {
	return string(re)
}


func (g *Nfa) NewNode() NodeId {
	g.MaxNode++
	g.Nodes[g.MaxNode] = make(NfaNode)
	return g.MaxNode
}

func (g *Nfa) AddTransition(label byte, from, to NodeId) {
	node, ok := g.Nodes[from]
	if !ok {
		node = make(NfaNode)
		g.Nodes[from] = node
	}

	if _, ok := node[label]; !ok {
		node[label] = make([]NodeId, 0, 1)
	}
	node[label] = append(node[label], to)
}


func ParseRegex(source string) regex {
	reader := bytes.NewBufferString(source + ")")
	var parse1 func(r *bytes.Buffer) regex
	readChar := func(r *bytes.Buffer, preEscaped bool) (ch byte, escaped bool, err os.Error) {
		// TODO(thequux): Hook unicode handling in here
		escaped = preEscaped

		ch, err = r.ReadByte()
		if err != nil {
			return
		}
		if !escaped && ch == '\\' {
			escaped = true
			ch, err = r.ReadByte()
			if err != nil {
				return
			}
		}
		return
	}

	parse1 = func(r *bytes.Buffer) regex {
		top := make(alternateRe, 0, 1)
		current := make(sequenceRe, 0, 1)

		for {
			b, err := r.ReadByte()
			if err != nil {
				if err == os.EOF {
					panic("Unclosed '('")
				}
				panic(err)
			}
			switch b {
			case '(':
				current = append(current, parse1(r))
			case ')':
				goto end
			case ']':
				panic("unexpected ']'")
			case '|':
				top = append(top, current)
				current = make(sequenceRe, 0, 1)

			case '[':
				buf := make(alternateRe, 0, 1)
				for {
					ch, escaped, ok := readChar(r, false)
					// BUG(thequux): Does not handle multibyte
					if ok != nil {
						panic("Unexpected EOF")
					} else if ch == ']' && !escaped {
						current = append(current, buf)
						log.Printf("Read %#v", buf)
						break
					} else {
						// TODO(thequux): handle '-', '^'
						buf = append(buf, literal(ch))
					}
				}
			case '\\':
				ch, _, ok := readChar(r, true)
				if ok != nil {
					panic("Unexpected EOF")
				}
				// BUG(thequux): Does not handle multibyte
				current = append(current, literal(ch))
			case '*':
				current[len(current)-1] = starRe{current[len(current)-1]}
			case '+':
				current = append(current, starRe{current[len(current)-1]})
			case '?':
				current[len(current)-1] = alternateRe{
					sequenceRe{},
					current[len(current)-1]}

			default:
				current = append(current, literal(b))

			}
		}
	end:
		top = append(top, current)
		return top
	}

	return parse1(reader)
}


//func addToGraph

func MakeNFA(regex string) (g Graph, err os.Error) {
	return
}
