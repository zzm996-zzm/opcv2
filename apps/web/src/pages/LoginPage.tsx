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
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [wechat, setWechat] = useState("");
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
    () => {
      const loginPhone = account.replace(/\D/g, "");
      const registerPasswordOk =
        password.trim().length >= 6 &&
        (!confirmPassword.trim() || confirmPassword === password);

      if (mode === "login") {
        return (
          phonePattern.test(loginPhone) &&
          password.trim().length >= 6 &&
          agreementAccepted &&
          status !== "submitting"
        );
      }

      return (
        account.trim().length >= 2 &&
        registerPasswordOk &&
        phonePattern.test(phone) &&
        agreementAccepted &&
        status !== "submitting"
      );
    },
    [account, agreementAccepted, confirmPassword, mode, password, phone, status]
  );

  async function sendCode() {
    const targetPhone = mode === "login" ? account.replace(/\D/g, "") : phone;
    if (!phonePattern.test(targetPhone) || countdown > 0) return;
    setError("");
    setStatus("sending");
    try {
      await authApi.sendCode(targetPhone);
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
    const loginPhone = account.replace(/\D/g, "");
    const effectivePhone = mode === "login" ? loginPhone : phone;
    const nickname = mode === "login" ? `智活用户${loginPhone.slice(-4)}` : account.trim();
    try {
      const result = await authApi.login({
        nickname,
        phone: effectivePhone,
        code: password.trim(),
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
        <Link className="brand login-brand auth-logo" to="/" aria-label="智活AI OPC 首页">
          <span className="auth-logo-symbol" aria-hidden="true" />
          <span>智活AI</span>
          <small>OPC V4.0</small>
        </Link>
        <div className="login-story-copy">
          <p className="section-eyebrow">企业增长智能引擎</p>
          <h1>让每个决策更智能<br />让每次增长更确定</h1>
          <p>智活AI 助力企业打通数据、洞察与执行，驱动可持续增长</p>
        </div>

        <div className="auth-stage" aria-hidden="true">
          <div className="orbit orbit-a" />
          <div className="orbit orbit-b" />
          <span className="orbit-dot dot-a" />
          <span className="orbit-dot dot-b" />
          <span className="glass-cube cube-a" />
          <span className="glass-cube cube-b" />
          <div className="a-platform">
            <span className="platform-ring ring-one" />
            <span className="platform-ring ring-two" />
            <span className="platform-ring ring-three" />
            <span className="floating-a" />
          </div>
        </div>

        <div className="auth-feature-list">
          <article>
            <span>▥</span>
            <div>
              <strong>全域数据智能洞察</strong>
              <small>整合多源数据，洞察业务关键机会</small>
            </div>
          </article>
          <article>
            <span>ϟ</span>
            <div>
              <strong>AI赋能高效决策</strong>
              <small>智能分析与预测，辅助科学决策</small>
            </div>
          </article>
          <article>
            <span>◇</span>
            <div>
              <strong>业务闭环持续增长</strong>
              <small>从洞察到执行，沉淀增长方法论</small>
            </div>
          </article>
        </div>

        <p className="auth-copyright">© 2024 智活AI 版权所有 | 京ICP备2023001234号-1</p>
      </section>

      <section className="login-panel">
        <Link className="auth-help" to="/tools">? 帮助中心</Link>
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
            <label className="field auth-input">
              <span>{mode === "login" ? "账号/用户名" : <>账号 <em>*</em></>}</span>
              <input
                aria-label={mode === "login" ? "账号/用户名" : "账号"}
                autoComplete="username"
                maxLength={50}
                onChange={(event) => setAccount(event.target.value)}
                placeholder={mode === "login" ? "请输入手机号作为账号" : "请输入账号，支持字母、数字、下划线，4-20位"}
                value={account}
              />
            </label>

            <label className="field auth-input">
              <span>{mode === "login" ? "密码" : <>密码 <em>*</em></>}</span>
              <input
                aria-label="密码"
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                maxLength={20}
                onChange={(event) => setPassword(event.target.value)}
                placeholder={mode === "login" ? "请输入密码" : "请输入密码，8-20位，需包含字母和数字"}
                type="password"
                value={password}
              />
            </label>

            {mode === "register" && (
              <>
                <label className="field auth-input">
                  <span>确认密码 <small>（选填）</small></span>
                  <input
                    aria-label="确认密码"
                    autoComplete="new-password"
                    maxLength={20}
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
                  <div className="code-control">
                    <input
                      aria-label="手机号"
                      autoComplete="tel"
                      inputMode="numeric"
                      maxLength={11}
                      onChange={(event) => setPhone(event.target.value.replace(/\D/g, ""))}
                      placeholder="请输入手机号"
                      value={phone}
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

                <label className="field auth-input">
                  <span>微信 / 企业微信 <small>（选填）</small></span>
                  <input
                    aria-label="微信 / 企业微信"
                    autoComplete="off"
                    onChange={(event) => setWechat(event.target.value)}
                    placeholder="请输入微信号或企业微信号（选填）"
                    value={wechat}
                  />
                </label>
              </>
            )}

            {mode === "login" && (
              <div className="login-inline-action">
                <button type="button">忘记密码</button>
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

            {mode === "login" ? (
              <>
                <div className="auth-divider"><span>其他登录方式</span></div>
                <div className="auth-secondary-actions">
                  <button type="button">微信快捷登录</button>
                  <button
                    disabled={!phonePattern.test(account.replace(/\D/g, "")) || countdown > 0 || status === "sending"}
                    onClick={sendCode}
                    type="button"
                  >
                    {countdown > 0 ? `${countdown}s` : status === "sending" ? "发送中" : "手机号快捷登录"}
                  </button>
                </div>
                <p className="auth-switch-text">
                  还没有账号？
                  <button onClick={() => setMode("register")} type="button">立即注册</button>
                </p>
              </>
            ) : (
              <p className="auth-switch-text">
                已有账号？
                <button onClick={() => setMode("login")} type="button">立即登录</button>
              </p>
            )}

            <input aria-hidden="true" hidden readOnly value={email} />
            <input aria-hidden="true" hidden readOnly value={wechat} />
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
