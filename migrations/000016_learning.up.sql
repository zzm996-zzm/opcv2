CREATE TABLE learning_courses (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    level TEXT NOT NULL,
    hours INTEGER NOT NULL DEFAULT 0,
    learners INTEGER NOT NULL DEFAULT 0,
    price_label TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    outline JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX learning_courses_category_idx
    ON learning_courses (category);

CREATE TABLE learning_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_slug TEXT NOT NULL REFERENCES learning_courses(slug) ON DELETE CASCADE,
    percent INTEGER NOT NULL DEFAULT 0 CHECK (percent >= 0 AND percent <= 100),
    last_lesson TEXT NOT NULL DEFAULT '',
    recommended_action TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, course_slug)
);

CREATE INDEX learning_progress_user_updated_idx
    ON learning_progress (user_id, updated_at DESC);

CREATE TABLE learning_diagnoses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal TEXT NOT NULL DEFAULT '',
    project TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    overall_score INTEGER NOT NULL DEFAULT 0 CHECK (overall_score >= 0 AND overall_score <= 100),
    dimensions JSONB NOT NULL DEFAULT '[]'::JSONB,
    recommendations JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX learning_diagnoses_user_created_idx
    ON learning_diagnoses (user_id, created_at DESC);

INSERT INTO learning_courses (slug, title, description, category, level, hours, learners, price_label, tags, outline)
VALUES
    ('ai-basics', 'AI基础入门：从0到1了解AI', '快速掌握 AI 核心概念与应用场景', '入门', '入门', 18, 12400, '免费', '["AI基础","认知框架"]'::JSONB, '["AI核心概念","常见应用场景","能力边界与风险"]'::JSONB),
    ('prompt-engineering', '提示词工程实战', '掌握高质量提示词设计与优化技巧', '实战', '实战', 24, 8700, '会员免费', '["提示词","实战"]'::JSONB, '["任务拆解","上下文设计","质量评估"]'::JSONB),
    ('customer-service-ai', '智能客服应用案例解析', '打造高效客户服务，提升满意度', '行业', '行业', 20, 6300, '会员免费', '["智能客服","行业案例"]'::JSONB, '["客服场景拆解","知识库建设","转人工策略"]'::JSONB),
    ('ai-market-analysis', 'AI行业分析方法', '用 AI 洞察市场趋势与竞争格局', '工具', '实战', 22, 9100, '会员免费', '["市场分析","竞品分析"]'::JSONB, '["行业地图","竞品拆解","机会判断"]'::JSONB);
