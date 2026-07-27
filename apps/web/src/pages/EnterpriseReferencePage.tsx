import { Check } from "lucide-react";
import { useState, type ReactNode } from "react";
import { Link } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

export type EnterpriseReferenceVariant = "home" | "form" | "cases" | "detail" | "success" | "contact";

type EnterpriseReferencePageProps = {
  variant?: EnterpriseReferenceVariant;
};

const services = [
  ["战略诊断", "深度洞察企业现状与市场环境，梳理增长机会与关键突破口。", "pulse"],
  ["产品与流程设计", "优化产品定位与业务流程，建构双向驱动的增长体系。", "cube"],
  ["增长策略", "制定可落地的增长策略，整合渠道与资源，实现规模化增长。", "chart"],
  ["落地陪跑", "专家团队全程陪跑落地，跟踪执行，确保目标达成与效果闭环。", "rocket"]
] as const;

const cases = [
  ["美妆品牌全域增长陪跑项目", "从品牌定位到全域增长体系搭建，实现品牌心智与销量双提升。", "全球营收增长", "300%+"],
  ["B2B SaaS企业增长陪跑", "构建从线索到付费的增长闭环，驱动订阅转化与客户续费双增长。", "付费客户增长", "200%+"],
  ["大健康品牌数字化转型陪跑", "重塑产品矩阵与数字化运营体系，实现用户规模与复购率提升。", "复购率提升", "220%+"],
  ["新能源车企渠道增长项目", "优化渠道布局与线索经营模型，提升线索质量与门店转化效率。", "线索转化率提升", "180%+"],
  ["服装零售品牌增长陪跑", "打造会员增长与私域运营体系，实现GMV与会员LTV双增长。", "GMV增长", "250%+"],
  ["物流科技企业增长陪跑", "明确增长路径与产品价值主张，提升市场份额与客户留存率。", "市场份额提升", "150%+"]
] as const;

function EnterpriseReferencePage({ variant = "home" }: EnterpriseReferencePageProps) {
  const [contactOpen, setContactOpen] = useState(variant === "contact");

  return (
    <V4PageShell className="enterprise-ref-shell" showCopilotMini={false}>
      <section className={`enterprise-ref-page ${variant}`} aria-label="企业定制化陪跑参考页面">
        {(variant === "home" || variant === "contact") && <EnterpriseHome onContact={() => setContactOpen(true)} />}
        {variant === "form" && <EnterpriseForm />}
        {variant === "cases" && <EnterpriseCases />}
        {variant === "detail" && <EnterpriseCaseDetail onContact={() => setContactOpen(true)} />}
        {variant === "success" && <EnterpriseSuccess />}
      </section>
      {contactOpen ? <ContactModal onClose={() => setContactOpen(false)} /> : null}
    </V4PageShell>
  );
}

function EnterpriseHome({ onContact }: { onContact: () => void }) {
  return (
    <div className="er-home-layout">
      <main>
        <section className="er-home-hero">
          <div>
            <div className="er-h1">企业定制化陪跑</div>
            <h2>为企业提供从战略、产品、增长到落地执行的深度陪跑服务</h2>
            <p>专业顾问团队 × 方法论体系 × 数据洞察，陪伴企业突破增长瓶颈，实现可持续增长。</p>
            <div className="er-actions">
              <button onClick={onContact} type="button">联系企业微信咨询</button>
              <Link to="/enterprise/cases">查看服务案例</Link>
            </div>
          </div>
          <div className="er-home-art" aria-hidden="true" />
        </section>

        <section className="er-summary-grid">
          <article><b>●</b><div><h3>适合谁</h3><p>成长型企业，寻求突破增长瓶颈</p><p>业务转型，探索新增长曲线</p><p>需要外部专业团队深度协同支持</p></div></article>
          <article><b>◇</b><div><h3>交付方式</h3><p>专家团队一对一定制陪跑</p><p>线上协作 + 线下工作坊结合</p><p>阶段性交付，可视化进度管理</p></div></article>
          <article><b>▦</b><div><h3>服务流程</h3><p>诊断评估，明确增长问题</p><p>定制方案，分阶段执行落地</p><p>持续优化，复盘与迭代升级</p></div></article>
        </section>

        <h2 className="er-section-title">服务内容</h2>
        <section className="er-service-grid">
          {services.map(([title, detail, icon]) => <article key={title}><i className={icon} /><div><h3>{title}</h3><p>{detail}</p></div></article>)}
        </section>

        <section className="er-home-bottom">
          <div>
            <h2 className="er-section-title">典型服务案例</h2>
            <div className="er-mini-cases">
              {cases.slice(0, 3).map(([title], index) => <Link key={title} to="/enterprise/cases/consumer-growth"><span style={{ backgroundImage: `url(/enterprise-reference/case-${index + 1}.jpg)` }} /><strong>{title}</strong><small>6个月实现全域增长突破</small></Link>)}
            </div>
          </div>
          <div>
            <h2 className="er-section-title">合作流程</h2>
            <div className="er-process">
              {["需求沟通", "方案定制", "执行落地", "复盘优化"].map((item, index) => <article key={item}><b>0{index + 1}</b><i>{index === 0 ? "●" : index === 1 ? "▤" : index === 2 ? "▲" : "▧"}</i><strong>{item}</strong><small>{index === 0 ? "深入沟通，明确目标与挑战" : index === 1 ? "诊断评估，定制陪跑方案" : index === 2 ? "分阶段执行，跟踪与支持" : "复盘总结，持续优化迭代"}</small></article>)}
            </div>
          </div>
        </section>
      </main>
      <ConsultantCard action={<Link to="/enterprise/form">填写需求表单</Link>} />
    </div>
  );
}

function EnterpriseForm() {
  return (
    <form className="er-form-page">
      <header><h1>填写需求表单</h1><p>请详细填写以下信息，以便我们更精准地理解您的需求，为您提供更贴合的定制化陪跑服务。</p></header>
      <div className="er-form-grid">
        <label><span>公司名称 *</span><input placeholder="请输入公司全称" /></label>
        <label><span>所属行业 *</span><select defaultValue=""><option value="" disabled>请选择所属行业</option><option>互联网 / 科技</option></select></label>
        <label><span>企业阶段 *</span><select defaultValue=""><option value="" disabled>请选择企业阶段</option><option>成长期</option></select></label>
        <label><span>联系人姓名 *</span><input placeholder="请输入联系人姓名" /></label>
        <label><span>联系方式 *</span><input placeholder="请输入手机号" /></label>
        <label><span>企业邮箱 *</span><input placeholder="请输入企业邮箱" /></label>
      </div>
      <label><span>需求描述 *</span><textarea placeholder="请详细描述您的业务背景、核心痛点、目标期望及希望我们协助解决的问题（建议不少于30字）" /></label>
      <fieldset><legend>需求类型（可多选）*</legend>{["战略诊断", "产品设计", "增长策略", "数字化转型", "组织机制", "其他"].map(item => <label key={item}><input type="checkbox" />{item}</label>)}</fieldset>
      <label><span>附件上传（可选）</span><div className="er-upload"><b>♧</b> 点击上传或将文件拖拽到此处<small>支持 PDF、PPT、Word、Excel、图片格式，单个文件不超过 20MB</small></div></label>
      <label className="er-agreement"><input type="checkbox" />我已阅读并同意《隐私政策》和《服务协议》，并同意智活AI在为我提供服务过程中处理相关信息。</label>
      <Link className="er-submit" to="/enterprise/success">提交需求</Link>
    </form>
  );
}

function EnterpriseCases() {
  return (
    <main className="er-cases-page">
      <header><h1>服务案例</h1><p>精选企业增长服务案例，展示真实项目成果与价值创造，助力企业找到适合的增长解决方案。</p></header>
      <div className="er-case-toolbar"><nav>{["全部案例", "战略诊断", "产品设计", "增长策略", "落地执行", "行业筛选⌄"].map((item, index) => <button className={index === 0 ? "active" : ""} key={item} type="button">{item}</button>)}</nav><label>⌕ <input placeholder="搜索案例标题/行业/关键词" /></label></div>
      <section className="er-case-grid">
        {cases.map(([title, detail, metric, value], index) => <Link key={title} to="/enterprise/cases/consumer-growth"><span style={{ backgroundImage: `url(/enterprise-reference/case-${index + 1}.jpg)` }} /><h2>{title}</h2><div><em>战略诊断</em><em>增长策略</em><em>落地执行</em></div><p>{detail}</p><footer><strong>{metric}</strong><b>{value}</b></footer></Link>)}
      </section>
      <footer className="er-pagination">‹ <b>1</b> 2 3 4 5 … 10 › <span>10条/页⌄</span><small>共 98 条案例</small></footer>
    </main>
  );
}

function EnterpriseCaseDetail({ onContact }: { onContact: () => void }) {
  return (
    <div className="er-detail-layout">
      <main>
        <Link className="er-back" to="/enterprise/cases">← 返回案例列表</Link>
        <section className="er-detail-hero"><div><h1>某快消品牌增长陪跑项目</h1><div><span>快消品</span><span>全域增长</span><span>新品上市</span><span>线下渠道</span></div><p>▣ 项目周期：2024.06 - 2024.12（6个月） ♙ 合作团队：品牌方增长团队 8人 ｜ 智活AI项目组 6人</p></div><i /></section>
        <section className="er-detail-card compact"><h2>▧ 项目背景</h2><p>客户为国内领先的快消品牌，计划在竞争激烈的饮料赛道推出新品。面临新品认知度低、渠道覆盖不足、线上线下联动弱、转化效率不高等挑战，需要一套数据驱动的增长策略与落地执行方案，快速实现新品破圈与销量增长。</p></section>
        <section className="er-detail-card solution"><h2>♧ 我们的解决方案</h2><i /><ul><li><b>市场与用户洞察：</b>多维数据分析，识别核心人群与场景机会，精准定位新品卖点。</li><li><b>全域策略制定：</b>制定“内容种草 + 直播转化 + 线下动销”一体化增长策略。</li><li><b>内容与投放优化：</b>基于数据持续优化内容策略与投放组合，提升ROI。</li><li><b>渠道与终端赋能：</b>优化渠道铺货策略，赋能门店动销，打通线上线下数据链路。</li><li><b>全程陪跑落地：</b>敏捷迭代执行方案，周度复盘优化，确保目标达成。</li></ul></section>
        <h2 className="er-section-title">项目成果</h2>
        <section className="er-result-grid">{[["300%+", "新品首月销售额增长", "对比目标达成率"], ["150%+", "全域ROI提升", "投放效率显著提升"], ["80%+", "新增核心用户增长", "核心人群规模大幅提升"], ["200%+", "线下动销增长", "终端动销率提升"]].map(([value, title, detail]) => <article key={title}><i>◈</i><div><strong>{value}</strong><b>{title}</b><small>{detail}</small></div></article>)}</section>
        <section className="er-testimonial"><h2>客户评价</h2><p>“智活AI团队对我们的业务理解非常深入，方案专业且落地性强。从策略制定到执行陪跑，全程高效协同，帮助我们新品快速打开市场，销售与品牌声量双双超预期，是值得长期合作的增长伙伴！”</p><strong>某快消品牌 <small>市场总监</small></strong></section>
      </main>
      <ConsultantCard action={<button onClick={onContact} type="button">联系企业增长顾问</button>} detail />
    </div>
  );
}

function EnterpriseSuccess() {
  return (
    <main className="er-success-page">
      <div className="er-success-icon" aria-hidden="true">
        <i /><i /><i /><i /><i /><i />
        <Check size={58} strokeWidth={5} />
      </div>
      <h1>需求提交成功！</h1>
      <p>感谢您的信任！我们已收到您的需求，专业顾问团队将在 <b>24 小时内</b> 与您联系，<br />为您提供专属解决方案。</p>
      <section className="er-success-info"><h2>您的提交信息</h2>{[["▥", "企业名称", "上海智活科技有限公司"], ["♟", "联系人", "张婧"], ["☎", "联系方式", "138 **** 5678"], ["▣", "提交时间", "2024-06-01 14:30:25"]].map(([icon, label, value]) => <article key={label}><i>{icon}</i><strong>{label}</strong><span>{value}</span></article>)}</section>
      <h2>我们将为您</h2><section className="er-success-services">{[["♙", "专业顾问 1对1联系沟通", "资深行业顾问将尽快与您取得联系，深入了解您的业务需求与目标。"], ["◇", "定制方案评估", "结合行业最佳实践与数据洞察，为您量身定制可落地的增长解决方案。"], ["▣", "排期跟踪服务", "全流程进度跟踪，关键节点及时同步，确保项目高效推进与落地。"]].map(([icon, title, detail]) => <article key={title}><i>{icon}</i><div><strong>{title}</strong><p>{detail}</p></div></article>)}</section>
      <Link to="/">⌂ 返回首页</Link>
    </main>
  );
}

function ConsultantCard({ action, detail = false }: { action: ReactNode; detail?: boolean }) {
  return <aside className={`er-consultant ${detail ? "detail" : ""}`}><h2>{detail ? "您的企业增长顾问" : "专属咨询顾问"}</h2><img alt="企业增长顾问张婧" src="/enterprise-reference/advisor.jpg" /><h3>张婧 · 智活AI</h3><span>企业增长顾问</span><p>10年+企业增长与数字化咨询经验，深耕快消、新消费领域，擅长从战略规划到执行的全链路增长陪跑，已助力100+企业实现持续增长。</p><hr /><h3>企业微信咨询</h3><img className="qr" alt="企业微信二维码" src="/public-components/service-qr.jpg" />{action}{detail ? <small>工作时间：9:00 - 18:00（工作日）</small> : null}</aside>;
}

function ContactModal({ onClose }: { onClose: () => void }) {
  return <div className="er-contact-backdrop"><section className="er-contact-modal" role="dialog" aria-modal="true" aria-label="联系专属咨询顾问"><button aria-label="关闭" onClick={onClose} type="button">×</button><h2>联系专属咨询顾问</h2><p>添加企业微信，获取专属服务</p><img alt="专属咨询顾问企业微信二维码" src="/public-components/service-qr.jpg" /><strong>▢ 微信扫一扫 添加好友</strong><span>◷ 工作时间：9:00-18:00（周一至周日）</span></section></div>;
}

export default EnterpriseReferencePage;
