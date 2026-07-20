import { Crown, LockKeyhole, X } from "lucide-react";
import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";

type SandboxQuotaDialogProps = {
  limit: number;
  onClose: () => void;
  used: number;
};

function SandboxQuotaDialog({ limit, onClose, used }: SandboxQuotaDialogProps) {
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    closeRef.current?.focus();
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [onClose]);

  return (
    <div className="sb-modal-scrim" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <section aria-label="本月沙盘次数已用尽" aria-modal="true" className="sb-quota-dialog" onMouseDown={(event) => event.stopPropagation()} role="dialog">
        <button aria-label="关闭额度弹窗" className="sb-dialog-close" onClick={onClose} ref={closeRef} type="button"><X size={19} /></button>
        <div className="sb-quota-lock" aria-hidden="true"><LockKeyhole size={48} strokeWidth={1.6} /></div>
        <div className="sb-quota-heading">
          <h2>本月沙盘次数已用尽</h2>
          <p>您已用完普通版每月可用的沙盘推演次数。</p>
        </div>
        <div className="sb-quota-compare">
          <article><strong>普通版</strong><b>{used}/{limit || 1}</b><small>次 / 月</small></article>
          <span>VS</span>
          <article className="is-member"><strong><Crown size={15} />会员版</strong><b>20</b><small>次 / 月</small></article>
        </div>
        <p className="sb-quota-reset">本月已使用：{used} / {limit || 1} 次，重置时间以会员中心为准</p>
        <div className="sb-quota-upgrade-title">升级会员版，立即享受更多权益</div>
        <div className="sb-quota-benefits">
          <span><b>20</b><small>次/月推演额度</small></span>
          <span><b>长期</b><small>历史记录保存</small></span>
          <span><b>完整</b><small>结构化报告</small></span>
          <span><b>优先</b><small>更快生成速度</small></span>
        </div>
        <footer>
          <Link to="/membership/upgrade"><Crown size={16} />升级套餐</Link>
          <Link className="is-secondary" to="/help"><span aria-hidden="true">?</span>联系顾问</Link>
        </footer>
        <button className="sb-quota-later" onClick={onClose} type="button">稍后再说</button>
      </section>
    </div>
  );
}

export default SandboxQuotaDialog;

