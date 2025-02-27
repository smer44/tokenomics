use std::collections::HashMap;

use super::common::*;

#[derive(PartialEq, Clone, Copy)]
enum PartStatus {
    Completed,
    Processing,
    Rejected,
    Unknown,
}

struct Part {
    capacity: Capacity,
    status: PartStatus,
}

#[derive(Clone)]
enum Investment {
    Restoration,
    Upgrade,
}

#[derive(Clone)]
struct CustomerRequest {
    customer_id: CustomerId,
    product: Product,
    tokens: Tokens,
}

#[derive(Clone)]
struct ProducerRequest {
    producer_id: ProducerId,
    product: Product,
    tokens: Tokens,
    cut_off_price: CapacityUnitPrice,
    investement_type: Investment,
}

#[derive(Clone)]
pub enum Request {
    FromCustomer(CustomerRequest),
    FromProducer(ProducerRequest),
}

#[derive(Hash, Clone, PartialEq, Eq)]
pub struct OrderId(String);
impl OrderId {
    fn new(val: &str) -> Self {
        return OrderId(String::from(val));
    }
}

pub struct Order {
    id: OrderId,
    tokens: Tokens,
    parts: HashMap<CapacityType, Part>,
    request: Request,
    cycle: u8,
}

pub struct OrderInfo {
    id: OrderId,
    tokens: Tokens,
    required: HashMap<CapacityType, Capacity>,
}

impl Order {
    pub fn new(id: OrderId, ps: &ProcessSheet, request: Request) -> Self {
        let parts = ps
            .require
            .iter()
            .map(|(t, capacity)| {
                (
                    *t,
                    Part {
                        capacity: *capacity,
                        status: PartStatus::Unknown,
                    },
                )
            })
            .collect();
        let tokens = match &request {
            Request::FromCustomer(r) => r.tokens,
            Request::FromProducer(r) => r.tokens,
        };
        Self {
            id,
            tokens,
            parts,
            request,
            cycle: 0,
        }
    }

    pub fn info(&self) -> OrderInfo {
        let required = self
            .parts
            .iter()
            .filter(|(_, part)| part.status != PartStatus::Processing && part.status != PartStatus::Completed)
            .map(|(k, v)| (*k, v.capacity))
            .collect();
        return OrderInfo {
            id: self.id.clone(),
            tokens: self.tokens,
            required: required,
        };
    }

    pub fn agent_id(&self) -> OrderingAgentId {
        match &self.request {
            Request::FromCustomer(r) => OrderingAgentId::from_customer(r.customer_id.clone()),
            Request::FromProducer(r) => OrderingAgentId::from_producer(r.producer_id.clone()),
        }
    }

    pub fn requires_funding(&self) -> bool {
        self.tokens.value() == 0
    }

    pub fn cut_off_price(&self) -> CapacityUnitPrice {
        if let Request::FromProducer(r) = &self.request {
            return r.cut_off_price;
        }
        panic!("not an investment order")
    }

    pub fn fund(&mut self, tokens: Tokens) {
        if !self.requires_funding() {
            panic!("not in funding status");
        }
        self.tokens = tokens;
    }

    fn get_part_status(&self, bid: &Bid) -> PartStatus {
        if bid.order_id != self.id {
            panic!("wrong order id");
        }
        self.parts.get(&bid.capacity_type).expect("ErrNotFound").status
    }

    pub fn mark_processing(&mut self, bid: &Bid) {
        match self.get_part_status(bid) {
            PartStatus::Processing => return,
            PartStatus::Completed => panic!("ErrWrongState"),
            _ => {}
        }
        self.tokens.add(bid.tokens);
        self.parts.get_mut(&bid.capacity_type).unwrap().status = PartStatus::Processing;
    }

    pub fn mark_completed(&mut self, bid: &Bid) {
        match self.get_part_status(bid) {
            PartStatus::Processing => {}
            PartStatus::Completed => panic!("ErrWrongState"),
            _ => self.tokens.sub(bid.tokens),
        }
        self.parts.get_mut(&bid.capacity_type).unwrap().status = PartStatus::Completed;
    }

    pub fn makr_rejected(&mut self, bid: &Bid) {
        match self.get_part_status(bid) {
            PartStatus::Processing | PartStatus::Completed => panic!("ErrWrongState"),
            _ => {}
        }
        self.parts.get_mut(&bid.capacity_type).unwrap().status = PartStatus::Rejected;
    }

    pub fn end_cycle(&mut self) -> (Score, OrderEvent) {
        let rejected_count = self.parts.values().filter(|p| p.status == PartStatus::Rejected).count();
        let completed_count = self.parts.values().filter(|p| p.status == PartStatus::Completed).count();
        if completed_count == self.parts.len() {
            let score = Score(0);
            return match &self.request {
                Request::FromCustomer(r) => (score, OrderEvent::CustomerRequestCompleted(r.clone())),
                Request::FromProducer(r) => (score, OrderEvent::ProducerRequestCompleted(r.clone())),
            };
        }
        if rejected_count == self.parts.len() || self.cycle == 3 {
            let score = Score(3);
            return match &self.request {
                Request::FromCustomer(r) => (score, OrderEvent::CustomerRequestRejected(self.tokens, r.clone())),
                Request::FromProducer(r) => (score, OrderEvent::ProducerRequestRejected(self.tokens, r.clone())),
            };
        }
        self.cycle += 1;
        (Score(1), OrderEvent::StillProcessing)
    }
}

pub enum OrderEvent {
    CustomerRequestCompleted(CustomerRequest),
    ProducerRequestCompleted(ProducerRequest),
    CustomerRequestRejected(Tokens, CustomerRequest),
    ProducerRequestRejected(Tokens, ProducerRequest),
    StillProcessing,
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn smoke() {
        let product = Product(1);
        let request = CustomerRequest {
            customer_id: CustomerId("a".to_owned()),
            product: product,
            tokens: Tokens::new(100),
        };
        let ps = ProcessSheet {
            product: product,
            require: HashMap::from([(CapacityType(1), Capacity(100))]),
        };
        let order_id = OrderId::new("abc");
        let mut order = Order::new(order_id.clone(), &ps, Request::FromCustomer(request.clone()));
        assert!(!order.requires_funding());
        let bid = Bid {
            capacity_type: CapacityType(1),
            capacity: Capacity(100),
            tokens: Tokens::new(100),
            order_id: order_id.clone(),
        };
        order.mark_completed(&bid);
        let (score, event) = order.end_cycle();
        assert_eq!(Score(0), score);
        assert!(matches!(event, 
                OrderEvent::CustomerRequestCompleted(CustomerRequest { customer_id, .. }) 
                if customer_id == request.customer_id));
    }
}
