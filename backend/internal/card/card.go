package card

import "fmt"

type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

func (s Suit) String() string {
	return [...]string{"♠", "♥", "♦", "♣"}[s]
}

type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

func (r Rank) String() string {
	switch r {
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

type Card struct {
	Rank Rank
	Suit Suit
}

func New(r Rank, s Suit) Card {
	return Card{Rank: r, Suit: s}
}

func (c Card) String() string {
	return c.Rank.String() + c.Suit.String()
}

func (c Card) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, c.String())), nil
}
