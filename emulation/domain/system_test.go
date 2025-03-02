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

type testConfig struct {
	config          *Configuration
	cpt1            CapacityType
	cpt2            CapacityType
	consumerProduct Product
}

func setupTestConfig() testConfig {
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

	pac1 := ProducingAgentConfig{"p1", cpt1, 100, 1, Restoration{}, Upgrade{investmentProduct, 50}}
	pac2 := ProducingAgentConfig{"p2", cpt2, 110, 1, Restoration{}, Upgrade{}}
	producerConfigs := []ProducingAgentConfig{pac1, pac2}

	return testConfig{
		config: &Configuration{
			CycleEmission:   100,
			ProcessSheets:   sheets,
			ProducerConfigs: producerConfigs,
		},
		cpt1:            cpt1,
		cpt2:            cpt2,
		consumerProduct: consumerProduct,
	}
}

func TestSystem(t *testing.T) {
	require.True(t, cmp.Equal(UndefinedPrice, UndefinedPrice))

	cfg := setupTestConfig()
	consumer1 := TestConsumer{id: "c1", products: []Product{cfg.consumerProduct}}

	t.Run(`Given the empty system
		When consumer orders a product
		And the power request is less than the producer's capacity
		Then the product is produced in a single cycle`, func(t *testing.T) {

		system := NewSystem(&TestIdGenerator{}, cfg.config, map[ConsumerId]Consumer{"c1": &consumer1})
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
				"0": {cfg.cpt1: 10},
			},
			Producers: map[CapacityType]map[ProducerId]ProducerInfo{
				cfg.cpt1: {"p1": ProducerInfo{"p1", cfg.cpt1, 100, 100, UndefinedPrice}},
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

		system := NewSystem(&TestIdGenerator{}, cfg.config, map[ConsumerId]Consumer{})
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
				"0": {cfg.cpt2: 200},
			},
			Producers: map[CapacityType]map[ProducerId]ProducerInfo{
				cfg.cpt2: {"p2": ProducerInfo{"p2", cfg.cpt2, 110, 110, UndefinedPrice}},
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
