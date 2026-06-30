import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { leadsApi, type LeadTask } from "../lib/leadsApi";
import { CdkTopNav } from "./AnalysisPage";

type LeadCompany = {
  name: string;
  industry: string;
  score: string;
  stage: string;
  signals: readonly string[];
  next: string;
};

const leadStats = [
  ["可触达企业", "328"],
  ["高意向线索", "46"],
  ["今日待跟进", "12"],
  ["预计机会额", "¥86万"]
] as const;

const crmStats: ReadonlyArray<readonly [string, string]> = [
  ...leadStats,
  ["跟进中", "1,426"],
  ["已成交", "532"]
];

const leadCompanies = [
  {
    name: "星桥教育集团",
    industry: "连锁教育 / 私域运营",
    score: "92",
    stage: "高意向",
    signals: ["近期招聘客服主管", "公众号强调招生转化", "企微矩阵活跃"],
    next: "发送智能客服 + 企微转化方案"
  },
  {
    name: "橙果职业培训",
    industry: "职业培训 / 成人教育",
    score: "84",
    stage: "可开发",
    signals: ["新增校区", "课程咨询量增长", "官网表单更新"],
    next: "补充行业案例后预约演示"
  },
  {
    name: "领航企业内训",
    industry: "企业培训 / B端服务",
    score: "78",
    stage: "待验证",
    signals: ["投放关键词变化", "内容聚焦客户成功", "销售岗位扩张"],
    next: "先用公开信息完成需求画像"
  }
] as const;

const sourceChannels = [
  ["天眼查企业库", "主体、规模、行业、联系方式", "已接入"],
  ["公开网页搜索", "官网、新闻、案例、招聘", "采集中"],
  ["内容平台", "公众号、视频号、小红书", "排队中"],
  ["AI意图判断", "匹配需求、预算和时机", "已生成"]
] as const;

const developmentPath = [
  ["1", "定义客群", "用行业、规模、区域和触发事件圈定目标企业"],
  ["2", "抓取信号", "从工商、招聘、内容和官网变化里提取购买意图"],
  ["3", "AI评分", "按匹配度、时机、预算和触达难度给出优先级"],
  ["4", "加入CRM", "生成跟进话术、下一步动作和提醒时间"]
] as const;

const followups = [
  ["今天 14:00", "星桥教育集团", "发送智能客服选型清单"],
  ["明天 10:30", "橙果职业培训", "预约增长负责人演示"],
  ["06-25 16:00", "领航企业内训", "补充企业内训行业案例"]
] as const;

const scoringRules = [
  ["匹配度", "行业、规模和场景与当前项目是否吻合"],
  ["购买时机", "招聘、投放、扩张、内容变化是否出现近期信号"],
  ["触达质量", "是否能找到电话、邮箱、官网或关键联系人入口"],
  ["成交价值", "客单价、复购潜力和交付复杂度综合判断"]
] as const;

const statusLabels: Record<LeadTask["status"], string> = {
  queued: "排队中",
  running: "采集中",
  succeeded: "已完成",
  failed: "失败",
  cancelled: "已取消",
  refunded: "已退回"
};

function toLeadCompany(task: LeadTask): LeadCompany {
  return {
    name: task.query,
    industry: `线索任务 #${task.id}`,
    score: task.status === "succeeded" ? "90" : "72",
    stage: statusLabels[task.status],
    signals: [`消耗 ${task.credit_cost} 点额度`, `任务状态：${statusLabels[task.status]}`, `更新于 ${new Date(task.updated_at).toLocaleDateString("zh-CN")}`],
    next: task.status === "succeeded" ? "查看采集结果并筛选高优先级客户" : "任务已提交，等待线索采集和 AI 评分"
  };
}

function LeadDevelopmentPage() {
  const [query, setQuery] = useState("");
  const [tasks, setTasks] = useState<LeadTask[]>([]);
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    leadsApi
      .listTasks(20)
      .then((payload) => {
        if (active) setTasks(payload.tasks);
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取线索任务"));
      });
    return () => {
      active = false;
    };
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!query.trim() || status === "submitting") return;
    setStatus("submitting");
    setError("");
    try {
      const task = await leadsApi.createTask({
        query,
        idempotencyKey: `lead-task-${Date.now()}`
      });
      setTasks((current) => [task, ...current.filter((item) => item.id !== task.id)]);
      setQuery("");
    } catch (error) {
      setError(apiErrorMessage(error, "暂时无法创建线索任务，请稍后重试"));
    } finally {
      setStatus("idle");
    }
  }

  const companies = tasks.length > 0 ? tasks.map(toLeadCompany) : leadCompanies;

  return (
    <main className="cdk-analysis-page cdk-leads-page">
      <CdkTopNav active="VIP获客" />

      <section className="cdk-leads-hero" aria-label="AI线索开发">
        <div className="cdk-crown-art" aria-hidden="true" />
        <div>
          <h1>VIP获客</h1>
          <h2>AI线索开发</h2>
          <p>基于行业、地域、关键词与客户角色，AI 为你寻找高意向、可触达的精准客户</p>
        </div>
        <aside className="cdk-leads-formula" aria-label="线索公式">
          <strong><span aria-hidden="true">◎</span>线索公式</strong>
          <div>
            <b>军人群<small>行业+地域+角色</small></b>
            <i>×</i>
            <b>购买信号<small>需求表达+行为信号</small></b>
            <i>×</i>
            <b>可核实联系方式<small>多源验证，真实有效</small></b>
          </div>
        </aside>
      </section>

      <form className="cdk-leads-search-card" onSubmit={submit}>
        <div className="cdk-leads-query-icon" aria-hidden="true">✦</div>
        <label htmlFor="lead-target">帮我找成都市高新区少儿英语暑期班的高意向客户，优先家长求推荐信号和可核实联系方式。</label>
        <textarea
          id="lead-target"
          aria-label="描述目标客户画像"
          onChange={(event) => setQuery(event.target.value)}
          placeholder="例如：华东地区、连锁教育培训机构、正在扩张校区、需要提升客服响应和私域转化..."
          value={query}
        />
        <div className="cdk-leads-tags" aria-label="线索条件">
          {["行业", "关键词", "地域", "客户角色"].map((tag) => <button key={tag} type="button">{tag}</button>)}
        </div>
        <div className="cdk-leads-actions">
          <span>示例</span>
          <span>附件</span>
          <button aria-label={status === "submitting" ? "生成中..." : "生成线索池"} disabled={!query.trim() || status === "submitting"} type="submit">
            {status === "submitting" ? "生成中..." : "开始查找线索"}
          </button>
        </div>
        {error && <p className="form-error" role="alert">{error}</p>}
        <p className="cdk-leads-safe">AI 可能会继续询问：客户的区间、业务形态（线上/线下/混合）、决策人等信息，以便更精准地匹配线索。</p>
      </form>

      <section className="cdk-leads-crm-card" aria-label="CRM客户管理">
        <div className="cdk-leads-crm-copy">
          <h2>CRM客户管理</h2>
          <p>统一管理线索、跟进商机，提升转化效率</p>
          <Link className="cdk-leads-primary" to="/crm">进入CRM →</Link>
          <button className="module-primary-action" type="button">新建线索任务</button>
          <Link className="cdk-leads-secondary" to="/leads">查看获客任务</Link>
        </div>
        <div className="cdk-leads-stat-grid">
          {crmStats.map(([label, value], index) => (
            <article key={label}>
              <i className={`stat-${index + 1}`} aria-hidden="true" />
              <small>{label}</small>
              <strong>{value}</strong>
            </article>
          ))}
        </div>
        <aside className="cdk-leads-follow-table">
          <header>
            <h2>近期待跟进</h2>
            <Link to="/crm">查看全部 ›</Link>
          </header>
          <div className="cdk-leads-table-head">
            <span>客户/公司</span><span>阶段</span><span>负责人</span><span>下次跟进时间</span>
          </div>
          {followups.map(([time, company, action], index) => (
            <article key={`${time}-${company}`}>
              <strong>{company}</strong>
              <span>{index === 1 ? "已沟通" : "跟进中"}</span>
              <small>{index === 0 ? "李明" : index === 1 ? "陈晨" : "赵磊"}</small>
              <time>{time}</time>
              <em>{action}</em>
            </article>
          ))}
        </aside>
      </section>

      <section className="cdk-leads-result-card" aria-label="高意向线索">
        <div className="cdk-section-head">
          <div>
            <h2>高意向线索</h2>
            <p>按 AI 评分排序，优先处理购买时机明确、触达入口清晰的企业</p>
          </div>
          <div className="module-chip-row compact">
            {["全部", "高意向", "可开发", "待验证"].map((view, index) => (
              <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
            ))}
          </div>
        </div>

        <div className="leads-company-list">
          {companies.map((company) => (
            <article key={company.name}>
              <header>
                <div>
                  <h3>{company.name}</h3>
                  <small>{company.industry}</small>
                </div>
                <strong>{company.score}</strong>
              </header>
              <p>{company.next}</p>
              <div className="tool-tags">
                {company.signals.map((signal) => <span key={signal}>{signal}</span>)}
              </div>
              <footer>
                <span className={company.stage === "高意向" ? "hot" : ""}>{company.stage}</span>
                <Link to="/crm">加入CRM</Link>
              </footer>
            </article>
          ))}
        </div>
      </section>

      <section className="cdk-leads-source-section">
        <h2>线索来源与合规保障</h2>
        <div>
          {sourceChannels.slice(0, 3).map(([source, detail, status]) => (
            <article key={source}>
              <i aria-hidden="true" />
              <strong>{source}</strong>
              <p>{detail}</p>
              <div><span>{status}</span><span>更多</span></div>
            </article>
          ))}
        </div>
      </section>

      <section className="cdk-leads-rule-section">
        <strong>合规与安全</strong>
        <p>系统仅使用公开的企业信息、用户主动公开的内容及授权的线索采集渠道，不抓取私人隐私信息，严格遵守相关法律法规。</p>
        <Link to="/terms">了解更多合规说明 ›</Link>
      </section>

      <section className="leads-lower-grid cdk-leads-extra">
        <div className="leads-path-card">
          <div className="module-section-head">
            <div>
              <h2>开发路径</h2>
              <p>从目标客群到 CRM 跟进，拆成可追踪的四步流程</p>
            </div>
          </div>
          <div className="leads-path-grid">
            {developmentPath.map(([step, title, detail]) => (
              <article key={step}>
                <b>{step}</b>
                <strong>{title}</strong>
                <small>{detail}</small>
              </article>
            ))}
          </div>
        </div>

        <aside className="leads-follow-card" aria-label="跟进提醒">
          <h2>评分规则</h2>
          {scoringRules.map(([title, detail]) => (
            <article key={title}>
              <strong>{title}</strong>
              <small>{detail}</small>
            </article>
          ))}
        </aside>
      </section>
    </main>
  );
}

export default LeadDevelopmentPage;
