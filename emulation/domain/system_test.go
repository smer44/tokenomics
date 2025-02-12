package domain

import "testing"

func TestOrderingAgent(t *testing.T) {
	p1 := Product(1)
	p2 := Product(2)
	productSheet := ProcessSheet{p1, map[Product]Power{
		p2: Power{p2, 10},
	}}

	t.Run(`Given an empty ordering agent
		When new order is placed
		And State is called
		Then returns agent state`, func(t *testing.T) {
		agent := NewOrderingAgent()
		order := NewOrder("1", 10, &productSheet)
		agent.Place(order)
		state := agent.State()
		require.True(t, ))
		
	}
}
