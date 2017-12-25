package parser

func ComputeFirsts(prods []Production,
	tokens map[string]int,
	maxterm int) map[string][]int {
	firsts := make(map[string][]int)

	firsts[""] = []int{0}
	for tok, id := range tokens {
		if id <= maxterm {
			firsts[tok] = []int{id}
		}
		if _, b := firsts[tok]; !b {
			firsts[tok] = make([]int, 0)
		}
	}
	parserLog("Firsts (After terms): %v", firsts)
	changed := false
	var tmpBool bool
	for {
		for _, prod := range prods {
			lhs := firsts[prod.name]
		Inner:
			for _, sym := range prod.body {
				lhs, tmpBool = mergeSets(lhs, firsts[sym])
				changed = changed || tmpBool
				if indexValue(firsts[sym], 0) != -1 {
					continue
				} else {
					break Inner
				}
			}
			if len(prod.body) == 0 {
				lhs, tmpBool = mergeSets(lhs, firsts[""])
				changed = changed || tmpBool
			}
			firsts[prod.name] = lhs
		}
		if !changed {
			break
		}
		changed = false
		parserLog("After A Round, Firsts:\n%v\n", firsts)
	}
	return firsts
}

func ComputeFollows(prods []Production, tokens map[string]int, firsts map[string][]int) map[string][]int {
	follows := make(map[string][]int)
	follows[prods[0].name] = []int{1}

	changed := false
	var tmpBool bool
	for {
		for _, prod := range prods {
			if len(prod.body) == 0 {
				continue
			}
			nbody := len(prod.body) - 1
			emptyTail := true
			if tokens[prod.body[nbody]] > MAXTOKEN {
				lhs := follows[prod.body[nbody]]
				rhs := follows[prod.name]
				lhs, tmpBool = mergeSetsNoE(lhs, rhs)
				follows[prod.body[nbody]] = lhs
				changed = changed || tmpBool

				if indexValue(firsts[prod.body[nbody]], 0) == -1 {
					emptyTail = false
				}
			} else {
				emptyTail = false
			}
			for i := nbody - 1; i >= 0; i-- {
				tokName := prod.body[i]
				if tokens[tokName] > MAXTOKEN {
					lhs := follows[tokName]
					rhs := firsts[prod.body[i+1]]
					lhs, tmpBool = mergeSetsNoE(lhs, rhs)
					follows[tokName] = lhs
					changed = changed || tmpBool
				}
				if emptyTail {
					lhs := follows[tokName]
					rhs := follows[prod.name]
					lhs, tmpBool = mergeSetsNoE(lhs, rhs)
					follows[tokName] = lhs
					changed = changed || tmpBool
					if indexValue(firsts[prod.body[i]], 0) == -1 {
						emptyTail = false
					}
				}
			}
		}
		if !changed {
			break
		}
		changed = false
		parserLog("After a round, Follows:\n%v", follows)
	}

	return follows
}

func ComputeLLTable(prods []Production,
	tokens map[string]int,
	firsts map[string][]int,
	follows map[string][]int,
	symBegin, symEnd int) map[int][]int {
	lltable := make(map[int][]int)
	// init table
	for i := symBegin; i <= symEnd; i++ {
		lltable[i] = make([]int, symBegin)
		for idx := range lltable[i] {
			lltable[i][idx] = -1
		}
	}

	for i, prod := range prods {
		nullable := true

		for _, body := range prod.body {
			for _, tok := range firsts[body] {
				lltable[tokens[prod.name]][tok] = i
			}
			if indexValue(firsts[body], 0) == -1 {
				nullable = false
			}
			if !nullable {
				break
			}
		}
		if nullable {
			for _, tok := range follows[prod.name] {
				lltable[tokens[prod.name]][tok] = i
			}
		}
	}

	return lltable
}

func mergeSets(lhs []int, rhs []int) ([]int, bool) {
	merged := false
	for _, v := range rhs {
		if indexValue(lhs, v) == -1 {
			lhs = append(lhs, v)
			merged = true
		}
	}
	return lhs, merged
}

func mergeSetsNoE(lhs []int, rhs []int) ([]int, bool) {
	merged := false
	for _, v := range rhs {
		if v == 0 {
			continue
		}
		if indexValue(lhs, v) == -1 {
			lhs = append(lhs, v)
			merged = true
		}
	}
	return lhs, merged
}

func indexValue(lhs []int, val int) int {
	for i, v := range lhs {
		if v == val {
			return i
		}
	}
	return -1
}
