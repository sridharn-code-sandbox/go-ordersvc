import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createOrder } from "../api/client";
import type { CreateOrderItem } from "../api/types";

const emptyItem = (): CreateOrderItem => ({
  product_id: "",
  name: "",
  quantity: 1,
  price: 0,
});

export default function CreateOrderForm() {
  const navigate = useNavigate();
  const [customerId, setCustomerId] = useState("");
  const [items, setItems] = useState<CreateOrderItem[]>([emptyItem()]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const updateItem = (idx: number, patch: Partial<CreateOrderItem>) => {
    setItems((prev) => prev.map((it, i) => (i === idx ? { ...it, ...patch } : it)));
  };

  const removeItem = (idx: number) => {
    setItems((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const order = await createOrder({ customer_id: customerId, items });
      navigate(`/orders/${order.id}`);
    } catch (err) {
      setError(String(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <h1 className="mb-4 text-xl font-semibold">Create Order</h1>
      {error && <p className="mb-4 text-red-600">{error}</p>}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium">Customer ID</label>
          <input
            required
            value={customerId}
            onChange={(e) => setCustomerId(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm"
            placeholder="e.g. cust-001"
          />
        </div>

        <div>
          <div className="mb-2 flex items-center justify-between">
            <label className="text-sm font-medium">Items</label>
            <button
              type="button"
              onClick={() => setItems((prev) => [...prev, emptyItem()])}
              className="text-sm text-blue-600 hover:underline"
            >
              + Add item
            </button>
          </div>

          {items.map((item, idx) => (
            <div key={idx} className="mb-2 grid grid-cols-[1fr_1fr_80px_100px_auto] gap-2">
              <input
                required
                placeholder="Product ID"
                value={item.product_id}
                onChange={(e) => updateItem(idx, { product_id: e.target.value })}
                className="rounded border px-2 py-1 text-sm"
              />
              <input
                required
                placeholder="Name"
                value={item.name}
                onChange={(e) => updateItem(idx, { name: e.target.value })}
                className="rounded border px-2 py-1 text-sm"
              />
              <input
                required
                type="number"
                min={1}
                placeholder="Qty"
                value={item.quantity}
                onChange={(e) => updateItem(idx, { quantity: Number(e.target.value) })}
                className="rounded border px-2 py-1 text-sm"
              />
              <input
                required
                type="number"
                min={0}
                step={0.01}
                placeholder="Price"
                value={item.price || ""}
                onChange={(e) => updateItem(idx, { price: Number(e.target.value) })}
                className="rounded border px-2 py-1 text-sm"
              />
              <button
                type="button"
                onClick={() => removeItem(idx)}
                disabled={items.length === 1}
                className="text-red-500 hover:text-red-700 disabled:opacity-30"
              >
                &times;
              </button>
            </div>
          ))}
        </div>

        <button
          type="submit"
          disabled={submitting}
          className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-40"
        >
          {submitting ? "Creating..." : "Create Order"}
        </button>
      </form>
    </div>
  );
}
