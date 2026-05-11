# Phase 20: Blog Preview - UI Design

**Created:** 2026-05-11
**Phase:** 20 - Blog Preview
**Requirements:** PREV-01~06

## Overview

在添加 Blog 之前，用户可以预览 Feed 解析结果，确认文章列表正确后再保存。预览功能通过在 Add Blog 表单添加 Preview 按钮实现，点击后跳转到临时预览页面显示最多 20 条文章。

## Design Decisions

### D-01: 预览按钮位置 — Inline 并列

**Decision:** Preview 和 Add Blog 按钮并排在表单底部。

**Why:**
- 操作直观，用户可自主选择先预览或直接添加
- 与现有表单风格一致（按钮并排布局）
- 两个功能平等可见，不隐藏次要操作

**How to apply:**
`add-blog-form.gohtml` 在 `.form-actions` 区域添加两个按钮：
```html
<div class="form-actions">
    <button type="button" class="btn-action" hx-post="/blogs/preview">Preview</button>
    <button type="submit" class="btn-action btn-primary">Add Blog</button>
</div>
```

### D-02: 预览页面布局 — 卡片列表

**Decision:** 每篇文章一个独立卡片，显示标题、时间和链接，与主页文章卡片风格一致。

**Why:**
- 视觉层次清晰，信息密度适中
- 与现有 UI 风格一致，用户熟悉
- 可清晰展示文章标题、发布时间、原文链接

**How to apply:**
创建 `preview-page.gohtml`，使用与 `article-card.gohtml` 相似的卡片样式：
- 标题（可点击打开原文）
- 发布时间（relative time format）
- 原文链接图标（target="_blank"）

最多显示 20 条文章，超过时显示 "Showing 20 of N articles" 提示。

### D-03: 错误显示方式 — Inline 错误提示

**Decision:** 在预览页面内显示红色错误框，同时提供 "Back to Edit" 和 "Save Anyway" 选项。

**Why:**
- 用户可自主选择下一步（返回修改或强制保存）
- 错误信息详细，可显示具体原因（如 "timeout fetching feed"）
- 允许用户保存无效 URL，稍后手动修复（灵活性）

**How to apply:**
创建 `preview-error.gohtml` fragment：
```html
<div class="error-message">
    <h4>⚠ Feed 解析失败</h4>
    <p>{{.ErrorMessage}}</p>
</div>
<div class="error-actions">
    <button hx-get="/blogs/add" hx-target="#main-content">← Back to Edit</button>
    <button hx-post="/blogs/preview/save" class="btn-primary">Save Anyway</button>
</div>
```

### D-04: 保存后跳转 — Settings 页面

**Decision:** 保存成功后跳转到 Settings 页面，显示成功消息。

**Why:**
- 流程闭环：Settings → Add Blog → Preview → Save → Settings
- 用户可立即验证博客已添加（在 Tracked Blogs 列表中看到）
- 符合现有 Add Blog 流程的成功跳转行为

**How to apply:**
`handleBlogPreviewSave` 保存成功后：
1. 返回 Settings 页面
2. 显示成功消息 "Blog 'XXX' added successfully"
3. 后台触发 sync（与现有 Add Blog 行为一致）

## Architecture

### Data Flow

```
Add Blog Form (name + url input)
    │
    ├─→ [Preview] ─→ POST /blogs/preview
    │                   │
    │                   ├─→ rss.ParseFeed(url) → []FeedArticle (max 20)
    │                   │
    │                   ├─→ Success: render preview-page.gohtml
    │                   │
    │                   └─→ Error: render preview-page.gohtml with error
    │
    ├─→ [Add Blog] ─→ POST /blogs/add (existing flow)
    │
    └─→ [Back to Edit] ─→ GET /blogs/add (preserve input)

Preview Page
    │
    ├─→ [Save as Blog] ─→ POST /blogs/preview/save
    │                       │
    │                       ├─→ BlogService.AddBlog()
    │                       │
    │                       └─→ Redirect to Settings + success message
    │
    └─→ [Save Anyway] (error case) ─→ POST /blogs/preview/save
```

### New Components

| Component | Purpose |
|-----------|---------|
| `handleBlogPreview` | 处理 Preview 按钮，解析 Feed，返回预览页面 |
| `handleBlogPreviewSave` | 处理 Save as Blog，调用 BlogService，跳转 Settings |
| `preview-page.gohtml` | 预览页面模板（文章卡片列表 + 操作按钮） |
| `preview-article-card.gohtml` | 单个预览文章卡片（可复用 article-card 样式） |

### Existing Components Extended

| Component | Change |
|-----------|--------|
| `add-blog-form.gohtml` | 添加 Preview 按钮，与 Add Blog 并排 |
| `routes.go` | 新增 `/blogs/preview` POST 路由 |
| `handlers.go` | 新增 `handleBlogPreview` 和 `handleBlogPreviewSave` |

## Error Handling

| Error Scenario | Display | Actions |
|----------------|---------|---------|
| Feed URL 无效（非 HTTP/HTTPS） | "Invalid URL format" | Back to Edit only |
| Feed URL 无法访问（404/500） | "Failed to fetch feed: status 404" | Back to Edit / Save Anyway |
| Feed 解析失败（非 RSS/Atom） | "Failed to parse feed: invalid format" | Back to Edit / Save Anyway |
| 网络超时 | "Timeout fetching feed" | Back to Edit / Save Anyway |
| Feed URL 为空 | "Please enter a URL to preview" | Back to Edit only |

## UI Specifications

### Add Blog Form (Extended)

```
┌─────────────────────────────────────┐
│ Add New Blog                        │
├─────────────────────────────────────┤
│ Blog Name: [________________]       │
│ Blog URL:  [________________]       │
│                                     │
│ [Preview] [Add Blog]                │
└─────────────────────────────────────┘
```

### Preview Page (Success)

```
┌─────────────────────────────────────┐
│ Preview: My Favorite Blog           │
├─────────────────────────────────────┤
│ ┌─────────────────────────────────┐ │
│ │ 文章标题 1                       │ │
│ │ 2 hours ago • [↗ Open]          │ │
│ └─────────────────────────────────┘ │
│ ┌─────────────────────────────────┐ │
│ │ 文章标题 2                       │ │
│ │ 1 day ago • [↗ Open]            │ │
│ └─────────────────────────────────┘ │
│ ... (最多 20 条)                    │
│ Showing 20 of 45 articles           │
├─────────────────────────────────────┤
│ [← Back to Edit] [Save as Blog]     │
└─────────────────────────────────────┘
```

### Preview Page (Error)

```
┌─────────────────────────────────────┐
│ Preview: My Favorite Blog           │
├─────────────────────────────────────┤
│ ┌─────────────────────────────────┐ │
│ │ ⚠ Feed 解析失败                  │ │
│ │ 无法从 https://example.com/feed  │ │
│ │ 获取文章。请检查 URL 是否正确。   │ │
│ └─────────────────────────────────┘ │
│                                     │
│ [← Back to Edit] [Save Anyway]      │
└─────────────────────────────────────┘
```

## Testing Scenarios

1. **Preview 成功**
   - 输入有效 Blog URL → Preview → 看到文章列表 → Save → Settings 显示新博客

2. **Preview 失败 → 返回修改**
   - 输入无效 URL → Preview → 看到错误 → Back to Edit → 修改 URL → 重新 Preview

3. **Preview 失败 → 强制保存**
   - 输入无效 URL → Preview → 看到错误 → Save Anyway → Settings 显示新博客（无 Feed URL）

4. **直接 Add Blog**
   - 输入 URL → Add Blog → 跳过预览 → Settings 显示新博客（现有流程不变）

## Requirements Coverage

| Requirement | Covered by |
|-------------|------------|
| PREV-01 | D-01: Preview 按钮 |
| PREV-02 | handleBlogPreview + rss.ParseFeed |
| PREV-03 | D-02: 卡片列表，最多 20 条 |
| PREV-04 | D-03: Inline 错误提示 |
| PREV-05 | handleBlogPreviewSave |
| PREV-06 | Back to Edit 按钮 |

---

*Design created: 2026-05-11*
*Decisions validated through visual brainstorming*