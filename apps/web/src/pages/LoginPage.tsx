import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { authApi } from "../lib/authApi";
import { authSession } from "../lib/authSession";

const phonePattern = /^1[3-9]\d{9}$/;

const errorMessages: Record<string, string> = {
  invalid_phone: "请输入正确的中国大陆手机号",
  invalid_code: "验证码错误或已过期",
  code_rate_limited: "发送太频繁，请稍后再试",
  agreement_required: "请先同意用户协议和隐私政策",
  nickname_required: "请输入昵称"
};

function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [nickname, setNickname] = useState("");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [agreementAccepted, setAgreementAccepted] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [status, setStatus] = useState<"idle" | "sending" | "submitting" | "success">(
    "idle"
  );
  const [error, setError] = useState("");

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = window.setInterval(() => {
      setCountdown((current) => Math.max(0, current - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [countdown]);

  const canSubmit = useMemo(
    () =>
      nickname.trim().length > 0 &&
      phonePattern.test(phone) &&
      /^\d{6}$/.test(code) &&
      agreementAccepted &&
      status !== "submitting",
    [agreementAccepted, code, nickname, phone, status]
  );

  async function sendCode() {
    if (!phonePattern.test(phone) || countdown > 0) return;
    setError("");
    setStatus("sending");
    try {
      await authApi.sendCode(phone);
      setCountdown(60);
      setStatus("idle");
    } catch (requestError) {
      setStatus("idle");
      setError(resolveError(requestError));
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!canSubmit) return;
    setError("");
    setStatus("submitting");
    try {
      const result = await authApi.login({
        nickname: nickname.trim(),
        phone,
        code,
        agreementAccepted
      });
      authSession.set(result);
      setStatus("success");
      const returnTo = (location.state as { returnTo?: string } | null)?.returnTo;
      window.setTimeout(() => navigate(returnTo ?? "/"), 600);
    } catch (requestError) {
      setStatus("idle");
      setError(resolveError(requestError));
    }
  }

  return (
    <main className="login-page">
      <section className="login-story" aria-label="产品价值">
        <Link className="brand login-brand" to="/" aria-label="智活AI OPC 首页">
          <span className="brand-mark">智</span>
          <span>智活AI</span>
        </Link>
        <div className="login-story-copy">
          <p className="section-eyebrow">AI 商业行动伙伴</p>
          <h1>先找到方向，再找到客户。</h1>
          <p>
            从一句话开始，补齐关键信息，获得可执行的商业分析和真实企业线索。
          </p>
        </div>
        <dl className="trust-grid">
          <div>
            <dt>3 个</dt>
            <dd>清晰方向建议</dd>
          </div>
          <div>
            <dt>7 天</dt>
            <dd>具体行动计划</dd>
          </div>
          <div>
            <dt>公开来源</dt>
            <dd>企业信息可追溯</dd>
          </div>
          <div>
            <dt>持续跟进</dt>
            <dd>线索进入 CRM</dd>
          </div>
        </dl>
      </section>

      <section className="login-panel">
        <div className="login-card">
          <p className="section-eyebrow">欢迎开始</p>
          <h2>登录或创建账号</h2>
          <p className="login-subtitle">首次登录会自动创建账号，无需设置密码。</p>

          <form onSubmit={submit}>
            <label className="field">
              <span>昵称</span>
              <input
                aria-label="昵称"
                autoComplete="nickname"
                maxLength={50}
                onChange={(event) => setNickname(event.target.value)}
                placeholder="怎么称呼你"
                value={nickname}
              />
            </label>

            <label className="field">
              <span>手机号</span>
              <input
                aria-label="手机号"
                autoComplete="tel"
                inputMode="numeric"
                maxLength={11}
                onChange={(event) => setPhone(event.target.value.replace(/\D/g, ""))}
                placeholder="请输入手机号"
                value={phone}
              />
            </label>

            <label className="field">
              <span>验证码</span>
              <div className="code-control">
                <input
                  aria-label="验证码"
                  autoComplete="one-time-code"
                  inputMode="numeric"
                  maxLength={6}
                  onChange={(event) => setCode(event.target.value.replace(/\D/g, ""))}
                  placeholder="6 位验证码"
                  value={code}
                />
                <button
                  disabled={!phonePattern.test(phone) || countdown > 0 || status === "sending"}
                  onClick={sendCode}
                  type="button"
                >
                  {countdown > 0 ? `${countdown}s` : status === "sending" ? "发送中" : "获取验证码"}
                </button>
              </div>
            </label>

            <label className="agreement-row">
              <input
                aria-label="同意用户协议和隐私政策"
                checked={agreementAccepted}
                onChange={(event) => setAgreementAccepted(event.target.checked)}
                type="checkbox"
              />
              <span>
                我已阅读并同意 <Link to="/terms">《用户协议》</Link> 和{" "}
                <Link to="/privacy">《隐私政策》</Link>
              </span>
            </label>

            {error && <p className="form-error" role="alert">{error}</p>}
            {status === "success" && (
              <p className="form-success" role="status">登录成功，正在进入工作台</p>
            )}

            <button className="login-submit" disabled={!canSubmit} type="submit">
              {status === "submitting" ? "正在登录" : "开始使用"}
            </button>
          </form>
        </div>
      </section>
    </main>
  );
}

function resolveError(error: unknown) {
  if (!(error instanceof Error)) return "暂时无法完成操作，请稍后重试";
  return errorMessages[error.message] ?? "暂时无法完成操作，请稍后重试";
}

export default LoginPage;
