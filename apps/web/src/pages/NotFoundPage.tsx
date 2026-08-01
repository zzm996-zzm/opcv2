import { Link, useNavigate } from "react-router-dom";

import V4PageShell from "../components/V4PageShell";

function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <V4PageShell className="not-found-shell">
      <section className="not-found-card" aria-labelledby="not-found-title">
        <span className="not-found-code" aria-hidden="true">404</span>
        <div>
          <h1 id="not-found-title">页面不存在</h1>
          <p>这个地址可能已经变化，或者页面暂时不可用。</p>
          <div className="not-found-actions">
            <Link className="module-primary-action" to="/">返回工作台</Link>
            <button onClick={() => navigate(-1)} type="button">返回上一页</button>
          </div>
        </div>
      </section>
    </V4PageShell>
  );
}

export default NotFoundPage;
