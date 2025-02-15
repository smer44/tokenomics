package domain

type ConsumerId uint

type ProductRequest struct {
	Product Product
	Tokens  Tokens
}

type Consumer interface {
	Id() ConsumerId
	Order() []ProductRequest
	Emit(Tokens)
}
