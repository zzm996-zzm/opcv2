import { FormEvent, useMemo, useState } from "react";
import {
  BarChart3,
  CircleHelp,
  Eye,
  EyeOff,
  LockKeyhole,
  Mail,
  MessageCircleMore,
  ShieldCheck,
  Smartphone,
  UserRound,
  Zap
} from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { ApiRequestError } from "../lib/apiRequest";
import { authApi } from "../lib/authApi";
import { authSession } from "../lib/authSession";

type AuthMode = "login" | "register";

type LoginPageProps = {
  initialMode?: AuthMode;
};

const errorMessages: Record<string, string> = {
  invalid_phone: "请输入正确的中国大陆手机号",
  invalid_code: "验证码错误或已过期",
  code_rate_limited: "发送太频繁，请稍后再试",
  agreement_required: "请先同意用户协议和隐私政策",
  nickname_required: "请输入昵称",
  invalid_account: "账号需为4-32位字母、数字或下划线",
  invalid_password: "密码需为6-72位",
  invalid_credentials: "账号或密码错误",
  account_exists: "账号已存在，请直接登录"
};

function LoginPage({ initialMode = "login" }: LoginPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [mode, setMode] = useState<AuthMode>(initialMode);
  const [account, setAccount] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [wechat, setWechat] = useState("");
  const [agreementAccepted, setAgreementAccepted] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [status, setStatus] = useState<"idle" | "submitting" | "success">("idle");
  const [error, setError] = useState("");

  const canSubmit = useMemo(() => {
    if (!agreementAccepted || status === "submitting") return false;
    const accountReady = account.trim().length >= 4;
    const passwordReady = password.trim().length >= 6;
    if (mode === "register") {
      return accountReady && passwordReady && confirmPassword === password;
    }
    return accountReady && passwordReady;
  }, [account, agreementAccepted, confirmPassword, mode, password, status]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!canSubmit) return;
    setError("");
    setStatus("submitting");
    try {
      const credentials = {
        account: account.trim(),
        password: password.trim()
      };
      const result = mode === "register"
        ? await authApi.register({
            nickname: account.trim(),
            ...credentials,
            agreementAccepted
          })
        : await authApi.login(credentials);
      authSession.set(result);
      setStatus("success");
      const returnTo = (location.state as { returnTo?: string } | null)?.returnTo;
      window.setTimeout(() => navigate(returnTo ?? "/"), 600);
    } catch (requestError) {
      setStatus("idle");
      setError(resolveError(requestError, mode));
    }
  }

  return (
    <main className="login-page public-component-auth">
      <Link className="auth-help" to="/help">
        <CircleHelp aria-hidden="true" size={22} />
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
          <span className="login-scene-orbit orbit-a"><i /></span>
          <span className="login-scene-orbit orbit-b"><i /></span>
          <span className="login-scene-cube cube-a" />
          <span className="login-scene-cube cube-b" />
          <span className="login-scene-cube cube-c" />
          <div className="login-scene-stage">
            <span className="stage-ring stage-ring-back" />
            <span className="stage-ring stage-ring-mid" />
            <span className="stage-ring stage-ring-front" />
            <span className="stage-core" />
          </div>
          <div className="login-scene-logo">
            <span className="login-logo-stroke left" />
            <span className="login-logo-stroke right" />
            <i />
          </div>
          <span className="login-scene-glow" />
        </div>

        <div className="auth-feature-list">
          <article>
            <i className="auth-feature-icon" aria-hidden="true"><BarChart3 size={24} /></i>
            <div>
              <strong>全域数据智能洞察</strong>
              <small>整合多源数据，洞察业务关键机会</small>
            </div>
          </article>
          <article>
            <i className="auth-feature-icon" aria-hidden="true"><Zap size={24} /></i>
            <div>
              <strong>AI赋能高效决策</strong>
              <small>智能分析与预测，辅助科学决策</small>
            </div>
          </article>
          <article>
            <i className="auth-feature-icon" aria-hidden="true"><ShieldCheck size={24} /></i>
            <div>
              <strong>业务闭环持续增长</strong>
              <small>从洞察到执行，沉淀增长方法论</small>
            </div>
          </article>
        </div>

        <p className="auth-copyright">© 2026 智活AI 版权所有 | <a href="https://beian.miit.gov.cn/" rel="noreferrer" target="_blank">京ICP备2023001234号-1</a></p>

      </section>

      <section className="login-panel" aria-label="登录注册">
        <div className={`login-card ${mode === "register" ? "register" : ""}`}>
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
                  <span>账号 / 用户名</span>
                  <span className="auth-field-control">
                    <UserRound aria-hidden="true" size={19} />
                    <input
                      aria-label="账号"
                      autoComplete="username"
                      maxLength={32}
                      onChange={(event) => setAccount(event.target.value)}
                      placeholder="请输入账号 / 用户名"
                      value={account}
                    />
                  </span>
                </label>

                <label className="field auth-input">
                  <span>密码</span>
                  <span className="auth-field-control has-action">
                    <LockKeyhole aria-hidden="true" size={19} />
                    <input
                      aria-label="密码"
                      autoComplete="current-password"
                      onChange={(event) => setPassword(event.target.value)}
                      placeholder="请输入密码"
                      type={showPassword ? "text" : "password"}
                      value={password}
                    />
                    <button
                      aria-label={showPassword ? "隐藏密码" : "显示密码"}
                      className="auth-password-toggle"
                      onClick={(event) => {
                        event.preventDefault();
                        setShowPassword((visible) => !visible);
                      }}
                      type="button"
                    >
                      {showPassword ? <EyeOff aria-hidden="true" size={19} /> : <Eye aria-hidden="true" size={19} />}
                    </button>
                  </span>
                </label>
                <Link className="forgot-password-link" to="/help">忘记密码</Link>
              </>
            ) : (
              <div className="register-field-grid">
                <label className="field auth-input">
                  <span>账号 <em>*</em></span>
                  <span className="auth-field-control">
                    <UserRound aria-hidden="true" size={19} />
                    <input
                      aria-label="账号"
                      autoComplete="username"
                      maxLength={32}
                      onChange={(event) => setAccount(event.target.value)}
                      placeholder="请输入账号（4-32位）"
                      aria-required="true"
                      required
                      value={account}
                    />
                  </span>
                </label>
                <label className="field auth-input">
                  <span>密码 <em>*</em></span>
                  <span className="auth-field-control has-action">
                    <LockKeyhole aria-hidden="true" size={19} />
                    <input
                      aria-label="密码"
                      autoComplete="new-password"
                      onChange={(event) => setPassword(event.target.value)}
                      placeholder="请输入密码，至少6位"
                      type={showPassword ? "text" : "password"}
                      aria-required="true"
                      required
                      value={password}
                    />
                    <button
                      aria-label={showPassword ? "隐藏密码" : "显示密码"}
                      className="auth-password-toggle"
                      onClick={(event) => {
                        event.preventDefault();
                        setShowPassword((visible) => !visible);
                      }}
                      type="button"
                    >
                      {showPassword ? <EyeOff aria-hidden="true" size={19} /> : <Eye aria-hidden="true" size={19} />}
                    </button>
                  </span>
                </label>
                <label className="field auth-input">
                  <span>确认密码 <em>*</em></span>
                  <span className="auth-field-control has-action">
                    <LockKeyhole aria-hidden="true" size={19} />
                    <input
                      aria-label="确认密码"
                      autoComplete="new-password"
                      onChange={(event) => setConfirmPassword(event.target.value)}
                      placeholder="请再次输入密码"
                      type={showConfirmPassword ? "text" : "password"}
                      aria-required="true"
                      required
                      value={confirmPassword}
                    />
                    <button
                      aria-label={showConfirmPassword ? "隐藏确认密码" : "显示确认密码"}
                      className="auth-password-toggle"
                      onClick={(event) => {
                        event.preventDefault();
                        setShowConfirmPassword((visible) => !visible);
                      }}
                      type="button"
                    >
                      {showConfirmPassword ? <EyeOff aria-hidden="true" size={19} /> : <Eye aria-hidden="true" size={19} />}
                    </button>
                  </span>
                  {confirmPassword && confirmPassword !== password ? <small className="field-error" role="alert">两次密码不一致</small> : null}
                </label>
                <label className="field auth-input">
                  <span>邮箱 <small>（选填）</small></span>
                  <span className="auth-field-control">
                    <Mail aria-hidden="true" size={19} />
                    <input
                      aria-label="邮箱"
                      autoComplete="email"
                      onChange={(event) => setEmail(event.target.value)}
                      placeholder="请输入常用邮箱地址"
                      type="email"
                      value={email}
                    />
                  </span>
                </label>
                <label className="field auth-input">
                  <span>手机号 <small>（选填）</small></span>
                  <span className="auth-field-control">
                    <Smartphone aria-hidden="true" size={19} />
                    <input
                      aria-label="手机号"
                      autoComplete="tel"
                      inputMode="tel"
                      maxLength={11}
                      onChange={(event) => setPhone(event.target.value)}
                      placeholder="请输入手机号（选填）"
                      value={phone}
                    />
                  </span>
                </label>
                <label className="field auth-input">
                  <span>微信 / 企业微信 <small>（选填）</small></span>
                  <span className="auth-field-control">
                    <MessageCircleMore aria-hidden="true" size={19} />
                    <input
                      aria-label="微信或企业微信"
                      autoComplete="off"
                      onChange={(event) => setWechat(event.target.value)}
                      placeholder="请输入微信号或企业微信号（选填）"
                      value={wechat}
                    />
                  </span>
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

            {mode === "login" && (
              <>
                <div className="auth-divider">其他登录方式</div>
                <div className="auth-secondary-actions">
                  <button type="button">
                    <MessageCircleMore className="wechat-mark" aria-hidden="true" size={20} />
                    微信快捷登录
                  </button>
                  <button type="button">
                    <Smartphone className="phone-mark" aria-hidden="true" size={20} />
                    手机号快捷登录
                  </button>
                </div>
              </>
            )}

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

function resolveError(error: unknown, mode: AuthMode) {
  const fallback = mode === "login" ? "暂时无法完成登录" : "暂时无法完成注册";
  if (error instanceof ApiRequestError) return error.message;
  if (!(error instanceof Error)) return fallback;
  return errorMessages[error.message] ?? fallback;
}

export default LoginPage;
