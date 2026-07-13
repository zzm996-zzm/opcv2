import { apiRequest } from "./apiRequest";

export type LearningCourse = {
  id: number;
  slug: string;
  title: string;
  description: string;
  category: string;
  level: string;
  hours: number;
  learners: number;
  price_label: string;
  tags: string[];
  outline: string[];
  created_at: string;
  updated_at: string;
};

export type LearningCourseMaterial = {
  id: number;
  course_slug: string;
  title: string;
  material_type: string;
  content_url: string;
  position: number;
  downloadable: boolean;
  created_at: string;
  updated_at: string;
};

export type LearningProgress = {
  id: number;
  user_id: number;
  course_slug: string;
  course_title: string;
  percent: number;
  last_lesson: string;
  recommended_action: string;
  updated_at: string;
};

export type UpdateLearningProgressInput = {
  percent: number;
  last_lesson: string;
  recommended_action: string;
};

export type LearningDimension = {
  name: string;
  score: number;
  gap: number;
  summary: string;
};

export type LearningAssessmentAnswer = {
  key: string;
  question: string;
  answer: string;
};

export type LearningEvidenceSource = {
  type: string;
  label: string;
  captured_at: string;
};

type LearningProvenance = {
  basis: "model_assessment" | "legacy_estimate" | string;
  disclaimer: string;
  assumptions: string[];
  evidence_sources: LearningEvidenceSource[];
};

export type LearningDiagnosis = LearningProvenance & {
  id: number;
  user_id: number;
  goal: string;
  project: string;
  focus_abilities: string[];
  weekly_time: string;
  bottleneck: string;
  answers: LearningAssessmentAnswer[];
  status: string;
  overall_score: number;
  dimensions: LearningDimension[];
  recommendations: string[];
  created_at: string;
  updated_at: string;
};

export type CreateLearningDiagnosisInput = {
  goal: string;
  project: string;
  focus_abilities: string[];
  weekly_time: string;
  bottleneck: string;
  answers?: LearningAssessmentAnswer[];
};

export type LearningGapItem = {
  name: string;
  current: number;
  target: number;
  gap: number;
  priority: string;
  summary: string;
  evidence: string;
  recommended: string;
};

export type LearningGaps = LearningProvenance & {
  diagnosis_id: number;
  goal: string;
  project: string;
  overall_score: number;
  gaps: LearningGapItem[];
  evidence: string[];
  generated_at: string;
};

export type LearningRecommendationFocus = {
  name: string;
  priority: string;
  summary: string;
};

export type LearningMethod = {
  title: string;
  value: string;
  detail: string;
};

export type LearningRecommendations = LearningProvenance & {
  diagnosis_id: number;
  goal: string;
  project: string;
  focus: LearningRecommendationFocus[];
  recommendations: string[];
  methods: LearningMethod[];
  generated_at: string;
};

export type LearningPlanStage = {
  number: number;
  title: string;
  status: string;
  courses: string[];
  duration: string;
  goal: string;
  milestone: string;
};

export type LearningPlanItem = {
  id: number;
  user_id: number;
  diagnosis_id: number;
  stage_number: number;
  title: string;
  completed: boolean;
  completed_at?: string;
  updated_at: string;
};

export type LearningPlan = LearningProvenance & {
  diagnosis_id: number;
  title: string;
  description: string;
  recommendations: string[];
  stages: LearningPlanStage[];
  items: LearningPlanItem[];
  estimated_hours: number;
  weekly_suggestion: string;
  generated_at: string;
};

export type LearningReport = LearningProvenance & {
  diagnosis_id: number;
  goal: string;
  project: string;
  overall_score: number;
  dimensions: LearningDimension[];
  priority_gaps: LearningGapItem[];
  recommendations: string[];
  evidence: string[];
  generated_at: string;
};

export type CourseFilter = {
  category?: string;
  limit?: number;
};

function courseQuery(filter: CourseFilter = {}) {
  const params = new URLSearchParams();
  if (filter.category) params.set("category", filter.category);
  if (filter.limit) params.set("limit", String(filter.limit));
  const query = params.toString();
  return query ? `?${query}` : "";
}

export const learningApi = {
  listCourses(filter: CourseFilter = {}) {
    return apiRequest<{ courses: LearningCourse[] }>(`/api/v1/learning/courses${courseQuery(filter)}`, {
      method: "GET"
    });
  },

  getCourse(slug: string) {
    return apiRequest<LearningCourse>(`/api/v1/learning/courses/${encodeURIComponent(slug)}`, {
      method: "GET"
    });
  },

  listCourseMaterials(slug: string) {
    return apiRequest<{ materials: LearningCourseMaterial[] }>(`/api/v1/learning/courses/${encodeURIComponent(slug)}/materials`, {
      method: "GET"
    });
  },

  listProgress() {
    return apiRequest<{ progress: LearningProgress[] }>("/api/v1/learning/progress", {
      method: "GET"
    });
  },

  getProgress(courseSlug: string) {
    return apiRequest<LearningProgress>(`/api/v1/learning/progress/${encodeURIComponent(courseSlug)}`, {
      method: "GET"
    });
  },

  updateProgress(courseSlug: string, input: UpdateLearningProgressInput) {
    return apiRequest<LearningProgress>(`/api/v1/learning/progress/${encodeURIComponent(courseSlug)}`, {
      method: "PUT",
      body: JSON.stringify(input)
    });
  },

  createDiagnosis(input: CreateLearningDiagnosisInput) {
    return apiRequest<LearningDiagnosis>("/api/v1/learning/diagnoses", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  submitAssessment(input: CreateLearningDiagnosisInput) {
    return apiRequest<LearningDiagnosis>("/api/v1/learning/assessments", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  getAssessment(id: number) {
    return apiRequest<LearningDiagnosis>(`/api/v1/learning/assessments/${id}`, { method: "GET" });
  },

  getLatestAssessment() {
    return apiRequest<LearningDiagnosis>("/api/v1/learning/assessments/latest", { method: "GET" });
  },

  getLatestDiagnosis() {
    return apiRequest<LearningDiagnosis>("/api/v1/learning/diagnoses/latest", {
      method: "GET"
    });
  },

  getLatestGaps() {
    return apiRequest<LearningGaps>("/api/v1/learning/diagnoses/latest/gaps", {
      method: "GET"
    });
  },

  getLatestRecommendations() {
    return apiRequest<LearningRecommendations>("/api/v1/learning/diagnoses/latest/recommendations", {
      method: "GET"
    });
  },

  getLatestPlan() {
    return apiRequest<LearningPlan>("/api/v1/learning/diagnoses/latest/plan", {
      method: "GET"
    });
  },

  getPlan(diagnosisId: number) {
    return apiRequest<LearningPlan>(`/api/v1/learning/diagnoses/${diagnosisId}/plan`, { method: "GET" });
  },

  updatePlanItem(diagnosisId: number, stageNumber: number, completed: boolean) {
    return apiRequest<LearningPlanItem>(`/api/v1/learning/diagnoses/${diagnosisId}/plan/items/${stageNumber}`, {
      method: "PUT",
      body: JSON.stringify({ completed })
    });
  },

  getLatestReport() {
    return apiRequest<LearningReport>("/api/v1/learning/diagnoses/latest/report", {
      method: "GET"
    });
  }
};
