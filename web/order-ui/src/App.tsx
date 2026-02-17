import { Routes, Route, Link } from "react-router-dom";
import OrderList from "./components/OrderList";
import OrderDetail from "./components/OrderDetail";
import CreateOrderForm from "./components/CreateOrderForm";

export default function App() {
  return (
    <div className="min-h-screen">
      <nav className="bg-white shadow">
        <div className="mx-auto max-w-5xl px-4 py-3 flex items-center gap-6">
          <Link to="/" className="text-lg font-semibold text-gray-800">
            Order Service
          </Link>
          <Link
            to="/orders/new"
            className="rounded bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700"
          >
            New Order
          </Link>
        </div>
      </nav>
      <main className="mx-auto max-w-5xl px-4 py-6">
        <Routes>
          <Route path="/" element={<OrderList />} />
          <Route path="/orders/new" element={<CreateOrderForm />} />
          <Route path="/orders/:id" element={<OrderDetail />} />
        </Routes>
      </main>
    </div>
  );
}
