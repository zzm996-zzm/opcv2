import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { membershipApi, type MembershipSnapshot } from "../lib/membershipApi";

const errorMessages: Record<string, string> = {
  code_not_found: "兑换码不存在",
  code_expired: "兑换码已过期",
  code_exhausted: "兑换码已被领完",
  invalid_code: "请输入有效兑换码"
};

function MembershipPage() {
  const [snapshot, setSnapshot] = useState<MembershipSnapshot | null>(null);
  const [code, setCode] = useState("");
  const [status, setStatus] = useState<"loading" | "idle" | "submitting">("loading");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    membershipApi
      .current()
      .then((result) => {
        if (active) {
          setSnapshot(result);
          setStatus("idle");
        }
      })
      .catch(() => {
        if (active) {
          setError("暂时无法读取会员信息");
          setStatus("idle");
        }
      });
    return () => {
      active = false;
    };
  }, []);

  async function redeem(event: FormEvent) {
    event.preventDefault();
    if (!code.trim() || status === "submitting") return;
    setStatus("submitting");
    setError("");
    setMessage("");
    try {
      const result = await membershipApi.redeem(code);
      setSnapshot(result.snapshot);
      setMessage(result.already_redeemed ? "这个兑换码已经使用过，权益没有重复发放" : "兑换成功");
      setCode("");
    } catch (requestError) {
      setError(resolveError(requestError));
    } finally {
      setStatus("idle");
    }
  }

  return (
    <main className="membership-page">
      <header className="placeholder-header">
        <Link className="brand" to="/" aria-label="返回智活AI OPC 首页">
          <span className="brand-mark" aria-hidden="true" />
          <span>智活AI · OPC</span>
        </Link>
        <Link to="/">返回首页</Link>
      </header>

      <section className="membership-card">
        <p className="section-eyebrow">会员兑换</p>
        <h1>使用兑换码激活会员权益或兑换积分</h1>
        <div className="membership-grid">
          <article>
            <span>当前会员</span>
            <strong>{status === "loading" ? "读取中" : snapshot?.plan.name ?? "免费版"}</strong>
            <p>每月分析 {snapshot?.plan.monthly_analysis_limit ?? 0} 次</p>
          </article>
          <article>
            <span>当前积分</span>
            <strong>{snapshot?.credit_balance ?? 0}</strong>
            <p>后续 AI 分析和获客任务会从这里扣减</p>
          </article>
          <article>
            <span>获客额度</span>
            <strong>{snapshot?.plan.lead_export_limit ?? 0}</strong>
            <p>由服务端权益判断，前端只展示结果</p>
          </article>
        </div>

        <form className="redeem-card" onSubmit={redeem}>
          <div className="redeem-tabs" aria-hidden="true">
            <span>兑换会员/权益</span>
            <span>兑换积分</span>
          </div>
          <label className="field">
            <span>兑换码</span>
            <input
              aria-label="兑换码"
              onChange={(event) => setCode(event.target.value)}
              placeholder="输入运营发放的兑换码"
              value={code}
            />
          </label>
          <button disabled={!code.trim() || status === "submitting"} type="submit">
            {status === "submitting" ? "兑换中" : "兑换"}
          </button>
        </form>
        {message && <p className="form-success" role="status">{message}</p>}
        {error && <p className="form-error" role="alert">{error}</p>}
      </section>
    </main>
  );
}

function resolveError(error: unknown) {
  if (!(error instanceof Error)) return "暂时无法完成兑换";
  return errorMessages[error.message] ?? "暂时无法完成兑换";
}

export default MembershipPage;
