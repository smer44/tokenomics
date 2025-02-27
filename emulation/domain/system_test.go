package domain

import (
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
func (t *TestConsumer) Order() []ConsumerRequest {
	i := t.idx
	t.idx = (t.idx + 1) % len(t.products)
	return []ConsumerRequest{{t.id, t.products[i], t.tokens}}
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
	power1 := CapacityType(1)
	product1 := Product(1)
	processSheet1 := ProcessSheet{product1, map[CapacityType]Capacity{
		power1: 10,
	}}
	consumer1 := TestConsumer{id: "c1", products: []Product{product1}}

	t.Run(`Given the empty system
		When consumer orders a product
		And the power request is less than the producer's capacity
		Then the product is produced in a single cycle`, func(t *testing.T) {
		pac := ProducingAgentConfig{"p1", power1, 100, 0, Restoration{}, Upgrade{}}

		system := NewSystem(&TestIdGenerator{}, 100, []ProcessSheet{processSheet1}, []ProducingAgentConfig{pac}, map[ConsumerId]Consumer{"c1": &consumer1})
		// Investment
		pav, err := system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 100, 100, 0, 0, 0, 0, false, false}, pav)
		system.ProducingAgentAction("p1", ProducingAgentCommand{})
		// oav, err := system.OrderingAgentView("c1")
		// require.NoError(t, err)
		// require.Equal(t, OrderingAgentView{
		// 	Incoming: map[OrderId]Request{
		// 		"0": {50, []CapacityValue{{power1, 10}}},
		// 	},
		// 	InProgress: map[OrderId]Request{},
		// 	Producers:  map[CapacityType][]ProducerInfo{power1: {{"p1", power1, 100, 0}}},
		// }, oav)
		err = system.OrderingAgentAction("c1", OrderingAgentCommand{
			Orders: map[OrderId]map[ProducerId]Tokens{
				"1": {"p1": 100},
			}})
		require.NoError(t, err)
		scores, err := system.EndCycle()
		require.NoError(t, err)
		require.Equal(t, CycleResult{0}, scores)
		pav, err = system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 100, 100, 10, 0, 0, 0, false, false}, pav)
	})
}
