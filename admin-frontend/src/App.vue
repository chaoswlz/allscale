<template>
  <main class="page">
    <section v-if="!isLoggedIn" class="login-card">
      <h1>Admin Login</h1>
      <form class="form" @submit.prevent="submit">
        <label class="field">
          <span>Username</span>
          <input v-model.trim="username" type="text" placeholder="Enter username" required />
        </label>
        <label class="field">
          <span>Password</span>
          <input
            v-model="password"
            type="password"
            placeholder="Enter password"
            required
          />
        </label>
        <button class="button" type="submit" :disabled="loading">
          {{ loading ? "Signing in..." : "Sign in" }}
        </button>
      </form>
      <p class="hint">Demo login: <strong>admin</strong> / <strong>demo</strong></p>
      <p v-if="message" class="message" :class="{ error: isError }">{{ message }}</p>
    </section>

    <section v-else-if="viewMode === 'detail'" class="detail-page">
      <header class="detail-topbar">
        <button class="back-link" type="button" @click="backToDashboard">
          ← Back to Operations List
        </button>
        <button class="ops-button" type="button">OPS VIEW</button>
      </header>

      <section class="detail-header-card">
        <div class="detail-left">
          <div class="detail-title">
            <h1>Payout Order</h1>
            <span class="pill">{{ selectedOrder.status }}</span>
          </div>
          <div class="detail-meta-row">
            <span class="order-code">{{ selectedOrder.orderId }}</span>
            <span class="muted">|</span>
            <span class="muted">Created: {{ selectedOrder.createdAt }}</span>
          </div>
        </div>
        <div class="detail-right">
          <div class="detail-customer">
            <span class="muted">CUSTOMER</span>
            <strong>{{ selectedOrder.customer }}</strong>
          </div>
        </div>
      </section>

      <section class="detail-card">
        <header class="detail-card-header">
          <div class="detail-section-title">CUSTOMER SUBMISSION</div>
          <span class="chip">Read Only</span>
        </header>
        <div class="detail-body">
          <div class="detail-grid">
            <div>
              <div class="label">NETWORK / ASSET</div>
              <div class="value">
                <span class="badge">{{ selectedOrder.network }}</span>
                {{ selectedOrder.asset }}
              </div>
            </div>
            <div>
              <div class="label">DECLARED AMOUNT</div>
              <div class="value">{{ selectedOrder.amount }}</div>
            </div>
          </div>
          <div class="detail-row">
            <div class="label">TRANSACTION HASH</div>
            <div class="value hash">
              <span class="hash-text">{{ selectedOrder.txid }}</span>
              <div class="icon-row">
                <button class="icon-button small" type="button" @click="copyTxid">📋</button>
                <button class="icon-button small" type="button" @click="openExplorer">↗</button>
              </div>
            </div>
          </div>

          <div class="detail-subsection">
            <div class="detail-section-title">BENEFICIARY DETAILS</div>
            <button class="link" type="button" @click="toggleReveal">
              {{ revealDetails ? "Hide Details" : "Reveal Details" }}
            </button>
          </div>
          <div class="detail-grid">
            <div>
              <div class="label">BENEFICIARY NAME</div>
              <div class="value">
                {{ selectedOrder.beneficiaryName }}
              </div>
            </div>
            <div>
              <div class="label">BANK COUNTRY</div>
              <div class="value">
                {{ selectedOrder.bankCountry }}
              </div>
            </div>
            <div>
              <div class="label">ACCOUNT NUMBER</div>
              <div class="value">
                {{ revealDetails ? selectedOrder.accountNumber : "**** ****" }}
              </div>
            </div>
            <div>
              <div class="label">SWIFT / ROUTING</div>
              <div class="value">
                {{ revealDetails ? selectedOrder.swift : "** **** ****" }}
              </div>
            </div>
            <div>
              <div class="label">BANK NAME</div>
              <div class="value">
                {{ selectedOrder.bankName }}
              </div>
            </div>
            <div>
              <div class="label">CUSTOMER NOTE</div>
              <div class="value note">
                {{ selectedOrder.note }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="detail-card">
        <header class="detail-card-header">
          <div class="detail-section-title">MANUAL STATUS UPDATE</div>
        </header>
        <div class="detail-body">
          <div class="status-panel">
            <div>
              <div class="label">CURRENT STATUS</div>
              <span class="pill">{{ selectedOrder.status }}</span>
            </div>
          </div>
          <p class="note-text">
            <strong>Note:</strong> Update status only after confirming external steps (e.g. bank
            transfer). These updates are visible to the customer.
          </p>
          <div class="status-actions">
            <button class="status-action processing" type="button" @click="updateOrderStatus('Processing')">
              Mark Processing <span>→</span>
            </button>
            <button class="status-action request" type="button" @click="updateOrderStatus('Summitted')">
              Request Info <span>→</span>
            </button>
            <button class="status-action completed" type="button" @click="updateOrderStatus('Paid')">
              Mark Completed <span>→</span>
            </button>
            <button class="status-action failed" type="button" @click="updateOrderStatus('Failed')">
              Mark Failed <span>→</span>
            </button>
          </div>
        </div>
      </section>
    </section>

    <section v-else class="dashboard">
      <header class="topbar">
        <div class="brand">
          <div class="logo">
            <span>⌗</span>
          </div>
          <h1>Payout Operations</h1>
        </div>
        <div class="topbar-actions">
          <div class="search">
            <span class="icon">🔍</span>
            <input placeholder="Search orders..." />
          </div>
          <button class="icon-button" type="button" @click="refreshData">↻</button>
          <div class="avatar-group">
            <button class="avatar-button" type="button" @click="toggleMenu">
              <div class="avatar">JD</div>
            </button>
            <div v-if="menuOpen" class="menu">
              <button class="menu-item" type="button" @click="logout">Logout</button>
            </div>
          </div>
        </div>
      </header>

      <section class="stats">
        <div v-for="card in statCards" :key="card.label" class="stat-card" :class="card.tone">
          <div class="stat-header">
            <span class="stat-label">{{ card.label }}</span>
            <span class="stat-icon">{{ card.icon }}</span>
          </div>
          <div class="stat-value">{{ card.value }}</div>
        </div>
      </section>

      <p v-if="apiError" class="message error">{{ apiError }}</p>

      <section class="table-card">
        <header class="table-header">
          <div class="section-title">
            <span class="dot"></span>
            Ready for Processing
            <span class="count">{{ readyTotal }}</span>
          </div>
        </header>
        <table class="table">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>User</th>
              <th>Asset</th>
              <th>Amount</th>
              <th>Time Received</th>
              <th class="align-right">Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in processingRows" :key="row.id" class="row" @click="openDetail(row)">
              <td>{{ row.id }}</td>
              <td>{{ row.user }}</td>
              <td>
                <strong>{{ row.asset }}</strong>
                <span class="muted asset-space">{{ row.network }}</span>
              </td>
              <td>{{ row.amount }}</td>
              <td class="muted">{{ row.time }}</td>
              <td class="align-right"><button class="link" type="button">OPEN</button></td>
            </tr>
          </tbody>
        </table>
        <div class="table-pagination">
          <button class="page-btn" type="button" :disabled="readyPage === 1" @click="prevReadyPage">
            Previous
          </button>
          <span>Page {{ readyPage }}</span>
          <button
            class="page-btn"
            type="button"
            :disabled="readyPage * readyPageSize >= readyTotal"
            @click="nextReadyPage"
          >
            Next
          </button>
        </div>
      </section>

      <section class="table-card">
        <header class="table-header">
          <div class="section-title">Recent Orders</div>
          <button class="filter-button" type="button">Filter</button>
        </header>
        <table class="table">
          <thead>
            <tr>
              <th>Order ID</th>
              <th>Status</th>
              <th>User</th>
              <th>Network</th>
              <th>Amount</th>
              <th>Last Update</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in recentRows" :key="row.id" class="row" @click="openDetail(row)">
              <td>{{ row.id }}</td>
              <td><span class="status" :class="row.statusTone">{{ row.status }}</span></td>
              <td>{{ row.user }}</td>
              <td class="muted">{{ row.network }}</td>
              <td>{{ row.amount }} <span class="muted">{{ row.asset }}</span></td>
              <td class="muted">{{ row.update }}</td>
            </tr>
          </tbody>
        </table>
        <div class="table-pagination">
          <button class="page-btn" type="button" :disabled="recentPage === 1" @click="prevRecentPage">
            Previous
          </button>
          <span>Page {{ recentPage }}</span>
          <button
            class="page-btn"
            type="button"
            :disabled="recentPage * recentPageSize >= recentTotal"
            @click="nextRecentPage"
          >
            Next
          </button>
        </div>
      </section>
    </section>
  </main>
</template>

<script setup>
import { computed, ref } from "vue";

const API_BASE = import.meta.env.VITE_API_BASE || "http://localhost:8080";

const username = ref("");
const password = ref("");
const loading = ref(false);
const message = ref("");
const isError = ref(false);
const token = ref(localStorage.getItem("admin_token") || "");

const isLoggedIn = computed(() => Boolean(token.value));
const apiError = ref("");
const readyPage = ref(1);
const readyPageSize = 5;
const readyTotal = ref(0);
const recentPage = ref(1);
const recentPageSize = 5;
const recentTotal = ref(0);
const menuOpen = ref(false);
const viewMode = ref("dashboard");
const selectedOrder = ref(null);
const revealDetails = ref(false);

const submit = async () => {
  message.value = "";
  isError.value = false;
  token.value = "";
  loading.value = true;
  try {
    if (username.value === "admin" && password.value === "demo") {
      token.value = "demo-admin-token";
      localStorage.setItem("admin_token", token.value);
      message.value = "Login successful";
      loadDemoData();
      return;
    }
    const resp = await fetch(`${API_BASE}/admin/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username: username.value,
        password: password.value
      })
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data?.error || "Login failed");
    }
    token.value = data.token || "";
    if (token.value) {
      localStorage.setItem("admin_token", token.value);
    }
    message.value = "Login successful";
    await loadAdminData();
  } catch (err) {
    isError.value = true;
    message.value = err?.message || "Login failed";
  } finally {
    loading.value = false;
  }
};

const logout = () => {
  token.value = "";
  localStorage.removeItem("admin_token");
  menuOpen.value = false;
  viewMode.value = "dashboard";
  selectedOrder.value = null;
};

const toggleMenu = () => {
  menuOpen.value = !menuOpen.value;
};

const refreshData = async () => {
  await loadAdminData();
};

const openDetail = async (row) => {
  apiError.value = "";
  selectedOrder.value = {
    orderId: row.id,
    status: row.status || "-",
    createdAt: row.time || row.update || "-",
    customer: row.user || "-",
    network: row.network || "-",
    asset: row.asset || "-",
    amount: row.amount || "-",
    txid: row.txid || "-",
    beneficiaryName: row.beneficiaryName || "-",
    bankCountry: row.bankCountry || "-",
    bankName: row.bankName || "-",
    swift: row.swift || "zhiheng chas",
    accountNumber: row.accountNumber || "54528892",
    note: row.note || "-"
  };
  viewMode.value = "detail";
  revealDetails.value = false;

  if (!token.value) {
    return;
  }

  if (token.value === "demo-admin-token") {
    selectedOrder.value = {
      orderId: row.id,
      status: row.status || "Funds Received",
      createdAt: "2023-10-24 10:30:00",
      customer: row.user || "Enterprise Corp",
      network: row.network || "Ethereum",
      asset: row.asset || "USDT",
      amount: row.amount || "12,500.00",
      txid: "0x7a23c...9b21",
      beneficiaryName: "TechVentures LLC",
      bankCountry: "United States",
      bankName: "Chase Bank International",
      swift: "zhiheng chas",
      accountNumber: "54528892",
      note: "\"Please process ASAP, urgent vendor payment.\""
    };
    return;
  }

  try {
    const resp = await fetch(`${API_BASE}/admin/order?id=${row.id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data?.error || "Failed to load order");
    }
    const order = data.order || data;
    selectedOrder.value = mapAdminOrderDetail(order);
  } catch (err) {
    apiError.value = err?.message || "Failed to load order";
  }
};

const backToDashboard = () => {
  viewMode.value = "dashboard";
  selectedOrder.value = null;
  revealDetails.value = false;
};

const toggleReveal = () => {
  revealDetails.value = !revealDetails.value;
};

const copyTxid = async () => {
  const txid = selectedOrder.value?.txid || "";
  if (!txid) {
    return;
  }
  try {
    await navigator.clipboard.writeText(txid);
  } catch {
    const fallback = document.createElement("textarea");
    fallback.value = txid;
    fallback.style.position = "fixed";
    fallback.style.opacity = "0";
    document.body.appendChild(fallback);
    fallback.select();
    document.execCommand("copy");
    document.body.removeChild(fallback);
  }
};

const openExplorer = () => {
  const txid = selectedOrder.value?.txid || "";
  if (!txid) {
    return;
  }
  const network = (selectedOrder.value?.network || "").toLowerCase();
  let baseUrl = "https://etherscan.io/tx/";
  if (network.includes("tron")) {
    baseUrl = "https://tronscan.org/index.html#/transaction/";
  } else if (network.includes("bsc") || network.includes("binance")) {
    baseUrl = "https://bscscan.com/tx/";
  }
  window.open(`${baseUrl}${txid}`, "_blank", "noopener");
};

const maskText = (value) => {
  if (!value) {
    return "••••";
  }
  return "••••••";
};

const formatDateTime = (value) => {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};

const mapAdminOrderDetail = (order) => ({
  orderId: order.order_id ?? order.id ?? "-",
  status: order.status || "-",
  createdAt: formatDateTime(order.created_at),
  customer: order.merchant_name || "-",
  network: order.transaction_network || order.network || "-",
  asset: order.transaction_asset || order.asset || "-",
  amount: order.amount != null ? Number(order.amount).toLocaleString() : "-",
  txid: order.txid || "-",
  beneficiaryName: order.beneficiary_name || "-",
  bankCountry: order.bank_country || "-",
  bankName: order.bank_name || "-",
  swift: order.swift || "zhiheng chas",
  accountNumber: order.iban || "54528892",
  note: order.reference_note || "-"
});

const syncRowStatus = (status) => {
  const id = selectedOrder.value?.orderId;
  if (!id) {
    return;
  }
  processingRows.value = processingRows.value.filter(
    (row) => row.id !== id || status === "Processing"
  );
  recentRows.value = recentRows.value.map((row) =>
    row.id === id ? { ...row, status, statusTone: statusTone(status) } : row
  );
};

const updateOrderStatus = async (status) => {
  apiError.value = "";
  if (!selectedOrder.value) {
    return;
  }

  if (token.value === "demo-admin-token") {
    selectedOrder.value = { ...selectedOrder.value, status };
    syncRowStatus(status);
    return;
  }

  try {
    const resp = await fetch(`${API_BASE}/admin/order/status`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token.value}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        id: selectedOrder.value.orderId,
        status
      })
    });
    const data = await resp.json();
    if (!resp.ok) {
      throw new Error(data?.error || "Failed to update status");
    }
    const order = data.order || data;
    selectedOrder.value = mapAdminOrderDetail(order);
    syncRowStatus(selectedOrder.value.status);
  } catch (err) {
    apiError.value = err?.message || "Failed to update status";
  }
};

const statCards = ref([
  { label: "FUNDS RECEIVED", value: 0, icon: "◔", tone: "info" },
  { label: "PROCESSING", value: 0, icon: "↗", tone: "default" },
  { label: "ACTION REQUIRED", value: 0, icon: "!", tone: "warn" },
  { label: "AWAITING", value: 0, icon: "◷", tone: "default" },
  { label: "COMPLETED TODAY", value: 0, icon: "✓", tone: "success" }
]);

const processingRows = ref([]);

const recentRows = ref([]);

const loadAdminData = async () => {
  apiError.value = "";
  if (!token.value) {
    return;
  }

  try {
    const [statsResp, readyResp, recentResp] = await Promise.all([
      fetch(`${API_BASE}/admin/stats`, {
        headers: { Authorization: `Bearer ${token.value}` }
      }),
      fetch(
        `${API_BASE}/admin/ready-processing?page=${readyPage.value}&page_size=${readyPageSize}`,
        { headers: { Authorization: `Bearer ${token.value}` } }
      ),
      fetch(
        `${API_BASE}/admin/recent-orders?page=${recentPage.value}&page_size=${recentPageSize}`,
        { headers: { Authorization: `Bearer ${token.value}` } }
      )
    ]);

    const statsData = await statsResp.json();
    const readyData = await readyResp.json();
    const recentData = await recentResp.json();

    if (!statsResp.ok || !readyResp.ok || !recentResp.ok) {
      throw new Error(statsData?.error || readyData?.error || recentData?.error || "Load failed");
    }

    statCards.value = [
      { label: "FUNDS RECEIVED", value: statsData.funds_received, icon: "◔", tone: "info" },
      { label: "PROCESSING", value: statsData.processing, icon: "↗", tone: "default" },
      { label: "ACTION REQUIRED", value: statsData.action_required, icon: "!", tone: "warn" },
      { label: "AWAITING", value: statsData.awaiting, icon: "◷", tone: "default" },
      { label: "COMPLETED TODAY", value: statsData.completed_today, icon: "✓", tone: "success" }
    ];

    readyTotal.value = readyData.total || 0;
    processingRows.value = (readyData.items || []).map((row) => ({
      id: row.order_id,
      user: row.merchant_name,
      asset: row.asset,
      network: row.network,
      amount: row.amount ? Number(row.amount).toLocaleString() : "-",
      time: formatRelativeTime(row.time_received)
    }));

    recentTotal.value = recentData.total || 0;
    recentRows.value = (recentData.items || []).map((row) => ({
      id: row.order_id,
      status: row.status,
      statusTone: statusTone(row.status),
      user: row.merchant_name,
      network: row.network,
      amount: row.amount ? Number(row.amount).toLocaleString() : "-",
      asset: row.asset,
      update: formatRelativeTime(row.last_update)
    }));
  } catch (err) {
    apiError.value = err?.message || "Failed to load data";
    loadDemoData();
  }
};

const loadDemoData = () => {
  statCards.value = [
    { label: "FUNDS RECEIVED", value: 3, icon: "◔", tone: "info" },
    { label: "PROCESSING", value: 2, icon: "↗", tone: "default" },
    { label: "ACTION REQUIRED", value: 2, icon: "!", tone: "warn" },
    { label: "AWAITING", value: 1, icon: "◷", tone: "default" },
    { label: "COMPLETED TODAY", value: 4, icon: "✓", tone: "success" }
  ];

  readyTotal.value = 3;
  processingRows.value = [
    {
      id: "PYT-9921-AK1",
      user: "Enterprise Corp",
      asset: "USDT",
      network: "Tron",
      amount: "12,500.00",
      time: "2 mins ago"
    },
    {
      id: "PYT-3321-PL0",
      user: "John Doe",
      asset: "ETH",
      network: "Arbitrum",
      amount: "1,200.00",
      time: "5 mins ago"
    },
    {
      id: "PYT-9988-MN5",
      user: "Liam Neeson",
      asset: "USDT",
      network: "Tron",
      amount: "3,300.00",
      time: "Just now"
    }
  ];

  recentTotal.value = 2;
  recentRows.value = [
    {
      id: "PYT-9921-AK1",
      status: "FUNDS RECEIVED",
      statusTone: "received",
      user: "Enterprise Corp",
      network: "Tron",
      amount: "12,500.00",
      asset: "USDT",
      update: "2 mins ago"
    },
    {
      id: "PYT-8829-XJ2",
      status: "PROCESSING",
      statusTone: "processing",
      user: "Alice Freeman",
      network: "Ethereum",
      amount: "4,250.00",
      asset: "USDC",
      update: "15 mins ago"
    }
  ];
};

const formatRelativeTime = (value) => {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  const diff = Date.now() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) {
    return "Just now";
  }
  if (minutes < 60) {
    return `${minutes} mins ago`;
  }
  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours} hrs ago`;
  }
  const days = Math.floor(hours / 24);
  return `${days} days ago`;
};

const statusTone = (status) => {
  if (status === "Funds Received") {
    return "received";
  }
  if (status === "Processing") {
    return "processing";
  }
  return "processing";
};

const nextReadyPage = async () => {
  if (readyPage.value * readyPageSize >= readyTotal.value) {
    return;
  }
  readyPage.value += 1;
  await loadAdminData();
};

const prevReadyPage = async () => {
  if (readyPage.value === 1) {
    return;
  }
  readyPage.value -= 1;
  await loadAdminData();
};

const nextRecentPage = async () => {
  if (recentPage.value * recentPageSize >= recentTotal.value) {
    return;
  }
  recentPage.value += 1;
  await loadAdminData();
};

const prevRecentPage = async () => {
  if (recentPage.value === 1) {
    return;
  }
  recentPage.value -= 1;
  await loadAdminData();
};
</script>

<style scoped>
.page {
  font-family: "Inter", Arial, Helvetica, sans-serif;
  min-height: 100vh;
  background: #f5f7fb;
  padding: 32px;
}

.login-card {
  width: 360px;
  margin: 120px auto 0;
  background: #ffffff;
  border-radius: 14px;
  padding: 24px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.1);
}

.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
  color: #334155;
  font-weight: 600;
}

.field input {
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 14px;
  background: #ffffff;
  color: #0f172a;
  box-sizing: border-box;
}

.button {
  background: #111827;
  border: none;
  color: #ffffff;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
}

.button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message {
  margin-top: 12px;
  font-size: 14px;
  color: #2563eb;
}

.message.error {
  color: #dc2626;
}

.hint {
  margin-top: 12px;
  font-size: 13px;
  color: #64748b;
}

.dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 32px;
  border-bottom: 1px solid #e7edf4;
  background: #ffffff;
  margin: -32px -32px 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo {
  width: 36px;
  height: 36px;
  background: #eef2ff;
  color: #4f46e5;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 700;
}

.brand h1 {
  margin: 0;
  font-size: 20px;
  color: #0f172a;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #2b2f36;
  color: #e5e7eb;
  padding: 8px 12px;
  border-radius: 12px;
}

.search input {
  border: none;
  background: transparent;
  color: #f8fafc;
  outline: none;
  width: 200px;
}

.icon-button {
  border: none;
  background: #f8fafc;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.08);
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  background: #e6effe;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #1d4ed8;
}

.avatar-group {
  position: relative;
  display: flex;
  align-items: center;
}

.avatar-button {
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
}

.menu {
  position: absolute;
  right: 0;
  top: 46px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 10px 20px rgba(15, 23, 42, 0.12);
  padding: 6px;
  min-width: 120px;
  z-index: 5;
}

.menu-item {
  width: 100%;
  text-align: left;
  border: none;
  background: transparent;
  color: #0f172a;
  padding: 8px 10px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.menu-item:hover {
  background: #f1f5f9;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.stat-card {
  background: #ffffff;
  border-radius: 14px;
  padding: 14px 16px;
  border: 1px solid #e7edf4;
  box-shadow: 0 6px 14px rgba(15, 23, 42, 0.04);
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 700;
}

.stat-value {
  font-size: 24px;
  color: #0f172a;
  margin-top: 6px;
  font-weight: 700;
}

.stat-card.warn {
  border-color: #fde68a;
  box-shadow: 0 0 0 2px #fef3c7 inset;
}

.stat-card.success {
  border-color: #bbf7d0;
  box-shadow: 0 0 0 2px #ecfdf3 inset;
}

.table-card {
  background: #ffffff;
  border-radius: 16px;
  border: 1px solid #e7edf4;
  padding: 0 0 8px;
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #eef2f7;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  color: #0f172a;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #a5b4fc;
}

.count {
  background: #eef2ff;
  color: #4f46e5;
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 12px;
}

.filter-button {
  border: none;
  background: transparent;
  color: #64748b;
  font-weight: 600;
  cursor: pointer;
}

.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.table th,
.table td {
  padding: 14px 20px;
  border-bottom: 1px solid #eef2f7;
  text-align: left;
}

.table th {
  color: #64748b;
  font-weight: 600;
}

.table td {
  color: #0f172a;
}

.align-right {
  text-align: right;
}

.muted {
  color: #6b7280;
}

.asset-space {
  margin-left: 6px;
}

.link {
  border: none;
  background: transparent;
  color: #0ea5e9;
  font-weight: 700;
  cursor: pointer;
}

.status {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.status.received {
  background: #eef2ff;
  color: #4338ca;
}

.status.processing {
  background: #e0f2fe;
  color: #0369a1;
}

.table-pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 12px 20px 16px;
  color: #64748b;
  font-size: 13px;
}

.page-btn {
  border: 1px solid #e2e8f0;
  background: #ffffff;
  color: #0f172a;
  padding: 6px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.row {
  cursor: pointer;
}

.row:hover {
  background: #f8fafc;
}

.detail-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-topbar {
  background: #0f172a;
  color: #cbd5f5;
  padding: 14px 20px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.back-link {
  background: transparent;
  border: none;
  color: #cbd5f5;
  font-weight: 600;
  cursor: pointer;
}

.ops-button {
  border: 1px solid #1e293b;
  background: #0b1220;
  color: #e2e8f0;
  padding: 6px 10px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.detail-header-card {
  background: #ffffff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.detail-left {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  text-align: right;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-title h1 {
  margin: 0;
  font-size: 22px;
  color: #0f172a;
}

.pill {
  padding: 4px 10px;
  border-radius: 999px;
  background: #eef2ff;
  color: #4338ca;
  font-size: 12px;
  font-weight: 700;
}

.order-code {
  background: #f1f5f9;
  color: #475569;
  padding: 6px 10px;
  border-radius: 10px;
  font-weight: 700;
}

.detail-customer {
  color: #0f172a;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-meta-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 0 0;
  color: #64748b;
  font-weight: 600;
}

.detail-card {
  background: #ffffff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

.detail-card-header {
  padding: 14px 20px;
  border-bottom: 1px solid #eef2f7;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chip {
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #94a3b8;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.detail-body {
  padding: 16px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.status-panel {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 16px 18px;
  display: flex;
  justify-content: center;
}

.note-text {
  color: #64748b;
  font-size: 13px;
  margin: 0;
}

.status-actions {
  display: grid;
  gap: 10px;
}

.status-action {
  width: 100%;
  border: none;
  border-radius: 12px;
  padding: 12px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 700;
  color: #ffffff;
  cursor: pointer;
}

.status-action.processing {
  background: #2563eb;
}

.status-action.request {
  background: #f59e0b;
}

.status-action.completed {
  background: #059669;
}

.status-action.failed {
  background: #dc2626;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.label {
  font-size: 12px;
  color: #94a3b8;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  font-weight: 700;
}

.value {
  font-size: 15px;
  color: #0f172a;
  font-weight: 600;
  margin-top: 6px;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 999px;
  background: #eef2ff;
  color: #4338ca;
  font-size: 12px;
  font-weight: 700;
  margin-right: 8px;
}

.hash {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  padding: 10px 12px;
  border-radius: 10px;
  font-family: "SFMono-Regular", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
    monospace;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.hash-text {
  overflow: hidden;
  text-overflow: ellipsis;
}

.icon-row {
  display: flex;
  gap: 8px;
}

.icon-button.small {
  width: 32px;
  height: 32px;
  border-radius: 8px;
}

.detail-subsection {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 8px;
  border-top: 1px solid #eef2f7;
}

.detail-section-title {
  color: #0f172a;
  font-weight: 700;
  font-size: 13px;
  letter-spacing: 0.04em;
}

.note {
  color: #475569;
  font-style: italic;
}
</style>
