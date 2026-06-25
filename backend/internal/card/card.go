package card

import (
	"encoding/json"
	"fmt"
)

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

func (c *Card) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	runes := []rune(s)
	if len(runes) < 2 {
		return fmt.Errorf("invalid card string: %q", s)
	}
	suitRune := runes[len(runes)-1]
	rankStr := string(runes[:len(runes)-1])

	switch suitRune {
	case '♠':
		c.Suit = Spades
	case '♥':
		c.Suit = Hearts
	case '♦':
		c.Suit = Diamonds
	case '♣':
		c.Suit = Clubs
	default:
		return fmt.Errorf("invalid suit rune: %c", suitRune)
	}

	switch rankStr {
	case "2":
		c.Rank = Two
	case "3":
		c.Rank = Three
	case "4":
		c.Rank = Four
	case "5":
		c.Rank = Five
	case "6":
		c.Rank = Six
	case "7":
		c.Rank = Seven
	case "8":
		c.Rank = Eight
	case "9":
		c.Rank = Nine
	case "10":
		c.Rank = Ten
	case "J":
		c.Rank = Jack
	case "Q":
		c.Rank = Queen
	case "K":
		c.Rank = King
	case "A":
		c.Rank = Ace
	default:
		return fmt.Errorf("invalid rank string: %q", rankStr)
	}
	return nil
}

