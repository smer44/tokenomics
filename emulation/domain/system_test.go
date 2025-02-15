package domain

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestConsumer struct {
	id       ConsumerId
	tokens   Tokens
	products []Product
	idx      int
}

// Emit implements Consumer.
func (t *TestConsumer) Emit(val Tokens) {
	t.tokens = val
}

// Id implements Consumer.
func (t *TestConsumer) Id() ConsumerId {
	return t.id
}

// Order implements Consumer.
func (t *TestConsumer) Order() []ProductRequest {
	fmt.Println("ORDER")
	i := t.idx
	t.idx = (t.idx + 1) % len(t.products)
	return []ProductRequest{{t.products[i], t.tokens}}
}

var _ Consumer = &TestConsumer{}

type TestIdGenerator struct {
	val int
}

// New implements OrderIdGenerator.
func (t *TestIdGenerator) New() OrderId {
	id := OrderId(strconv.Itoa(t.val))
	t.val++
	return id
}

var _ OrderIdGenerator = &TestIdGenerator{}

func TestSystem(t *testing.T) {
	pow1 := PowerType(1)
	prod1 := Product(1)
	ps := ProcessSheet{prod1, map[PowerType]Power{
		pow1: 10,
	}}
	consumer := TestConsumer{id: 1, products: []Product{prod1}}
	pac := ProducingAgentConfig{1, pow1, 100, 0, Restoration{}, Upgrade{}}
	system := NewSystem(&TestIdGenerator{}, 100, []ProcessSheet{ps}, []ProducingAgentConfig{pac}, []Consumer{&consumer})
	oav, err := system.OrderingAgentView(1)
	require.NoError(t, err)
	require.Equal(t, OrderingAgentView{
		Incoming: map[OrderId]Request{
			"0": {50, []PowerValue{{pow1, 10}}},
		},
		InProgress: map[OrderId]Request{},
	}, oav)
	// err = system.OrderingAgentAction(1, OrderingAgentCommand{})
	// require.NoError(t, err)
	// system.FixBids()
	// _, _ = system.ProducingAgentView(1)
	// system.ProducingAgentAction(1, ProducingAgentCommand{})
	// _, _ = system.EndTurn()
}
