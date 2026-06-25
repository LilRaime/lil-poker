package deck

import (
	"crypto/rand"
	"lil-poker/internal/card"
	"math/big"
)

type Deck struct {
	cards []card.Card
}

func New() *Deck {
	d := &Deck{}
	for _, suit := range []card.Suit{card.Spades, card.Hearts, card.Diamonds, card.Clubs} {
		for rank := card.Two; rank <= card.Ace; rank++ {
			d.cards = append(d.cards, card.New(rank, suit))
		}
	}
	d.Shuffle()
	return d
}

func (d *Deck) Shuffle() {
	for i := len(d.cards) - 1; i > 0; i-- {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic("crypto/rand failure: " + err.Error())
		}
		j := int(nBig.Int64())
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
}

func (d *Deck) Deal(n int) []card.Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	dealt := d.cards[:n]
	d.cards = d.cards[n:]
	return dealt
}

func (d *Deck) Remaining() int {
	return len(d.cards)
}
