import { FormEvent, useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { crmApi } from "../lib/crmApi";
import { leadsApi, type LeadResult, type LeadTask, type LeadTaskDetail } from "../lib/leadsApi";
import { membershipApi, type MembershipUsageItem } from "../lib/membershipApi";
import { quotaKeys, quotaSummary } from "../lib/quotaUsage";
import { CdkTopNav } from "./AnalysisPage";

type LeadCompany = {
  leadResultID?: number;
  name: string;
  industry: string;
  score: string;
  stage: string;
  signals: readonly string[];
  next: string;
};

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

function toLeadCompanyFromResult(result: LeadResult): LeadCompany {
  const evidence = result.evidence ?? [];
  const signals = evidence.length > 0
    ? evidence.slice(0, 3).map((item) => item.title)
    : [
      result.phone ? `电话：${result.phone}` : "联系方式待补充",
      result.email ? `邮箱：${result.email}` : "邮箱待补充",
      result.website ? "官网入口已识别" : "官网待验证"
    ];
  const score = 82 + Math.min(evidence.length * 4, 12) + (result.phone || result.email ? 4 : 0);
  return {
    leadResultID: result.id,
    name: result.name,
    industry: result.website || "AI线索采集结果",
    score: String(Math.min(score, 98)),
    stage: score >= 90 ? "高意向" : "可开发",
    signals,
    next: result.phone || result.email ? "加入 CRM 并安排首轮触达" : "先补充联系人和公开业务信号"
  };
}

function LeadDevelopmentPage() {
  const [query, setQuery] = useState("");
  const [tasks, setTasks] = useState<LeadTask[]>([]);
  const [taskDetail, setTaskDetail] = useState<LeadTaskDetail | null>(null);
  const [results, setResults] = useState<LeadResult[]>([]);
  const [selectedLeadResultIDs, setSelectedLeadResultIDs] = useState<number[]>([]);
  const [importingLeadID, setImportingLeadID] = useState<number | null>(null);
  const [batchImporting, setBatchImporting] = useState(false);
  const [importedLeadIDs, setImportedLeadIDs] = useState<number[]>([]);
  const [crmImportStatus, setCrmImportStatus] = useState("");
  const [crmImportError, setCrmImportError] = useState("");
  const [status, setStatus] = useState<"idle" | "submitting">("idle");
  const [error, setError] = useState("");
  const [usage, setUsage] = useState<MembershipUsageItem[]>([]);
  const leadTaskQuota = quotaSummary(usage, quotaKeys.leadTasks, "AI线索任务");

  useEffect(() => {
    let active = true;
    async function loadUsage() {
      try {
        const payload = await membershipApi.usage();
        if (active) setUsage(payload.usage ?? []);
      } catch {
        if (active) setUsage([]);
      }
    }
    async function loadTasks() {
      try {
        const payload = await leadsApi.listTasks(20);
        if (!active) return;
        setTasks(payload.tasks);
        const completed = payload.tasks.find((task) => task.status === "succeeded");
        if (!completed) {
          setTaskDetail(null);
          setResults([]);
          setSelectedLeadResultIDs([]);
          return;
        }
        const [detailPayload, resultsPayload] = await Promise.all([
          leadsApi.getTask(completed.id).catch(() => null),
          leadsApi.listResults(completed.id, 20).catch(() => null)
        ]);
        if (!active) return;
        setTaskDetail(detailPayload);
        const nextResults = resultsPayload?.results ?? [];
        setResults(nextResults);
        setSelectedLeadResultIDs([]);
      } catch (error) {
        if (active) setError(apiErrorMessage(error, "暂时无法读取线索任务"));
      }
    }
    void loadTasks();
    void loadUsage();
    return () => {
      active = false;
    };
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!query.trim() || status === "submitting") return;
    if (leadTaskQuota.blocked) {
      setError("本月 AI线索任务额度已用完，请升级套餐或等待下月重置。");
      return;
    }
    setStatus("submitting");
    setError("");
    try {
      const task = await leadsApi.createTask({
        query,
        idempotencyKey: `lead-task-${Date.now()}`
      });
      setTasks((current) => [task, ...current.filter((item) => item.id !== task.id)]);
      setTaskDetail(null);
      setResults([]);
      setQuery("");
      const usagePayload = await membershipApi.usage().catch(() => null);
      if (usagePayload) setUsage(usagePayload.usage ?? []);
    } catch (error) {
      setError(apiErrorMessage(error, "暂时无法创建线索任务，请稍后重试"));
    } finally {
      setStatus("idle");
    }
  }

  const companies = results.length > 0
    ? results.map(toLeadCompanyFromResult)
    : tasks.length > 0
      ? tasks.map(toLeadCompany)
      : [];
  const crmStats: ReadonlyArray<readonly [string, string]> = companies.length > 0
    ? [
      ["可触达企业", String(results.length || companies.length)],
      ["高意向线索", String(companies.filter((company) => company.stage === "高意向").length)],
      ["线索任务", String(tasks.length)],
      ["待导入CRM", String(results.length)]
    ]
    : [];
  const leadResultSummary = taskDetail
    ? `${taskDetail.message} 已发现 ${taskDetail.results_count} 条候选线索。`
    : companies.length > 0
      ? "展示后端返回的线索任务和采集结果，优先处理购买时机明确、触达入口清晰的企业"
      : "暂无线索任务，提交目标客户画像后这里会展示采集结果。";
  const selectedLeadResults = results.filter((result) => selectedLeadResultIDs.includes(result.id));

  function toggleLeadResult(id: number) {
    setSelectedLeadResultIDs((current) => (
      current.includes(id) ? current.filter((item) => item !== id) : [...current, id]
    ));
  }

  async function importLeadResult(result: LeadResult) {
    setImportingLeadID(result.id);
    setCrmImportStatus("");
    setCrmImportError("");
    try {
      await crmApi.importLead({
        leadResultId: result.id,
        name: result.name,
        phone: result.phone,
        email: result.email,
        website: result.website
      });
      setImportedLeadIDs((current) => current.includes(result.id) ? current : [...current, result.id]);
      setCrmImportStatus(`${result.name} 已加入 CRM`);
    } catch (error) {
      setCrmImportError(apiErrorMessage(error, "暂时无法加入 CRM"));
    } finally {
      setImportingLeadID(null);
    }
  }

  async function importSelectedLeadResults() {
    if (selectedLeadResults.length === 0) return;
    setBatchImporting(true);
    setCrmImportStatus("");
    setCrmImportError("");
    try {
      await Promise.all(selectedLeadResults.map((result) => crmApi.importLead({
        leadResultId: result.id,
        name: result.name,
        phone: result.phone,
        email: result.email,
        website: result.website
      })));
      setImportedLeadIDs((current) => Array.from(new Set([...current, ...selectedLeadResults.map((result) => result.id)])));
      setCrmImportStatus(`已批量加入 ${selectedLeadResults.length} 条线索到 CRM`);
      setSelectedLeadResultIDs([]);
    } catch (error) {
      setCrmImportError(apiErrorMessage(error, "暂时无法批量加入 CRM"));
    } finally {
      setBatchImporting(false);
    }
  }

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
          <div className={leadTaskQuota.blocked ? "cdk-leads-quota depleted" : "cdk-leads-quota"}>
            <small>{leadTaskQuota.label}</small>
            <strong>{leadTaskQuota.value}</strong>
            {leadTaskQuota.blocked ? <Link to="/membership">升级套餐</Link> : <span>{leadTaskQuota.unit}</span>}
          </div>
          <span>示例</span>
          <span>附件</span>
          <button aria-label={status === "submitting" ? "生成中..." : "生成线索池"} disabled={!query.trim() || status === "submitting" || leadTaskQuota.blocked} type="submit">
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
          {crmStats.length === 0 ? (
            <div className="module-empty-state" role="status">暂无CRM统计</div>
          ) : crmStats.map(([label, value], index) => (
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
          <p className="module-empty-state" role="status">暂无待跟进客户</p>
        </aside>
      </section>

      <section className="cdk-leads-result-card" aria-label="高意向线索">
        <div className="cdk-section-head">
          <div>
            <h2>高意向线索</h2>
            <p>{leadResultSummary}</p>
          </div>
          <div className="module-chip-row compact">
            {results.length > 0 && (
              <button disabled={selectedLeadResults.length === 0 || batchImporting} onClick={() => void importSelectedLeadResults()} type="button">
                {batchImporting ? "导入中..." : `批量加入CRM (${selectedLeadResults.length})`}
              </button>
            )}
            {["全部", "高意向", "可开发", "待验证"].map((view, index) => (
              <button className={index === 0 ? "active" : ""} key={view} type="button">{view}</button>
            ))}
          </div>
        </div>
        {crmImportStatus && <p className="form-success" role="status">{crmImportStatus}</p>}
        {crmImportError && <p className="form-error" role="alert">{crmImportError}</p>}

        <div className="leads-company-list">
          {companies.length === 0 ? (
            <div className="module-empty-state" role="status">暂无高意向线索</div>
          ) : companies.map((company) => {
              const result = company.leadResultID ? results.find((item) => item.id === company.leadResultID) : null;
              const imported = company.leadResultID ? importedLeadIDs.includes(company.leadResultID) : false;
              return (
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
                  {company.leadResultID && (
                    <label className="lead-import-check">
                      <input
                        aria-label={`选择线索 ${company.name}`}
                        checked={selectedLeadResultIDs.includes(company.leadResultID)}
                        onChange={() => toggleLeadResult(company.leadResultID!)}
                        type="checkbox"
                      />
                      选择
                    </label>
                  )}
                  <span className={company.stage === "高意向" ? "hot" : ""}>{company.stage}</span>
                  {result ? (
                    <button disabled={importingLeadID === result.id || imported} onClick={() => void importLeadResult(result)} type="button">
                      {imported ? "已加入CRM" : importingLeadID === result.id ? "加入中..." : "加入CRM"}
                    </button>
                  ) : (
                    <Link to="/crm">加入CRM</Link>
                  )}
                </footer>
              </article>
              );
            })}
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
