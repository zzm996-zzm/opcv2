import { useEffect, useState } from "react";
import { Link } from "react-router-dom";

import { apiErrorMessage } from "../lib/apiErrors";
import { crmApi, type CrmCustomer, type CrmStage } from "../lib/crmApi";
import { CdkTopNav } from "./AnalysisPage";

type CrmPageProps = {
  variant?: "customers" | "followUps";
};

type CustomerCard = {
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

const crmStats = [
  ["总客户", "12,845", "user"],
  ["高意向", "2,318", "star"],
  ["今日待跟进", "68", "calendar"],
  ["跟进中", "1,426", "person"],
  ["已成交", "532", "check"]
] as const;

const customers = [
  {
    name: "张女士 · 成都蓝鲸教育",
    owner: "李明",
    value: "¥36万",
    stage: "跟进中",
    health: "高意向",
    next: "明天 10:00",
    location: "成都高新区 · 少儿英语",
    contact: "138****6789",
    email: "zhang***@qq.com",
    tags: ["英语培训", "K12"]
  },
  {
    name: "王先生 · 星辰国际学校",
    owner: "陈晨",
    value: "¥18万",
    stage: "已沟通",
    health: "可推进",
    next: "后天 16:00",
    location: "成都高新区 · 国际教育",
    contact: "139****8821",
    email: "wang***@qq.com",
    tags: ["国际学校", "留学"]
  },
  {
    name: "刘女士 · 乐学英语中心",
    owner: "赵磊",
    value: "¥24万",
    stage: "跟进中",
    health: "高意向",
    next: "明天 14:00",
    location: "成都锦江区 · 成人英语",
    contact: "136****0192",
    email: "liu***@qq.com",
    tags: ["英语培训", "少儿英语"]
  },
  {
    name: "李先生 · 启航教育科技",
    owner: "孙悦",
    value: "待评估",
    stage: "新线索",
    health: "待决策",
    next: "3天后 10:00",
    location: "成都武侯区 · K12辅导",
    contact: "待补充",
    email: "待补充",
    tags: ["教育科技", "SaaS"]
  },
  {
    name: "陈女士 · 阳光少儿英语",
    owner: "李明",
    value: "¥12万",
    stage: "跟进中",
    health: "高意向",
    next: "明天 09:00",
    location: "成都青羊区 · 出国留学",
    contact: "137****7721",
    email: "chen***@qq.com",
    tags: ["少儿英语", "口语"]
  },
  {
    name: "周先生 · 优学国际教育",
    owner: "陈晨",
    value: "¥28万",
    stage: "高意向",
    health: "高意向",
    next: "今天 16:30",
    location: "成都天府新区 · 职业培训",
    contact: "135****9134",
    email: "zhou***@qq.com",
    tags: ["国际教育", "留学"]
  },
  {
    name: "黄女士 · 巴蜀文化学校",
    owner: "赵磊",
    value: "¥9万",
    stage: "已成交",
    health: "已成交",
    next: "-",
    location: "成都金牛区 · 语言培训",
    contact: "138****6520",
    email: "huang***@qq.com",
    tags: ["学校", "K12"]
  },
  {
    name: "吴先生 · 博睿教育咨询",
    owner: "孙悦",
    value: "待评估",
    stage: "新线索",
    health: "可推进",
    next: "2天后 15:00",
    location: "成都高新区 · 素质教育",
    contact: "待补充",
    email: "wu***@qq.com",
    tags: ["教育咨询", "升学规划"]
  },
  {
    name: "星桥教育集团",
    owner: "张婧",
    value: "¥36万",
    stage: "方案演示",
    health: "高意向",
    next: "今天 14:00",
    location: "连锁教育 · 私域运营",
    contact: "138****1024",
    email: "contact***@example.com",
    tags: ["连锁教育", "企微转化"]
  }
] as const;

const followRows = [
  ["张女士 · 成都蓝鲸教育", "成都高新区 · 少儿英语", "跟进中", "已介绍课程体系与师资，客户对外教课程有兴趣，需发送详细课程介绍。", "李明", "今天 15:00", "高", "待跟进"],
  ["王先生 · 星辰国际学校", "成都高新区 · 国际教育", "已沟通", "客户对课程方案认可，正在内部评估预算，预计下周给答复。", "陈晨", "明天 10:30", "高", "待跟进"],
  ["刘女士 · 乐学英语中心", "成都锦江区 · 成人英语", "跟进中", "发送了试听安排，客户反馈老师讲解清晰，满意度较高。", "赵磊", "后天 16:00", "中", "待跟进"],
  ["赵磊 · 优学教育", "成都武侯区 · K12辅导", "意向确认", "正在确认合作模式与费用细节，客户关注开课时间与排课安排。", "周文", "2024-05-28 11:00", "中", "已逾期"],
  ["陈女士 · 启航教育", "成都青羊区 · 出国留学", "已沟通", "已沟通留学规划方案，客户想了解申请流程与成功案例。", "李明", "2024-05-27 14:30", "高", "已逾期"],
  ["周先生 · 未来学院", "成都天府新区 · 职业培训", "跟进中", "已提供课程大纲与就业数据，客户比较其他机构价格。", "王芳", "2024-05-30 10:00", "低", "待跟进"],
  ["孙女士 · 博雅教育", "成都高新区 · 素质教育", "意向确认", "意向较强，计划周末到校参观，需安排接待与课程体验。", "陈晨", "2024-05-31 09:30", "中", "待跟进"],
  ["黄先生 · 新航道学校", "成都金牛区 · 语言培训", "已成交", "已签订合作协议，正在推进开课准备与教材采购。", "赵磊", "2024-06-05 10:00", "低", "已完成"]
] as const;

const reminders = [
  ["15:00", "张女士 · 成都蓝鲸教育", "成都高新区 · 少儿英语", "跟进中"],
  ["16:30", "周先生 · 未来学院", "成都天府新区 · 职业培训", "跟进中"],
  ["17:00", "杨先生 · 智学教育", "成都锦江区 · K12辅导", "意向确认"],
  ["18:00", "吴女士 · 启明星教育", "成都青羊区 · 素质教育", "已沟通"],
  ["19:00", "郑先生 · 环球留学", "成都高新区 · 出国留学", "意向确认"]
] as const;

const recentUpdates = [
  ["李明 记录跟进：张女士 · 成都蓝鲸教育", "已介绍课程体系，客户感兴趣，待发资料", "1小时前"],
  ["陈晨 记录跟进：王先生 · 星辰国际学校", "客户认可方案，内部评估预算中", "2小时前"],
  ["赵磊 改期跟进：赵磊 · 优学教育", "客户临时有事，将跟进时间改至 5/28 11:00", "3小时前"],
  ["王芳 完成跟进：吴先生 · 博文教育", "已签订合作协议，进入开课准备阶段", "5小时前"]
] as const;

const stageLabels: Record<CrmStage, string> = {
  new: "新线索",
  contacted: "需求确认",
  qualified: "方案演示",
  proposal: "报价谈判",
  won: "已成交",
  lost: "已流失"
};

function toCustomerCard(customer: CrmCustomer): CustomerCard {
  const nextDate = customer.next_follow_up_at
    ? new Date(customer.next_follow_up_at).toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" })
    : "待安排";
  return {
    name: customer.name,
    owner: "张婧",
    value: "待评估",
    stage: stageLabels[customer.stage],
    health: customer.stage === "proposal" || customer.stage === "qualified" ? "高意向" : "可推进",
    next: `${nextDate} 跟进客户进展`,
    location: customer.source === "lead" ? "AI线索 · 待补地区" : customer.source,
    contact: customer.phone || "待补充",
    email: customer.email || "待补充",
    tags: [customer.source === "lead" ? "AI线索" : customer.source, customer.phone ? "电话可触达" : "待补联系方式"]
  };
}

function CrmPage({ variant = "customers" }: CrmPageProps) {
  const [dueCustomers, setDueCustomers] = useState<CrmCustomer[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    crmApi
      .listDueCustomers(20)
      .then((payload) => {
        if (active) setDueCustomers(payload.customers);
      })
      .catch((error) => {
        if (active) setError(apiErrorMessage(error, "暂时无法读取 CRM 客户"));
      });
    return () => {
      active = false;
    };
  }, []);

  const visibleCustomers = dueCustomers.length > 0 ? dueCustomers.map(toCustomerCard) : customers;

  if (variant === "followUps") {
    return <FollowUpsPage />;
  }

  const selectedCustomer = visibleCustomers[0];

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
        {crmStats.map(([label, value, icon]) => (
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
              {["全部客户", "高意向", "跟进中", "已成交"].map((item, index) => (
                <button className={index === 0 ? "active" : ""} key={item} type="button">{item}</button>
              ))}
            </div>
            <label>
              <span aria-hidden="true">⌕</span>
              <input aria-label="搜索客户" placeholder="搜索客户/公司/电话" />
            </label>
            <button type="button">来源筛选⌄</button>
          </header>

          <h2 className="sr-only">客户列表</h2>
          {error && <p className="form-error" role="alert">{error}</p>}
          <div className="cdk-crm-table" role="table" aria-label="客户列表">
            <div className="cdk-crm-table-head" role="row">
              {["", "客户 / 公司", "阶段", "标签", "最近跟进", "负责人", "下次跟进", "操作"].map((item) => (
                <span key={item} role="columnheader">{item}</span>
              ))}
            </div>
            {visibleCustomers.map((customer) => (
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
                <span>{customer.stage === "新线索" ? "昨天 09:15" : "今天 15:00"}</span>
                <span>{customer.owner}</span>
                <span>{customer.next}</span>
                <div className="cdk-crm-actions">
                  <Link to="/crm/follow-ups">查看</Link>
                  <button type="button" aria-label={`更多操作 ${customer.name}`}>•••</button>
                </div>
              </article>
            ))}
          </div>

          <footer className="cdk-crm-pagination">
            <span>共 12,845 条数据</span>
            <button type="button">10 条/页⌄</button>
            {[1, 2, 3, 4, 5].map((page) => <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>)}
            <button type="button">1285</button>
          </footer>
        </div>

        <aside className="cdk-crm-detail-card" aria-label="客户详情">
          <header>
            <h2>客户详情</h2>
            <span aria-hidden="true">⌖ ×</span>
          </header>
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
            <div><dt>来源</dt><dd>线索查找（关键词：少儿英语培训）</dd></div>
            <div><dt>需求摘要</dt><dd>希望为3-12岁孩子提供系统化英语课程，提升口语表达与应试能力。<button type="button">展开⌄</button></dd></div>
          </dl>
          <section className="cdk-crm-timeline">
            <h2>跟进看板</h2>
            {[
              ["今天 15:00", "电话沟通，介绍课程体系与教学服务，客户对外教口语课感兴趣，约定明天发送课程方案。"],
              ["昨天 10:30", "添加微信，初步了解需求，客户计划暑期班提升口语。"],
              ["05-23 16:45", "首次电话沟通，了解机构基本情况与需求。"]
            ].map(([time, detail]) => (
              <article key={time}>
                <time>{time} 李明</time>
                <p>{detail}</p>
              </article>
            ))}
            <Link to="/crm/follow-ups">查看全部跟进记录 ›</Link>
          </section>
          <footer>
            <button type="button">拨打电话</button>
            <button type="button">发消息</button>
            <Link to="/crm/follow-ups">记录跟进</Link>
          </footer>
        </aside>
      </section>
    </main>
  );
}

function FollowUpsPage() {
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
          ["今日待跟进", "68", "calendar"],
          ["本周待跟进", "356", "chart"],
          ["已沟通", "1,426", "message"],
          ["已成交跟进", "532", "check"]
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
              {["全部", "今日待跟进", "本周待跟进", "已沟通", "已成交"].map((item, index) => (
                <button className={index === 0 ? "active" : ""} key={item} type="button">{item}</button>
              ))}
            </div>
            <label>
              <input aria-label="搜索跟进记录" placeholder="搜索客户/公司/负责人" />
              <span aria-hidden="true">⌕</span>
            </label>
            <button type="button">阶段筛选⌄</button>
            <button type="button">时间范围⌄</button>
          </header>
          <div className="cdk-followups-table" role="table" aria-label="全部跟进列表">
            <div className="cdk-followups-row head" role="row">
              {["客户 / 公司", "当前阶段", "最近跟进内容", "负责人", "下次跟进时间", "优先级", "跟进状态", "操作"].map((item) => (
                <span key={item} role="columnheader">{item}</span>
              ))}
            </div>
            {followRows.map(([name, sub, stage, note, owner, next, priority, status]) => (
              <article className="cdk-followups-row" key={name} role="row">
                <div><strong>{name}</strong><small>{sub}</small></div>
                <span className={`stage ${stageTone(stage)}`}>{stage}</span>
                <p>{note}</p>
                <span>{owner}</span>
                <time>{next}<small>{next.includes("今天") ? "2小时后" : next.includes("明天") ? "21小时后" : ""}</small></time>
                <b className={`priority ${priority}`}>{priority}</b>
                <span className="follow-status">{status}</span>
                <div><Link to="/crm">查看详情</Link><button type="button">记录跟进</button><button type="button">改期</button></div>
              </article>
            ))}
          </div>
          <footer className="cdk-crm-pagination">
            <span>共 1,426 条记录</span>
            <button type="button">10 条/页⌄</button>
            {[1, 2, 3, 4, 5].map((page) => <button className={page === 1 ? "active" : ""} key={page} type="button">{page}</button>)}
            <button type="button">143</button>
          </footer>
        </div>

        <aside className="cdk-followups-sidebar">
          <section>
            <h2>跟进提醒</h2>
            <header><strong>今日待跟进（5）</strong><Link to="/crm/follow-ups">查看全部</Link></header>
            {reminders.map(([time, name, sub, status]) => (
              <article key={`${time}-${name}`}>
                <time>{time}</time>
                <span><strong>{name}</strong><small>{sub}</small></span>
                <b>{status}</b>
              </article>
            ))}
            <Link to="/tasks">查看全部日程</Link>
          </section>
          <section>
            <h2>最近更新</h2>
            {recentUpdates.map(([title, detail, time], index) => (
              <article className="update" key={title}>
                <i className={`dot-${index + 1}`} aria-hidden="true" />
                <span><strong>{title}</strong><small>{detail}</small></span>
                <time>{time}</time>
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
