use std::{collections::HashMap, i32, u32};

use super::order::OrderId;

#[derive(Clone, Copy)]
pub struct Tokens(i32);
impl Tokens {
    pub fn new(val: i32) -> Self {
        Tokens(val)
    }
    pub fn value(&self) -> i32 {
        self.0
    }

    pub fn add(&mut self, tokens: Tokens) {
        self.0 += tokens.0;
    }
    pub fn sub(&mut self, tokens: Tokens) {
        self.0 -= tokens.0;
    }
}

#[derive(Clone, Copy, PartialEq)]
pub struct Product(pub u32);
#[derive(Clone, Copy, Hash, PartialEq, Eq)]
pub struct CapacityType(pub u32);
#[derive(Clone, Copy, PartialEq)]
pub struct Capacity(pub i32);
#[derive(Clone, PartialEq)]
pub struct CustomerId(pub String);
#[derive(Clone, PartialEq)]
pub struct ProducerId(pub String);
#[derive(Clone, Copy, PartialEq)]
pub struct CapacityUnitPrice(pub f32);
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Score(pub u32);

#[derive(Hash, Clone, PartialEq, Eq)]
pub struct OrderingAgentId(String);
impl OrderingAgentId {
    pub fn from_producer(id: ProducerId) -> Self {
        OrderingAgentId(format!("p{}", id.0))
    }
    pub fn from_customer(id: CustomerId) -> Self {
        OrderingAgentId(format!("c{}", id.0))
    }
}

pub struct ProcessSheet {
    pub product: Product,
    pub require: HashMap<CapacityType, Capacity>,
}

#[derive(Clone)]
pub struct Bid {
    pub capacity_type: CapacityType,
    pub capacity: Capacity,
    pub tokens: Tokens,
    pub order_id: OrderId,
}
