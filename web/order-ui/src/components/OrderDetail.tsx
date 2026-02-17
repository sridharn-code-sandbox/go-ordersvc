import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import useSWR from "swr";
import { getOrder, updateOrderStatus, deleteOrder } from "../api/client";
import type { OrderStatus } from "../api/types";
import StatusBadge from "./StatusBadge";

const STATUSES: OrderStatus[] = [
  "pending",
  "confirmed",
  "processing",
  "shipped",
  "delivered",
  "cancelled",
];

export default function OrderDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: order, error, mutate } = useSWR(id ? `order-${id}` : null, () => getOrder(id!));
  const [newStatus, setNewStatus] = useState("");
  const [updating, setUpdating] = useState(false);

  if (error) return <p className="text-red-600">Failed to load order.</p>;
  if (!order) return <p className="text-gray-500">Loading...</p>;

  const handleStatusUpdate = async () => {
    if (!newStatus) return;
    setUpdating(true);
    try {
      await updateOrderStatus(order.id, { status: newStatus });
      await mutate();
      setNewStatus("");
    } catch (err) {
      alert(`Update failed: ${err}`);
    } finally {
      setUpdating(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this order?")) return;
    await deleteOrder(order.id);
    navigate("/");
  };

  return (
    <div>
      <h1 className="mb-4 text-xl font-semibold">Order {order.id.slice(0, 8)}...</h1>

      <div className="mb-6 grid grid-cols-2 gap-4 rounded border p-4 text-sm">
        <div>
          <span className="text-gray-500">Customer</span>
          <p>{order.customer_id}</p>
        </div>
        <div>
          <span className="text-gray-500">Status</span>
          <p><StatusBadge status={order.status} /></p>
        </div>
        <div>
          <span className="text-gray-500">Total</span>
          <p>${order.total.toFixed(2)}</p>
        </div>
        <div>
          <span className="text-gray-500">Version</span>
          <p>{order.version}</p>
        </div>
        <div>
          <span className="text-gray-500">Created</span>
          <p>{new Date(order.created_at).toLocaleString()}</p>
        </div>
        <div>
          <span className="text-gray-500">Updated</span>
          <p>{new Date(order.updated_at).toLocaleString()}</p>
        </div>
      </div>

      <h2 className="mb-2 text-lg font-medium">Items</h2>
      <table className="mb-6 w-full text-left text-sm">
        <thead className="border-b text-xs uppercase text-gray-500">
          <tr>
            <th className="py-2">Product</th>
            <th>Name</th>
            <th className="text-right">Qty</th>
            <th className="text-right">Price</th>
            <th className="text-right">Subtotal</th>
          </tr>
        </thead>
        <tbody>
          {order.items.map((item) => (
            <tr key={item.id} className="border-b">
              <td className="py-2">{item.product_id}</td>
              <td>{item.name}</td>
              <td className="text-right">{item.quantity}</td>
              <td className="text-right">${item.price.toFixed(2)}</td>
              <td className="text-right">${item.subtotal.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex items-center gap-3">
        <select
          value={newStatus}
          onChange={(e) => setNewStatus(e.target.value)}
          className="rounded border px-2 py-1 text-sm"
        >
          <option value="">Change status...</option>
          {STATUSES.filter((s) => s !== order.status).map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
        <button
          disabled={!newStatus || updating}
          onClick={handleStatusUpdate}
          className="rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700 disabled:opacity-40"
        >
          {updating ? "Updating..." : "Update"}
        </button>
        <button
          onClick={handleDelete}
          className="rounded bg-red-600 px-3 py-1 text-sm text-white hover:bg-red-700"
        >
          Delete
        </button>
      </div>
    </div>
  );
}
