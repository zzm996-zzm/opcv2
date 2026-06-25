import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";
import { leadsApi, type LeadTask } from "../lib/leadsApi";

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
      .catch(() => {
        if (active) setError("暂时无法读取线索任务");
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
    } catch {
      setError("暂时无法创建线索任务，请稍后重试");
    } finally {
      setStatus("idle");
    }
  }

  const companies = tasks.length > 0 ? tasks.map(toLeadCompany) : leadCompanies;

  return (
    <V4PageShell className="lead-development-shell">
      <section className="module-page lead-development-page" aria-label="AI线索开发">
        <div className="page-title-row">
          <div>
            <h1>AI线索开发</h1>
            <p>输入目标客户画像，系统自动找到公开企业线索、识别购买信号、生成评分和跟进动作</p>
          </div>
          <button className="module-primary-action" type="button">新建线索任务</button>
        </div>

        <section className="module-overview-card leads-hero">
          <div className="module-overview-copy">
            <span className="module-kicker">教育培训机构 · 智能客服机会</span>
            <h2>把散落的公开企业信息变成可跟进线索</h2>
            <p>AI 会结合行业、招聘、内容动态、官网变化和联系方式，判断哪些企业更可能需要你的产品，并同步到 CRM 跟进。</p>
            <div className="module-stat-strip">
              {leadStats.map(([label, value]) => (
                <article key={label}>
                  <small>{label}</small>
                  <strong>{value}</strong>
                </article>
              ))}
            </div>
          </div>

          <form className="module-ai-box compact leads-query-card" onSubmit={submit}>
            <label htmlFor="lead-target">描述目标客户画像</label>
            <textarea
              id="lead-target"
              aria-label="描述目标客户画像"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="例如：华东地区、连锁教育培训机构、正在扩张校区、需要提升客服响应和私域转化..."
              value={query}
            />
            <button disabled={!query.trim() || status === "submitting"} type="submit">
              {status === "submitting" ? "生成中..." : "生成线索池"}
            </button>
            {error && <p className="form-error" role="alert">{error}</p>}
          </form>
        </section>

        <section className="leads-workbench">
          <div className="leads-main-card">
            <div className="module-section-head">
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
          </div>

          <aside className="leads-source-card" aria-label="数据源状态">
            <h2>数据源状态</h2>
            {sourceChannels.map(([source, detail, status]) => (
              <article key={source}>
                <span>
                  <strong>{source}</strong>
                  <small>{detail}</small>
                </span>
                <em>{status}</em>
              </article>
            ))}
          </aside>
        </section>

        <section className="leads-lower-grid">
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
            <h2>跟进提醒</h2>
            {followups.map(([time, company, action]) => (
              <article key={`${time}-${company}`}>
                <time>{time}</time>
                <strong>{company}</strong>
                <small>{action}</small>
              </article>
            ))}
          </aside>
        </section>

        <section className="leads-rule-section">
          <div className="module-section-head">
            <div>
              <h2>评分规则</h2>
              <p>第一版先展示评分逻辑，后续可接入真实数据源和异步任务</p>
            </div>
          </div>
          <div className="leads-rule-grid">
            {scoringRules.map(([title, detail]) => (
              <article key={title}>
                <strong>{title}</strong>
                <p>{detail}</p>
              </article>
            ))}
          </div>
        </section>
      </section>
    </V4PageShell>
  );
}

export default LeadDevelopmentPage;
