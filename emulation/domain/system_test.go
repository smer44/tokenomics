package domain

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	require.True(t, cmp.Equal(UndefinedPrice, UndefinedPrice))

	cpt1 := CapacityType("1")
	cpt2 := CapacityType("2")
	consumerProduct := Product(1)
	investmentProduct := Product(2)
	sheets := []ProcessSheet{
		{consumerProduct, map[CapacityType]Capacity{
			cpt1: 10,
		}},
		{investmentProduct, map[CapacityType]Capacity{
			cpt2: 200,
		}},
	}
	consumer1 := TestConsumer{id: "c1", products: []Product{consumerProduct}}
	pac1 := ProducingAgentConfig{"p1", cpt1, 100, 1, Restoration{}, Upgrade{investmentProduct, 50}}
	pac2 := ProducingAgentConfig{"p2", cpt2, 110, 1, Restoration{}, Upgrade{}}
	producerConfigs := []ProducingAgentConfig{pac1, pac2}

	t.Run(`Given the empty system
		When consumer orders a product
		And the power request is less than the producer's capacity
		Then the product is produced in a single cycle`, func(t *testing.T) {

		system := NewSystem(&TestIdGenerator{}, 100, sheets, producerConfigs, map[ConsumerId]Consumer{"c1": &consumer1})
		// Investment
		pav, err := system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 100, 100, 0, 1, 50, 0, false, false}, pav)
		err = system.ProducingAgentAction("p1", ProducingAgentCommand{})
		require.NoError(t, err)
		err = system.StartOrdering()
		require.NoError(t, err)
		oav, err := system.OrderingAgentView("c1")
		require.NoError(t, err)
		require.True(t, cmp.Equal(OrderingAgentView{
			Incoming: map[OrderId]map[CapacityType]Capacity{
				"0": {cpt1: 10},
			},
			Producers: map[CapacityType]map[ProducerId]ProducerInfo{
				cpt1: {"p1": ProducerInfo{"p1", cpt1, 100, 100, UndefinedPrice}},
			},
		}, oav))
		err = system.OrderingAgentAction("c1", OrderingAgentCommand{
			Orders: map[OrderId]map[ProducerId]Tokens{
				"0": {"p1": 50},
			}})
		require.NoError(t, err)
		scores, err := system.CompleteCycle()
		require.NoError(t, err)
		require.Equal(t, CycleResult{0}, scores)
		pav, err = system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 100, 99, 10, 1, 50, 0, false, false}, pav)
	})

	t.Run(`Given the empty system
		When producer orders an Upgrade
		And the capacity request is greater than the producer's capacity
		Then the Upgrade is produced in 2 cycles
		And capacity of the ordered producer is increased`, func(t *testing.T) {

		system := NewSystem(&TestIdGenerator{}, 100, sheets, producerConfigs, map[ConsumerId]Consumer{})
		// Investment
		pav, err := system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 100, 100, 0, 1, 50, 0, false, false}, pav)
		err = system.ProducingAgentAction("p1", ProducingAgentCommand{DoUpgrade: true})
		require.NoError(t, err)
		err = system.StartOrdering()
		require.NoError(t, err)
		oav, err := system.OrderingAgentView("p1")
		require.NoError(t, err)
		require.True(t, cmp.Equal(OrderingAgentView{
			Incoming: map[OrderId]map[CapacityType]Capacity{
				"0": {cpt2: 200},
			},
			Producers: map[CapacityType]map[ProducerId]ProducerInfo{
				cpt2: {"p2": ProducerInfo{"p2", cpt2, 110, 110, UndefinedPrice}},
			},
		}, oav))
		err = system.OrderingAgentAction("p1", OrderingAgentCommand{
			Orders: map[OrderId]map[ProducerId]Tokens{
				"0": {"p2": 50},
			}})
		require.NoError(t, err)
		scores, err := system.CompleteCycle()
		require.NoError(t, err)
		require.Equal(t, CycleResult{1}, scores)

		err = system.ProducingAgentAction("p1", ProducingAgentCommand{DoUpgrade: true})
		require.Error(t, err)

		err = system.StartOrdering()
		require.NoError(t, err)

		oav, err = system.OrderingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, OrderingAgentView{map[OrderId]map[CapacityType]Capacity{}, map[CapacityType]map[ProducerId]ProducerInfo{}}, oav)

		scores, err = system.CompleteCycle()
		require.NoError(t, err)
		require.Equal(t, CycleResult{Score(0)}, scores)

		pav, err = system.ProducingAgentView("p1")
		require.NoError(t, err)
		require.Equal(t, ProducingAgentView{"p1", 150, 148, 0, 2, 50, 0, false, false}, pav)
	})
}
