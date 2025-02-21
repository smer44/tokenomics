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
	power1 := PowerType(1)
	product1 := Product(1)
	processSheet1 := ProcessSheet{product1, map[PowerType]Power{
		power1: 10,
	}}
	consumer1 := TestConsumer{id: 1, products: []Product{product1}}

	t.Run(`Given the empty system
		When consumer orders a product
		And the power request is less than the producer's capacity
		Then the product is produced in a single cycle`, func(t *testing.T) {
		pac := ProducingAgentConfig{1, power1, 100, 0, Restoration{}, Upgrade{}}
		system := NewSystem(&TestIdGenerator{}, 100, []ProcessSheet{processSheet1}, []ProducingAgentConfig{pac}, []Consumer{&consumer1})
		oav, err := system.OrderingAgentView(1)
		require.NoError(t, err)
		require.Equal(t, OrderingAgentView{
			Incoming: map[OrderId]Request{
				"0": {50, []PowerValue{{power1, 10}}},
			},
			InProgress: map[OrderId]Request{},
			Producers:  map[PowerType][]ProducerInfo{power1: {{1, power1, 100, 0}}},
		}, oav)
		err = system.OrderingAgentAction(1, OrderingAgentCommand{
			Bids: map[ProducerId][]Bid{1: {Bid{
				Power:   10,
				Tokens:  100,
				OrderId: "1",
			}}},
		})
		require.NoError(t, err)
		require.NoError(t, system.FixBids())
		pav, err := system.ProducingAgentView(1)
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{
			1,
			100,
			10,
			0,
			0,
			0,
		}, pav)
		// system.ProducingAgentAction(1, ProducingAgentCommand{})
		// _, _ = system.EndTurn()

	})
}
