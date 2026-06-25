package room

import (
	"lil-poker/internal/card"
	"lil-poker/internal/hand"
)

var rankNames = [...]string{
	"Two", "Three", "Four", "Five", "Six", "Seven",
	"Eight", "Nine", "Ten", "Jack", "Queen", "King", "Ace",
}

func rankLongString(r card.Rank) string {
	idx := int(r) - int(card.Two)
	if idx >= 0 && idx < len(rankNames) {
		return rankNames[idx]
	}
	return "Unknown"
}

func rankPluralString(r card.Rank) string {
	if r == card.Six {
		return "Sixes"
	}
	return rankLongString(r) + "s"
}

func getHandDescription(res hand.Result) string {
	switch res.Rank {
	case hand.RoyalFlush:
		return "Royal Flush"
	case hand.StraightFlush:
		if len(res.TieBreak) > 0 {
			return "Straight Flush, " + rankLongString(card.Rank(res.TieBreak[0])) + " High"
		}
		return "Straight Flush"
	case hand.FourOfAKind:
		if len(res.TieBreak) > 0 {
			return "Four of a Kind, " + rankPluralString(card.Rank(res.TieBreak[0]))
		}
		return "Four of a Kind"
	case hand.FullHouse:
		if len(res.TieBreak) > 1 {
			return "Full House, " + rankPluralString(card.Rank(res.TieBreak[0])) + " full of " + rankPluralString(card.Rank(res.TieBreak[1]))
		}
		return "Full House"
	case hand.Flush:
		if len(res.TieBreak) > 0 {
			return "Flush, " + rankLongString(card.Rank(res.TieBreak[0])) + " High"
		}
		return "Flush"
	case hand.Straight:
		if len(res.TieBreak) > 0 {
			return "Straight, " + rankLongString(card.Rank(res.TieBreak[0])) + " High"
		}
		return "Straight"
	case hand.ThreeOfAKind:
		if len(res.TieBreak) > 0 {
			return "Three of a Kind, " + rankPluralString(card.Rank(res.TieBreak[0]))
		}
		return "Three of a Kind"
	case hand.TwoPair:
		if len(res.TieBreak) > 1 {
			return "Two Pair, " + rankPluralString(card.Rank(res.TieBreak[0])) + " and " + rankPluralString(card.Rank(res.TieBreak[1]))
		}
		return "Two Pair"
	case hand.OnePair:
		if len(res.TieBreak) > 0 {
			return "Pair of " + rankPluralString(card.Rank(res.TieBreak[0]))
		}
		return "Pair"
	default:
		if len(res.TieBreak) > 0 {
			return "High Card, " + rankLongString(card.Rank(res.TieBreak[0]))
		}
		return "High Card"
	}
}
