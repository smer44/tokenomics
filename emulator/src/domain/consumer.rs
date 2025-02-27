use super::common::*;

pub struct ConsumerId(String);

pub struct ConsumerRequest {
    ConsumerId: ConsumerId,
    Product: Product,
    Tokens: Tokens,
}

// type Consumer interface {
// 	Id() ConsumerId
// 	Order() []ConsumerRequest
// 	Emit(Tokens)
// }
