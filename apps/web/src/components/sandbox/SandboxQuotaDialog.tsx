import { Clock3, LockKeyhole, X } from "lucide-react";
import { useEffect, useRef } from "react";

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
      <section aria-label="沙盘额度功能即将开放" aria-modal="true" className="sb-quota-dialog" onMouseDown={(event) => event.stopPropagation()} role="dialog">
        <button aria-label="关闭额度弹窗" className="sb-dialog-close" onClick={onClose} ref={closeRef} type="button"><X size={19} /></button>
        <div className="sb-quota-lock" aria-hidden="true"><LockKeyhole size={48} strokeWidth={1.6} /></div>
        <div className="sb-quota-heading">
          <h2>额度功能即将开放</h2>
          <p>商业沙盘 V1.2 当前不限次数，正常推演不会触发本弹窗。</p>
        </div>
        <div className="sb-quota-compare">
          <article><strong>当前版本</strong><b>不限</b><small>推演次数</small></article>
          <span>VS</span>
          <article className="is-member"><strong><Clock3 size={15} />后续版本</strong><b>待定</b><small>以正式公告为准</small></article>
        </div>
        <p className="sb-quota-reset">兼容字段：{used} / {limit}，本期不用于拦截</p>
        <div className="sb-quota-upgrade-title">本期所有登录用户均可使用完整功能</div>
        <div className="sb-quota-benefits">
          <span><b>不限</b><small>推演次数</small></span>
          <span><b>不限</b><small>历史记录保存</small></span>
          <span><b>完整</b><small>结构化报告</small></span>
          <span><b>优先</b><small>更快生成速度</small></span>
        </div>
        <footer>
          <button disabled type="button"><Clock3 size={16} />即将开放</button>
          <button className="is-secondary" onClick={onClose} type="button">我知道了</button>
        </footer>
        <button className="sb-quota-later" onClick={onClose} type="button">稍后再说</button>
      </section>
    </div>
  );
}

export default SandboxQuotaDialog;
