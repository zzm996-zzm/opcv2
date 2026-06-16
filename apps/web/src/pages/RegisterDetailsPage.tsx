import { Link } from "react-router-dom";

const sections = [
  ["1", "基本身份", [["姓名", "请输入您的姓名"], ["职位", "请输入您的职位"], ["所在组织", "请输入您所在的部门或团队"]]],
  ["2", "我的业务 / 公司", [["公司名称", "请输入公司全称"], ["所在行业", "请选择所属行业"], ["公司规模", "请选择公司规模"]]],
  ["3", "我的产品", [["主营产品 / 服务", "请输入主营产品或服务"], ["产品阶段", "请选择产品阶段"], ["核心客户群", "请选择核心客户群"]]],
  ["4", "能力与资源", [["核心能力", "请输入您的核心能力"], ["可用资源", "请输入可用资源"], ["合作伙伴", "请输入主要合作伙伴（可选）"]]],
  ["5", "目标与诉求", [["核心目标", "请选择核心目标"], ["关键诉求", "请输入当前最关注的诉求"], ["期望合作", "请选择期望的合作方式"]]],
  ["6", "偏好", [["关注领域", "请选择关注的领域（可多选）"], ["内容偏好", "请选择内容偏好"], ["联系偏好", "请选择联系偏好"]]]
] as const;

function RegisterDetailsPage() {
  return (
    <main className="register-details-page">
      <header className="register-details-top">
        <Link className="auth-brand" to="/" aria-label="智活AI OPC V4.0 首页">
          <span className="v4-logo" aria-hidden="true" />
          <strong>智活AI</strong>
          <small>OPC V4.0</small>
        </Link>
        <nav>
          <Link to="/help">帮助中心</Link>
          <Link className="details-user" to="/profile">
            <span aria-hidden="true" />
            张婧 · 智活AI
          </Link>
        </nav>
      </header>

      <section className="register-details-card" aria-label="完善企业资料">
        <div className="details-card-head">
          <div>
            <h1>完善企业资料 <span>可选</span></h1>
            <p>完善的企业资料有助于智活AI更好地理解您的业务与需求，为您提供更精准的产品推荐、内容与服务。</p>
          </div>
          <div className="details-progress" aria-label="资料完成进度 0%">
            <i>0%</i>
            <small>已完成 0 / 6 步</small>
          </div>
          <Link to="/">先跳过，进入工作台 <span aria-hidden="true">›</span></Link>
        </div>

        <div className="details-section-grid">
          {sections.map(([index, title, fields]) => (
            <article key={title} className="details-section">
              <h2><span>{index}</span>{title} <small>（可选）</small></h2>
              {fields.map(([label, placeholder]) => (
                <label key={label}>
                  <span>{label}</span>
                  <input placeholder={placeholder} />
                </label>
              ))}
            </article>
          ))}
        </div>

        <footer className="details-save-bar">
          <span>您的信息将被严格保密，仅用于提升智活AI为您提供的服务体验。</span>
          <div>
            <Link to="/">稍后再说</Link>
            <button type="button">保存并进入工作台 <span aria-hidden="true">›</span></button>
          </div>
        </footer>
      </section>
    </main>
  );
}

export default RegisterDetailsPage;
