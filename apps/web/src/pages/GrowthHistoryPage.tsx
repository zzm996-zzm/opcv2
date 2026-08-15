import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, RefreshCw, Search } from "lucide-react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { growthApi, type GrowthModelPage } from "../lib/growthApi";

const pageSize = 10;

type HistoryPeriod = "all" | "7" | "30" | "90";
type BusinessType = "all" | "saas" | "ecommerce" | "education" | "service" | "content" | "other";
type RiskLevel = "all" | "low" | "medium" | "high";

function pageFromQuery(value: string | null) {
  const page = Number(value);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 0
  }).format(value);
}

function formatPercent(value: number) {
  return `${(value * 100).toFixed(1)}%`;
}

function dateFromPeriod(period: HistoryPeriod) {
  if (period === "all") return undefined;
  const date = new Date();
  date.setHours(0, 0, 0, 0);
  date.setDate(date.getDate() - Number(period) + 1);
  return date.toISOString().slice(0, 10);
}

function filterValue<T extends string>(value: string | null, values: readonly T[], fallback: T) {
  return value && values.includes(value as T) ? value as T : fallback;
}

const businessTypeLabels: Record<Exclude<BusinessType, "all">, string> = {
  saas: "SaaS/软件",
  ecommerce: "电商/零售",
  education: "教育/培训",
  service: "咨询/服务",
  content: "内容/社群",
  other: "其他"
};

const riskLabels: Record<Exclude<RiskLevel, "all">, string> = {
  low: "低风险",
  medium: "中风险",
  high: "高风险"
};

function GrowthHistoryPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q")?.trim() ?? "";
  const currentPage = pageFromQuery(searchParams.get("page"));
  const periodValue = searchParams.get("period") ?? "all";
  const period = filterValue(periodValue, ["all", "7", "30", "90"] as const, "all");
  const businessType = filterValue(searchParams.get("business_type"), ["all", "saas", "ecommerce", "education", "service", "content", "other"] as const, "all");
  const risk = filterValue(searchParams.get("risk"), ["all", "low", "medium", "high"] as const, "all");
  const status = searchParams.get("status") === "completed" ? "completed" : "all";
  const sort = filterValue(searchParams.get("sort"), ["created_desc", "created_asc", "revenue_desc", "revenue_asc", "margin_desc", "margin_asc"] as const, "created_desc");
  const [searchInput, setSearchInput] = useState(query);
  const [page, setPage] = useState<GrowthModelPage>({ models: [], total: 0, limit: pageSize, offset: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [recalculatingID, setRecalculatingID] = useState<number | null>(null);
  const [selectedModelIDs, setSelectedModelIDs] = useState<number[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    growthApi.listModels({
      q: query || undefined,
      businessType: businessType === "all" ? undefined : businessType,
      status: status === "all" ? undefined : status,
      risk: risk === "all" ? undefined : risk,
      sort: sort === "created_desc" ? undefined : sort,
      from: dateFromPeriod(period),
      limit: pageSize,
      offset: (currentPage - 1) * pageSize
    })
      .then((result) => {
        if (!active) return;
        setPage({
          models: result.models ?? [],
          total: result.total ?? result.models?.length ?? 0,
          limit: result.limit ?? pageSize,
          offset: result.offset ?? (currentPage - 1) * pageSize
        });
        setSelectedModelIDs((current) => current.filter((id) => (result.models ?? []).some((model) => model.id === id)));
      })
      .catch((requestError) => {
        if (!active) return;
        setPage({ models: [], total: 0, limit: pageSize, offset: 0 });
        setError(apiErrorMessage(requestError, "暂时无法读取测算历史"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [businessType, currentPage, period, query, risk, sort, status]);

  function applySearch() {
    const next = new URLSearchParams();
    const normalized = searchInput.trim();
    if (normalized) next.set("q", normalized);
    if (period !== "all") next.set("period", period);
    if (businessType !== "all") next.set("business_type", businessType);
    if (status !== "all") next.set("status", status);
    if (risk !== "all") next.set("risk", risk);
    if (sort !== "created_desc") next.set("sort", sort);
    setSearchParams(next);
  }

  function changePeriod(nextPeriod: string) {
    const next = new URLSearchParams(searchParams);
    if (nextPeriod === "all") next.delete("period");
    else next.set("period", nextPeriod);
    next.delete("page");
    setSearchParams(next);
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    if (nextPage <= 1) next.delete("page");
    else next.set("page", String(nextPage));
    setSearchParams(next);
  }

  function changeFilter(key: "business_type" | "status" | "risk" | "sort", value: string) {
    const next = new URLSearchParams(searchParams);
    if (value === "all" || (key === "sort" && value === "created_desc")) next.delete(key);
    else next.set(key, value);
    next.delete("page");
    setSearchParams(next);
  }

  async function recalculate(modelID: number) {
    if (recalculatingID !== null) return;
    setRecalculatingID(modelID);
    setError("");
    try {
      const result = await growthApi.recalculateModel(modelID);
      navigate(`/growth-calculator/report?model_id=${result.model.id}`);
    } catch (requestError) {
      setError(apiErrorMessage(requestError, "再次测算失败，请稍后重试"));
    } finally {
      setRecalculatingID(null);
    }
  }

  function toggleModel(modelID: number) {
    setSelectedModelIDs((current) => {
      if (current.includes(modelID)) return current.filter((id) => id !== modelID);
      if (current.length >= 4) return current;
      return [...current, modelID];
    });
  }

  function compareSelected() {
    if (selectedModelIDs.length < 2 || selectedModelIDs.length > 4) return;
    navigate(`/growth-calculator/compare?model_ids=${selectedModelIDs.join(",")}`);
  }

  const pageCount = Math.max(1, Math.ceil(page.total / pageSize));

  return (
    <V4PageShell className="growth-calculator-shell">
      <section className="module-page growth-history-page" aria-label="增长测算历史">
        <header className="growth-history-heading">
          <div>
            <span className="module-kicker">增长测算</span>
            <h1>测算历史</h1>
            <p>查看已保存的模型结果，打开指定报告或沿用参数再次测算。</p>
          </div>
          <Link className="module-primary-action" to="/growth-calculator">新建测算</Link>
        </header>

        <form className="growth-history-toolbar" onSubmit={(event) => {
          event.preventDefault();
          applySearch();
        }}>
          <label>
            <Search aria-hidden="true" size={17} />
            <input
              aria-label="搜索测算历史"
              maxLength={100}
              onChange={(event) => setSearchInput(event.target.value)}
              placeholder="搜索测算名称"
              value={searchInput}
            />
          </label>
          <button className="primary" type="submit">查询</button>
          <button onClick={() => {
            setSearchInput("");
            setSearchParams({});
          }} type="button">
            <RefreshCw aria-hidden="true" size={16} />重置
          </button>
          <label className="growth-history-period-filter">
            <span>时间范围</span>
            <select aria-label="历史时间范围" onChange={(event) => changePeriod(event.target.value)} value={period}>
              <option value="all">全部</option>
              <option value="7">近 7 天</option>
              <option value="30">近 30 天</option>
              <option value="90">近 90 天</option>
            </select>
          </label>
          <label className="growth-history-period-filter">
            <span>业务类型</span>
            <select aria-label="业务类型筛选" onChange={(event) => changeFilter("business_type", event.target.value)} value={businessType}>
              <option value="all">全部</option>
              {Object.entries(businessTypeLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select>
          </label>
          <label className="growth-history-period-filter">
            <span>风险等级</span>
            <select aria-label="风险等级筛选" onChange={(event) => changeFilter("risk", event.target.value)} value={risk}>
              <option value="all">全部</option>
              {Object.entries(riskLabels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
            </select>
          </label>
          <label className="growth-history-period-filter">
            <span>状态</span>
            <select aria-label="模型状态筛选" onChange={(event) => changeFilter("status", event.target.value)} value={status}>
              <option value="all">全部</option>
              <option value="completed">已完成</option>
            </select>
          </label>
          <label className="growth-history-period-filter">
            <span>排序</span>
            <select aria-label="测算结果排序" onChange={(event) => changeFilter("sort", event.target.value)} value={sort}>
              <option value="created_desc">最新创建</option>
              <option value="created_asc">最早创建</option>
              <option value="revenue_desc">收入从高到低</option>
              <option value="revenue_asc">收入从低到高</option>
              <option value="margin_desc">利润率从高到低</option>
              <option value="margin_asc">利润率从低到高</option>
            </select>
          </label>
          <button className="primary" disabled={loading || selectedModelIDs.length < 2} onClick={compareSelected} type="button">
            对比测算{selectedModelIDs.length > 0 ? ` (${selectedModelIDs.length}/4)` : ""}
          </button>
        </form>

        {error ? <p className="form-error" role="alert">{error}</p> : null}
        {loading ? <div className="module-empty-state" role="status">正在读取测算历史...</div> : null}
        {!loading && !error && page.models.length === 0 ? (
          <div className="module-empty-state" role="status">{query ? "没有匹配的测算记录" : "暂无测算记录"}</div>
        ) : null}

        {!loading && page.models.length > 0 ? (
          <section className="growth-live-history-table" aria-label="测算历史列表">
            <header>
              <span>选择</span>
              <span>测算名称</span>
              <span>创建时间</span>
              <span>核心输入</span>
              <span>预计月收入</span>
              <span>净利润率</span>
              <span>状态</span>
              <span>操作</span>
            </header>
            {page.models.map((model) => (
              <article key={model.id}>
                <label className="growth-history-select">
                  <input
                    aria-label={`选择${model.name}`}
                    checked={selectedModelIDs.includes(model.id)}
                    onChange={() => toggleModel(model.id)}
                    type="checkbox"
                  />
                </label>
                <span><strong>{model.name}</strong><small>模型 #{model.id}</small></span>
                <time>{new Date(model.created_at).toLocaleString("zh-CN")}</time>
                <span>{model.assumptions.monthly_visits.toLocaleString("zh-CN")} 访问 · {formatPercent(model.assumptions.deal_rate)} 成交</span>
                <strong>{formatCurrency(model.result.monthly_revenue)}</strong>
                <span>{formatPercent(model.result.net_margin)}</span>
                <em>{model.status === "completed" ? "已完成" : model.status || "未知状态"} · {riskLabels[model.risk_level ?? "low"]}</em>
                <span className="growth-history-actions">
                  <Link to={`/growth-calculator/report?model_id=${model.id}`}>查看报告</Link>
                  <button disabled={recalculatingID !== null} onClick={() => void recalculate(model.id)} type="button">
                    {recalculatingID === model.id ? "测算中..." : "再次测算"}
                  </button>
                </span>
              </article>
            ))}
          </section>
        ) : null}

        <footer className="growth-history-pagination">
          <span>共 {page.total} 条</span>
          <div>
            <button aria-label="上一页" disabled={loading || currentPage <= 1} onClick={() => goToPage(currentPage - 1)} type="button"><ArrowLeft size={16} /></button>
            <strong>{currentPage} / {pageCount}</strong>
            <button aria-label="下一页" disabled={loading || currentPage >= pageCount} onClick={() => goToPage(currentPage + 1)} type="button"><ArrowRight size={16} /></button>
          </div>
        </footer>
      </section>
    </V4PageShell>
  );
}

export default GrowthHistoryPage;
