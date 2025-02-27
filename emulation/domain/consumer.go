package domain

type ConsumerId string

type ConsumerRequest struct {
	ConsumerId ConsumerId
	Product    Product
	Tokens     Tokens
}

type Consumer interface {
	Id() ConsumerId
	Order() []ConsumerRequest
	Emit(Tokens)
}
