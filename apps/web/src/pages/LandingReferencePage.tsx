import { useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

export type LandingModule = "tasks" | "data" | "monitoring" | "growth";
export type LandingView =
  | "list" | "board" | "calendar" | "create" | "create-menu" | "ai" | "detail" | "list-menu"
  | "home" | "progress" | "history" | "overview" | "content" | "live" | "product" | "audience" | "ads" | "sentiment" | "compare"
  | "analysis" | "questions" | "report";

type LandingReferencePageProps = {
  module: LandingModule;
  view: LandingView;
};

const taskRows = [
  ["完善销售漏斗及商机OPC店铺", "李明", "6月14日", "执行中", "中", "调研分析"],
  ["优化官网栏目用户体验", "王芳", "6月15日", "执行中", "高", "方案设计"],
  ["制定智能体全年执行计划", "张强", "6月15日", "待开始", "中", "计划制定"],
  ["补齐 SaaS 产品功能需求池", "张云", "6月16日", "已完成", "中", "需求收集"],
  ["优化线索分配规则（区域版）", "刘佳", "6月17日", "已逾期", "高", "上线验证"],
  ["更新销售话术培训", "周成", "6月17日", "已完成", "低", "培训赋能"],
  ["制定下季度直播物料计划", "钱芳", "6月13日", "待开始", "中", "策划"],
  ["输出 MVP 交互原型", "李明", "6月14日", "执行中", "高", "MVP"]
] as const;

const dataTabs = ["综合分析", "内容分析", "直播分析", "商品分析", "用户画像", "投放分析", "舆情分析", "对标分析"];
const resultCopy: Record<string, string> = {
  overview: "综合分析", content: "内容分析", live: "直播分析", product: "商品分析",
  audience: "用户画像", ads: "投放分析", sentiment: "舆情分析", compare: "对标分析"
};

function LandingReferencePage({ module, view }: LandingReferencePageProps) {
  const [isCopilotCollapsed, setIsCopilotCollapsed] = useState(false);

  return (
    <V4PageShell className={`landing-ref-shell landing-ref-${module}`} showCopilotMini={false}>
      <div className={`landing-ref-layout${isCopilotCollapsed ? " copilot-collapsed" : ""}`}>
        <section className="landing-ref-main">
          {module === "tasks" && <TasksReference view={view} />}
          {module === "data" && <DataReference view={view} />}
          {module === "monitoring" && <MonitoringReference view={view} />}
          {module === "growth" && <GrowthReference view={view} />}
        </section>
        <LandingCopilot
          collapsed={isCopilotCollapsed}
          module={module}
          onToggle={() => setIsCopilotCollapsed((collapsed) => !collapsed)}
          view={view}
        />
      </div>
    </V4PageShell>
  );
}

function VisualTitle({ title, subtitle }: { title: string; subtitle: string }) {
  return <header className="landing-ref-title"><div className="landing-ref-h1">{title}</div><p>{subtitle}</p></header>;
}

function TasksReference({ view }: { view: LandingView }) {
  if (view === "board") return <TaskBoard />;
  if (view === "calendar") return <TaskCalendar />;
  if (view === "create" || view === "create-menu") return <TaskCreate showMenu={view === "create-menu"} />;
  if (view === "ai") return <TaskAI />;
  if (view === "detail") return <TaskDetail />;
  return <TaskList showMenu={view === "list-menu"} />;
}

function TaskViewNav({ active }: { active: "list" | "board" | "calendar" }) {
  return <nav className="task-ref-view-nav">{[["列表", "/tasks", "list"], ["看板", "/tasks/board", "board"], ["日历", "/tasks/calendar", "calendar"]].map(([label, href, key]) => <Link className={active === key ? "active" : ""} key={key} to={href}>{label}</Link>)}</nav>;
}

function TaskList({ showMenu }: { showMenu: boolean }) {
  return <>
    <VisualTitle title="任务中心" subtitle="把目标拆解可执行任务，让推进更有节奏" />
    <div className="task-ref-toolbar"><label>⌕<input placeholder="搜索任务、负责人、进度阶段、标签" /></label><Link className="primary" to="/tasks/new">＋ 新建任务 ⌄</Link><TaskViewNav active="list" /><button>筛选</button><button>排序</button><button>分组</button><button>字段配置</button></div>
    {showMenu && <div className="task-ref-dropdown"><Link to="/tasks/new"><b>＋</b><span><strong>手动新建任务</strong><small>填写任务信息，快速创建</small></span></Link><Link to="/tasks/ai"><b>✦</b><span><strong>AI 快捷生成任务</strong><small>描述目标，自动拆解执行任务</small></span></Link></div>}
    <section className="task-ref-stats">{[["★", "今日新增", "8", "较昨日 -2"], ["⌛", "进行中", "24", "较昨日 +3"], ["▣", "即将到期", "6", "3天内到期"]].map(([icon, label, value, note]) => <article key={label}><i>{icon}</i><span>{label}<strong>{value}</strong><small>{note}</small></span></article>)}<div>{["负责人", "状态", "进度", "来源"].map((item) => <button key={item}>{item}<span>全部⌄</span></button>)}</div></section>
    <section className="task-ref-table"><div className="task-ref-head"><span>□ 任务标题</span><span>负责人</span><span>截止时间</span><span>状态</span><span>优先级</span><span>进展阶段</span><span>操作</span></div>{taskRows.map((row, index) => <article key={row[0]}><span>□ <strong>{row[0]}</strong></span><span><i className="mini-person">{row[1][0]}</i>{row[1]}</span><span>{row[2]}</span><span><b className={`status s${index % 4}`}>{row[3]}⌄</b></span><span><b className={`priority p${index % 3}`}>{row[4]}</b></span><span><em>{row[5]}</em></span><span><Link to="/tasks/detail">任务详情</Link></span></article>)}</section>
    <footer className="landing-ref-pagination"><span>共 10 条任务</span><div>‹ <b>1</b> 2 ›  20 条/页⌄</div></footer>
  </>;
}

function TaskBoard() {
  const columns = [
    ["待开始", "4", taskRows.slice(2, 5)], ["进行中", "5", taskRows.slice(0, 5)], ["评审中", "3", taskRows.slice(1, 4)], ["已完成", "6", taskRows.slice(3, 8)]
  ] as const;
  return <><VisualTitle title="任务视图" subtitle="切换不同视图，实时掌握任务进展，高效协同执行" /><div className="task-ref-filter"><button>负责人 全部⌄</button><button>优先级 全部⌄</button><button>项目 全部项目⌄</button><button>更多筛选</button><TaskViewNav active="board" /></div><section className="task-ref-board">{columns.map(([title, count, rows], columnIndex) => <div className={`board-col c${columnIndex}`} key={title}><header>{title}<b>{count}</b></header>{rows.map((row, index) => <article key={`${title}-${row[0]}`}><strong>{row[0]}</strong><small>（{index % 2 ? "增长测算" : "项目超市"}）</small><em>● {row[4]}</em><footer><i className="mini-person">{row[1][0]}</i>{row[1]}  {row[2]} <span>{row[5]}</span></footer></article>)}<Link to="/tasks/new">＋ 新建任务</Link></div>)}</section></>;
}

function TaskCalendar() {
  const days = Array.from({ length: 35 }, (_, index) => index < 7 ? 25 + index : index - 6);
  const events: Record<number, string> = { 3: "完善项目定位与用户画像", 8: "复盘竞品本周投放变化", 10: "制定下周渠道投放计划", 12: "设计 MVP 关键页面原型", 13: "搭建数据看板与监控指标", 17: "季度复盘与策略优化建议", 18: "验收 MVP 关键假设", 20: "完成测试与复盘", 23: "输出促销海报模板", 24: "项目里程碑" };
  return <><VisualTitle title="日历视图" subtitle="按时间维度查看任务安排，掌握时间分布与周期计划" /><div className="task-ref-filter"><button>负责人 全部⌄</button><button>优先级 全部⌄</button><button>项目 全部项目⌄</button><button>更多筛选</button><TaskViewNav active="calendar" /></div><section className="calendar-stats">{[["本月任务", "18"], ["即将到期", "5"], ["已完成", "6"], ["逾期任务", "2"]].map(([label, value]) => <article key={label}><i>◷</i><span>{label}<strong>{value}<small> 个任务</small></strong></span></article>)}</section><section className="task-calendar"><header><button>‹ ›</button><strong>2025年6月 ⌄</strong><button>今天</button><Link to="/tasks">导出日历</Link></header><div className="calendar-week">{["周日", "周一", "周二", "周三", "周四", "周五", "周六"].map((day) => <span key={day}>{day}</span>)}</div><div className="calendar-grid">{days.map((day, index) => <article key={index}><b>{day}</b>{events[day] && <span className={`e${index % 5}`}>{events[day]}</span>}</article>)}</div></section></>;
}

function TaskCreate({ showMenu }: { showMenu: boolean }) {
  return <><VisualTitle title="新建任务" subtitle="手动创建执行任务，明确任务目标、负责人和截止时间，推动高效执行" /><form className="task-ref-create"><section><h2>基本信息</h2><div className="create-two"><label><span>任务标题 *</span><input placeholder="请输入任务标题，清晰简洁地描述任务目标" /></label><label><span>来源模块 *</span><button type="button">请选择来源模块 ⌄</button>{showMenu && <div className="field-menu">项目超市<br />商业沙盘<br />竞品全盘数据破解<br />竞品动态监测<br />增长测算</div>}</label></div><label><span>任务描述</span><textarea placeholder="请输入任务描述，说明任务背景、目标、范围、验收标准等详细信息" /></label></section><section><h2>执行信息</h2><div className="create-four">{["负责人 *", "截止时间 *", "优先级 *", "状态", "标签", "预计工时", "可见范围"].map((label) => <label key={label}><span>{label}</span><button type="button">请选择 ⌄</button></label>)}</div></section><section><h2>关联信息</h2><div className="subtask-lines">{taskRows.slice(0, 3).map((row, index) => <p key={row[0]}>☷ □ {index + 1}. {row[0]} <span>{row[1]} ▣ {row[2]} 待开始 ×</span></p>)}<button type="button">＋ 新增子任务</button></div></section><div className="create-bottom"><label><span>备注</span><textarea placeholder="其他补充说明（可选）" /></label><label><span>附件</span><div>♧ 点击或拖拽文件到此处上传</div></label></div><footer><button className="primary" type="button">保存任务</button><button type="button">保存并继续添加</button><Link to="/tasks">取消</Link></footer></form></>;
}

function TaskAI() {
  return <><VisualTitle title="AI生成任务结果" subtitle="基于你提供的当前情况与目标，AI 已为你生成任务计划，请确认并采纳" /><section className="ai-task-summary"><h2>你的输入摘要</h2><div><article><b>当前情况</b><p>我们是一家B2B SaaS公司，目标客户为中大型制造业，当前线索增长慢、转化率偏低。</p></article><article><b>目标</b><p>未来 90 天内，线索量提升 30%，MQL 转化率提升 20%，建立可复制的线索增长路径。</p></article></div></section><section className="ai-task-result"><header><div><h2>✦ AI 已为你生成 12 项任务</h2><p>已按优先级与执行顺序为你规划，支持编辑与调整</p></div><span>生成质量：<b>优秀</b> 置信度：92%</span><button>重新生成</button><button className="primary">一键采纳全部</button><button>采纳选中任务</button></header><div className="task-ref-table ai">{["本周重点（4）", "后续执行（8）"].map((group, groupIndex) => <section key={group}><h3>⌄ {group}</h3>{taskRows.slice(groupIndex * 4, groupIndex * 4 + 4).map((row) => <article key={row[0]}><span>□ <strong>{row[0]}</strong></span><span>{row[5]}</span><span><i className="mini-person">{row[1][0]}</i>{row[1]}</span><span>{row[2]}</span><span><b className="priority p1">{row[4]}</b></span><span>官网优化 转化</span><span>编辑 删除 采纳</span></article>)}</section>)}</div></section></>;
}

function TaskDetail() {
  return <><div className="detail-ref-top"><Link to="/tasks">‹ 返回任务列表</Link><div><button>✓ 标记完成</button><button className="primary">保存更改</button><button className="danger">删除任务</button></div></div><div className="landing-ref-h1 small">任务详情</div><input className="detail-title-input" value="完善项目定位与用户画像（项目超市）" readOnly /><section className="detail-info"><h2>基本信息</h2><div>{["来源模块 项目超市", "关联项目 智能营销 SaaS 平台", "负责人 李明", "截止时间 2025-06-14", "优先级 高", "状态 进行中", "标签 用户研究 定位", "进度 60%"].map((item) => <button key={item}>{item} ⌄</button>)}</div><h3>任务描述</h3><div className="editor-box">↶ ↷ 正文 14  B I U S  A ⌁<p>通过对目标市场、用户需求、竞品格局的系统分析，明确本项目的市场定位、核心价值主张及关键用户画像。</p></div></section><section className="detail-columns"><article><h2>子任务清单 <small>（3/5）</small></h2>{taskRows.slice(0,5).map((row,index)=><p key={row[0]}>□ {row[0]} <span>{row[1]} {index < 2 ? "已完成" : "进行中"} {index < 2 ? "100%" : "60%"}</span></p>)}</article><article><h2>协作备注</h2>{["王芳 已完成用户基础信息收集", "张强 竞品分析已完成初稿", "李明 收到，已更新以上内容"].map((item)=><p key={item}><i className="mini-person">{item[0]}</i>{item}</p>)}</article><article><h2>操作记录</h2>{["创建任务 6月11日 09:30", "更新任务 6月12日 10:33", "更新子任务 6月12日 16:45", "提醒事件 6月12日 09:00"].map((item)=><p key={item}>● {item}</p>)}</article></section></>;
}

function DataReference({ view }: { view: LandingView }) {
  if (view === "progress") return <DataProgress />;
  if (view === "history") return <DataHistory />;
  if (["overview", "content", "live", "product", "audience", "ads", "sentiment", "compare"].includes(view)) return <DataResult section={view} />;
  return <DataHome />;
}

function DataHome() {
  return <><section className="data-ref-hero"><div><div className="landing-ref-h1 light">竞品全盘数据破解 🔒</div><h2>多平台脚本代查 + AI 深度解读</h2><p>快速获取竞品在多平台的数据，深度增长策略与内容打法，为你的决策提供真实实据。</p><div className="hero-pills"><span>多平台查询</span><span>高并发数据采集</span><span>AI深度解读</span><span>结构化输出建议</span></div></div><div className="data-hero-art" /></section><section className="data-query-card"><h2>发起数据查询</h2><p>选择平台并填写竞品账号、链接或品牌关键词，我们将为你自动脚本抓取与 AI 分析。</p><div><button>♪ 抖音 ⌄</button><input placeholder="粘贴竞品账号链接 / 输入账号昵称 / 输入品牌关键词" /><Link to="/competitor-data/progress">发起查询</Link></div><footer>◷ 查询将进入队列，完成后为你生成数据报告与 AI 解读  预计 5–15 分钟完成</footer></section><section className="data-unlock"><p>🔒 获取能力付费能力，查询结果与完整数据解读报告查看。</p><button>立即升级</button></section><h2>解锁后，你将获得</h2><section className="data-benefits">{[["全盘数据总览", "粉丝、互动、内容、直播、商品等核心指标"], ["AI 深度解读", "基于海量数据与模型，内容打法、优势与机会点"], ["后续动作建议", "针对性输出行动计划与方案清单"]].map(([title, desc]) => <article key={title}><b>{title}</b><p>{desc}</p><i>▣</i></article>)}</section></>;
}

function DataProgress() {
  const steps = ["排队中", "连接数据源", "数据抓取中", "数据清洗中", "AI 分析中", "生成结果中"];
  return <><VisualTitle title="查询处理中" subtitle="系统正在为你生成竞品全盘分析结果，请稍候..." /><section className="data-progress"><div className="progress-steps">{steps.map((step,index)=><article className={index<3?"done":""} key={step}><i>{index<2?"✓":"○"}</i><strong>{step}</strong><small>{index===2?"多维数据抓取中":"处理进行中"}</small></article>)}</div><div className="progress-body"><article><span>当前进度</span><strong>38%</strong><i><b /></i><small>数据抓取中...</small></article><article><span>当前队列位置</span><strong>第 <b>3</b> 位</strong><small>前方还有 2 个任务</small></article><div className="data-process-art" /></div><section><h2>本次查询信息</h2><p>查询平台  ♪ 抖音</p><p>目标账号  @ 美妆小白的日常</p><p>关注维度  商品、直播、内容、流量、达人、店铺、投放</p><p>发起时间  2024-06-01 10:30:15</p></section></section><aside className="progress-tip">你可以先切换到其他页面，系统会继续处理；<Link to="/competitor-data/history">查看查询历史 →</Link></aside></>;
}

function DataHistory() {
  const rows = ["完美日记官方旗舰店", "花西子Florasis官方旗舰店", "珀莱雅PROYA官方旗舰店", "小米官方旗舰店", "三只松鼠旗舰店", "蕉内Bananain官方旗舰店", "元气森林官方旗舰店"];
  return <><VisualTitle title="查询历史" subtitle="在这里查看你过往的查询记录，随时回顾结果、重新查询或导出数据。" /><div className="history-filter"><button>全部平台⌄</button><button>全部状态⌄</button><button>开始日期 → 结束日期</button><input placeholder="输入账号昵称 / 品牌 / 关键词" /><button className="primary">查询</button><button>重置</button></div><p className="info-strip">ⓘ 小贴士：历史记录保留 180 天，支持随时查看结果、重新查询或导出数据。</p><section className="history-table"><header><span>账号 / 品牌名称</span><span>平台</span><span>查询类型</span><span>创建时间</span><span>状态</span><span>完成时间</span><span>操作</span></header>{rows.map((name,index)=><article key={name}><span><i className="platform-icon">♪</i><b>{name}</b><small>抖音号：PerfectDiary</small></span><span>♪ 抖音</span><span>全盘数据破解</span><span>2025-05-{19-index} 14:32:18</span><span><em className={`state${index%3}`}>{index%3===0?"已完成":index%3===1?"处理中":"失败"}</em></span><span>2025-05-{19-index} 14:37:52</span><span><Link to="/competitor-data/results/overview">查看结果</Link> 重新查询 ⌫</span></article>)}</section><footer className="landing-ref-pagination"><span>共 28 条记录</span><div>‹ <b>1</b> 2 3 ... 6 ›  10 条/页</div></footer></>;
}

function DataResult({ section }: { section: LandingView }) {
  const activeLabel = resultCopy[section] || "综合分析";
  return <><header className="result-ref-heading"><div><small>竞品全盘数据破解 / 查询结果</small><div className="landing-ref-h1 small">竞品账号全盘数据结果 <em>✓ 查询完成</em></div><p>基于多平台数据抓取与 AI 深度分析，为你的决策提供全面、客观的参考依据。</p></div><button>重新查询</button><button>导出报告</button></header><section className="result-account"><i className="platform-icon large">♪</i><span>目标账号<strong>完美日记官方旗舰店 <small>抖音</small></strong></span><span>查询时间<strong>2024-06-08 15:42:31</strong></span><span>数据时间范围<strong>近30天（2024-05-09 - 2024-06-07）</strong></span><span>数据范围<strong>抖音账号数据</strong></span></section><section className="result-stats">{[["粉丝总量", "1,286.7万", "+12.5%"], ["作品总数", "213", "+3.4%"], ["获赞总数", "3,245.6万", "+10.7%"], ["爆款作品数", "21", "+2"], ["带货商品数", "1,268", "+86"], ["预估销售额", "¥3,245.6万", "+18.6%"]].map(([label,value,trend],index)=><article key={label}><i className={`metric-ball m${index}`}>◇</i><span>{label}<strong>{value}</strong><small>较上期 {trend}</small></span></article>)}</section><nav className="result-tabs">{dataTabs.map((tab,index)=><Link className={tab===activeLabel?"active":""} key={tab} to={`/competitor-data/results/${["overview","content","live","product","audience","ads","sentiment","compare"][index]}`}>{tab}</Link>)}</nav><ResultSection section={section} /></>;
}

function ResultSection({ section }: { section: LandingView }) {
  if (section === "audience") return <AudienceResult />;
  if (section === "compare") return <CompareResult />;
  const title = resultCopy[section] || "综合分析";
  return <><section className="chart-grid"><article><h2>{title}趋势（近30天）</h2><div className="line-chart"><i /><i /><i /></div><p>近30天核心指标稳定增长，整体表现高于行业均值。</p></article><article><h2>互动数据趋势（近30天）</h2><div className="multi-line"><i /><i /><i /><i /></div><p>平均互动率 6.82%，较上期提升 0.73pp。</p></article><article><h2>内容类型分布（近30天）</h2><div className="donut-chart"><strong>213<small>总作品数</small></strong></div></article></section><section className="insight-grid">{[["AI 核心洞察", "账号近30天净增粉丝 77.5万，增长趋势稳定向上。"], ["高表现内容 TOP3", "这支唇釉也太显白了吧！ 9.85%"], ["热销商品 TOP3", "完美日记小细跟口红 ¥586.2万"], ["后续动作建议", "持续优化短视频内容质量，提升完播与互动率。"]].map(([heading,body])=><article key={heading}><h2>{heading}</h2><p>✓ {body}</p><p>✓ 重点关注差异化内容机会与高潜力用户。</p><p>✓ 建议转化为执行任务，持续跟踪。</p></article>)}</section></>;
}

function AudienceResult() {
  return <section className="audience-grid"><article><h2>粉丝性别分布</h2><div className="donut-chart gender"><strong>62%<small>女性</small></strong></div></article><article><h2>粉丝年龄分布</h2><div className="bar-chart">{[72,88,54,32].map((v,i)=><i key={i} style={{height:`${v}%`}} />)}</div></article><article><h2>地域分布 TOP10</h2>{["广东 18.6%", "江苏 12.4%", "浙江 10.7%", "上海 8.9%", "北京 7.3%"].map((item)=><p key={item}>{item}<i /></p>)}</article><article className="wide"><h2>AI 用户画像洞察</h2><p>核心人群为 18-34 岁女性，集中在一二线城市，对新品、妆效和达人测评内容更敏感。</p><div className="tag-cloud">美妆兴趣 悦己消费 新品尝鲜 成分关注 品质生活 社交分享</div></article></section>;
}

function CompareResult() {
  return <><section className="compare-table"><header><span>核心指标</span><span>完美日记</span><span>花西子</span><span>行业均值</span></header>{[["粉丝总量","1,286.7万","1,105.2万","865.4万"],["近30天涨粉","77.5万","61.2万","28.6万"],["互动率","6.82%","5.76%","3.45%"],["预估销售额","¥3,245.6万","¥2,918.4万","¥1,462万"]].map((row)=><article key={row[0]}>{row.map((item)=><span key={item}>{item}</span>)}</article>)}</section><section className="insight-grid"><article><h2>核心优势</h2><p>✓ 粉丝增长领先竞品</p><p>✓ 内容更新稳定</p></article><article><h2>差距与风险</h2><p>✓ 直播转化仍有提升空间</p><p>✓ 同质化内容增多</p></article><article><h2>对标建议</h2><p>✓ 强化差异化新品叙事</p><p>✓ 提升达人矩阵效率</p></article></section></>;
}

function MonitoringReference({ view }: { view: LandingView }) {
  if (view === "progress") return <MonitoringProgress />;
  if (view === "history") return <MonitoringHistory />;
  if (view === "analysis") return <MonitoringAnalysis />;
  return <MonitoringHome />;
}

function MonitoringHome() {
  return <><section className="monitoring-ref-hero"><div><div className="landing-ref-h1">竞品动态监测</div><h2>持续追踪竞品招聘、内容、投放与新品动态</h2><p>✓ 多平台实时扫描</p><p>✓ 发现竞品最新动向</p><p>✓ AI 解读意图与影响</p></div><div className="monitoring-hero-art" /></section><section className="monitoring-launch"><h2>发起动态监测</h2><label><span>监测对象</span><input placeholder="输入竞品公司 / 账号 / 关键词 / 官网" /></label><label><span>监测维度</span><div>{["招聘信息", "内容发布", "营销投放", "新品动态"].map((item)=><button className="selected" key={item}>{item} ●</button>)}</div></label><label><span>监测平台</span><div>{["♪ 抖音", "小红书", "◎ 官网", "▣ 视频平台", "＋ 添加平台"].map((item)=><button key={item}>{item}</button>)}</div></label><footer><span>◉ 一次性监测 ○ 定时监测</span><Link to="/competitor-monitoring/progress">开始监测 ›</Link></footer></section><h2>监测能力一览</h2><section className="monitoring-benefits">{[["动态时间线","监测完成后，查看竞品动态时间线"],["AI 意图解读","AI 分析竞品动作意图、策略方向"],["应对建议","获得针对性建议，一键生成行动策略"],["转任务执行","将建议直接转为任务，落地执行"]].map(([title,desc])=><article key={title}><i>◇</i><b>{title}</b><p>{desc}</p></article>)}</section><div className="history-entry"><b>监测历史</b><span>查看和管理你创建的监测任务与结果</span><Link to="/competitor-monitoring/history">进入监测历史 ›</Link></div></>;
}

function MonitoringProgress() {
  return <><VisualTitle title="监测进行中" subtitle="正在持续采集完美日记官方旗舰店的最新动态" /><section className="monitoring-progress"><div className="radar-live"><i /><strong>实时扫描中</strong><small>已运行 12 分 36 秒</small></div><div className="scan-stats">{[["已扫描平台","5"],["发现动态","42"],["高影响动态","8"],["分析完成","34"]].map(([label,value])=><article key={label}><span>{label}</span><strong>{value}</strong></article>)}</div><div className="live-log"><h2>实时监测日志</h2>{["正在扫描抖音账号内容更新", "发现 2 条新品相关动态", "正在解析小红书营销投放内容", "官网产品页面发生变化", "AI 正在分析动态影响"].map((item,index)=><p key={item}><time>10:{15+index}</time><i className={`l${index}`} />{item}<b>{index<3?"已完成":"分析中"}</b></p>)}</div><footer><Link to="/competitor-monitoring/history">后台运行，前往监测历史</Link><button>停止监测</button></footer></section></>;
}

function MonitoringHistory() {
  return <><VisualTitle title="监测目录" subtitle="查看和管理竞品动态监测任务，快速进入监测结果与 AI 分析" /><div className="history-filter"><button>全部平台⌄</button><button>全部状态⌄</button><input placeholder="搜索监测对象 / 关键词" /><button className="primary">查询</button></div><section className="monitoring-directory">{["完美日记官方旗舰店","花西子官方旗舰店","珀莱雅官方旗舰店","小米官方旗舰店","三只松鼠旗舰店"].map((name,index)=><article key={name}><i className="platform-icon">{index%2?"◎":"♪"}</i><div><h2>{name}</h2><p>招聘信息、内容发布、营销投放、新品动态</p></div><span>近30天 最近更新：今天 10:30</span><b className={index?"running":"done"}>{index?"监测中":"已分析"}</b><Link to="/competitor-monitoring/analysis">查看结果</Link><button>•••</button></article>)}</section></>;
}

function MonitoringAnalysis() {
  return <><header className="analysis-ref-heading"><i className="brand-thumb">PERFECT<br/>DIARY</i><div><small>竞品动态监测 / 竞品结果</small><div className="landing-ref-h1 small">完美日记官方旗舰店 动态监测结果 <em>监测中</em></div><p>监测维度： 招聘信息 内容发布 营销投放 新品动态  最近更新：2025-05-23 10:30:00</p></div></header><nav className="monitor-tabs"><button className="active">全部</button><button>店铺信息</button><button>内容发布</button><button>营销投放</button><button>新品动态</button><button>近30天（04-23 - 05-23）⌄</button><button>✦ 分析当前动态 ⌄</button></nav><section className="monitoring-summary">{[["新增动态数","42","+15.7%"],["高影响动态","8","+33.3%"],["已分析条目","42","+15.7%"]].map(([label,value,trend])=><article key={label}><span>{label}</span><strong>{value}</strong><small>较上期 ↑ {trend}</small><i /></article>)}</section><section className="monitor-timeline">{[["10:15","♪ 抖音","内容发布","发布新品「小细跟口红」适合轻熟"],["昨天20:30","小红书","营销投放","小红书KOL合作投放｜520礼遇种草"],["05-21 15:45","公众号推文","内容发布","公众号推文｜夏日底妆清透攻略"],["05-20 09:00","官网动态","新品动态","上新「小细跟口红」系列"],["05-19 18:20","百度引擎","营销投放","百度引擎信息流投放｜新品系列品牌推广"]].map((row,index)=><article key={row[0]}><time>{row[0]}</time><i className={`t${index}`} /><b>{row[1]}</b><em>{row[2]}</em><span><strong>{row[3]}</strong><small>竞品动态内容摘要，点击查看原文与 AI 分析建议。</small></span><button>查看原文 ↗</button><Link to="/competitor-monitoring/analysis">AI分析建议</Link><mark>{index%2?"中影响":"高影响"}</mark></article>)}</section><h2>重点动态摘要</h2><section className="monitoring-benefits"><article><b>新品推进加速</b><p>新品系列结合密集投放，提升细分销售转化发力。</p></article><article><b>营销投放密集</b><p>多平台集中曝光，围绕重点节点加大推广。</p></article><article><b>内容持续输出</b><p>通过短视频与公众号持续种草用户。</p></article></section></>;
}

function GrowthReference({ view }: { view: LandingView }) {
  if (view === "questions") return <GrowthQuestions />;
  if (view === "history") return <GrowthHistory />;
  if (view === "report") return <GrowthReport />;
  return <GrowthHome />;
}

function GrowthHome() {
  return <><section className="growth-ref-hero"><div><div className="landing-ref-h1">增长测算</div><h2>用最少输入，快速测算未来营收、增长空间与关键风险</h2><section><h3>请描述你的业务情况</h3><textarea placeholder="请描述你的产品/服务、当前阶段、目标用户、当前收入/用户量、增长目标、预算/团队资源，以及你希望达成的增长目标..." /><footer><Link to="/growth-calculator/questions">✦ 开始测算</Link><Link to="/growth-calculator/history">◷ 测算历史</Link></footer></section></div><div className="growth-hero-art" /></section><p className="try-copy">试试这样写： 我是一家SaaS产品，想提升续费率  美妆电商，估算下一季度营收  线下培训机构，评估招生增长空间</p><h2>测算结果包含</h2><section className="growth-benefits">{[["营收预测","基于行业与关键参数，辅助目标制定与资源分配"],["增长空间判断","识别增长瓶颈与潜力"],["风险提示","识别潜在风险与不确定性"],["优化建议","基于数据输出可落地的增长策略"]].map(([title,desc])=><article key={title}><i>◇</i><b>{title}</b><p>{desc}</p></article>)}</section><h2>如何使用增长测算</h2><section className="growth-steps">{[["1","描述业务情况","输入关键业务数据，让 AI 快速理解你的业务。"],["2","AI 建立测算模型","基于行业模型与数据算法，构建专属测算模型。"],["3","查看结果与建议","查看营收预测、增长空间、风险与优化建议。"]].map(([n,title,desc])=><article key={n}><i>{n}</i><b>{title}</b><p>{desc}</p></article>)}</section></>;
}

function GrowthQuestions() {
  return <><VisualTitle title="补充提问" subtitle="为了让测算结果更准确，请补充以下关键业务信息" /><section className="growth-question-card"><div className="question-progress"><span>信息完整度</span><strong>65%</strong><i><b /></i><small>已完成 5 / 8 项</small></div>{[["当前月均营收","请输入近3个月平均营收","元"],["当前月均客户数","请输入每月平均付费客户数","家"],["平均客单价","请输入单客户平均收入","元"],["客户续费率","请输入当前续费率","%"],["获客成本","请输入单客户获客成本","元"],["团队规模","请输入核心团队人数","人"]].map(([label,placeholder,unit],index)=><label key={label}><span>{index<4?"*":""} {label}</span><div><input placeholder={placeholder} defaultValue={index===0?"2800000":""}/><b>{unit}</b></div><small>{index<4?"AI 已从描述中识别，可调整":"补充后可提升测算准确度"}</small></label>)}<footer><Link to="/growth-calculator">上一步</Link><Link className="primary" to="/growth-calculator/report">生成测算报告</Link></footer></section></>;
}

function GrowthHistory() {
  const rows=["智能客服SaaS增长测算","企业知识平台增长测算","AI营销助手增长测算","CRM升级版增长测算","数据分析平台增长测算","低代码开发平台增长测算","移动办公套件增长测算","开放平台API增长测算"];
  return <><header className="history-ref-title"><VisualTitle title="测算历史" subtitle="查看、对比并继续历史测算，快速追踪增长机会与趋势。"/><label>搜索测算名称或业务关键词 ⌕</label><Link to="/growth-calculator">＋ 新建测算</Link></header><div className="history-filter"><button>时间范围 近90天 ▣</button><button>业务类型 全部⌄</button><button>状态 全部⌄</button><button>排序 创建时间 ↓⌄</button><button>⟳ 重置</button></div><section className="growth-history-table"><header><span>测算名称 / 业务概况</span><span>创建时间</span><span>预测周期</span><span>核心预测营收</span><span>增长空间</span><span>风险等级</span><span>状态</span><span>操作</span></header>{rows.map((name,index)=><article key={name}><span><i>▣</i><b>{name}</b><small>面向中小企业的增长解决方案</small></span><span>2024-05-{16-index} 14:32</span><span>2024.05–2025.04</span><span>¥ {index%2?"2,180.3":"3,280.6"} 万</span><span>+{18+index}.2%</span><span><em className={`risk${index%3}`}>{["中","低","高"][index%3]}</em></span><span><b>已完成</b></span><span><Link to="/growth-calculator/report">查看报告</Link> 再次测算 ⋮</span></article>)}</section><footer className="landing-ref-pagination"><span>共 38 条</span><div>‹ <b>1</b> 2 3 ... 5 ›</div></footer></>;
}

function GrowthReport() {
  return <><header className="report-ref-head"><div><div className="landing-ref-h1 small">增长测算报告 <em>测算完成</em></div><p>基于深度的业务数据与行业模型，输出多维度增长分析与建议。</p></div><button>⇩ 导出报告</button><button>生成PPT</button><button>重新测算</button><button>☆ 收藏</button></header><section className="report-meta">{[["测算对象","智能客服 SaaS增长测算"],["测算时间","2024-05-16 14:32"],["预测周期","未来12个月"],["业务类型","SaaS 软件服务"]].map(([label,value])=><span key={label}>{label}<strong>{value}</strong></span>)}</section><section className="report-stats">{[["未来12个月营业总额","¥2,850.6万","+27.8%"],["毛利预测","¥1,312.4万","+2.4pp"],["ROI（投资回报率）","186.5%","+28.7pp"],["回本周期","5.2个月","-1.1个月"],["风险等级","中等","下降"]].map(([label,value,trend])=><article key={label}><span>{label}</span><strong>{value}</strong><small>较上期 {trend}</small><i /></article>)}</section><section className="report-ai"><h2>AI综合判断</h2>{[["增长空间","成长型（加速扩张期）"],["主要增长驱动","获客成本与转化效率"],["增长信心评分","72 /100"],["最优杠杆","提升转化率 & 复购率"]].map(([label,value])=><article key={label}><i>◇</i><span>{label}<strong>{value}</strong><small>基于测算模型与行业数据</small></span></article>)}</section><h2>核心增长驱动</h2><section className="driver-grid">{[["转化率","12.8%","大"],["客单价 (LTV)","¥3,210","中"],["复购率","26.7%","中"],["投放效率 (CAC)","¥198","中"]].map(([label,value,impact])=><article key={label}><b>{label}</b><p>提升潜力 10%，带来关键增长</p><strong>{value}</strong><span>影响程度 {impact}</span></article>)}</section><h2>关键风险诊断</h2><GenericReportTable /><h2>AI优化建议</h2><GenericReportTable /><h2>情景模拟</h2><section className="scenario-report"><div className="line-chart"><i/><i/><i/></div><table><tbody><tr><td>保守</td><td>¥2,150万</td><td>-24.6%</td></tr><tr><td>基准</td><td>¥2,850.6万</td><td>-27.6%</td></tr><tr><td>乐观</td><td>¥3,740.2万</td><td>+40.3%</td></tr></tbody></table></section><h2>90天行动计划</h2><section className="action-plan">{[["第1阶段：校准基础模型","0–30天"],["第2阶段：优化获客与转化","31–60天"],["第3阶段：提升复购与协同","61–90天"]].map(([title,time])=><article key={title}><b>{title}</b><small>{time}</small><p>□ 完善基础数据口径</p><p>□ 建立日常监测与复盘机制</p><p>□ 持续优化核心指标</p></article>)}</section></>;
}

function GenericReportTable(){return <section className="generic-report-table">{["获客成本增长过快","转化率不足预期","复购不足","交付与扩配压力"].map((item,index)=><article key={item}><b>{item}</b><em className={`risk${index%3}`}>{index<2?"高":"中"}</em><span>影响范围 8%–25%</span><p>建议优化渠道组合，提升有效流量占比，建立持续监测机制。</p></article>)}</section>}

function LandingCopilot({ collapsed, module, onToggle, view }: { collapsed: boolean; module: LandingModule; onToggle: () => void; view: LandingView }) {
  const copy = {
    tasks: ["任务中心", "我可以帮你快速定位任务、安排时间计划，并把目标拆解成可执行任务。"],
    data: ["竞品数据破解", "我可以帮你通过脚本代查，获取竞品核心数据，并输出 AI 洞察与系统化解读。"],
    monitoring: ["动态监测", "输入竞品名称或账号，我可以持续关注招聘、内容发布、营销投放和新品动态。"],
    growth: ["增长测算", "你的专属 AI 助手，帮助你高效管理历史测算并解读测算结果。"]
  }[module];
  const links = module === "tasks" ? [["拆解任务","/tasks/ai"],["安排时间计划","/tasks/calendar"],["推荐相关工具","/tools"]] : module === "data" ? [["查询竞品账号","/competitor-data"],["对比两个竞品","/competitor-data/results/compare"],["分析增长策略","/competitor-data/results/overview"]] : module === "monitoring" ? [["帮我设置监测","/competitor-monitoring"],["查看监测历史","/competitor-monitoring/history"],["了解监测维度","/competitor-monitoring/analysis"]] : [["补全关键数据","/growth-calculator/questions"],["查看测算历史","/growth-calculator/history"],["生成测算模型","/growth-calculator/report"]];
  return <aside className={`landing-ref-copilot${collapsed ? " is-collapsed" : ""}`} aria-label="智活 Copilot">
    {!collapsed && <header>
      <strong><i>✦</i> 智活 Copilot</strong>
      <button className="copilot-toggle" type="button" aria-expanded="true" aria-label="收起智活 Copilot" title="收起智活 Copilot" onClick={onToggle}>›</button>
    </header>}
    {collapsed ? <button className="copilot-rail-trigger" type="button" onClick={onToggle} aria-expanded="false" aria-label="展开智活 Copilot" title="展开智活 Copilot"><i>✦</i><span>Copilot</span><b>‹</b></button> : <div className="copilot-content">
      <p>你的全能 AI 助手，随时为你提供帮助</p>
      <article className="mine"><i>我</i><p>{view === "home" ? "请问可以帮我做些什么？" : `请帮我分析当前${copy[0]}页面`}</p></article>
      <article className="bot"><i>✦</i><div><strong>智活 Copilot</strong><p>{copy[1]}</p><ul><li>快速定位关键信息</li><li>发现机会与风险</li><li>生成可执行的操作方案</li></ul></div></article>
      <h3>你可以这样问我：</h3>
      {links.map(([label,href])=><Link key={label} to={href}>{label}<span>›</span></Link>)}
      <label><input aria-label="询问落地 Copilot" placeholder="询问任何问题..."/><button type="button" aria-label="发送消息">➤</button></label>
    </div>}
  </aside>;
}

export default LandingReferencePage;
