package main

import (
	"fmt"
	"io"
	"os"
	_ "strings"
	"bytes"
	"log"
	"tabwriter"
	"strconv"
)

type (
	NodeId int


	NfaNode struct {
		Nodes map[byte][]NodeId
		Epsilons []NodeId
	}

	Graph interface {
		NewNode() NodeId
		AddTransition(label byte, from, to NodeId)
	}
	Nfa interface {
		Graph
		AddEpsilon(from, to NodeId)
	}
	
	nfa struct {
		Nodes     map[NodeId]*NfaNode
		Start     NodeId
		Accepting map[NodeId]string
		MaxNode   NodeId
	}

	regex interface {
		AddToGraph(g Nfa, from, to NodeId)
		StringPrec(prec int) string
		//WriteTo(w io.Writer) (n int, err os.Error)
	}

	alternateRe []regex
	plusRe      struct {
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

func (re plusRe) StringPrec(prec int) string {
	const Prec = 11
	str := re.regex.StringPrec(Prec)
	if prec > Prec {
		return "(" + str + ")+"
	}
	return str + "+"
}

func (anyRe) StringPrec(int) string {
	return "."
}

func (re literal) StringPrec(int) string {
	return string(re)
}

func MakeNfaNode() *NfaNode {
	return &NfaNode{
	Nodes: make(map[byte][]NodeId, 256),
	Epsilons: make([]NodeId, 0, 1)}
}

func (g *nfa) NewNode() NodeId {
	g.MaxNode++
	g.Nodes[g.MaxNode] = MakeNfaNode()
	return g.MaxNode
}

func (g *nfa) AddTransition(label byte, from, to NodeId) {
	node, ok := g.Nodes[from]
	if !ok {
		node = MakeNfaNode()
		g.Nodes[from] = node
	}

	if _, ok := node.Nodes[label]; !ok {
		node.Nodes[label] = make([]NodeId, 0, 1)
	}
	node.Nodes[label] = append(node.Nodes[label], to)
}

func (g *nfa) AddEpsilon(from, to NodeId) {
	node := g.Nodes[from]
	node.Epsilons = append (node.Epsilons, to)
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
				current[len(current)-1] = alternateRe{
					plusRe{current[len(current)-1]},
					sequenceRe{}}
			case '+':
				current[len(current)-1] = plusRe{current[len(current)-1]}
			case '?':
				current[len(current)-1] = alternateRe{
					sequenceRe{},
					current[len(current)-1]}
			case '.':
				current = append(current, anyRe{})
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

func MakeNFA(regexes []string) (Nfa, os.Error) {
	g := &nfa{
	Nodes: map[NodeId]*NfaNode{},
 	Accepting: map[NodeId]string{},
	MaxNode: NodeId(-1),
	}
	g.Start = g.NewNode()
	
	for _, re := range regexes {
		end := g.NewNode()
		g.Accepting[end] = re
		ParseRegex(re).AddToGraph(g, g.Start, end)
	}
	return g, nil
}

func (re plusRe) AddToGraph(g Nfa, from, to NodeId) {
	n1 := g.NewNode()
	n2 := g.NewNode()
	g.AddEpsilon(from, n1)
	g.AddEpsilon(n2, to)
	g.AddEpsilon(n2,n1)
	re.regex.AddToGraph(g, n1, n2)
}

func (re anyRe) AddToGraph(g Nfa, from, to NodeId) {
	for i := 0 ; i < 2; i++ {
		g.AddTransition(uint8(i), from, to)
	}
}

func (re alternateRe) AddToGraph(g Nfa, from, to NodeId) {
	for _, sre := range re {
		sre.AddToGraph(g,from,to)
	}
}

func (re sequenceRe) AddToGraph(g Nfa, from, to NodeId) {
	if len(re) == 0 {
		g.AddEpsilon(from, to)
		return
	}
	for i, sre := range re {
		var next NodeId
		if i == len(re)-1 {
			next = to
		} else 
			next = g.NewNode()
		sre.AddToGraph (g, from, next)
		from = next
	}
}

func (lit literal) AddToGraph (g Nfa, from, to NodeId) {
	g.AddTransition (byte(lit), from, to)
}

func (g *nfa) ToDot(w io.Writer) {
	tw := tabwriter.NewWriter(w, 4, 8, 1, '\t', 0)
	w.Write([]byte("digraph {\n"))
	for i, _ := range g.Nodes {
		
		fmt.Fprintf(tw, 
			"\tn%d\t[label=\"%d\"",
			i, i)
		
		if i == g.Start {
			fmt.Fprintf(tw, ",style=solid,fillcolor=green")
		}

		if _, ok := g.Accepting[i]; ok {
			fmt.Fprintf(tw, ",color=red")
		}
		fmt.Fprintf(tw, "];\n")
	}
	for i, node := range g.Nodes {
		for label, to_list := range node.Nodes {
			for _, to := range to_list {
				fmt.Fprintf(tw, "\tn%d -> n%d\t[label=%#v];\n",
					i, to, strconv.Quote(string(byte(label))))
			}
		}
		for _, to := range node.Epsilons {
			fmt.Fprintf(tw, "\tn%d -> n%d\t[label=\"eps\"];\n", i, to)
		}
	}
	tw.Flush()
	fmt.Fprintf(w, "}\n")
}