package hand

import (
	"fmt"
	"lil-poker/internal/card"
)

type Rank int

const (
	HighCard Rank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (r Rank) String() string {
	return [...]string{
		"High Card", "Pair", "Two Pair", "Three of a Kind",
		"Straight", "Flush", "Full House", "Square of a Kind",
		"Straight Flush", "Royal Flush",
	}[r]
}

type Result struct {
	Rank     Rank
	Cards    []card.Card
	TieBreak []int
}

func (r Result) String() string {
	return fmt.Sprintf("%s %v", r.Rank, r.Cards)
}

var comboIndices7 = [21][5]int{
	{0, 1, 2, 3, 4}, {0, 1, 2, 3, 5}, {0, 1, 2, 3, 6}, {0, 1, 2, 4, 5}, {0, 1, 2, 4, 6},
	{0, 1, 2, 5, 6}, {0, 1, 3, 4, 5}, {0, 1, 3, 4, 6}, {0, 1, 3, 5, 6}, {0, 1, 4, 5, 6},
	{0, 2, 3, 4, 5}, {0, 2, 3, 4, 6}, {0, 2, 3, 5, 6}, {0, 2, 4, 5, 6}, {0, 3, 4, 5, 6},
	{1, 2, 3, 4, 5}, {1, 2, 3, 4, 6}, {1, 2, 3, 5, 6}, {1, 2, 4, 5, 6}, {1, 3, 4, 5, 6},
	{2, 3, 4, 5, 6},
}

var comboIndices6 = [6][5]int{
	{0, 1, 2, 3, 4}, {0, 1, 2, 3, 5}, {0, 1, 2, 4, 5},
	{0, 1, 3, 4, 5}, {0, 2, 3, 4, 5}, {1, 2, 3, 4, 5},
}

type evalResult struct {
	rank     Rank
	tieBreak [5]int
	tbLen    int
	cards    [5]card.Card
}

func compareEvalResult(r1, r2 evalResult) int {
	if r1.rank > r2.rank {
		return 1
	}
	if r1.rank < r2.rank {
		return -1
	}
	for i := 0; i < r1.tbLen && i < r2.tbLen; i++ {
		if r1.tieBreak[i] > r2.tieBreak[i] {
			return 1
		}
		if r1.tieBreak[i] < r2.tieBreak[i] {
			return -1
		}
	}
	return 0
}

func Evaluate(cards []card.Card) Result {
	if len(cards) < 5 {
		return Result{Rank: HighCard, Cards: cards}
	}

	best := evalResult{rank: -1}
	var temp [5]card.Card

	if len(cards) == 7 {
		for _, indices := range comboIndices7 {
			temp[0] = cards[indices[0]]
			temp[1] = cards[indices[1]]
			temp[2] = cards[indices[2]]
			temp[3] = cards[indices[3]]
			temp[4] = cards[indices[4]]
			r := evaluate5Array(temp)
			if compareEvalResult(r, best) > 0 {
				best = r
			}
		}
	} else if len(cards) == 6 {
		for _, indices := range comboIndices6 {
			temp[0] = cards[indices[0]]
			temp[1] = cards[indices[1]]
			temp[2] = cards[indices[2]]
			temp[3] = cards[indices[3]]
			temp[4] = cards[indices[4]]
			r := evaluate5Array(temp)
			if compareEvalResult(r, best) > 0 {
				best = r
			}
		}
	} else {
		temp[0] = cards[0]
		temp[1] = cards[1]
		temp[2] = cards[2]
		temp[3] = cards[3]
		temp[4] = cards[4]
		best = evaluate5Array(temp)
	}

	resCards := make([]card.Card, 5)
	copy(resCards, best.cards[:])
	resTB := make([]int, best.tbLen)
	copy(resTB, best.tieBreak[:best.tbLen])
	return Result{Rank: best.rank, Cards: resCards, TieBreak: resTB}
}

func rankCountsArray(cards [5]card.Card) [15]int {
	var counts [15]int
	for _, c := range cards {
		counts[c.Rank]++
	}
	return counts
}

func evaluate5Array(cards [5]card.Card) evalResult {
	sorted := cards
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 5; j++ {
			if sorted[i].Rank < sorted[j].Rank {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	sortedSlice := sorted[:]
	counts := rankCountsArray(sorted)
	isFlush := checkFlush(sortedSlice)
	isStraight, straightHigh := checkStraight(sortedSlice)

	var res evalResult
	res.cards = sorted

	switch {
	case isFlush && isStraight && straightHigh == card.Ace:
		res.rank = RoyalFlush
		res.tieBreak[0] = int(card.Ace)
		res.tbLen = 1

	case isFlush && isStraight:
		res.rank = StraightFlush
		res.tieBreak[0] = int(straightHigh)
		res.tbLen = 1

	case hasFourOfAKindC(counts):
		quad, kick := getFourOfAKindC(counts)
		res.rank = FourOfAKind
		res.tieBreak[0] = int(quad)
		res.tieBreak[1] = int(kick)
		res.tbLen = 2

	case hasFullHouseC(counts):
		three, two := getFullHouseC(counts)
		res.rank = FullHouse
		res.tieBreak[0] = int(three)
		res.tieBreak[1] = int(two)
		res.tbLen = 2

	case isFlush:
		res.rank = Flush
		tb := rankList(sortedSlice)
		copy(res.tieBreak[:], tb)
		res.tbLen = len(tb)

	case isStraight:
		res.rank = Straight
		res.tieBreak[0] = int(straightHigh)
		res.tbLen = 1

	case hasThreeOfAKindC(counts):
		three, kicks := getThreeOfAKindC(counts)
		res.rank = ThreeOfAKind
		res.tieBreak[0] = int(three)
		copy(res.tieBreak[1:], kicks)
		res.tbLen = 1 + len(kicks)

	case hasTwoPairC(counts):
		high, low, kick := getTwoPairC(counts)
		res.rank = TwoPair
		res.tieBreak[0] = int(high)
		res.tieBreak[1] = int(low)
		res.tieBreak[2] = int(kick)
		res.tbLen = 3

	case hasOnePairC(counts):
		pair, kicks := getOnePairC(counts)
		res.rank = OnePair
		res.tieBreak[0] = int(pair)
		copy(res.tieBreak[1:], kicks)
		res.tbLen = 1 + len(kicks)

	default:
		res.rank = HighCard
		tb := rankList(sortedSlice)
		copy(res.tieBreak[:], tb)
		res.tbLen = len(tb)
	}
	return res
}

func checkFlush(cards []card.Card) bool {
	suit := cards[0].Suit
	for _, c := range cards[1:] {
		if c.Suit != suit {
			return false
		}
	}
	return true
}

func checkStraight(cards []card.Card) (bool, card.Rank) {
	ranks := rankList(cards)
	isSeq := true
	for i := 1; i < len(ranks); i++ {
		if ranks[i-1]-ranks[i] != 1 {
			isSeq = false
			break
		}
	}
	if isSeq {
		return true, card.Rank(ranks[0])
	}
	if ranks[0] == int(card.Ace) && ranks[1] == int(card.Five) &&
		ranks[2] == int(card.Four) && ranks[3] == int(card.Three) && ranks[4] == int(card.Two) {
		return true, card.Five
	}
	return false, 0
}

func rankList(cards []card.Card) []int {
	result := make([]int, len(cards))
	for i, c := range cards {
		result[i] = int(c.Rank)
	}
	return result
}

func hasFourOfAKindC(counts [15]int) bool {
	for _, v := range counts {
		if v == 4 {
			return true
		}
	}
	return false
}

func getFourOfAKindC(counts [15]int) (card.Rank, card.Rank) {
	var quad, kick card.Rank
	for r, v := range counts {
		if v == 4 {
			quad = card.Rank(r)
		} else if v > 0 {
			kick = card.Rank(r)
		}
	}
	return quad, kick
}

func hasFullHouseC(counts [15]int) bool {
	hasThree, hasPair := false, false
	for _, v := range counts {
		if v == 3 {
			hasThree = true
		}
		if v == 2 {
			hasPair = true
		}
	}
	return hasThree && hasPair
}

func getFullHouseC(counts [15]int) (card.Rank, card.Rank) {
	var three, two card.Rank
	for r, v := range counts {
		switch v {
		case 3:
			three = card.Rank(r)
		case 2:
			two = card.Rank(r)
		}
	}
	return three, two
}

func hasThreeOfAKindC(counts [15]int) bool {
	for _, v := range counts {
		if v == 3 {
			return true
		}
	}
	return false
}

func getThreeOfAKindC(counts [15]int) (card.Rank, []int) {
	var three card.Rank
	var kicks [2]int
	n := 0
	for r := len(counts) - 1; r >= 0; r-- {
		if counts[r] == 3 {
			three = card.Rank(r)
		} else if counts[r] == 1 && n < 2 {
			kicks[n] = r
			n++
		}
	}
	return three, kicks[:n]
}

func hasTwoPairC(counts [15]int) bool {
	pairs := 0
	for _, v := range counts {
		if v == 2 {
			pairs++
		}
	}
	return pairs == 2
}

func getTwoPairC(counts [15]int) (card.Rank, card.Rank, card.Rank) {
	var pairs [2]card.Rank
	var kick card.Rank
	n := 0
	for r := len(counts) - 1; r >= 0; r-- {
		if counts[r] == 2 && n < 2 {
			pairs[n] = card.Rank(r)
			n++
		} else if counts[r] == 1 {
			kick = card.Rank(r)
		}
	}
	return pairs[0], pairs[1], kick
}

func hasOnePairC(counts [15]int) bool {
	for _, v := range counts {
		if v == 2 {
			return true
		}
	}
	return false
}

func getOnePairC(counts [15]int) (card.Rank, []int) {
	var pair card.Rank
	var kicks [3]int
	n := 0
	for r := len(counts) - 1; r >= 0; r-- {
		if counts[r] == 2 {
			pair = card.Rank(r)
		} else if counts[r] == 1 && n < 3 {
			kicks[n] = r
			n++
		}
	}
	return pair, kicks[:n]
}

func (r Result) Description() string {
	if len(r.TieBreak) == 0 {
		return r.Rank.String()
	}
	pluralRanks := map[int]string{
		2: "Twos", 3: "Threes", 4: "Fours", 5: "Fives", 6: "Sixes", 7: "Sevens",
		8: "Eights", 9: "Nines", 10: "Tens", 11: "Jacks", 12: "Queens", 13: "Kings", 14: "Aces",
	}
	singularRanks := map[int]string{
		2: "Two", 3: "Three", 4: "Four", 5: "Five", 6: "Six", 7: "Seven",
		8: "Eight", 9: "Nine", 10: "Ten", 11: "Jack", 12: "Queen", 13: "King", 14: "Ace",
	}

	getRankName := func(rankVal int, plural bool) string {
		if plural {
			if name, ok := pluralRanks[rankVal]; ok {
				return name
			}
		} else {
			if name, ok := singularRanks[rankVal]; ok {
				return name
			}
		}
		return fmt.Sprintf("%d", rankVal)
	}

	switch r.Rank {
	case RoyalFlush:
		return "Royal Flush"
	case StraightFlush:
		return fmt.Sprintf("Straight Flush, %s High", getRankName(r.TieBreak[0], false))
	case FourOfAKind:
		return fmt.Sprintf("Four of a Kind, %s", getRankName(r.TieBreak[0], true))
	case FullHouse:
		return fmt.Sprintf("Full House, %s full of %s", getRankName(r.TieBreak[0], true), getRankName(r.TieBreak[1], true))
	case Flush:
		return fmt.Sprintf("Flush, %s High", getRankName(r.TieBreak[0], false))
	case Straight:
		return fmt.Sprintf("Straight, %s High", getRankName(r.TieBreak[0], false))
	case ThreeOfAKind:
		return fmt.Sprintf("Three of a Kind, %s", getRankName(r.TieBreak[0], true))
	case TwoPair:
		return fmt.Sprintf("Two Pair, %s and %s", getRankName(r.TieBreak[0], true), getRankName(r.TieBreak[1], true))
	case OnePair:
		return fmt.Sprintf("Pair of %s", getRankName(r.TieBreak[0], true))
	default:
		return fmt.Sprintf("High Card, %s", getRankName(r.TieBreak[0], false))
	}
}

func (r Result) GetCombinationCards() []card.Card {
	if len(r.TieBreak) == 0 {
		return r.Cards
	}
	var filtered []card.Card

	switch r.Rank {
	case RoyalFlush, StraightFlush, FullHouse, Flush, Straight:
		return r.Cards

	case FourOfAKind:
		quadRank := r.TieBreak[0]
		for _, c := range r.Cards {
			if int(c.Rank) == quadRank {
				filtered = append(filtered, c)
			}
		}
		return filtered

	case ThreeOfAKind:
		threeRank := r.TieBreak[0]
		for _, c := range r.Cards {
			if int(c.Rank) == threeRank {
				filtered = append(filtered, c)
			}
		}
		return filtered

	case TwoPair:
		highRank := r.TieBreak[0]
		lowRank := r.TieBreak[1]
		for _, c := range r.Cards {
			if int(c.Rank) == highRank || int(c.Rank) == lowRank {
				filtered = append(filtered, c)
			}
		}
		return filtered

	case OnePair:
		pairRank := r.TieBreak[0]
		for _, c := range r.Cards {
			if int(c.Rank) == pairRank {
				filtered = append(filtered, c)
			}
		}
		return filtered

	case HighCard:
		highRank := r.TieBreak[0]
		for _, c := range r.Cards {
			if int(c.Rank) == highRank {
				filtered = append(filtered, c)
				break
			}
		}
		return filtered

	default:
		return r.Cards
	}
}
