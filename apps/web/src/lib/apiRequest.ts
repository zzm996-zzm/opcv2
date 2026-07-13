import { authSession } from "./authSession";

const errorMessages: Record<string, string> = {
  invalid_request: "请求参数有误，请检查后重试",
  invalid_limit: "列表数量参数有误",
  invalid_offset: "分页位置参数有误",
  invalid_access_token: "登录状态已过期，请重新登录",
  invalid_refresh_token: "登录状态已过期，请重新登录",
  invalid_token: "登录状态已过期，请重新登录",
  invalid_phone: "请输入正确的中国大陆手机号",
  invalid_code: "验证码错误或已过期",
  code_rate_limited: "发送太频繁，请稍后再试",
  sms_unavailable: "短信服务暂时不可用，请稍后再试",
  agreement_required: "请先同意用户协议和隐私政策",
  nickname_required: "请输入昵称",
  invalid_account: "账号需为4-32位字母、数字或下划线",
  invalid_password: "密码需为6-72位",
  invalid_credentials: "账号或密码错误",
  account_exists: "账号已存在，请直接登录",
  user_not_found: "账号不存在或已失效",
  user_disabled: "账号已被停用",
  logout_failed: "退出登录失败，请稍后重试",
  rate_limited: "请求过于频繁，请稍后再试",
  service_not_ready: "服务暂时不可用，请稍后再试",
  internal_error: "服务开小差了，请稍后再试",
  invalid_ai_result: "AI 结果暂时不可用，请稍后重试",
  request_failed: "请求失败，请稍后重试",
  unauthorized: "请先登录后再操作",
  admin_required: "当前账号没有管理权限",
  invalid_content_input: "内容参数有误，请检查后重试",
  article_not_found: "文章不存在或已下架",
  tool_not_found: "工具不存在或已下架",
  code_not_found: "兑换码不存在",
  code_expired: "兑换码已过期",
  code_exhausted: "兑换码已被使用完",
  invalid_task_input: "线索任务参数有误，请检查后重试",
  task_not_found: "任务不存在或已无权限访问",
  invalid_task_id: "任务 ID 不正确",
  invalid_priority: "优先级筛选参数有误",
  quota_exceeded: "当前会员额度不足，请升级或下月重置后再试",
  quota_not_configured: "当前功能额度暂未配置，请联系管理员",
  session_not_found: "记录不存在或已无权限访问",
  invalid_session_id: "记录 ID 不正确",
  model_not_found: "测算模型不存在或已无权限访问",
  invalid_model_id: "测算模型 ID 不正确",
  scan_not_found: "竞品扫描不存在或已无权限访问",
  invalid_scan_id: "竞品扫描 ID 不正确",
  course_not_found: "课程不存在或已下架",
  diagnosis_not_found: "暂时没有诊断报告",
  action_item_not_found: "行动项不存在或已无权限访问",
  invalid_action_item_id: "行动项 ID 不正确",
  match_not_found: "项目匹配记录不存在或已无权限访问",
  invalid_match_id: "项目匹配 ID 不正确",
  customer_not_found: "客户不存在或已无权限访问",
  invalid_customer_id: "客户 ID 不正确",
  invalid_crm_input: "客户跟进参数有误，请检查后重试",
  thread_not_found: "对话不存在或已无权限访问",
  memory_not_found: "记忆不存在或已无权限访问",
  file_not_found: "文件不存在或已无权限访问",
  file_too_large: "单个文件不能超过 10MB",
  unsupported_file_type: "暂不支持该文件类型",
  invalid_file_encoding: "文件内容无法识别，请上传 UTF-8 文本或有效 DOCX",
  streaming_not_supported: "当前服务暂不支持流式回复",
  stream_failed: "流式回复中断，请重试",
  invalid_thread_id: "对话 ID 不正确",
  invalid_memory_id: "记忆 ID 不正确",
  invalid_file_id: "文件 ID 不正确",
  notification_not_found: "消息不存在或已无权限访问",
  membership_required: "循环提醒仅限会员使用，请升级后重试"
};

export class ApiRequestError extends Error {
  code: string;
  status: number;

  constructor(code: string, status: number) {
    super(errorMessages[code] ?? errorMessages.request_failed);
    this.name = "ApiRequestError";
    this.code = code;
    this.status = status;
  }
}

function apiUrl(path: string) {
  const baseUrl = import.meta.env.VITE_API_BASE_URL?.trim();
  if (!baseUrl) return path;
  return `${baseUrl.replace(/\/+$/, "")}/${path.replace(/^\/+/, "")}`;
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await apiStreamRequest(path, init);
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export async function apiStreamRequest(path: string, init: RequestInit = {}) {
  const token = authSession.get().accessToken;
  const isFormData = typeof FormData !== "undefined" && init.body instanceof FormData;
  const response = await fetch(apiUrl(path), {
    ...init,
    credentials: "include",
    headers: {
      ...(isFormData ? {} : { "Content-Type": "application/json" }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers
    }
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as { error?: string };
    throw new ApiRequestError(payload.error ?? "request_failed", response.status);
  }
  return response;
}
