import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { crmApi, type CrmActivity, type CrmCustomer, type CrmFollowUp, type CrmPipelineStats, type CrmStage } from "../lib/crmApi";
import { CdkTopNav } from "./AnalysisPage";

type CrmPageProps = {
  variant?: "customers" | "followUps";
};

type CustomerCard = {
  id: number;
  name: string;
  owner: string;
  value: string;
  stage: string;
  health: string;
  next: string;
  tags: readonly string[];
  location: string;
  contact: string;
  email: string;
};

type StatCard = readonly [string, string, string];

type FollowRow = readonly [string, string, string, string, string, string, string, string, number];
type TimelineRow = readonly [string, string];
type CustomerSourceFilter = "all" | "lead" | "enterprise";
type CustomerStageFilter = "all" | CrmStage;
type FollowUpDueFilter = "all" | "today" | "week" | "overdue";

const stageLabels: Record<CrmStage, string> = {
  new: "新线索",
  contacted: "需求确认",
  qualified: "方案演示",
  proposal: "报价谈判",
  won: "已成交",
  lost: "已流失"
};

const sourceLabels: Record<string, string> = {
  lead: "AI线索",
  enterprise: "企业交付"
};

function toCustomerCard(customer: CrmCustomer): CustomerCard {
  const nextDate = customer.next_follow_up_at
    ? new Date(customer.next_follow_up_at).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" })
    : "待安排";
  return {
    id: customer.id,
    name: customer.name,
    owner: "张婧",
    value: "待评估",
    stage: stageLabels[customer.stage],
    health: customer.stage === "proposal" || customer.stage === "qualified" ? "高意向" : "可推进",
    next: `${nextDate} 跟进客户进展`,
    location: `${sourceLabels[customer.source] ?? customer.source} · 待补地区`,
    contact: customer.phone || "待补充",
    email: customer.email || "待补充",
    tags: [sourceLabels[customer.source] ?? customer.source, customer.phone ? "电话可触达" : "待补联系方式"]
  };
}

function toStatCards(stats: CrmPipelineStats | null): readonly StatCard[] {
  if (!stats) {
    return [
      ["总客户", "0", "user"],
      ["高意向", "0", "star"],
      ["今日待跟进", "0", "calendar"],
      ["跟进中", "0", "person"],
      ["已成交", "0", "check"]
    ];
  }
  return [
    ["总客户", String(stats.total), "user"],
    ["高意向", String(stats.qualified + stats.proposal), "star"],
    ["今日待跟进", String(stats.due_today), "calendar"],
    ["跟进中", String(stats.contacted + stats.qualified + stats.proposal), "person"],
    ["已成交", String(stats.won), "check"]
  ];
}

function followUpStatus(followUp: CrmFollowUp) {
  const dueAt = new Date(followUp.next_follow_up_at);
  return dueAt.getTime() < Date.now() ? "已逾期" : "待跟进";
}

function toFollowRow(followUp: CrmFollowUp): FollowRow {
  const next = new Date(followUp.next_follow_up_at).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
  return [
    `客户 #${followUp.customer_id}`,
    "CRM客户",
    "跟进中",
    followUp.note,
    "张婧",
    next,
    "中",
    followUpStatus(followUp),
    followUp.customer_id
  ];
}

function activityTitle(activity: CrmActivity) {
  if (activity.type === "customer_updated") return "客户资料更新";
  if (activity.type === "stage_changed") return "阶段变更";
  if (activity.type === "follow_up_recorded") return "记录跟进";
  return "客户动态";
}

function toTimelineRow(activity: CrmActivity): TimelineRow {
  const time = new Date(activity.created_at).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
  return [`${time} 张婧`, activity.note || activityTitle(activity)] as const;
}

function defaultFollowUpDateTime() {
  const date = new Date(Date.now() + 24 * 60 * 60 * 1000);
  const offsetDate = new Date(date.getTime() - date.getTimezoneOffset() * 60 * 1000);
  return offsetDate.toISOString().slice(0, 16);
}

function CrmPage({ variant = "customers" }: CrmPageProps) {
  const location = useLocation();
  const [apiCustomers, setApiCustomers] = useState<CrmCustomer[]>([]);
  const [activities, setActivities] = useState<CrmActivity[]>([]);
  const [stats, setStats] = useState<CrmPipelineStats | null>(null);
  const [error, setError] = useState("");
  const [sourceFilter, setSourceFilter] = useState<CustomerSourceFilter>("all");
  const [stageFilter, setStageFilter] = useState<CustomerStageFilter>("all");
  const [customerQuery, setCustomerQuery] = useState("");
  const [followUpGoal, setFollowUpGoal] = useState("推进下一次沟通");
  const [followUpNote, setFollowUpNote] = useState("");
  const [followUpNextAt, setFollowUpNextAt] = useState(defaultFollowUpDateTime);
  const [followUpSaving, setFollowUpSaving] = useState(false);
  const [followUpCopyGenerating, setFollowUpCopyGenerating] = useState(false);
  const [followUpStatus, setFollowUpStatus] = useState("");
  const [followUpError, setFollowUpError] = useState("");
  const [stageUpdateValue, setStageUpdateValue] = useState<CrmStage>("contacted");
  const [stageUpdating, setStageUpdating] = useState(false);
  const [stageUpdateStatus, setStageUpdateStatus] = useState("");
  const [stageUpdateError, setStageUpdateError] = useState("");
  const [profileName, setProfileName] = useState("");
  const [profilePhone, setProfilePhone] = useState("");
  const [profileEmail, setProfileEmail] = useState("");
  const [profileWebsite, setProfileWebsite] = useState("");
  const [profileSaving, setProfileSaving] = useState(false);
  const [profileStatus, setProfileStatus] = useState("");
  const [profileError, setProfileError] = useState("");
  const requestedCustomerID = Number(new URLSearchParams(location.search).get("customer_id") ?? 0);

  useEffect(() => {
    let active = true;
    setError("");
    const customersRequest = requestedCustomerID > 0
      ? crmApi.getCustomer(requestedCustomerID).then((customer) => ({ customers: [customer] }))
      : crmApi.listCustomers({
        limit: 20,
        stage: stageFilter === "all" ? undefined : stageFilter,
        source: sourceFilter === "all" ? undefined : sourceFilter,
        q: customerQuery.trim() || undefined
      });
    Promise.all([customersRequest, crmApi.pipelineStats()])
      .then(([customersPayload, statsPayload]) => {
        if (active) {
          setApiCustomers(customersPayload.customers);
          setStats(statsPayload);
        }
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取 CRM 客户"));
      });
    return () => {
      active = false;
    };
  }, [sourceFilter, stageFilter, customerQuery, requestedCustomerID]);

  useEffect(() => {
    const customer = apiCustomers[0];
    if (!customer) {
      setActivities([]);
      return;
    }
    let active = true;
    crmApi
      .listActivities(customer.id, 20)
      .then((payload) => {
        if (active) setActivities(payload.activities);
      })
      .catch(() => {
        if (active) setActivities([]);
      });
    return () => {
      active = false;
    };
  }, [apiCustomers]);

  const visibleCustomers = apiCustomers.map(toCustomerCard);
  const visibleStats = toStatCards(stats);
  const selectedCustomer = visibleCustomers[0];
  const selectedApiCustomer = apiCustomers[0];
  const visibleTimelineRows = activities.map(toTimelineRow);
  const selectedFollowUpsPath = selectedApiCustomer ? `/crm/follow-ups?customer_id=${selectedApiCustomer.id}` : "/crm/follow-ups";

  useEffect(() => {
    if (!selectedApiCustomer) return;
    setStageUpdateValue(selectedApiCustomer.stage);
    setStageUpdateStatus("");
    setStageUpdateError("");
    setProfileName(selectedApiCustomer.name);
    setProfilePhone(selectedApiCustomer.phone ?? "");
    setProfileEmail(selectedApiCustomer.email ?? "");
    setProfileWebsite(selectedApiCustomer.website ?? "");
    setProfileStatus("");
    setProfileError("");
    setFollowUpGoal("推进下一次沟通");
    setFollowUpStatus("");
    setFollowUpError("");
  }, [selectedApiCustomer?.id]);

  if (variant === "followUps") {
    return <FollowUpsPage />;
  }

  const recordSelectedFollowUp = async () => {
    if (!selectedApiCustomer) return;
    const normalizedNote = followUpNote.trim();
    if (!normalizedNote) {
      setFollowUpStatus("");
      setFollowUpError("请输入跟进内容");
      return;
    }
    setFollowUpSaving(true);
    setFollowUpStatus("");
    setFollowUpError("");
    try {
      const followUp = await crmApi.recordFollowUp(selectedApiCustomer.id, {
        note: normalizedNote,
        nextFollowUpAt: new Date(followUpNextAt).toISOString()
      });
      setApiCustomers((current) => current.map((customer) => (
        customer.id === followUp.customer_id ? { ...customer, next_follow_up_at: followUp.next_follow_up_at } : customer
      )));
      setFollowUpNote("");
      setFollowUpNextAt(defaultFollowUpDateTime());
      setFollowUpStatus("跟进已记录");
    } catch (error) {
      setFollowUpError(apiErrorMessage(error, "暂时无法记录跟进"));
    } finally {
      setFollowUpSaving(false);
    }
  };

  const generateSelectedFollowUpCopy = async () => {
    if (!selectedApiCustomer) return;
    const normalizedGoal = followUpGoal.trim();
    if (!normalizedGoal) {
      setFollowUpStatus("");
      setFollowUpError("请输入跟进目标");
      return;
    }
    setFollowUpCopyGenerating(true);
    setFollowUpStatus("");
    setFollowUpError("");
    try {
      const copy = await crmApi.generateFollowUpCopy(selectedApiCustomer.id, { goal: normalizedGoal });
      setFollowUpNote(copy.body);
      setFollowUpStatus(`已生成${copy.channel}话术：${copy.subject}`);
    } catch (error) {
      setFollowUpError(apiErrorMessage(error, "暂时无法生成跟进话术"));
    } finally {
      setFollowUpCopyGenerating(false);
    }
  };

  const updateSelectedStage = async () => {
    if (!selectedApiCustomer) return;
    setStageUpdating(true);
    setStageUpdateStatus("");
    setStageUpdateError("");
    try {
      const updatedCustomer = await crmApi.updateStage(selectedApiCustomer.id, {
        stage: stageUpdateValue,
        note: `阶段更新为${stageLabels[stageUpdateValue]}`
      });
      setApiCustomers((current) => current.map((customer) => (
        customer.id === updatedCustomer.id ? updatedCustomer : customer
      )));
      const nextStats = await crmApi.pipelineStats();
      setStats(nextStats);
      setStageUpdateStatus("阶段已更新");
    } catch (error) {
      setStageUpdateError(apiErrorMessage(error, "暂时无法更新阶段"));
    } finally {
      setStageUpdating(false);
    }
  };

  const saveSelectedProfile = async () => {
    if (!selectedApiCustomer) return;
    setProfileSaving(true);
    setProfileStatus("");
    setProfileError("");
    try {
      const updatedCustomer = await crmApi.updateCustomer(selectedApiCustomer.id, {
        name: profileName.trim(),
        phone: profilePhone.trim(),
        email: profileEmail.trim(),
        website: profileWebsite.trim()
      });
      setApiCustomers((current) => current.map((customer) => (
        customer.id === updatedCustomer.id ? updatedCustomer : customer
      )));
      setProfileStatus("客户资料已更新");
    } catch (error) {
      setProfileError(apiErrorMessage(error, "暂时无法更新客户资料"));
    } finally {
      setProfileSaving(false);
    }
  };

  return (
    <main className="cdk-analysis-page cdk-crm-page">
      <CdkTopNav active="VIP获客" />
      <section className="cdk-crm-head" aria-label="CRM客户管理">
        <div>
          <h1>CRM客户管理</h1>
          <p>统一管理线索、客户与跟进流程，提升转化效率</p>
        </div>
        <div>
          <button type="button">导入客户</button>
          <button type="button">新建客户</button>
          <Link to="/crm/follow-ups">新建跟进</Link>
        </div>
      </section>

      <section className="cdk-crm-stats" aria-label="CRM关键指标">
        {visibleStats.map(([label, value, icon]) => (
          <article key={label}>
            <i className={`crm-stat-${icon}`} aria-hidden="true" />
            <span>{label}</span>
            <strong>{value}</strong>
          </article>
        ))}
      </section>

      <section className="cdk-crm-layout">
        <div className="cdk-crm-table-card">
          <header>
            <div className="cdk-crm-tabs">
              <button className={sourceFilter === "all" ? "active" : ""} type="button" onClick={() => setSourceFilter("all")}>全部客户</button>
              <button className={sourceFilter === "lead" ? "active" : ""} type="button" onClick={() => setSourceFilter("lead")}>AI线索</button>
              <button className={sourceFilter === "enterprise" ? "active" : ""} type="button" onClick={() => setSourceFilter("enterprise")}>企业交付</button>
            </div>
            <label>
              <span aria-hidden="true">⌕</span>
              <input
                aria-label="搜索客户"
                placeholder="搜索客户/公司/电话"
                value={customerQuery}
                onChange={(event) => setCustomerQuery(event.target.value)}
                disabled={requestedCustomerID > 0}
              />
            </label>
            <select
              aria-label="客户阶段筛选"
              value={stageFilter}
              onChange={(event) => setStageFilter(event.target.value as CustomerStageFilter)}
              disabled={requestedCustomerID > 0}
            >
              <option value="all">全部阶段</option>
              <option value="new">新线索</option>
              <option value="contacted">需求确认</option>
              <option value="qualified">方案演示</option>
              <option value="proposal">报价谈判</option>
              <option value="won">已成交</option>
              <option value="lost">已流失</option>
            </select>
          </header>

          <h2 className="sr-only">客户列表</h2>
          {error && <p className="form-error" role="alert">{error}</p>}
          <div className="cdk-crm-table" role="table" aria-label="客户列表">
            <div className="cdk-crm-table-head" role="row">
              {["", "客户 / 公司", "阶段", "标签", "最近跟进", "负责人", "下次跟进", "操作"].map((item) => (
                <span key={item} role="columnheader">{item}</span>
              ))}
            </div>
            {visibleCustomers.length === 0 ? (
              <div className="module-empty-state" role="status">暂无CRM客户</div>
            ) : visibleCustomers.map((customer) => (
                <article key={customer.name} role="row">
                  <input aria-label={`选择${customer.name}`} type="checkbox" />
                  <div>
                    <h3>{customer.name}</h3>
                    <small>{customer.location}</small>
                  </div>
                  <span className={`stage ${stageTone(customer.stage)}`}>{customer.stage}</span>
                  <div className="cdk-crm-tags">
                    {customer.tags.map((tag) => <b key={tag}>{tag}</b>)}
                  </div>
                  <span>{customer.stage === "新线索" ? "暂无最近跟进" : "最近已更新"}</span>
                  <span>{customer.owner}</span>
                  <span>{customer.next}</span>
                  <div className="cdk-crm-actions">
                    <Link to={`/crm?customer_id=${customer.id}`}>查看</Link>
                    <button type="button" aria-label={`更多操作 ${customer.name}`}>•••</button>
                  </div>
                </article>
              ))}
          </div>

          <footer className="cdk-crm-pagination">
            <span>共 {visibleCustomers.length} 条数据</span>
            <button type="button">10 条/页⌄</button>
            <button className="active" type="button">1</button>
          </footer>
        </div>

        <aside className="cdk-crm-detail-card" aria-label="客户详情">
          <header>
            <h2>客户详情</h2>
            <span aria-hidden="true">⌖ ×</span>
          </header>
          {selectedCustomer ? (
            <>
              <section className="cdk-crm-profile">
                <i aria-hidden="true" />
                <div>
                  <strong>{selectedCustomer.name.split(" · ")[0]}</strong>
                  <span>{selectedCustomer.stage}</span>
                  <p>{selectedCustomer.name.split(" · ")[1] || selectedCustomer.location}</p>
                  <div>{selectedCustomer.tags.map((tag) => <b key={tag}>{tag}</b>)}</div>
                </div>
              </section>
              <dl className="cdk-crm-detail-list">
                <div><dt>联系方式</dt><dd>{selectedCustomer.contact} {selectedCustomer.email}</dd></div>
                <div><dt>来源</dt><dd>{selectedCustomer.location}</dd></div>
                <div><dt>需求摘要</dt><dd>{selectedCustomer.health}，下一步：{selectedCustomer.next}</dd></div>
              </dl>
              <section className="cdk-crm-profile-form" aria-label="编辑客户资料">
                <label>
                  <span>客户名称</span>
                  <input value={profileName} onChange={(event) => setProfileName(event.target.value)} />
                </label>
                <label>
                  <span>电话</span>
                  <input value={profilePhone} onChange={(event) => setProfilePhone(event.target.value)} />
                </label>
                <label>
                  <span>邮箱</span>
                  <input value={profileEmail} onChange={(event) => setProfileEmail(event.target.value)} />
                </label>
                <label>
                  <span>网站</span>
                  <input value={profileWebsite} onChange={(event) => setProfileWebsite(event.target.value)} />
                </label>
                {profileStatus && <p className="form-success" role="status">{profileStatus}</p>}
                {profileError && <p className="form-error" role="alert">{profileError}</p>}
                <button type="button" onClick={saveSelectedProfile} disabled={profileSaving}>
                  {profileSaving ? "保存中..." : "保存资料"}
                </button>
              </section>
              <section className="cdk-crm-stage-form" aria-label="更新客户阶段">
                <label>
                  <span>客户阶段</span>
                  <select value={stageUpdateValue} onChange={(event) => setStageUpdateValue(event.target.value as CrmStage)}>
                    <option value="new">新线索</option>
                    <option value="contacted">需求确认</option>
                    <option value="qualified">方案演示</option>
                    <option value="proposal">报价谈判</option>
                    <option value="won">已成交</option>
                    <option value="lost">已流失</option>
                  </select>
                </label>
                {stageUpdateStatus && <p className="form-success" role="status">{stageUpdateStatus}</p>}
                {stageUpdateError && <p className="form-error" role="alert">{stageUpdateError}</p>}
                <button type="button" onClick={updateSelectedStage} disabled={stageUpdating}>
                  {stageUpdating ? "更新中..." : "更新阶段"}
                </button>
              </section>
              <section className="cdk-crm-timeline">
                <h2>跟进看板</h2>
                {visibleTimelineRows.length === 0 ? (
                  <p className="module-empty-state">暂无跟进动态</p>
                ) : visibleTimelineRows.map(([time, detail]) => (
                    <article key={time}>
                      <time>{time}</time>
                      <p>{detail}</p>
                    </article>
                  ))}
                <Link to={selectedFollowUpsPath}>查看全部跟进记录 ›</Link>
              </section>
              <section className="cdk-crm-followup-form" aria-label="记录客户跟进">
                <label>
                  <span>跟进目标</span>
                  <input value={followUpGoal} onChange={(event) => setFollowUpGoal(event.target.value)} />
                </label>
                <button type="button" onClick={generateSelectedFollowUpCopy} disabled={followUpCopyGenerating}>
                  {followUpCopyGenerating ? "生成中..." : "生成话术"}
                </button>
                <label>
                  <span>跟进内容</span>
                  <textarea value={followUpNote} onChange={(event) => setFollowUpNote(event.target.value)} />
                </label>
                <label>
                  <span>下次跟进时间</span>
                  <input type="datetime-local" value={followUpNextAt} onChange={(event) => setFollowUpNextAt(event.target.value)} />
                </label>
                {followUpStatus && <p className="form-success" role="status">{followUpStatus}</p>}
                {followUpError && <p className="form-error" role="alert">{followUpError}</p>}
                <button type="button" onClick={recordSelectedFollowUp} disabled={followUpSaving}>
                  {followUpSaving ? "保存中..." : "保存跟进"}
                </button>
              </section>
              <footer>
                <button type="button">拨打电话</button>
                <button type="button">发消息</button>
                <Link to={selectedFollowUpsPath}>记录跟进</Link>
              </footer>
            </>
          ) : (
            <p className="module-empty-state" role="status">暂无客户详情</p>
          )}
        </aside>
      </section>
    </main>
  );
}

function FollowUpsPage() {
  const location = useLocation();
  const [apiFollowUps, setApiFollowUps] = useState<CrmFollowUp[]>([]);
  const [followUpQuery, setFollowUpQuery] = useState("");
  const [dueFilter, setDueFilter] = useState<FollowUpDueFilter>("all");
  const [error, setError] = useState("");
  const customerID = Number(new URLSearchParams(location.search).get("customer_id") ?? 0);

  useEffect(() => {
    let active = true;
    setError("");
    crmApi
      .listFollowUps({
        customerId: customerID > 0 ? customerID : undefined,
        q: followUpQuery.trim() || undefined,
        due: dueFilter === "all" ? undefined : dueFilter,
        limit: 20
      })
      .then((payload) => {
        if (active) setApiFollowUps(payload.follow_ups);
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取跟进记录"));
      });
    return () => {
      active = false;
    };
  }, [customerID, followUpQuery, dueFilter]);

  const visibleFollowRows = apiFollowUps.map(toFollowRow);
  const now = Date.now();
  const today = new Date();
  const dueTodayCount = apiFollowUps.filter((followUp) => {
    const dueAt = new Date(followUp.next_follow_up_at);
    return dueAt.getFullYear() === today.getFullYear() && dueAt.getMonth() === today.getMonth() && dueAt.getDate() === today.getDate();
  }).length;
  const dueThisWeekCount = apiFollowUps.filter((followUp) => {
    const dueAt = new Date(followUp.next_follow_up_at).getTime();
    return dueAt >= now && dueAt <= now + 7 * 24 * 60 * 60 * 1000;
  }).length;

  return (
    <main className="cdk-analysis-page cdk-crm-page cdk-followups-page">
      <CdkTopNav active="VIP获客" />
      <section className="cdk-crm-head cdk-followups-head" aria-label="全部跟进">
        <div>
          <Link to="/crm" aria-label="返回CRM">←</Link>
          <h1>全部跟进</h1>
          <p>查看所有待跟进、已沟通和已完成的客户跟进记录</p>
        </div>
        <div>
          <button type="button">导出记录</button>
          <button type="button">新建跟进计划 ＋</button>
        </div>
      </section>

      <section className="cdk-crm-stats cdk-followups-stats">
        {[
          ["今日待跟进", String(dueTodayCount), "calendar"],
          ["本周待跟进", String(dueThisWeekCount), "chart"],
          ["跟进记录", String(apiFollowUps.length), "message"],
          ["已逾期", String(visibleFollowRows.filter((row) => row[7] === "已逾期").length), "check"]
        ].map(([label, value, icon]) => (
          <article key={label}>
            <i className={`crm-stat-${icon}`} aria-hidden="true" />
            <span>{label}</span>
            <strong>{value}</strong>
          </article>
        ))}
      </section>

      <section className="cdk-followups-layout">
        <div className="cdk-followups-table-card">
          <header>
            <div className="cdk-crm-tabs">
              {[
                ["all", "全部"],
                ["today", "今日待跟进"],
                ["week", "本周待跟进"],
                ["overdue", "已逾期"]
              ].map(([value, label]) => (
                <button
                  className={dueFilter === value ? "active" : ""}
                  key={value}
                  type="button"
                  onClick={() => setDueFilter(value as FollowUpDueFilter)}
                >
                  {label}
                </button>
              ))}
            </div>
            <label>
              <input
                aria-label="搜索跟进记录"
                placeholder="搜索跟进内容/客户ID"
                value={followUpQuery}
                onChange={(event) => setFollowUpQuery(event.target.value)}
              />
              <span aria-hidden="true">⌕</span>
            </label>
            <button type="button">阶段筛选⌄</button>
            <button type="button">时间范围⌄</button>
          </header>
          <div className="cdk-followups-table" role="table" aria-label="全部跟进列表">
            {error && <p className="form-error" role="alert">{error}</p>}
            <div className="cdk-followups-row head" role="row">
              {["客户 / 公司", "当前阶段", "最近跟进内容", "负责人", "下次跟进时间", "优先级", "跟进状态", "操作"].map((item) => (
                <span key={item} role="columnheader">{item}</span>
              ))}
            </div>
            {visibleFollowRows.length === 0 ? (
              <div className="module-empty-state" role="status">暂无跟进记录</div>
            ) : visibleFollowRows.map(([name, sub, stage, note, owner, next, priority, status, customerID]) => (
                <article className="cdk-followups-row" key={`${name}-${next}`} role="row">
                  <div><strong>{name}</strong><small>{sub}</small></div>
                  <span className={`stage ${stageTone(stage)}`}>{stage}</span>
                  <p>{note}</p>
                  <span>{owner}</span>
                  <time>{next}</time>
                  <b className={`priority ${priority}`}>{priority}</b>
                  <span className="follow-status">{status}</span>
                  <div><Link to={`/crm?customer_id=${customerID}`}>查看详情</Link><button type="button">记录跟进</button><button type="button">改期</button></div>
                </article>
              ))}
          </div>
          <footer className="cdk-crm-pagination">
            <span>共 {visibleFollowRows.length} 条记录</span>
            <button type="button">10 条/页⌄</button>
            <button className="active" type="button">1</button>
          </footer>
        </div>

        <aside className="cdk-followups-sidebar">
          <section>
            <h2>跟进提醒</h2>
            <header><strong>今日待跟进（{dueTodayCount}）</strong><Link to="/crm/follow-ups">查看全部</Link></header>
            {visibleFollowRows.length === 0 ? (
              <p className="module-empty-state">暂无跟进提醒</p>
            ) : visibleFollowRows.slice(0, 5).map(([name, sub,, , , next,, status]) => (
                <article key={`${next}-${name}`}>
                  <time>{next}</time>
                  <span><strong>{name}</strong><small>{sub}</small></span>
                  <b>{status}</b>
                </article>
              ))}
            <Link to="/tasks">查看全部日程</Link>
          </section>
          <section>
            <h2>最近更新</h2>
            {apiFollowUps.length === 0 ? (
              <p className="module-empty-state">暂无最近更新</p>
            ) : apiFollowUps.slice(0, 4).map((followUp, index) => (
                <article className="update" key={followUp.id}>
                  <i className={`dot-${index + 1}`} aria-hidden="true" />
                  <span><strong>记录跟进：客户 #{followUp.customer_id}</strong><small>{followUp.note}</small></span>
                  <time>{new Date(followUp.created_at).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" })}</time>
                </article>
              ))}
            <Link to="/crm/follow-ups">查看更多记录</Link>
          </section>
        </aside>
      </section>
    </main>
  );
}

function stageTone(stage: string) {
  if (stage.includes("成交")) return "won";
  if (stage.includes("已沟通")) return "talked";
  if (stage.includes("高意向")) return "hot";
  if (stage.includes("意向")) return "intent";
  if (stage.includes("新")) return "new";
  return "active";
}

export default CrmPage;
