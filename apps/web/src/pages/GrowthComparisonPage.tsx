import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { apiErrorMessage } from "../lib/apiErrors";
import { growthApi, type GrowthModel } from "../lib/growthApi";

function parseModelIDs(value: string | null) {
  const ids = (value ?? "")
    .split(",")
    .map((item) => Number(item))
    .filter((id) => Number.isInteger(id) && id > 0);
  return [...new Set(ids)];
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

function GrowthComparisonPage() {
  const [searchParams] = useSearchParams();
  const modelIDs = useMemo(() => parseModelIDs(searchParams.get("model_ids")), [searchParams]);
  const modelIDsKey = modelIDs.join(",");
  const [models, setModels] = useState<GrowthModel[]>([]);
  const [generatedAt, setGeneratedAt] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    if (modelIDs.length < 2 || modelIDs.length > 4) {
      setModels([]);
      setGeneratedAt("");
      setError("请选择 2 到 4 个测算模型进行对比");
      setLoading(false);
      return () => {
        active = false;
      };
    }
    setLoading(true);
    setError("");
    growthApi.compareModels(modelIDs)
      .then((result) => {
        if (!active) return;
        setModels(result.models ?? []);
        setGeneratedAt(result.generated_at ?? "");
      })
      .catch((requestError) => {
        if (!active) return;
        setModels([]);
        setGeneratedAt("");
        setError(apiErrorMessage(requestError, "暂时无法读取测算对比"));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [modelIDsKey, modelIDs]);

  return (
    <V4PageShell className="growth-calculator-shell">
      <section className="module-page growth-comparison-page" aria-label="增长测算对比">
        <header className="growth-comparison-heading">
          <div>
            <span className="module-kicker">增长测算</span>
            <h1>测算对比</h1>
            <p>并列查看历史版本的收入、利润、回本周期和关键假设。</p>
          </div>
          <Link className="growth-comparison-back" to="/growth-calculator/history">返回测算历史</Link>
        </header>

        {loading ? <div className="module-empty-state" role="status">正在生成测算对比...</div> : null}
        {error ? <p className="form-error" role="alert">{error}</p> : null}
        {!loading && !error && models.length === 0 ? <div className="module-empty-state" role="status">暂无可对比模型</div> : null}

        {!loading && !error && models.length > 0 ? (
          <section className="growth-comparison-table-wrap" aria-label="测算对比表">
            <table className="growth-comparison-table">
              <thead>
                <tr>
                  <th scope="col">指标</th>
                  {models.map((model) => <th key={model.id} scope="col"><strong>{model.name}</strong><small>模型 #{model.id}</small></th>)}
                </tr>
              </thead>
              <tbody>
                <tr><th scope="row">预计月收入</th>{models.map((model) => <td key={model.id}>{formatCurrency(model.result.monthly_revenue)}</td>)}</tr>
                <tr><th scope="row">净利润率</th>{models.map((model) => <td key={model.id}>{formatPercent(model.result.net_margin)}</td>)}</tr>
                <tr><th scope="row">回本周期</th>{models.map((model) => <td key={model.id}>{model.result.payback_days} 天</td>)}</tr>
                <tr><th scope="row">月访问量</th>{models.map((model) => <td key={model.id}>{model.assumptions.monthly_visits.toLocaleString("zh-CN")}</td>)}</tr>
                <tr><th scope="row">线索转化率</th>{models.map((model) => <td key={model.id}>{formatPercent(model.assumptions.lead_rate)}</td>)}</tr>
                <tr><th scope="row">成交转化率</th>{models.map((model) => <td key={model.id}>{formatPercent(model.assumptions.deal_rate)}</td>)}</tr>
                <tr><th scope="row">平均客单价</th>{models.map((model) => <td key={model.id}>{formatCurrency(model.assumptions.average_order)}</td>)}</tr>
                <tr><th scope="row">获客成本</th>{models.map((model) => <td key={model.id}>{formatCurrency(model.assumptions.acquisition_cost)}</td>)}</tr>
                <tr><th scope="row">交付成本</th>{models.map((model) => <td key={model.id}>{formatCurrency(model.assumptions.delivery_cost)}</td>)}</tr>
              </tbody>
            </table>
            {generatedAt ? <p className="growth-comparison-meta">对比生成时间：{new Date(generatedAt).toLocaleString("zh-CN")}</p> : null}
          </section>
        ) : null}
      </section>
    </V4PageShell>
  );
}

export default GrowthComparisonPage;
