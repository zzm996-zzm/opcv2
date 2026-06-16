import V4PageShell from "../components/V4PageShell";

const helpTopics = ["账号与安全", "套餐与额度", "项目超市", "商业沙盘", "数据破解", "增长测算"];
const tickets = [
  ["关于项目超市筛选条件优化建议", "#FB202506250001", "2025-06-25 14:30", "已提交"],
  ["商业沙盘数据导出异常", "#FB202506240028", "2025-06-24 09:15", "处理中"],
  ["增长测算结果与预期不符", "#FB202506230017", "2025-06-23 16:45", "已解决"],
  ["AI线索开发联系人信息不全", "#FB202506220009", "2025-06-22 11:20", "已解决"],
  ["仪表盘图表显示异常", "#FB202506210006", "2025-06-21 10:05", "已解决"]
] as const;

function HelpPage() {
  return (
    <V4PageShell>
      <section className="help-page" aria-label="帮助与反馈">
        <div className="page-title-row">
          <div>
            <h1>帮助与反馈</h1>
            <p>自助答疑、问题反馈与人工咨询入口</p>
          </div>
        </div>

        <div className="help-grid">
          <section className="help-card help-docs">
            <div className="help-title">
              <span className="help-art doc" aria-hidden="true" />
              <div>
                <h2>1. 自助答疑 / 帮助文档</h2>
                <p>搜索常见问题或浏览帮助文档，快速找到解决方案</p>
              </div>
            </div>
            <label className="help-search">
              <span aria-hidden="true">⌕</span>
              <input aria-label="搜索帮助文档" placeholder="搜索帮助文档，如“如何创建项目”" />
            </label>
            <div className="help-topic-list">
              {helpTopics.map((topic) => (
                <button key={topic} type="button">
                  <span className="account-menu-icon help" aria-hidden="true" />
                  {topic}
                  <b aria-hidden="true">›</b>
                </button>
              ))}
            </div>
            <a href="/help">查看全部帮助文档 ›</a>
          </section>

          <section className="help-card feedback-form">
            <div className="help-title">
              <span className="help-art form" aria-hidden="true" />
              <div>
                <h2>2. 意见反馈 / 工单</h2>
                <p>请详细描述您的问题或建议，我们将尽快处理</p>
              </div>
            </div>
            <label>
              <span>问题类型</span>
              <button type="button">请选择问题类型 <b aria-hidden="true">⌄</b></button>
            </label>
            <label>
              <span>问题描述</span>
              <textarea placeholder="请详细描述您遇到的问题、建议或期望..." />
              <small>0/500</small>
            </label>
            <label>
              <span>附件上传（选填）</span>
              <div className="upload-box">
                <b aria-hidden="true">☁</b>
                点击或拖拽文件到此处上传
                <small>支持 jpg、png、pdf、doc、docx，单个文件不超过 10MB</small>
              </div>
            </label>
            <button className="submit-feedback" type="button">提交反馈</button>
          </section>

          <section className="help-card ticket-card">
            <div className="module-section-head">
              <div>
                <h2>3. 反馈记录 / 工单状态</h2>
              </div>
              <a href="/help">查看全部 ›</a>
            </div>
            <div className="ticket-list">
              {tickets.map(([title, no, time, state]) => (
                <article key={no}>
                  <strong>{title}</strong>
                  <span>工单号：{no}</span>
                  <span>提交时间：{time}</span>
                  <b className={state === "处理中" ? "processing" : state === "已提交" ? "submitted" : ""}>{state}</b>
                </article>
              ))}
            </div>
          </section>
        </div>

        <section className="wechat-service-card">
          <span className="service-headset" aria-hidden="true" />
          <div>
            <h2>4. 联系企业微信（人工支持 / 升级咨询）</h2>
            <p>如需人工协助、产品演示或升级咨询，请添加企业微信，我们的团队将为您提供专业支持。</p>
          </div>
          <div className="qr-box" aria-label="企业微信二维码" />
          <div>
            <strong>企业微信在线服务</strong>
            <span>人工答疑 · 快速响应</span>
            <span>产品演示 · 方案咨询</span>
            <span>版本升级 · 定制服务</span>
          </div>
          <button type="button">联系企业微信</button>
        </section>
      </section>
    </V4PageShell>
  );
}

export default HelpPage;
