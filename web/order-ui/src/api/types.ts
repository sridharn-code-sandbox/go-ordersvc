export interface OrderItem {
  id: string;
  product_id: string;
  name: string;
  quantity: number;
  price: number;
  subtotal: number;
}

export interface Order {
  id: string;
  customer_id: string;
  items: OrderItem[];
  status: OrderStatus;
  total: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled";

export interface ListOrdersResponse {
  orders: Order[];
  total: number;
  limit: number;
  offset: number;
}

export interface CreateOrderRequest {
  customer_id: string;
  items: CreateOrderItem[];
}

export interface CreateOrderItem {
  product_id: string;
  name: string;
  quantity: number;
  price: number;
}

export interface UpdateStatusRequest {
  status: string;
}
