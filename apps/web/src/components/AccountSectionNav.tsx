import { CreditCard, FolderOpen, Settings, SlidersHorizontal, UserRound } from "lucide-react";
import { Link } from "react-router-dom";

const accountSections = [
  ["个人中心", "/profile", UserRound],
  ["账号与资料设置", "/profile/settings", Settings],
  ["会员与账单", "/membership", CreditCard],
  ["我的内容", "/profile/content", FolderOpen],
  ["偏好设置", "/profile/preferences", SlidersHorizontal]
] as const;

type AccountSectionNavProps = {
  activeHref: string;
};

function AccountSectionNav({ activeHref }: AccountSectionNavProps) {
  return (
    <aside className="profile-side-tabs" aria-label="个人中心导航">
      {accountSections.map(([label, href, Icon]) => (
        <Link className={activeHref === href ? "active" : ""} key={label} to={href}>
          <Icon aria-hidden="true" size={18} strokeWidth={1.8} />
          <span>{label}</span>
          <b aria-hidden="true">›</b>
        </Link>
      ))}
    </aside>
  );
}

export default AccountSectionNav;
