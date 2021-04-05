package query

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulee
	ruleeor
	ruleeand
	ruleenot
	rulecondition
	ruletag
	ruleqvalue
	rulevalue
	rulenumber
	ruledigit
	ruletime
	ruledate
	ruleyear
	rulemonth
	ruleday
	ruleand
	ruleor
	rulenot
	ruleequal
	rulene
	rulecontains
	rulele
	rulege
	rulel
	ruleg
	ruleopen
	ruleclose
	rulesp
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	rulePegText
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"eor",
	"eand",
	"enot",
	"condition",
	"tag",
	"qvalue",
	"value",
	"number",
	"digit",
	"time",
	"date",
	"year",
	"month",
	"day",
	"and",
	"or",
	"not",
	"equal",
	"ne",
	"contains",
	"le",
	"ge",
	"l",
	"g",
	"open",
	"close",
	"sp",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"PegText",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type QueryParser struct {
	Expression

	Buffer string
	buffer []rune
	rules  [45]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *QueryParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *QueryParser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *QueryParser) Highlighter() {
	p.PrintSyntax()
}

func (p *QueryParser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.Operator(OpOr)
		case ruleAction1:
			p.Operator(OpAnd)
		case ruleAction2:
			p.Operator(OpNot)
		case ruleAction3:
			p.Operator(OpLessEqual)
		case ruleAction4:
			p.Operator(OpGreaterEqual)
		case ruleAction5:
			p.Operator(OpLess)
		case ruleAction6:
			p.Operator(OpGreater)
		case ruleAction7:
			p.Operator(OpEqual)
		case ruleAction8:
			p.Operator(OpNotEqual)
		case ruleAction9:
			p.Operator(OpContains)
		case ruleAction10:
			p.Tag(buffer[begin:end])
		case ruleAction11:
			p.Value(buffer[begin:end])
		case ruleAction12:
			p.Number(buffer[begin:end])
		case ruleAction13:
			p.Time(buffer[begin:end])
		case ruleAction14:
			p.Date(buffer[begin:end])

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *QueryParser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 e <- <(eor !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleeor]() {
					goto l0
				}
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
				depth--
				add(rulee, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 eor <- <(eand (or eand Action0)*)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !_rules[ruleeand]() {
					goto l3
				}
			l5:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !_rules[ruleor]() {
						goto l6
					}
					if !_rules[ruleeand]() {
						goto l6
					}
					if !_rules[ruleAction0]() {
						goto l6
					}
					goto l5
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				depth--
				add(ruleeor, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 eand <- <(enot (and enot Action1)*)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				if !_rules[ruleenot]() {
					goto l7
				}
			l9:
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if !_rules[ruleand]() {
						goto l10
					}
					if !_rules[ruleenot]() {
						goto l10
					}
					if !_rules[ruleAction1]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
				}
				depth--
				add(ruleeand, position8)
			}
			return true
		l7:
			position, tokenIndex, depth = position7, tokenIndex7, depth7
			return false
		},
		/* 3 enot <- <((not condition Action2) / condition)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !_rules[rulenot]() {
						goto l14
					}
					if !_rules[rulecondition]() {
						goto l14
					}
					if !_rules[ruleAction2]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[rulecondition]() {
						goto l11
					}
				}
			l13:
				depth--
				add(ruleenot, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 4 condition <- <((tag sp ((le (number / time / date) Action3) / (ge (number / time / date) Action4) / (l (number / time / date) Action5) / (g (number / time / date) Action6) / (equal (number / time / date / qvalue) Action7) / (ne (number / time / date / qvalue) Action8) / (contains qvalue Action9)) sp) / (open eor close))> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					if !_rules[ruletag]() {
						goto l18
					}
					if !_rules[rulesp]() {
						goto l18
					}
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						if !_rules[rulele]() {
							goto l20
						}
						{
							position21, tokenIndex21, depth21 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l22
							}
							goto l21
						l22:
							position, tokenIndex, depth = position21, tokenIndex21, depth21
							if !_rules[ruletime]() {
								goto l23
							}
							goto l21
						l23:
							position, tokenIndex, depth = position21, tokenIndex21, depth21
							if !_rules[ruledate]() {
								goto l20
							}
						}
					l21:
						if !_rules[ruleAction3]() {
							goto l20
						}
						goto l19
					l20:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[rulege]() {
							goto l24
						}
						{
							position25, tokenIndex25, depth25 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l26
							}
							goto l25
						l26:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
							if !_rules[ruletime]() {
								goto l27
							}
							goto l25
						l27:
							position, tokenIndex, depth = position25, tokenIndex25, depth25
							if !_rules[ruledate]() {
								goto l24
							}
						}
					l25:
						if !_rules[ruleAction4]() {
							goto l24
						}
						goto l19
					l24:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[rulel]() {
							goto l28
						}
						{
							position29, tokenIndex29, depth29 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l30
							}
							goto l29
						l30:
							position, tokenIndex, depth = position29, tokenIndex29, depth29
							if !_rules[ruletime]() {
								goto l31
							}
							goto l29
						l31:
							position, tokenIndex, depth = position29, tokenIndex29, depth29
							if !_rules[ruledate]() {
								goto l28
							}
						}
					l29:
						if !_rules[ruleAction5]() {
							goto l28
						}
						goto l19
					l28:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[ruleg]() {
							goto l32
						}
						{
							position33, tokenIndex33, depth33 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l34
							}
							goto l33
						l34:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
							if !_rules[ruletime]() {
								goto l35
							}
							goto l33
						l35:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
							if !_rules[ruledate]() {
								goto l32
							}
						}
					l33:
						if !_rules[ruleAction6]() {
							goto l32
						}
						goto l19
					l32:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[ruleequal]() {
							goto l36
						}
						{
							position37, tokenIndex37, depth37 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l38
							}
							goto l37
						l38:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if !_rules[ruletime]() {
								goto l39
							}
							goto l37
						l39:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if !_rules[ruledate]() {
								goto l40
							}
							goto l37
						l40:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if !_rules[ruleqvalue]() {
								goto l36
							}
						}
					l37:
						if !_rules[ruleAction7]() {
							goto l36
						}
						goto l19
					l36:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[rulene]() {
							goto l41
						}
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							if !_rules[rulenumber]() {
								goto l43
							}
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if !_rules[ruletime]() {
								goto l44
							}
							goto l42
						l44:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if !_rules[ruledate]() {
								goto l45
							}
							goto l42
						l45:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if !_rules[ruleqvalue]() {
								goto l41
							}
						}
					l42:
						if !_rules[ruleAction8]() {
							goto l41
						}
						goto l19
					l41:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
						if !_rules[rulecontains]() {
							goto l18
						}
						if !_rules[ruleqvalue]() {
							goto l18
						}
						if !_rules[ruleAction9]() {
							goto l18
						}
					}
				l19:
					if !_rules[rulesp]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
					if !_rules[ruleopen]() {
						goto l15
					}
					if !_rules[ruleeor]() {
						goto l15
					}
					if !_rules[ruleclose]() {
						goto l15
					}
				}
			l17:
				depth--
				add(rulecondition, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 5 tag <- <(<(!(' ' / '\t' / '\n' / '\r' / '\\' / '(' / ')' / '"' / '\'' / '=' / '>' / '<') .)+> sp Action10)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				{
					position48 := position
					depth++
					{
						position51, tokenIndex51, depth51 := position, tokenIndex, depth
						{
							position52, tokenIndex52, depth52 := position, tokenIndex, depth
							if buffer[position] != rune(' ') {
								goto l53
							}
							position++
							goto l52
						l53:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('\t') {
								goto l54
							}
							position++
							goto l52
						l54:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('\n') {
								goto l55
							}
							position++
							goto l52
						l55:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('\r') {
								goto l56
							}
							position++
							goto l52
						l56:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('\\') {
								goto l57
							}
							position++
							goto l52
						l57:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('(') {
								goto l58
							}
							position++
							goto l52
						l58:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune(')') {
								goto l59
							}
							position++
							goto l52
						l59:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('"') {
								goto l60
							}
							position++
							goto l52
						l60:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('\'') {
								goto l61
							}
							position++
							goto l52
						l61:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('=') {
								goto l62
							}
							position++
							goto l52
						l62:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('>') {
								goto l63
							}
							position++
							goto l52
						l63:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('<') {
								goto l51
							}
							position++
						}
					l52:
						goto l46
					l51:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
					}
					if !matchDot() {
						goto l46
					}
				l49:
					{
						position50, tokenIndex50, depth50 := position, tokenIndex, depth
						{
							position64, tokenIndex64, depth64 := position, tokenIndex, depth
							{
								position65, tokenIndex65, depth65 := position, tokenIndex, depth
								if buffer[position] != rune(' ') {
									goto l66
								}
								position++
								goto l65
							l66:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('\t') {
									goto l67
								}
								position++
								goto l65
							l67:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('\n') {
									goto l68
								}
								position++
								goto l65
							l68:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('\r') {
									goto l69
								}
								position++
								goto l65
							l69:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('\\') {
									goto l70
								}
								position++
								goto l65
							l70:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('(') {
									goto l71
								}
								position++
								goto l65
							l71:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune(')') {
									goto l72
								}
								position++
								goto l65
							l72:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('"') {
									goto l73
								}
								position++
								goto l65
							l73:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('\'') {
									goto l74
								}
								position++
								goto l65
							l74:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('=') {
									goto l75
								}
								position++
								goto l65
							l75:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('>') {
									goto l76
								}
								position++
								goto l65
							l76:
								position, tokenIndex, depth = position65, tokenIndex65, depth65
								if buffer[position] != rune('<') {
									goto l64
								}
								position++
							}
						l65:
							goto l50
						l64:
							position, tokenIndex, depth = position64, tokenIndex64, depth64
						}
						if !matchDot() {
							goto l50
						}
						goto l49
					l50:
						position, tokenIndex, depth = position50, tokenIndex50, depth50
					}
					depth--
					add(rulePegText, position48)
				}
				if !_rules[rulesp]() {
					goto l46
				}
				if !_rules[ruleAction10]() {
					goto l46
				}
				depth--
				add(ruletag, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 6 qvalue <- <('\'' value '\'' sp)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l77
				}
				position++
				if !_rules[rulevalue]() {
					goto l77
				}
				if buffer[position] != rune('\'') {
					goto l77
				}
				position++
				if !_rules[rulesp]() {
					goto l77
				}
				depth--
				add(ruleqvalue, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 7 value <- <(<(!('"' / '\'') .)*> Action11)> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				{
					position81 := position
					depth++
				l82:
					{
						position83, tokenIndex83, depth83 := position, tokenIndex, depth
						{
							position84, tokenIndex84, depth84 := position, tokenIndex, depth
							{
								position85, tokenIndex85, depth85 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l86
								}
								position++
								goto l85
							l86:
								position, tokenIndex, depth = position85, tokenIndex85, depth85
								if buffer[position] != rune('\'') {
									goto l84
								}
								position++
							}
						l85:
							goto l83
						l84:
							position, tokenIndex, depth = position84, tokenIndex84, depth84
						}
						if !matchDot() {
							goto l83
						}
						goto l82
					l83:
						position, tokenIndex, depth = position83, tokenIndex83, depth83
					}
					depth--
					add(rulePegText, position81)
				}
				if !_rules[ruleAction11]() {
					goto l79
				}
				depth--
				add(rulevalue, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 8 number <- <(<('0' / ([1-9] digit* ('.' digit*)?))> Action12)> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				{
					position89 := position
					depth++
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
						if buffer[position] != rune('0') {
							goto l91
						}
						position++
						goto l90
					l91:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
						if c := buffer[position]; c < rune('1') || c > rune('9') {
							goto l87
						}
						position++
					l92:
						{
							position93, tokenIndex93, depth93 := position, tokenIndex, depth
							if !_rules[ruledigit]() {
								goto l93
							}
							goto l92
						l93:
							position, tokenIndex, depth = position93, tokenIndex93, depth93
						}
						{
							position94, tokenIndex94, depth94 := position, tokenIndex, depth
							if buffer[position] != rune('.') {
								goto l94
							}
							position++
						l96:
							{
								position97, tokenIndex97, depth97 := position, tokenIndex, depth
								if !_rules[ruledigit]() {
									goto l97
								}
								goto l96
							l97:
								position, tokenIndex, depth = position97, tokenIndex97, depth97
							}
							goto l95
						l94:
							position, tokenIndex, depth = position94, tokenIndex94, depth94
						}
					l95:
					}
				l90:
					depth--
					add(rulePegText, position89)
				}
				if !_rules[ruleAction12]() {
					goto l87
				}
				depth--
				add(rulenumber, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 9 digit <- <[0-9]> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l98
				}
				position++
				depth--
				add(ruledigit, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 10 time <- <(('t' / 'T') ('i' / 'I') ('m' / 'M') ('e' / 'E') ' ' <(year '-' month '-' day 'T' digit digit ':' digit digit ':' digit digit ((('-' / '+') digit digit ':' digit digit) / 'Z'))> Action13)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l103
					}
					position++
					goto l102
				l103:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
					if buffer[position] != rune('T') {
						goto l100
					}
					position++
				}
			l102:
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if buffer[position] != rune('i') {
						goto l105
					}
					position++
					goto l104
				l105:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					if buffer[position] != rune('I') {
						goto l100
					}
					position++
				}
			l104:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l107
					}
					position++
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if buffer[position] != rune('M') {
						goto l100
					}
					position++
				}
			l106:
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l109
					}
					position++
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if buffer[position] != rune('E') {
						goto l100
					}
					position++
				}
			l108:
				if buffer[position] != rune(' ') {
					goto l100
				}
				position++
				{
					position110 := position
					depth++
					if !_rules[ruleyear]() {
						goto l100
					}
					if buffer[position] != rune('-') {
						goto l100
					}
					position++
					if !_rules[rulemonth]() {
						goto l100
					}
					if buffer[position] != rune('-') {
						goto l100
					}
					position++
					if !_rules[ruleday]() {
						goto l100
					}
					if buffer[position] != rune('T') {
						goto l100
					}
					position++
					if !_rules[ruledigit]() {
						goto l100
					}
					if !_rules[ruledigit]() {
						goto l100
					}
					if buffer[position] != rune(':') {
						goto l100
					}
					position++
					if !_rules[ruledigit]() {
						goto l100
					}
					if !_rules[ruledigit]() {
						goto l100
					}
					if buffer[position] != rune(':') {
						goto l100
					}
					position++
					if !_rules[ruledigit]() {
						goto l100
					}
					if !_rules[ruledigit]() {
						goto l100
					}
					{
						position111, tokenIndex111, depth111 := position, tokenIndex, depth
						{
							position113, tokenIndex113, depth113 := position, tokenIndex, depth
							if buffer[position] != rune('-') {
								goto l114
							}
							position++
							goto l113
						l114:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('+') {
								goto l112
							}
							position++
						}
					l113:
						if !_rules[ruledigit]() {
							goto l112
						}
						if !_rules[ruledigit]() {
							goto l112
						}
						if buffer[position] != rune(':') {
							goto l112
						}
						position++
						if !_rules[ruledigit]() {
							goto l112
						}
						if !_rules[ruledigit]() {
							goto l112
						}
						goto l111
					l112:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
						if buffer[position] != rune('Z') {
							goto l100
						}
						position++
					}
				l111:
					depth--
					add(rulePegText, position110)
				}
				if !_rules[ruleAction13]() {
					goto l100
				}
				depth--
				add(ruletime, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 11 date <- <(('d' / 'D') ('a' / 'A') ('t' / 'T') ('e' / 'E') ' ' <(year '-' month '-' day)> Action14)> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if buffer[position] != rune('D') {
						goto l115
					}
					position++
				}
			l117:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l120
					}
					position++
					goto l119
				l120:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
					if buffer[position] != rune('A') {
						goto l115
					}
					position++
				}
			l119:
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l122
					}
					position++
					goto l121
				l122:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if buffer[position] != rune('T') {
						goto l115
					}
					position++
				}
			l121:
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l124
					}
					position++
					goto l123
				l124:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if buffer[position] != rune('E') {
						goto l115
					}
					position++
				}
			l123:
				if buffer[position] != rune(' ') {
					goto l115
				}
				position++
				{
					position125 := position
					depth++
					if !_rules[ruleyear]() {
						goto l115
					}
					if buffer[position] != rune('-') {
						goto l115
					}
					position++
					if !_rules[rulemonth]() {
						goto l115
					}
					if buffer[position] != rune('-') {
						goto l115
					}
					position++
					if !_rules[ruleday]() {
						goto l115
					}
					depth--
					add(rulePegText, position125)
				}
				if !_rules[ruleAction14]() {
					goto l115
				}
				depth--
				add(ruledate, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 12 year <- <(('1' / '2') digit digit digit)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if buffer[position] != rune('1') {
						goto l129
					}
					position++
					goto l128
				l129:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
					if buffer[position] != rune('2') {
						goto l126
					}
					position++
				}
			l128:
				if !_rules[ruledigit]() {
					goto l126
				}
				if !_rules[ruledigit]() {
					goto l126
				}
				if !_rules[ruledigit]() {
					goto l126
				}
				depth--
				add(ruleyear, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 13 month <- <(('0' / '1') digit)> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l133
					}
					position++
					goto l132
				l133:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if buffer[position] != rune('1') {
						goto l130
					}
					position++
				}
			l132:
				if !_rules[ruledigit]() {
					goto l130
				}
				depth--
				add(rulemonth, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 14 day <- <(('0' / '1' / '2' / '3') digit)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l137
					}
					position++
					goto l136
				l137:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
					if buffer[position] != rune('1') {
						goto l138
					}
					position++
					goto l136
				l138:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
					if buffer[position] != rune('2') {
						goto l139
					}
					position++
					goto l136
				l139:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
					if buffer[position] != rune('3') {
						goto l134
					}
					position++
				}
			l136:
				if !_rules[ruledigit]() {
					goto l134
				}
				depth--
				add(ruleday, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 15 and <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') sp)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l143
					}
					position++
					goto l142
				l143:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if buffer[position] != rune('A') {
						goto l140
					}
					position++
				}
			l142:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l145
					}
					position++
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('N') {
						goto l140
					}
					position++
				}
			l144:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l147
					}
					position++
					goto l146
				l147:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('D') {
						goto l140
					}
					position++
				}
			l146:
				if !_rules[rulesp]() {
					goto l140
				}
				depth--
				add(ruleand, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 16 or <- <(('o' / 'O') ('r' / 'R') sp)> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l151
					}
					position++
					goto l150
				l151:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					if buffer[position] != rune('O') {
						goto l148
					}
					position++
				}
			l150:
				{
					position152, tokenIndex152, depth152 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l153
					}
					position++
					goto l152
				l153:
					position, tokenIndex, depth = position152, tokenIndex152, depth152
					if buffer[position] != rune('R') {
						goto l148
					}
					position++
				}
			l152:
				if !_rules[rulesp]() {
					goto l148
				}
				depth--
				add(ruleor, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 17 not <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') sp)> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l157
					}
					position++
					goto l156
				l157:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
					if buffer[position] != rune('N') {
						goto l154
					}
					position++
				}
			l156:
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l159
					}
					position++
					goto l158
				l159:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					if buffer[position] != rune('O') {
						goto l154
					}
					position++
				}
			l158:
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l161
					}
					position++
					goto l160
				l161:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if buffer[position] != rune('T') {
						goto l154
					}
					position++
				}
			l160:
				if !_rules[rulesp]() {
					goto l154
				}
				depth--
				add(rulenot, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 18 equal <- <('=' sp)> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('=') {
					goto l162
				}
				position++
				if !_rules[rulesp]() {
					goto l162
				}
				depth--
				add(ruleequal, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 19 ne <- <('!' '=' sp)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if buffer[position] != rune('!') {
					goto l164
				}
				position++
				if buffer[position] != rune('=') {
					goto l164
				}
				position++
				if !_rules[rulesp]() {
					goto l164
				}
				depth--
				add(rulene, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 20 contains <- <(('c' / 'C') ('o' / 'O') ('n' / 'N') ('t' / 'T') ('a' / 'A') ('i' / 'I') ('n' / 'N') ('s' / 'S') sp)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l169
					}
					position++
					goto l168
				l169:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if buffer[position] != rune('C') {
						goto l166
					}
					position++
				}
			l168:
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l171
					}
					position++
					goto l170
				l171:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
					if buffer[position] != rune('O') {
						goto l166
					}
					position++
				}
			l170:
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l173
					}
					position++
					goto l172
				l173:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('N') {
						goto l166
					}
					position++
				}
			l172:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l175
					}
					position++
					goto l174
				l175:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if buffer[position] != rune('T') {
						goto l166
					}
					position++
				}
			l174:
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l177
					}
					position++
					goto l176
				l177:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
					if buffer[position] != rune('A') {
						goto l166
					}
					position++
				}
			l176:
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if buffer[position] != rune('i') {
						goto l179
					}
					position++
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if buffer[position] != rune('I') {
						goto l166
					}
					position++
				}
			l178:
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l181
					}
					position++
					goto l180
				l181:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
					if buffer[position] != rune('N') {
						goto l166
					}
					position++
				}
			l180:
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l183
					}
					position++
					goto l182
				l183:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('S') {
						goto l166
					}
					position++
				}
			l182:
				if !_rules[rulesp]() {
					goto l166
				}
				depth--
				add(rulecontains, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 21 le <- <('<' '=' sp)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('<') {
					goto l184
				}
				position++
				if buffer[position] != rune('=') {
					goto l184
				}
				position++
				if !_rules[rulesp]() {
					goto l184
				}
				depth--
				add(rulele, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 22 ge <- <('>' '=' sp)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if buffer[position] != rune('>') {
					goto l186
				}
				position++
				if buffer[position] != rune('=') {
					goto l186
				}
				position++
				if !_rules[rulesp]() {
					goto l186
				}
				depth--
				add(rulege, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 23 l <- <('<' sp)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				if buffer[position] != rune('<') {
					goto l188
				}
				position++
				if !_rules[rulesp]() {
					goto l188
				}
				depth--
				add(rulel, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 24 g <- <('>' sp)> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				if buffer[position] != rune('>') {
					goto l190
				}
				position++
				if !_rules[rulesp]() {
					goto l190
				}
				depth--
				add(ruleg, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 25 open <- <('(' sp)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if buffer[position] != rune('(') {
					goto l192
				}
				position++
				if !_rules[rulesp]() {
					goto l192
				}
				depth--
				add(ruleopen, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 26 close <- <(')' sp)> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if buffer[position] != rune(')') {
					goto l194
				}
				position++
				if !_rules[rulesp]() {
					goto l194
				}
				depth--
				add(ruleclose, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 27 sp <- <(' ' / '\t')*> */
		func() bool {
			{
				position197 := position
				depth++
			l198:
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l201
						}
						position++
						goto l200
					l201:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
						if buffer[position] != rune('\t') {
							goto l199
						}
						position++
					}
				l200:
					goto l198
				l199:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
				}
				depth--
				add(rulesp, position197)
			}
			return true
		},
		/* 29 Action0 <- <{ p.Operator(OpOr) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 30 Action1 <- <{ p.Operator(OpAnd) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 31 Action2 <- <{ p.Operator(OpNot) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 32 Action3 <- <{ p.Operator(OpLessEqual) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 33 Action4 <- <{ p.Operator(OpGreaterEqual) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 34 Action5 <- <{ p.Operator(OpLess) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 35 Action6 <- <{ p.Operator(OpGreater) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 36 Action7 <- <{ p.Operator(OpEqual) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 37 Action8 <- <{ p.Operator(OpNotEqual) }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 38 Action9 <- <{ p.Operator(OpContains) }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		nil,
		/* 40 Action10 <- <{ p.Tag(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 41 Action11 <- <{ p.Value(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 42 Action12 <- <{ p.Number(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 43 Action13 <- <{ p.Time(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 44 Action14 <- <{ p.Date(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
	}
	p.rules = _rules
}
