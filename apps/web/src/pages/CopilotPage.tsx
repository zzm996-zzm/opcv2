import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

export type CopilotVariant = "home" | "new" | "models" | "files" | "compare" | "rename" | "delete";

const quickActions = [
  ["trend", "分析项目机会"],
  ["briefcase", "推荐工具"],
  ["doc", "制定落地计划"],
  ["page", "总结当前页面"]
] as const;

type ConversationItem = {
  title: string;
  desc: string;
  time: string;
  active?: boolean;
  menu?: boolean;
  group?: string;
};

const conversations: ConversationItem[] = [
  { title: "智能客服系统项目机会分析", desc: "分析市场机会、推荐工具与...", time: "10:32", active: true, menu: true },
  { title: "竞争对手监测方案设计", desc: "如何搭建竞品监测体系?", time: "09:15" },
  { title: "CRM客户管理落地计划", desc: "制定阶段性落地路线图", time: "08:47" },
  { title: "数据资产治理方法论", desc: "企业数据治理的5步进阶步骤", time: "16:22", group: "昨天" },
  { title: "GEO获客策略建议", desc: "针对SaaS产品的获客策略", time: "14:08" },
  { title: "AI教学课程内容设计", desc: "设计面向销售团队的AI课程", time: "11:30" },
  { title: "商业沙盘模拟复盘", desc: "本次沙盘的关键复盘点", time: "06-24", group: "更早" },
  { title: "增长测算模型搭建", desc: "建立业务增长测算模型", time: "06-23" }
] as const;

type CopilotModel = {
  name: string;
  icon: "swirl" | "ai" | "black";
  selected?: boolean;
};

const models: CopilotModel[] = [
  { name: "GPT-4o", icon: "swirl", selected: true },
  { name: "Claude opus4.8", icon: "ai" },
  { name: "Grok4.3", icon: "black" }
] as const;

type ReferenceItem = {
  section: string;
  title: string;
  meta: string;
  kind: "pdf" | "sheet" | "chat";
  checked: boolean;
  current?: boolean;
  time?: string;
};

const referenceItems: ReferenceItem[] = [
  { section: "当前页面", title: "智能客服市场分析报告.pdf", meta: "PDF · 1.8 MB", kind: "pdf", checked: true, current: true },
  { section: "最近上传文件", title: "智能客服竞品功能对比表.xlsx", meta: "XLSX · 320 KB", kind: "sheet", checked: false, time: "昨天" },
  { section: "历史对话", title: "竞争对手监测方案设计", meta: "对话 · 2024-06-20 09:15", kind: "chat", checked: false },
  { section: "我的内容", title: "客户成功案例：某政务热线升...", meta: "文档 · 昨天", kind: "sheet", checked: false }
] as const;

const comparisonAnswers = [
  {
    model: "GPT-4o",
    icon: "swirl",
    sections: [
      ["市场规模与增长趋势", "2024年中国智能客服市场规模约为 95.2 亿元，预计到 2027 年将达 181.6 亿元，年复合增长率约 24.0%。"],
      ["竞争格局", "头部集中且持续分化，阿里云、腾讯云、百度智能云、华为云等占据主要市场份额。"],
      ["客户需求", "降本增效、提升客户体验、全渠道整合与个性化服务成为核心诉求。"],
      ["机会点", "大模型驱动的智能化升级、垂直行业解决方案、出海与多语言服务是主要机会。"]
    ]
  },
  {
    model: "Claude opus4.8",
    icon: "ai",
    sections: [
      ["市场规模与增长趋势", "2024 市场规模约 92.3 亿元，受大模型普及推动，预计 2027 年达 175.8 亿元，CAGR 为 23.3%。"],
      ["竞争格局", "市场呈现“一超多强”格局，云厂商+AI厂商+SaaS厂商协同竞争。"],
      ["客户需求", "更注重智能化水平，尤其是 AI 理解与生成能力，和业务闭环效果。"],
      ["机会点", "AI 原生应用、行业 Know-how 沉淀、数据安全与合规能力将形成差异化壁垒。"]
    ]
  },
  {
    model: "Grok4.3",
    icon: "black",
    sections: [
      ["市场规模与增长趋势", "2024 年市场规模约 98.7 亿元，增长强劲，预计 2027 年突破 190 亿元，年复合增长率 24.8%。"],
      ["竞争格局", "竞争激烈，头部厂商加速布局大模型与全渠道，区域性厂商在细分行业突围。"],
      ["客户需求", "对实时响应、复杂问题解决和数据分析洞察的需求显著提升。"],
      ["机会点", "多模态交互、客服+营销一体化、智能体落地是关键机会。"]
    ]
  }
] as const;

function CopilotPage({ variant = "home" }: { variant?: CopilotVariant }) {
  const isNew = variant === "new";
  const isCompare = variant === "compare";
  const showModelPicker = variant === "models";
  const showReferencePicker = variant === "files";
  const showRename = variant === "rename";
  const showDelete = variant === "delete";

  return (
    <V4PageShell className="copilot-shell" showCopilotMini={false}>
      <section
        className={"copilot-workbench " + (isCompare ? "compare-mode " : "") + (showRename || showDelete ? "modal-open" : "")}
        aria-label="智活 Copilot 工作台"
      >
        <div className="copilot-main-panel">
          {isCompare ? <CompareHeader /> : <CopilotHeader />}

          <div className="copilot-chat-area">
            {isNew ? <EmptyConversation /> : isCompare ? <CompareConversation /> : <ChatThread />}
          </div>

          <Composer showModelPicker={showModelPicker} showReferencePicker={showReferencePicker} compare={isCompare} />
        </div>

        <ConversationSidebar showThreadMenu={showModelPicker} />
        {showRename && <RenameDialog />}
        {showDelete && <DeleteDialog />}
      </section>
    </V4PageShell>
  );
}

function CopilotHeader() {
  return (
    <header className="copilot-workbench-head">
      <div className="copilot-brand-mark" aria-hidden="true">
        <span className="v4-logo" />
      </div>
      <div>
        <h1>智活 Copilot</h1>
        <p>你的全球 AI 助手，随时为你提供专业的分析与建议</p>
      </div>
      <nav className="copilot-quick-actions" aria-label="Copilot 快捷指令">
        {quickActions.map(([icon, label]) => (
          <button key={label} type="button">
            <span className={"copilot-ui-icon " + icon} aria-hidden="true" />
            {label}
          </button>
        ))}
      </nav>
    </header>
  );
}

function CompareHeader() {
  return (
    <header className="copilot-compare-head">
      <div className="copilot-compare-icon" aria-hidden="true">
        <span className="copilot-ui-icon trend" />
      </div>
      <div>
        <h1>AI 对比分析</h1>
        <p>同时对比多个 AI 模型的回答，获得更全面、客观的洞察</p>
      </div>
      <div className="copilot-compare-settings" aria-label="对比设置">
        <strong>对比设置:</strong>
        <button type="button">2 模型</button>
        <button className="active" type="button">3 模型</button>
        {models.map((model) => (
          <button key={model.name} type="button">
            <span className={"model-glyph " + model.icon} aria-hidden="true" />
            {model.name}
            <i aria-hidden="true">×</i>
          </button>
        ))}
        <button type="button">
          <b aria-hidden="true">＋</b>
          选择模型
        </button>
      </div>
    </header>
  );
}

function ChatThread() {
  return (
    <div className="copilot-chat-thread" aria-label="会话内容">
      <article className="copilot-message user">
        <p>请帮我分析智能客服系统的市场机会和竞争格局。</p>
        <time>10:32 ✓✓</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>

      <article className="copilot-message assistant">
        <span className="v4-logo" aria-hidden="true" />
        <div className="copilot-bubble compact">
          <p>好的，我将从市场规模、增长趋势、竞争格局、客户需求与机会点四个维度为你分析智能客服系统的市场机会。</p>
          <time>10:32</time>
        </div>
      </article>

      <article className="copilot-message assistant report">
        <span className="v4-logo" aria-hidden="true" />
        <div className="copilot-bubble report-card">
          <h2>一、市场规模与增长趋势</h2>
          <ul>
            <li>2024年中国智能客服市场规模约为 95.2 亿元，预计 2027 年将达到 181.6 亿元，年复合增长率约 24.0%。</li>
            <li>受益于企业降本增效、用户体验提升与大模型技术普及，市场保持高速增长。</li>
          </ul>
          <h2>二、竞争格局</h2>
          <ul>
            <li>第一梯队：阿里云、腾讯云、百度智能云、华为云等，具备强大技术与生态能力。</li>
            <li>第二梯队：容联云、智齿科技、环信等，聚焦垂直场景与中大型客户。</li>
            <li>新兴玩家：大量AI原生创业公司，依托大模型+场景化能力切入细分赛道。</li>
          </ul>
          <div className="copilot-file-chip">
            <span className="pdf-thumb" aria-hidden="true">PDF</span>
            <span>
              <strong>智能客服市场分析报告.pdf</strong>
              <small>PDF · 1.8 MB</small>
            </span>
          </div>
          <time>10:33</time>
        </div>
      </article>

      <article className="copilot-message user lower">
        <p>请基于上面的分析，推荐适合我们的工具和落地路径。</p>
        <time>10:34 ✓✓</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>

      <article className="copilot-message assistant thinking">
        <span className="v4-logo" aria-hidden="true" />
        <div className="copilot-thinking">
          正在思考中
          <span />
          <span />
          <span />
          <span />
        </div>
      </article>
    </div>
  );
}

function EmptyConversation() {
  return (
    <div className="copilot-empty-state">
      <div className="copilot-empty-card">
        <div className="empty-orbit" aria-hidden="true">
          <span />
          <i />
        </div>
        <h1>开始一段新的对话</h1>
        <p>向智活 Copilot 提问，获取专业的分析与建议</p>
        <div className="copilot-prompt-list">
          {["分析一个新项目机会", "帮我制定执行计划", "总结当前页面"].map((prompt) => (
            <button key={prompt} type="button">
              {prompt}
              <span aria-hidden="true">↗</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

function CompareConversation() {
  return (
    <div className="copilot-comparison">
      <article className="copilot-message user compare-question">
        <p>请分析 2024 年中国智能客服市场的规模、增长趋势、竞争格局、客户需求与机会点。</p>
        <time>10:35 ✓</time>
        <span className="copilot-avatar user-avatar" aria-hidden="true">张</span>
      </article>
      <div className="comparison-grid">
        {comparisonAnswers.map((answer) => (
          <article key={answer.model} className="comparison-card">
            <header>
              <span className={"model-glyph " + answer.icon} aria-hidden="true" />
              <h2>{answer.model}</h2>
              <b>回答完成</b>
              <time>10:35</time>
            </header>
            {answer.sections.map(([title, body], index) => (
              <section key={title}>
                <h3>{index + 1}、{title}</h3>
                <p>{body}</p>
              </section>
            ))}
          </article>
        ))}
      </div>
      <button className="compare-summary-button" type="button">
        <span className="copilot-ui-icon doc" aria-hidden="true" />
        总结本次对比分析
      </button>
    </div>
  );
}

function Composer({
  showModelPicker,
  showReferencePicker,
  compare
}: {
  showModelPicker?: boolean;
  showReferencePicker?: boolean;
  compare?: boolean;
}) {
  return (
    <div className="copilot-composer-wrap">
      {showModelPicker && <ModelPicker />}
      {showReferencePicker && <ReferencePicker />}
      <form className="copilot-composer" aria-label="Copilot 输入框">
        <label className="sr-only" htmlFor="copilot-question">输入你的问题</label>
        <input id="copilot-question" placeholder="输入你的问题，Enter 发送、Shift + Enter 换行" />
        <div className="composer-toolbar">
          <button type="button">
            <span className="copilot-ui-icon clip" aria-hidden="true" />
            上传文件
          </button>
          <button className={showReferencePicker ? "active" : ""} type="button">
            <span className="copilot-ui-icon link" aria-hidden="true" />
            引用
          </button>
          <button className={showModelPicker ? "active" : ""} type="button" disabled={compare}>
            <span className="model-glyph swirl" aria-hidden="true" />
            GPT-4o
            <span aria-hidden="true">{showModelPicker ? "⌃" : "⌄"}</span>
          </button>
          <button className="compare-chip active" type="button">
            <span className="copilot-ui-icon doc" aria-hidden="true" />
            AI 对比分析
            <b>New</b>
          </button>
          <button type="button">
            <span className="copilot-ui-icon think" aria-hidden="true" />
            深度思考
            <b className="vip">VIP</b>
          </button>
          <span className="composer-spacer" />
          <button aria-label="语音输入" className="icon-only" type="button">
            <span className="mic-icon" aria-hidden="true" />
          </button>
          <button aria-label="发送" className="send-button" type="button">
            <span aria-hidden="true">↗</span>
          </button>
        </div>
      </form>
      <p className="ai-disclaimer">ⓘ 内容由 AI 生成，请注意甄别准确性</p>
    </div>
  );
}

function ModelPicker() {
  return (
    <div className="copilot-popover model-picker" role="dialog" aria-label="模型选择">
      {models.map((model) => (
        <button key={model.name} className={model.selected ? "selected" : ""} type="button">
          <span className={"model-glyph " + model.icon} aria-hidden="true" />
          {model.name}
          {model.selected && <b aria-hidden="true">✓</b>}
        </button>
      ))}
    </div>
  );
}

function ReferencePicker() {
  let previousSection = "";

  return (
    <div className="copilot-popover reference-picker" role="dialog" aria-label="引用">
      <h2>引用</h2>
      <label className="reference-search">
        <span aria-hidden="true">⌕</span>
        <input placeholder="搜索文件、对话或我的内容" />
      </label>
      <div className="reference-list">
        {referenceItems.map((item) => {
          const showSection = item.section !== previousSection;
          previousSection = item.section;

          return (
            <div key={item.title}>
              {showSection && <h3>{item.section}</h3>}
              <label className={item.checked ? "checked" : ""}>
                <input checked={item.checked} readOnly type="checkbox" />
                <span className={"reference-file " + item.kind} aria-hidden="true" />
                <span>
                  <strong>{item.title}</strong>
                  <small>{item.meta}</small>
                </span>
                {item.current && <b>当前</b>}
                {item.time && <em>{item.time}</em>}
              </label>
            </div>
          );
        })}
      </div>
      <footer>
        <span>已选择 1 项</span>
        <button type="button">插入引用</button>
      </footer>
    </div>
  );
}

function ConversationSidebar({ showThreadMenu }: { showThreadMenu?: boolean }) {
  let lastGroup = "";

  return (
    <aside className="copilot-history" aria-label="会话记录">
      <header>
        <button aria-label="收起会话记录" type="button">»</button>
        <Link to="/copilot/new">＋ 新建会话</Link>
      </header>
      <label className="history-search">
        <span aria-hidden="true">⌕</span>
        <input placeholder="搜索会话" />
      </label>
      <div className="history-list">
        {conversations.map((item, index) => {
          const group = item.group || (index < 3 ? "今天" : "");
          const showGroup = group && group !== lastGroup;
          if (group) lastGroup = group;

          return (
            <div key={item.title}>
              {showGroup && <h2>{group}</h2>}
              <article className={item.active ? "active" : ""}>
                <Link to="/copilot">
                  <strong>{item.title}</strong>
                  <span>{item.desc}</span>
                </Link>
                <time>{item.time}</time>
                {item.menu && (
                  <button aria-label="会话更多操作" className="history-more" type="button">
                    ⋮
                  </button>
                )}
              </article>
            </div>
          );
        })}
      </div>
      {showThreadMenu && (
        <div className="thread-action-menu" role="menu" aria-label="会话操作">
          <Link role="menuitem" to="/copilot/rename">重命名</Link>
          <Link role="menuitem" to="/copilot/delete">删除</Link>
        </div>
      )}
    </aside>
  );
}

function RenameDialog() {
  return (
    <div className="copilot-modal-backdrop">
      <section className="copilot-dialog rename-dialog" role="dialog" aria-modal="true" aria-label="重命名会话">
        <button aria-label="关闭" className="dialog-close" type="button">×</button>
        <h2>重命名会话</h2>
        <label className="sr-only" htmlFor="conversation-name">会话名称</label>
        <input id="conversation-name" defaultValue="智能客服系统项目机会分析" />
        <p>修改后仅影响会话标题，不影响对话内容</p>
        <footer>
          <button type="button">取消</button>
          <button className="primary" type="button">确认重命名</button>
        </footer>
      </section>
    </div>
  );
}

function DeleteDialog() {
  return (
    <div className="copilot-modal-backdrop">
      <section className="copilot-dialog delete-dialog" role="dialog" aria-modal="true" aria-label="删除会话">
        <button aria-label="关闭" className="dialog-close" type="button">×</button>
        <div className="delete-warning" aria-hidden="true">!</div>
        <h2>删除会话</h2>
        <p>删除后，该会话的所有内容（包括对话记录及引用的附件）将从历史会话列表中移除，且无法恢复。</p>
        <p>此操作不可撤销，请谨慎确认。</p>
        <label className="confirm-delete">
          <input type="checkbox" />
          我已确认删除此会话
        </label>
        <footer>
          <button type="button">取消</button>
          <button className="danger" type="button">确认删除</button>
        </footer>
      </section>
    </div>
  );
}

export default CopilotPage;
