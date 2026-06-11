import { Link } from "react-router-dom";

import { useAuthSession } from "../lib/authSession";

type ProductPlaceholderProps = {
  description: string;
  title: string;
};

function ProductPlaceholder({ description, title }: ProductPlaceholderProps) {
  const session = useAuthSession();

  return (
    <main className="product-placeholder">
      <header className="placeholder-header">
        <Link className="brand" to="/" aria-label="返回智活AI OPC 首页">
          <span className="brand-mark">智</span>
          <span>智活AI</span>
        </Link>
        <span>{session.user?.nickname}</span>
      </header>
      <section>
        <p className="section-eyebrow">第一版正在实现</p>
        <h1>{title}</h1>
        <p>{description}</p>
        <Link className="primary-action placeholder-action" to="/">
          返回首页
        </Link>
      </section>
    </main>
  );
}

export default ProductPlaceholder;
