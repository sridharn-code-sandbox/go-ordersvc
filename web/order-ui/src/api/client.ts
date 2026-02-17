import type {
  Order,
  ListOrdersResponse,
  CreateOrderRequest,
  UpdateStatusRequest,
} from "./types";

const BASE = "/api/v1/orders";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`${res.status}: ${body}`);
  }
  return res.json() as Promise<T>;
}

export function listOrders(params?: {
  status?: string;
  customer_id?: string;
  limit?: number;
  offset?: number;
}): Promise<ListOrdersResponse> {
  const q = new URLSearchParams();
  if (params?.status) q.set("status", params.status);
  if (params?.customer_id) q.set("customer_id", params.customer_id);
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.offset) q.set("offset", String(params.offset));
  const qs = q.toString();
  return request<ListOrdersResponse>(`${BASE}${qs ? `?${qs}` : ""}`);
}

export function getOrder(id: string): Promise<Order> {
  return request<Order>(`${BASE}/${id}`);
}

export function createOrder(data: CreateOrderRequest): Promise<Order> {
  return request<Order>(BASE, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function updateOrderStatus(
  id: string,
  data: UpdateStatusRequest,
): Promise<Order> {
  return request<Order>(`${BASE}/${id}/status`, {
    method: "PATCH",
    body: JSON.stringify(data),
  });
}

export function deleteOrder(id: string): Promise<void> {
  return fetch(`${BASE}/${id}`, { method: "DELETE" }).then((res) => {
    if (!res.ok) throw new Error(`${res.status}`);
  });
}
