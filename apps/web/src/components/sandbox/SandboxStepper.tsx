import { Check } from "lucide-react";

const steps = ["描述", "补充信息", "选择角色", "确认开始"] as const;
const smartSteps = [
  { label: "智能补充信息", step: 2 },
  { label: "选择推演角色", step: 3 },
  { label: "确认并开始推演", step: 4 }
] as const;

type SandboxStepperProps = {
  active: 1 | 2 | 3 | 4;
  smart?: boolean;
};

function SandboxStepper({ active, smart = false }: SandboxStepperProps) {
  const items = smart ? smartSteps : steps.map((label, index) => ({ label, step: index + 1 }));

  return (
    <ol className={`sb-stepper${smart ? " is-smart" : ""}`} aria-label="推演配置进度">
      {items.map(({ label, step }, index) => {
        const completed = step < active;
        return (
          <li className={step === active ? "is-active" : completed ? "is-complete" : ""} key={label}>
            <span aria-hidden="true">{completed ? <Check size={15} strokeWidth={3} /> : smart ? index + 1 : step}</span>
            <strong>{label}</strong>
          </li>
        );
      })}
    </ol>
  );
}

export default SandboxStepper;
