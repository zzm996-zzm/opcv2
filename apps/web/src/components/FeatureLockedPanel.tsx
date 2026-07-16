import { useState } from "react";
import { Link } from "react-router-dom";

import type { FeatureAccess } from "../lib/membershipApi";
import V4PageShell from "./V4PageShell";

type BoardThreeVariant = "geo" | "leads" | "dashboard" | "crm";

type FeatureLockedPanelProps = {
  feature: FeatureAccess;
  variant: BoardThreeVariant;
};

type FeatureCard = {
  icon: string;
  title: string;
  detail: string;
  primary: string;
  secondary: string;
};

type PreviewTable = {
  title: string;
  headers: string[];
  rows: string[][];
};

type BoardThreeConfig = {
  title: string;
  subtitle: string;
  description: string;
  badge?: string;
  heroKind: "target" | "crm" | "geo" | "dashboard";
  pills: string[];
  stats?: Array<[string, string, string]>;
  cards?: FeatureCard[];
  formTitle?: string;
  formRows?: Array<[string, string]>;
  primaryAction: string;
  secondaryAction?: string;
  previewTitle: string;
  previewNote: string;
  table?: PreviewTable;
  metricCards?: Array<[string, string, string, string]>;
  sourceCards?: Array<[string, string]>;
  flow: Array<[string, string, string]>;
  copilotIntro: string;
  copilotQuestion: string;
  copilotAnswer: string[];
  copilotActions: Array<[string, string, string]>;
};

const configs: Record<BoardThreeVariant, BoardThreeConfig> = {
  leads: {
    title: "AI线索开发",
    subtitle: "GEO智能获客引擎",
    description: "用AI帮你锁定最有可能购买的客户，低成本获得高质量线索",
    heroKind: "target",
    pills: ["更精准｜锁定高匹配人群", "更高效｜多源信号实时识别", "更可触达｜联系方式核验可用", "更低成本｜节省人工筛选成本"],
    stats: [["20+", "数据来源覆盖", ""], ["100万+", "企业线索库", ""], ["85%+", "联系方式可触达率", ""]],
    formTitle: "说清你的目标客户（一句话即可）",
    formRows: [["示例", "上海的连锁餐饮品牌，正在扩张门店，需要抖音同城获客服务"], ["AI识别预览", "导入客群 / 购买信号 / 潜在需求 / 决策人 / 可触达渠道"]],
    primaryAction: "一键生成线索池",
    previewTitle: "高意向线索池（样式预览）",
    previewNote: "预览样式，升级后生成真实线索",
    table: {
      title: "线索池预览",
      headers: ["企业 / 客户", "匹配度", "购买信号（30天内）", "来源验证", "联系方式可信度", "建议动作"],
      rows: [
        ["某连锁餐饮品牌", "95% 高意向", "新增门店 / 招聘新媒体运营", "官网动态 / 招聘平台", "92% 已验证", "优先投递门店增长方案"],
        ["某跨境物流服务商", "89% 高意向", "新增海外仓 / 行业展会参展", "官网动态 / 展会官网", "88% 已验证", "发送解决方案"],
        ["某口腔连锁品牌", "82% 高意向", "新增门店 / 招聘店长", "企查查 / 招聘平台", "78% 部分可联系", "电话联系了解需求"]
      ]
    },
    sourceCards: [["企业库", "主体、规模、联系方式"], ["搜索", "官网、新闻、招聘"], ["内容平台", "公众号、短视频"], ["AI判断", "意图、预算、时机"]],
    flow: [["1", "明确目标客户", "一句话描述想触达的细分客户"], ["2", "识别购买信号", "AI从多源信号里判断需求时机"], ["3", "核验联系方式", "验证电话、企业微信、官网入口"], ["4", "跟进与沉淀", "建议跟进动作，沉淀到CRM"]],
    copilotIntro: "你的AI增长顾问，帮你低成本找到高质量客户",
    copilotQuestion: "告诉我你的业务，我帮你拆解目标客户",
    copilotAnswer: ["拆解最精准的目标客群", "识别近期购买信号", "推荐可触达的联系方式", "生成高意向线索池"],
    copilotActions: [["我卖什么，帮我找客户", "识别目标客群", "✧"], ["帮我拆目标客群", "锁定细分行业", "◇"], ["查看线索池样式", "预览输出字段", "▣"], ["联系升级权限", "开通企业版功能", "○"]]
  },
  crm: {
    title: "CRM客户管理",
    subtitle: "",
    description: "AI驱动客户全生命周期管理，让销售跟进更高效、成交更清晰",
    heroKind: "crm",
    pills: [],
    stats: [["1,286", "客户总数", "较上月 ↑ 12.3%"], ["642", "跟进中", "较上月 ↑ 8.6%"], ["328", "已成交", "较上月 ↑ 15.4%"], ["87", "流失风险", "较上月 ↑ 5.0%"]],
    primaryAction: "联系升级权限",
    previewTitle: "客户列表预览",
    previewNote: "当前为示例预览，升级后接入真实客户数据",
    metricCards: [["意向沟通", "228", "客户", "18%"], ["需求确认", "172", "客户", "16%"], ["方案报价", "96", "客户", "12%"], ["谈判中", "64", "客户", "8%"]],
    table: {
      title: "客户列表预览",
      headers: ["客户名称", "当前阶段", "负责人", "更新时间"],
      rows: [
        ["杭州智创科技有限公司", "需求确认", "李明", "2024-05-20 10:30"],
        ["上海云联信息技术有限公司", "方案报价", "王琳", "2024-05-20 09:50"],
        ["广州星瀚贸易有限公司", "意向沟通", "张伟", "2024-05-19 16:20"],
        ["深圳数智未来科技有限公司", "谈判中", "刘婷", "2024-05-19 14:10"]
      ]
    },
    sourceCards: [["客户资产沉淀", "集中管理客户数据，构建企业客户资产库"], ["AI跟进建议", "智能分析客户行为，推荐最佳跟进策略"], ["团队协同管理", "共享客户信息，提升团队协作效率"]],
    flow: [["1", "客户沉淀", "把线索、咨询和手动录入合并"], ["2", "阶段推进", "按销售管道管理客户进度"], ["3", "AI推荐跟进", "识别重点客户并生成建议"], ["4", "团队协同", "共享客户资料与复盘记录"]],
    copilotIntro: "我可以帮你梳理客户阶段，推荐重点跟进对象，并生成跟进建议。",
    copilotQuestion: "生成今日跟进清单",
    copilotAnswer: ["优先跟进高意向客户", "识别流失风险客户", "联系升级权限"],
    copilotActions: [["生成今日跟进清单", "AI为你推荐优先跟进客户", "☑"], ["识别高意向客户", "发现潜力成交客户", "◎"], ["联系升级权限", "解锁更多CRM高级能力", "♢"]]
  },
  geo: {
    title: "GEO获客",
    subtitle: "",
    description: "通过 GEO（生成式引擎优化）提升品牌在 AI 搜索与推荐中的可见性，以更低成本获取更精准客户，实现内容驱动的增长。",
    heroKind: "geo",
    pills: ["提升 AI 搜索与推荐可见性", "降低内容获取成本", "精准触达目标客户"],
    cards: [
      { icon: "✎", title: "内容优化", detail: "优化内容质量与结构，更符合 AI 搜索与推荐偏好", primary: "使用功能", secondary: "升级后可用" },
      { icon: "AI", title: "文章生成", detail: "基于主题与关键词，一键生成优质文章", primary: "使用功能", secondary: "升级后可用" },
      { icon: "▥", title: "收录监测", detail: "监测内容在 AI 平台的收录与曝光情况", primary: "使用功能", secondary: "升级后可用" }
    ],
    formTitle: "输入关键词一键GEO",
    formRows: [["核心关键词", "请输入核心关键词，例如：固态存储、口腔门诊、功能体检"], ["业务 / 产品描述", "输入品牌、产品卖点、目标客户、服务区域等信息，系统将自动生成 GEO 优化方向与内容建议..."]],
    primaryAction: "一键开始 GEO（升级后使用）",
    secondaryAction: "查看关键词方案",
    previewTitle: "收录监测",
    previewNote: "当前为示例预览，升级后展示真实收录数据",
    table: {
      title: "收录监测",
      headers: ["平台", "收录状态", "展示次数", "时间"],
      rows: [["豆包", "已收录", "128 次", "10 分钟前"], ["DeepSeek", "已收录", "96 次", "25 分钟前"], ["Kimi", "部分收录", "43 次", "1 小时前"], ["腾讯元宝", "未收录", "18 次", "2 小时前"]]
    },
    flow: [["1", "明确目标", "确定关键词与目标受众"], ["2", "内容优化", "优化内容结构与表达"], ["3", "生成与发布", "生成优质内容并发布"], ["4", "监测与迭代", "监测收录与曝光，持续优化"]],
    copilotIntro: "你的全局 AI 助手，随时为你提供帮助",
    copilotQuestion: "什么是 GEO 获客？",
    copilotAnswer: ["GEO 是通过优化内容结构与表达，提升在 AI 搜索 / 推荐结果中的可见性与收录概率，从而带来更多品牌曝光与高质量流量的增长方法。"],
    copilotActions: [["了解 GEO 能力", "查看能力介绍", "□"], ["获取低成本获客方案", "生成开通建议", "▣"], ["联系升级权限", "开通企业版功能", "▤"]]
  },
  dashboard: {
    title: "增长仪表盘",
    subtitle: "功能预览中，当前为示例预览与空态展示",
    description: "全链路增长数据一站式洞察，整合 GEO 曝光、AI 线索、渠道转化、CRM 跟进与 ROI 回报，助力科学决策，驱动高效增长。",
    heroKind: "dashboard",
    pills: [],
    primaryAction: "解锁增长仪表盘",
    secondaryAction: "查看示例口径",
    previewTitle: "锁定数据预览",
    previewNote: "真实数据需解锁并接入后查看完整分析",
    metricCards: [["核心指标预览", "1286", "线索 / 成交 / 回报", "lock"], ["增长漏斗（示例预览）", "42%", "曝光到线索转化", "lock"], ["渠道贡献（示例预览）", "5类", "GEO / AI线索 / CRM", "lock"], ["ROI诊断（示例预览）", "待接入", "投入产出诊断", "lock"]],
    sourceCards: [["数据安全加密保护", "仅在解锁并接入数据后展示完整分析结果"]],
    flow: [["1", "GEO曝光", "AI曝光表现与趋势"], ["2", "AI线索", "线索获取量与质量"], ["3", "渠道转化", "各渠道转化效率"], ["4", "CRM跟进", "跟进进度与转化率"]],
    copilotIntro: "你的增长顾问与优化助手",
    copilotQuestion: "这个仪表盘可以帮我看什么？",
    copilotAnswer: ["GEO曝光：AI 曝光表现与趋势", "AI线索：线索获取量与质量", "渠道转化：各渠道转化效率", "CRM跟进：跟进进度与转化率", "ROI回报：投入产出与收益表现"],
    copilotActions: [["解锁增长仪表盘", "查看完整增长数据与分析", "▣"], ["查看示例口径", "了解指定定义与计算逻辑", "▤"], ["咨询增长陪跑", "专家 1 对 1 增长咨询", "♧"]]
  }
};

function FeatureLockedPanel({ feature, variant }: FeatureLockedPanelProps) {
  const [modalOpen, setModalOpen] = useState(false);
  const config = configs[variant];
  const lockedMessage = feature.message || "该模块当前版本仅开放入口展示，真实工作流暂未对外启用。";

  return (
    <V4PageShell className={`board3-shell board3-${variant}`} showCopilotMini={false}>
      <main className="board3-page">
        <section className="board3-main" aria-label={`${feature.label}占位首页`}>
        {config.subtitle && <p className="board3-prebadge">ⓘ {config.subtitle}</p>}
        <section className="board3-hero">
          <div className="board3-hero-copy">
            <h1>{config.title}</h1>
            {variant === "leads" ? <span className="board3-title-badge">{config.subtitle}</span> : null}
            <p>{config.description}</p>
            <p className="board3-lock-policy" role="status">当前不会读取业务数据，也不会创建任务、客户或分析请求。</p>
            {config.pills.length > 0 ? (
              <div className="board3-pills">
                {config.pills.map((pill) => <span key={pill}>✓ {pill}</span>)}
              </div>
            ) : null}
            {variant === "dashboard" ? (
              <div className="board3-hero-actions">
                <button onClick={() => setModalOpen(true)} type="button">🔒 {config.primaryAction}</button>
                <button type="button">{config.secondaryAction}</button>
                {feature.upgrade_url ? <Link to={feature.upgrade_url}>查看升级方案</Link> : null}
              </div>
            ) : null}
          </div>
          <HeroArt kind={config.heroKind} stats={config.stats} />
        </section>

        {config.stats && variant !== "leads" ? <StatsStrip stats={config.stats} /> : null}
        {config.stats && variant === "leads" ? <StatsStrip className="compact" stats={config.stats} /> : null}

        {config.cards ? (
          <div className="board3-card-row">
            {config.cards.map((card) => (
              <article key={card.title}>
                <span>{card.icon}</span>
                <div>
                  <h3>{card.title}</h3>
                  <p>{card.detail}</p>
                  <button onClick={() => setModalOpen(true)} type="button">{card.primary}</button>
                  <small>{card.secondary}</small>
                </div>
              </article>
            ))}
          </div>
        ) : null}

        <div className="board3-grid">
          {config.formTitle ? (
            <section className="board3-form-card">
              <h2>{config.formTitle}</h2>
              {config.formRows?.map(([label, value]) => (
                <label key={label}>
                  <span>{label}</span>
                  <textarea disabled value={value} readOnly />
                </label>
              ))}
              <div>
                <button onClick={() => setModalOpen(true)} type="button">{config.primaryAction}</button>
                {config.secondaryAction ? <button type="button">{config.secondaryAction}</button> : null}
              </div>
              <small>{lockedMessage}</small>
            </section>
          ) : null}

          <section className="board3-preview-card">
            <div className="board3-section-head">
              <div>
                <h2>{config.previewTitle}</h2>
                <p>{config.previewNote}</p>
              </div>
              {variant === "leads" ? <button type="button">导出样式</button> : null}
            </div>
            {config.metricCards ? <MetricPreview cards={config.metricCards} /> : null}
            {config.table ? <PreviewTable table={config.table} /> : null}
          </section>
        </div>

        {config.sourceCards ? (
          <div className="board3-source-row">
            {config.sourceCards.map(([title, detail]) => (
              <article key={title}>
                <span aria-hidden="true">✦</span>
                <div>
                  <strong>{title}</strong>
                  <p>{detail}</p>
                </div>
              </article>
            ))}
          </div>
        ) : null}

        <section className="board3-flow">
          <h2>{variant === "leads" ? "AI线索开发流程" : variant === "geo" ? "GEO获客使用流程" : variant === "crm" ? "CRM客户管理流程" : "增长数据链路"}</h2>
          <div>
            {config.flow.map(([step, title, detail]) => (
              <article key={step}>
                <span>{step}</span>
                <strong>{title}</strong>
                <p>{detail}</p>
              </article>
            ))}
          </div>
        </section>
        </section>

        <aside className="board3-copilot" aria-label="智活 Copilot">
        <header>
          <h2>✦ 智活 Copilot</h2>
          <span>⚙⌃</span>
        </header>
        <p>{config.copilotIntro}</p>
        <div className="board3-chat-question"><b>我</b><span>{config.copilotQuestion}</span></div>
        <div className="board3-chat-answer">
          <b>♛</b>
          <div>
            <strong>智活 Copilot</strong>
            <ul>
              {config.copilotAnswer.map((item) => <li key={item}>{item}</li>)}
            </ul>
          </div>
        </div>
        <div className="board3-copilot-actions">
          {config.copilotActions.map(([title, detail, icon]) => (
            <button key={title} onClick={title.includes("权限") || title.includes("解锁") ? () => setModalOpen(true) : undefined} type="button">
              <span>{icon}</span>
              <div>
                <strong>{title}</strong>
                <small>{detail}</small>
              </div>
              <em>›</em>
            </button>
          ))}
        </div>
        <label>
          <span className="sr-only">向智活提问</span>
          <input disabled placeholder="向智活提问或获取帮助..." />
          <i>➤</i>
        </label>
        </aside>

        <button className="board3-float" aria-label="打开智活 Copilot" type="button">A</button>

        {modalOpen ? <UpgradeModal feature={feature} onClose={() => setModalOpen(false)} /> : null}
      </main>
    </V4PageShell>
  );
}

function HeroArt({ kind, stats }: { kind: BoardThreeConfig["heroKind"]; stats?: Array<[string, string, string]> }) {
  if (kind === "target") {
    return (
      <div className="board3-art board3-art-target">
        <div className="board3-target-core">◎</div>
        <span>企</span><span>店</span><span>信</span>
      </div>
    );
  }
  if (kind === "crm") {
    return (
      <div className="board3-art board3-art-crm">
        <div className="board3-crm-card">●━━</div>
        <div className="board3-funnel" />
      </div>
    );
  }
  if (kind === "geo") {
    return (
      <div className="board3-art board3-art-geo">
        <div className="board3-geo-globe">⌕</div>
        <i>低成本获客</i><i>AI搜索曝光</i><i>高质量客户</i>
      </div>
    );
  }
  return (
    <div className="board3-art board3-art-dashboard">
      <div className="board3-dashboard-core">
        <span />
        <span />
        <span />
      </div>
      {(stats ?? []).slice(0, 4).map(([, label]) => <i key={label}>{label}</i>)}
    </div>
  );
}

function StatsStrip({ className = "", stats }: { className?: string; stats: Array<[string, string, string]> }) {
  return (
    <div className={`board3-stats ${className}`}>
      {stats.map(([value, label, delta]) => (
        <article key={label}>
          <strong>{value}</strong>
          <span>{label}</span>
          {delta ? <small>{delta}</small> : null}
        </article>
      ))}
    </div>
  );
}

function MetricPreview({ cards }: { cards: Array<[string, string, string, string]> }) {
  return (
    <div className="board3-metric-preview">
      {cards.map(([title, value, detail, delta]) => (
        <article key={title}>
          <h3>{title}</h3>
          <strong>{value}</strong>
          <span>{detail}</span>
          <small>{delta === "lock" ? "🔒 接入数据后查看完整分析" : `较上月 ↑ ${delta}`}</small>
        </article>
      ))}
    </div>
  );
}

function PreviewTable({ table }: { table: PreviewTable }) {
  return (
    <div className="board3-table" role="table" aria-label={table.title}>
      <div role="row" className="board3-table-head">
        {table.headers.map((header) => <span role="columnheader" key={header}>{header}</span>)}
      </div>
      {table.rows.map((row) => (
        <div role="row" key={row.join("-")}>
          {row.map((cell, index) => <span role="cell" key={`${cell}-${index}`}>{cell}</span>)}
        </div>
      ))}
    </div>
  );
}

function UpgradeModal({ feature, onClose }: { feature: FeatureAccess; onClose: () => void }) {
  return (
    <div className="board3-modal-backdrop" role="presentation">
      <section className="board3-modal" role="dialog" aria-modal="true" aria-label="联系企业微信升级权限">
        <button className="board3-modal-close" onClick={onClose} type="button" aria-label="关闭">×</button>
        <span className="board3-modal-kicker">♕ 企业版功能</span>
        <h2>联系企业微信升级权限</h2>
        <p>GEO、AI 线索开发、仪表盘、CRM 客户管理等功能，当前仅对企业版 / 定制化陪跑用户开放。添加企业微信后，我们将为你开通对应权限，并提供专属咨询服务。</p>
        <div className="board3-modal-grid">
          <ul>
            <li>开通企业版高级功能权限</li>
            <li>获取专属顾问咨询与使用指导</li>
            <li>支持定制化陪跑与增长方案</li>
            <li>优先体验后续高级能力</li>
          </ul>
          <div className="board3-qr">
            <strong>扫码添加企业微信</strong>
            <div>企微二维码</div>
            <small>企业微信号：zhihuo-opc</small>
          </div>
        </div>
        <footer>
          {feature.contact_url ? <Link to={feature.contact_url}>我已添加，申请开通</Link> : null}
          {feature.upgrade_url ? <Link to={feature.upgrade_url}>查看升级方案</Link> : null}
          <button onClick={onClose} type="button">稍后再说</button>
        </footer>
      </section>
    </div>
  );
}

export default FeatureLockedPanel;
