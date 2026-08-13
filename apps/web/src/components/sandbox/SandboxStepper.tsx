import { Check } from "lucide-react";

const steps = ["描述", "补充信息", "选择角色", "确认开始"] as const;

type SandboxStepperProps = {
  active: 1 | 2 | 3 | 4;
};

function SandboxStepper({ active }: SandboxStepperProps) {
  return (
    <ol className="sb-stepper" aria-label="推演配置进度">
      {steps.map((label, index) => {
        const step = (index + 1) as 1 | 2 | 3 | 4;
        const completed = step < active;
        return (
          <li className={step === active ? "is-active" : completed ? "is-complete" : ""} key={label}>
            <span aria-hidden="true">{completed ? <Check size={15} strokeWidth={3} /> : step}</span>
            <strong>{label}</strong>
          </li>
        );
      })}
    </ol>
  );
}

export default SandboxStepper;
