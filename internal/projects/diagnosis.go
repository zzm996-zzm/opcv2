package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zzm/opcv2/internal/ai"
)

func (s *Service) DiagnoseProject(ctx context.Context, input ProjectDiagnosisInput, ref string) (ProjectDiagnosis, error) {
	if s.repository == nil || input.UserID <= 0 {
		return ProjectDiagnosis{}, ErrServiceNotReady
	}
	s.countProjectAIUsage(ctx, input.UserID, fmt.Sprintf("project-diagnose-%s-%d", ref, s.now().UnixNano()))
	catalog, ok := s.repository.(CatalogRepository)
	if !ok {
		return ProjectDiagnosis{}, ErrServiceNotReady
	}
	project, err := catalog.GetProject(ctx, strings.TrimSpace(ref))
	if err != nil {
		return ProjectDiagnosis{}, err
	}
	profile := map[string]any{}
	if s.profile != nil {
		profileContext, profileErr := s.profile.GetProfileContext(ctx, input.UserID)
		if profileErr != nil {
			return ProjectDiagnosis{}, profileErr
		}
		for _, group := range profileContext.Groups {
			for key, value := range group.Fields {
				if strings.TrimSpace(value) != "" {
					profile[key] = value
				}
			}
		}
	}
	for key, value := range input.ProfilePatch {
		if strings.TrimSpace(fmt.Sprint(value)) != "" {
			profile[strings.TrimSpace(key)] = value
		}
	}
	if s.generator != nil {
		requestPayload := map[string]any{"project": project, "profile": profile}
		payload, marshalErr := json.Marshal(requestPayload)
		if marshalErr == nil {
			aiResult, generateErr := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
				UserID: input.UserID, Feature: "projects.diagnose", PromptVersion: "project_diagnose_v1",
				SystemPrompt: "你是项目可行性诊断助手。只根据项目和用户画像输出 JSON，不输出思维链。fit_score 必须是 0 到 100，verdict 只能是 recommended、conditional、not_recommended。",
				UserPrompt:   string(payload), SchemaName: "project_diagnosis", RepairAttempts: 1,
				Validate: func(content []byte) error {
					var result ProjectDiagnosis
					if err := json.Unmarshal(content, &result); err != nil {
						return err
					}
					if result.FitScore < 0 || result.FitScore > 100 || result.Verdict == "" {
						return errors.New("invalid project diagnosis")
					}
					return nil
				},
			})
			if generateErr == nil {
				var result ProjectDiagnosis
				if json.Unmarshal(aiResult.Content, &result) == nil && result.FitScore >= 0 && result.FitScore <= 100 && result.Verdict != "" {
					result.ProjectID, result.Profile, result.IsModelGenerated, result.Disclaimer = project.ID, profile, true, "AI 生成，仅供参考"
					return result, nil
				}
			}
		}
	}

	score := 45
	reasons := make([]string, 0, 4)
	prerequisites := append([]string(nil), project.ResourceRequirements...)
	nextSteps := []string{"先访谈 3 位目标用户，验证需求是否真实存在", "用最小版本完成一次付费或意向验证"}
	if value := firstProfile(profile, "budget", "budget_band", "available_budget"); value != "" {
		if strings.Contains(strings.ToLower(value), strings.ToLower(project.BudgetBand)) || strings.Contains(value, "万") || strings.Contains(value, "k") {
			score += 15
			reasons = append(reasons, "预算信息与项目启动区间基本匹配")
		} else {
			score -= 10
			prerequisites = appendUnique(prerequisites, "准备与启动预算匹配的现金流")
		}
	} else {
		prerequisites = appendUnique(prerequisites, "先确认可投入预算和回本周期")
	}
	if value := firstProfile(profile, "industry", "sector", "行业"); value != "" {
		if strings.Contains(strings.ToLower(project.Title+" "+project.Summary+" "+project.Track), strings.ToLower(value)) {
			score += 15
			reasons = append(reasons, "已有行业背景与项目目标客户相关")
		} else {
			score += 5
			reasons = append(reasons, "行业背景可迁移，但仍需做客户验证")
		}
	} else {
		prerequisites = appendUnique(prerequisites, "明确目标行业和首批客户画像")
	}
	if value := firstProfile(profile, "role", "experience", "能力"); value != "" {
		reasons = append(reasons, "已提供个人能力信息，可据此安排首轮执行")
		score += 10
	} else {
		prerequisites = appendUnique(prerequisites, "补充现有能力、资源和可投入时间")
	}
	if value := firstProfile(profile, "intent", "goal", "目标"); value != "" && strings.Contains(strings.ToLower(project.Title+" "+project.Summary), strings.ToLower(value)) {
		score += 10
		reasons = append(reasons, "项目方向与当前目标一致")
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	verdict := "not_recommended"
	if score >= 75 {
		verdict = "recommended"
	} else if score >= 55 {
		verdict = "conditional"
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "画像信息较少，当前结论需要更多事实验证")
	}
	if project.Difficulty != "" {
		nextSteps = append([]string{fmt.Sprintf("按“%s”难度拆出一周内可完成的验证任务", project.Difficulty)}, nextSteps...)
	}
	return ProjectDiagnosis{ProjectID: project.ID, FitScore: score, Verdict: verdict, Reasons: reasons, Prerequisites: prerequisites, NextSteps: nextSteps, Profile: profile, IsModelGenerated: true, Disclaimer: "AI 生成，仅供参考"}, nil
}

func firstProfile(profile map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(fmt.Sprint(profile[key])); value != "<nil>" && value != "" {
			return value
		}
	}
	return ""
}

func parseBudgetCents(value string) int64 {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimSuffix(value, "元")
	value = strings.TrimSuffix(value, "rmb")
	if value == "" {
		return 0
	}
	if strings.HasSuffix(value, "k") {
		number, _ := strconv.ParseFloat(strings.TrimSuffix(value, "k"), 64)
		return int64(number * 100000)
	}
	number, _ := strconv.ParseFloat(value, 64)
	return int64(number * 100)
}
