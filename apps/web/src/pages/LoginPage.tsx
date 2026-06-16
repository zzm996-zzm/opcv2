import { FormEvent, useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { authApi } from "../lib/authApi";
import { authSession } from "../lib/authSession";

const phonePattern = /^1[3-9]\d{9}$/;
type AuthMode = "login" | "register";

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
  const [mode, setMode] = useState<AuthMode>("login");
  const [account, setAccount] = useState("");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [email, setEmail] = useState("");
  const [wechat, setWechat] = useState("");
  const [agreementAccepted, setAgreementAccepted] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [status, setStatus] = useState<"idle" | "sending" | "submitting" | "success">("idle");
  const [error, setError] = useState("");

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = window.setInterval(() => {
      setCountdown((current) => Math.max(0, current - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [countdown]);

  const canSubmit = useMemo(() => {
    if (!agreementAccepted || status === "submitting") return false;
    if (mode === "register") {
      return (
        account.trim().length >= 4 &&
        password.trim().length >= 4 &&
        (!confirmPassword || confirmPassword === password) &&
        phonePattern.test(phone)
      );
    }
    return phonePattern.test(phone) && code.trim().length >= 4;
  }, [account, agreementAccepted, code, confirmPassword, mode, password, phone, status]);

  async function sendCode() {
    if (!phonePattern.test(phone) || countdown > 0 || status === "sending") return;
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
        nickname: mode === "register" ? account.trim() : `智活用户${phone.slice(-4)}`,
        phone,
        code: mode === "register" ? password.trim() : code.trim(),
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
      <Link className="auth-help" to="/help">
        <span aria-hidden="true">?</span>
        帮助中心
      </Link>

      <section className="login-story" aria-label="产品价值">
        <Link className="auth-brand" to="/" aria-label="智活AI OPC V4.0 首页">
          <span className="v4-logo" aria-hidden="true" />
          <strong>智活AI</strong>
          <small>OPC V4.0</small>
        </Link>

        <div className="login-story-copy">
          <p>企业增长智能引擎</p>
          <h1>让每个决策更智能<br />让每次增长更确定</h1>
          <span>智活AI 助力企业打通数据、洞察与执行，驱动可持续增长</span>
        </div>

        <div className="login-orbit" aria-hidden="true">
          <span className="orbit-line one" />
          <span className="orbit-line two" />
          <span className="orbit-dot one" />
          <span className="orbit-dot two" />
          <span className="brand-sculpture">
            <span className="v4-logo" />
          </span>
          <span className="login-cube a" />
          <span className="login-cube b" />
          <span className="login-cube c" />
        </div>

        <div className="auth-feature-list">
          <article>
            <i className="metric-icon chart" aria-hidden="true" />
            <div>
              <strong>全域数据智能洞察</strong>
              <small>整合多源数据，洞察业务关键机会</small>
            </div>
          </article>
          <article>
            <i className="metric-icon trend" aria-hidden="true" />
            <div>
              <strong>AI赋能高效决策</strong>
              <small>智能分析与预测，辅助科学决策</small>
            </div>
          </article>
          <article>
            <i className="metric-icon inbox" aria-hidden="true" />
            <div>
              <strong>业务闭环持续增长</strong>
              <small>从洞察到执行，沉淀增长方法论</small>
            </div>
          </article>
        </div>

        <p className="auth-copyright">© 2024 智活AI 版权所有 | 京ICP备2023001234号-1</p>
      </section>

      <section className="login-panel" aria-label="登录注册">
        <div className="login-card">
          <div className="auth-tabs" role="tablist" aria-label="登录注册切换">
            <button
              aria-selected={mode === "login"}
              onClick={() => {
                setMode("login");
                setError("");
              }}
              role="tab"
              type="button"
            >
              登录
            </button>
            <button
              aria-selected={mode === "register"}
              onClick={() => {
                setMode("register");
                setError("");
              }}
              role="tab"
              type="button"
            >
              注册
            </button>
          </div>

          <form onSubmit={submit}>
            {mode === "login" ? (
              <>
                <label className="field auth-input">
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

                <label className="field auth-input">
                  <span>验证码</span>
                  <div className="code-control">
                    <input
                      aria-label="验证码"
                      autoComplete="one-time-code"
                      inputMode="numeric"
                      maxLength={8}
                      onChange={(event) => setCode(event.target.value.replace(/\D/g, ""))}
                      placeholder="请输入验证码"
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
              </>
            ) : (
              <div className="register-field-grid">
                <label className="field auth-input">
                  <span>账号 <em>*</em></span>
                  <input
                    aria-label="账号"
                    autoComplete="username"
                    maxLength={20}
                    onChange={(event) => setAccount(event.target.value)}
                    placeholder="请输入账号，支持字母、数字、下划线，4-20位"
                    value={account}
                  />
                </label>
                <label className="field auth-input">
                  <span>密码 <em>*</em></span>
                  <input
                    aria-label="密码"
                    autoComplete="new-password"
                    onChange={(event) => setPassword(event.target.value)}
                    placeholder="请输入密码，8-20位，需包含字母和数字"
                    type="password"
                    value={password}
                  />
                </label>
                <label className="field auth-input">
                  <span>确认密码 <small>（选填）</small></span>
                  <input
                    aria-label="确认密码"
                    autoComplete="new-password"
                    onChange={(event) => setConfirmPassword(event.target.value)}
                    placeholder="请再次输入密码"
                    type="password"
                    value={confirmPassword}
                  />
                </label>
                <label className="field auth-input">
                  <span>邮箱 <small>（选填）</small></span>
                  <input
                    aria-label="邮箱"
                    autoComplete="email"
                    onChange={(event) => setEmail(event.target.value)}
                    placeholder="请输入常用邮箱地址"
                    type="email"
                    value={email}
                  />
                </label>
                <label className="field auth-input">
                  <span>手机号 <em>*</em></span>
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
                <label className="field auth-input">
                  <span>微信 / 企业微信 <small>（选填）</small></span>
                  <input
                    aria-label="微信或企业微信"
                    autoComplete="off"
                    onChange={(event) => setWechat(event.target.value)}
                    placeholder="请输入微信号或企业微信号（选填）"
                    value={wechat}
                  />
                </label>
              </div>
            )}

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
              <p className="form-success" role="status">
                {mode === "login" ? "登录成功，正在进入工作台" : "注册成功，正在进入工作台"}
              </p>
            )}

            <button className="login-submit" disabled={!canSubmit} type="submit">
              {status === "submitting"
                ? mode === "login" ? "正在登录" : "正在注册"
                : mode === "login" ? "登录" : "注册并创建账号"}
            </button>

            <div className="auth-divider"><span>其他登录方式</span></div>
            <div className="auth-secondary-actions">
              <button type="button">微信快捷登录</button>
              <button disabled={!phonePattern.test(phone) || countdown > 0 || status === "sending"} onClick={sendCode} type="button">
                手机号快捷登录
              </button>
            </div>

            <p className="auth-switch">
              {mode === "login" ? "还没有账号？" : "已有账号？"}
              <button onClick={() => setMode(mode === "login" ? "register" : "login")} type="button">
                {mode === "login" ? "立即注册" : "立即登录"}
              </button>
            </p>
          </form>
        </div>
      </section>
    </main>
  );
}

function resolveError(error: unknown) {
  if (!(error instanceof Error)) return "暂时无法完成登录";
  return errorMessages[error.message] ?? "暂时无法完成登录";
}

export default LoginPage;
