import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, RefreshCw, Search } from "lucide-react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { growthApi, type GrowthModelPage } from "../lib/growthApi";

const pageSize = 10;

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

function GrowthHistoryPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q")?.trim() ?? "";
  const currentPage = pageFromQuery(searchParams.get("page"));
  const [searchInput, setSearchInput] = useState(query);
  const [page, setPage] = useState<GrowthModelPage>({ models: [], total: 0, limit: pageSize, offset: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [recalculatingID, setRecalculatingID] = useState<number | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    growthApi.listModels({ q: query || undefined, limit: pageSize, offset: (currentPage - 1) * pageSize })
      .then((result) => {
        if (!active) return;
        setPage({
          models: result.models ?? [],
          total: result.total ?? result.models?.length ?? 0,
          limit: result.limit ?? pageSize,
          offset: result.offset ?? (currentPage - 1) * pageSize
        });
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
  }, [currentPage, query]);

  function applySearch() {
    const next = new URLSearchParams();
    const normalized = searchInput.trim();
    if (normalized) next.set("q", normalized);
    setSearchParams(next);
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    if (nextPage <= 1) next.delete("page");
    else next.set("page", String(nextPage));
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
        </form>

        {error ? <p className="form-error" role="alert">{error}</p> : null}
        {loading ? <div className="module-empty-state" role="status">正在读取测算历史...</div> : null}
        {!loading && !error && page.models.length === 0 ? (
          <div className="module-empty-state" role="status">{query ? "没有匹配的测算记录" : "暂无测算记录"}</div>
        ) : null}

        {!loading && page.models.length > 0 ? (
          <section className="growth-live-history-table" aria-label="测算历史列表">
            <header>
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
                <span><strong>{model.name}</strong><small>模型 #{model.id}</small></span>
                <time>{new Date(model.created_at).toLocaleString("zh-CN")}</time>
                <span>{model.assumptions.monthly_visits.toLocaleString("zh-CN")} 访问 · {formatPercent(model.assumptions.deal_rate)} 成交</span>
                <strong>{formatCurrency(model.result.monthly_revenue)}</strong>
                <span>{formatPercent(model.result.net_margin)}</span>
                <em>已完成</em>
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
