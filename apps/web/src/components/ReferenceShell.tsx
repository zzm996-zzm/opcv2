import type { ReactNode } from "react";

import V4PageShell from "./V4PageShell";

type ReferenceShellProps = {
  accountSlot?: ReactNode;
  children: ReactNode;
  className?: string;
  mainClassName?: string;
};

function ReferenceShell({ accountSlot, children, className = "", mainClassName = "" }: ReferenceShellProps) {
  return (
    <V4PageShell accountSlot={accountSlot} className={className} mainClassName={mainClassName}>
      {children}
    </V4PageShell>
  );
}

export default ReferenceShell;
