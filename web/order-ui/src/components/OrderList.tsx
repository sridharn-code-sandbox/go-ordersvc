import { useState } from "react";
import { Link } from "react-router-dom";
import useSWR from "swr";
import { listOrders } from "../api/client";
import type { OrderStatus } from "../api/types";
import StatusBadge from "./StatusBadge";
import Pagination from "./Pagination";

const PAGE_SIZE = 10;
const STATUSES: OrderStatus[] = [
  "pending",
  "confirmed",
  "processing",
  "shipped",
  "delivered",
  "cancelled",
];

export default function OrderList() {
  const [offset, setOffset] = useState(0);
  const [status, setStatus] = useState("");
  const [customerId, setCustomerId] = useState("");

  const { data, error, isLoading } = useSWR(
    ["orders", offset, status, customerId],
    () =>
      listOrders({
        limit: PAGE_SIZE,
        offset,
        status: status || undefined,
        customer_id: customerId || undefined,
      }),
  );

  return (
    <div>
      <h1 className="mb-4 text-xl font-semibold">Orders</h1>

      <div className="mb-4 flex gap-3">
        <select
          value={status}
          onChange={(e) => {
            setStatus(e.target.value);
            setOffset(0);
          }}
          className="rounded border px-2 py-1 text-sm"
        >
          <option value="">All statuses</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <input
          type="text"
          placeholder="Customer ID"
          value={customerId}
          onChange={(e) => {
            setCustomerId(e.target.value);
            setOffset(0);
          }}
          className="rounded border px-2 py-1 text-sm"
        />
      </div>

      {error && (
        <p className="text-red-600">Failed to load orders: {String(error)}</p>
      )}
      {isLoading && <p className="text-gray-500">Loading...</p>}

      {data && (
        <>
          <table className="w-full text-left text-sm">
            <thead className="border-b text-xs uppercase text-gray-500">
              <tr>
                <th className="py-2">ID</th>
                <th>Customer</th>
                <th>Status</th>
                <th className="text-right">Total</th>
                <th className="text-right">Created</th>
              </tr>
            </thead>
            <tbody>
              {data.orders.map((o) => (
                <tr key={o.id} className="border-b hover:bg-gray-50">
                  <td className="py-2">
                    <Link
                      to={`/orders/${o.id}`}
                      className="text-blue-600 hover:underline"
                    >
                      {o.id.slice(0, 8)}...
                    </Link>
                  </td>
                  <td>{o.customer_id}</td>
                  <td>
                    <StatusBadge status={o.status} />
                  </td>
                  <td className="text-right">${o.total.toFixed(2)}</td>
                  <td className="text-right">
                    {new Date(o.created_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
              {data.orders.length === 0 && (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-gray-400">
                    No orders found
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          <Pagination
            total={data.total}
            limit={data.limit}
            offset={data.offset}
            onChange={setOffset}
          />
        </>
      )}
    </div>
  );
}
