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
  invalid_idempotency_key: "幂等键格式有误，请稍后重试",
  invalid_status_transition: "当前状态不允许这样流转，请刷新任务后重试",
  invalid_progress: "任务进度必须是 0 到 100 的整数",
  task_version_conflict: "任务已被其他操作更新，请重新打开后再保存",
  task_ai_draft_not_found: "AI 任务草稿不存在或已无权访问",
  task_ai_draft_already_adopted: "该 AI 草稿已经采纳，请刷新任务列表",
  invalid_task_ai_draft: "AI 草稿内容不完整，请检查后再采纳",
  task_not_found: "任务不存在或已无权限访问",
  invalid_task_id: "任务 ID 不正确",
  invalid_priority: "优先级筛选参数有误",
  invalid_sort: "任务排序参数有误，请重置排序后重试",
  invalid_group: "任务分组参数有误，请重置分组后重试",
  quota_exceeded: "当前会员额度不足，请升级或下月重置后再试",
  quota_not_configured: "当前功能额度暂未配置，请联系管理员",
  session_not_found: "记录不存在或已无权限访问",
  invalid_session_id: "记录 ID 不正确",
  invalid_intake: "补充信息有误，请检查后重试",
  intake_incomplete: "请完成或明确跳过所有补充问题",
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
  membership_required: "循环提醒仅限会员使用，请升级后重试",
  compare_limit_reached: "最多只能同时对比 5 个项目",
  project_file_not_found: "文件不存在或已无权访问",
  project_file_expired: "文件已过期，请重新上传",
  project_file_too_large: "单个文件不能超过 20MB",
  project_file_count_exceeded: "最多只能上传 10 个文件",
  project_file_total_size_exceeded: "文件总大小不能超过 50MB",
  unsupported_project_file_type: "暂不支持该文件类型",
  project_file_mime_mismatch: "文件类型与内容不一致",
  invalid_project_file_name: "文件名不安全，请重命名后上传",
  unsafe_project_file: "文件未通过安全扫描",
  project_file_not_ready: "文件尚未解析完成",
  project_file_duplicate: "该文件已上传或已绑定到其他匹配记录",
  task_attachment_not_found: "附件不存在或已无权限访问",
  task_attachment_too_large: "单个附件不能超过 10MB",
  task_attachment_count_exceeded: "一个任务最多上传 10 个附件",
  unsupported_task_attachment_type: "暂不支持该附件类型",
  task_attachment_mime_mismatch: "附件类型与内容不一致",
  unsafe_task_attachment: "附件未通过安全扫描",
  invalid_attachment_signature: "附件下载链接已失效，请重新下载",
  invalid_task_attachment: "附件参数有误，请重新上传"
};

export class ApiRequestError extends Error {
  code: string;
  status: number;

  constructor(code: string, status: number, message?: string) {
    super(message?.trim() || errorMessages[code] || errorMessages.request_failed);
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
    const payload = (await response.json().catch(() => ({}))) as { error?: string; message?: string };
    throw new ApiRequestError(payload.error ?? "request_failed", response.status, payload.message);
  }
  return response;
}
