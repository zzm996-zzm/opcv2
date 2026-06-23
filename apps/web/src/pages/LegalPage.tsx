import { Link } from "react-router-dom";

type LegalPageProps = {
  kind: "terms" | "privacy";
};

const content = {
  terms: {
    eyebrow: "使用说明",
    title: "用户协议",
    intro:
      "欢迎使用智活AI OPC。第一版服务用于帮助用户整理商业方向、行动计划和公开企业线索。",
    sections: [
      ["服务边界", "AI 生成内容仅供商业分析参考，不构成投资、法律或财务建议。"],
      ["账号使用", "请使用本人账号与密码登录，并妥善保管账号访问权限。"],
      ["公开数据", "企业线索来自公开或已获授权的数据源，用户应自行判断其准确性与适用性。"],
      ["合规联系", "用户对后续联系行为负责，不得发送骚扰信息或以违法方式使用线索。"]
    ]
  },
  privacy: {
    eyebrow: "数据说明",
    title: "隐私政策",
    intro:
      "我们只收集运行第一版产品所必需的信息，并尽量缩短数据链路和访问范围。",
    sections: [
      ["账号信息", "账号、密码与可选联系方式仅用于登录、安全校验和必要的服务通知。"],
      ["输入内容", "你提交的商业背景与需求用于生成分析结果，不会作为公开资料展示。"],
      ["Cookie", "Refresh Token 通过 HttpOnly Cookie 保存，前端脚本无法直接读取。"],
      ["公开企业信息", "产品展示来源链接与采集时间，方便用户回到原始页面核对。"]
    ]
  }
} as const;

function LegalPage({ kind }: LegalPageProps) {
  const page = content[kind];

  return (
    <main className="legal-page">
      <Link className="brand" to="/" aria-label="返回智活AI OPC 首页">
        <span className="brand-mark">智</span>
        <span>智活AI</span>
      </Link>
      <article className="legal-card">
        <p className="section-eyebrow">{page.eyebrow}</p>
        <h1>{page.title}</h1>
        <p className="legal-intro">{page.intro}</p>
        {page.sections.map(([title, body]) => (
          <section key={title}>
            <h2>{title}</h2>
            <p>{body}</p>
          </section>
        ))}
        <p className="legal-note">
          当前为 MVP 说明版本，正式上线前将根据实际服务范围与法律意见更新。
        </p>
      </article>
    </main>
  );
}

export default LegalPage;
