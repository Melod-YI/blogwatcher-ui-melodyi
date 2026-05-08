# Phase 17: Category Management UI - 设计文档

**日期:** 2026-05-08
**阶段:** Phase 17 - Category Management UI
**状态:** 已确认

## 概述

为 BlogWatcher UI 设置页面添加分类管理功能，允许用户创建、编辑、删除分类，并为 blog 分配分类。

## Requirements 覆盖

| Requirement | 覆盖方式 |
|-------------|----------|
| CATG-03 | 设置页面添加分类管理区（独立区域，位于 Tracked Blogs 之前） |
| CATG-04 | 用户可创建新分类（inline 输入，点击 "+ New Category"） |
| CATG-05 | 用户可编辑分类名称（inline 编辑，点击名称直接编辑） |
| CATG-06 | 用户可删除分类（确认 dialog，删除时 blog.category_id 置空） |
| CATG-07 | Blog 编辑时可选择分类（编辑表单内下拉选择） |

## 设计决策

### 1. 分类管理区位置

**选择:** Tracked Blogs 之前（推荐）

**理由:**
- 逻辑上：先有分类，再分配 blog
- 用户新增 blog 时，可以先看到分类选项
- 分类作为组织单元，优先展示更符合认知顺序

### 2. 分类管理交互风格

**选择:** Inline 操作（推荐）

**理由:**
- 与现有 blog name 编辑风格一致（click-to-edit）
- 流畅直观，无需弹窗打断
- HTMX 友好，易于实现动态交互

**交互详情:**
- **创建:** 点击 "+ New Category" → inline 输入框出现 → 输入名称 → Save/Cancel
- **编辑:** 点击分类名称 → inline 编辑 → Enter/Save 保存
- **删除:** 点击 Delete 按钮 → 确认 dialog → 确认后删除

### 3. Blog 分类选择方式

**选择:** 编辑表单内下拉选择（推荐）

**理由:**
- 与现有编辑流程一致
- 点击 Edit 进入编辑模式后，统一编辑 name + category
- 避免功能冗余（display 行内下拉会与 Edit 按钮功能重叠）

**编辑表单字段:**
```
[Name: _____________]
[Category: [下拉选择]]
[Save] [Cancel]
```

### 4. 删除分类行为

**选择:** 简单删除 + 自动置空（推荐）

**理由:**
- 符合 CATG-06 要求："删除时 blog.category_id 置空"
- 操作简单，与删除 blog 的 cascade delete 风格一致
- 确认 dialog 显示受影响的 blog 数量

**确认 Dialog 内容:**
```
Delete Category?
[分类名称]

This category has [N] blogs.
Deleting it will set these blogs to "未分类".

[Blog1] → 未分类
[Blog2] → 未分类
...

[Delete] [Cancel]
```

## UI 结构

### Categories Section 布局

```html
<section class="category-section">
  <h2>Categories</h2>

  <!-- 分类列表 -->
  <div class="category-list">
    <div class="category-item" id="category-{id}">
      <span class="category-name">{name}</span>
      <span class="category-count">({count} blogs)</span>
      <button class="btn-edit">Edit</button>
      <button class="btn-delete">Delete</button>
    </div>
  </div>

  <!-- 新建分类 inline 输入 -->
  <div class="add-category-form" id="add-category-form">
    <input type="text" placeholder="New category name">
    <button class="btn-save">Save</button>
    <button class="btn-cancel">Cancel</button>
  </div>

  <button class="btn-add-category">+ New Category</button>
</section>
```

### Blog Edit Form 扩展

```html
<div class="blog-edit-form" id="blog-{id}">
  <input type="text" name="name" value="{blog_name}">

  <!-- 新增：分类下拉选择 -->
  <select name="category_id">
    <option value="">-- 未分类 --</option>
    {{range .Categories}}
    <option value="{{.ID}}">{{.Name}}</option>
    {{end}}
  </select>

  <button class="btn-save">Save</button>
  <button class="btn-cancel">Cancel</button>
</div>
```

## 数据流

### 创建分类

1. 用户点击 "+ New Category"
2. HTMX: `hx-get="/categories/new"` → 返回 inline 输入表单
3. 用户输入名称，点击 Save
4. HTMX: `hx-post="/categories"` → 创建分类，返回分类列表行
5. 触发 `HX-Trigger: categoryListUpdated` → 更新下拉选项

### 编辑分类名称

1. 用户点击分类名称（click-to-edit）
2. HTMX: `hx-get="/categories/{id}/edit"` → 返回 inline 编辑输入框
3. 用户修改名称，点击 Save 或按 Enter
4. HTMX: `hx-put="/categories/{id}"` → 更新分类，返回 display 行

### 删除分类

1. 用户点击 Delete 按钮
2. 打开确认 dialog（显示关联 blog 数量）
3. 用户确认删除
4. HTMX: `hx-delete="/categories/{id}"` → 删除分类，blog.category_id 置空
5. 触发 `HX-Trigger: categoryListUpdated` + `blogListUpdated`

### Blog 编辑选择分类

1. 用户点击 blog 的 Edit 按钮
2. HTMX: `hx-get="/blogs/{id}/edit"` → 返回编辑表单（包含 category 下拉）
3. 用户选择分类，点击 Save
4. HTMX: `hx-put="/blogs/{id}"` → 更新 blog（name + category_id），返回 display 行

## 技术实现要点

### 数据库方法（Phase 16 已完成）

- `CreateCategory(name)` → 创建分类
- `ListCategories()` → 获取所有分类
- `ListCategoriesWithBlogCount()` → 分类 + blog 数量
- `UpdateCategoryName(id, name)` → 更新分类名称
- `DeleteCategory(id)` → 删除分类（置空关联 blog.category_id）
- `UpdateBlogCategory(blogID, categoryID)` → 更新 blog 分类

### Handler 路由

| 路由 | 方法 | 说明 |
|------|------|------|
| `/categories/new` | GET | 返回新建分类 inline 表单 |
| `/categories` | POST | 创建分类 |
| `/categories/{id}/edit` | GET | 返回编辑分类 inline 输入框 |
| `/categories/{id}` | PUT | 更新分类名称 |
| `/categories/{id}` | DELETE | 删除分类 + 置空关联 blog |
| `/blogs/{id}/edit` | GET | 返回编辑 blog 表单（含 category 下拉） |
| `/blogs/{id}` | PUT | 更新 blog（name + category_id） |

### Template 文件

| 文件 | 说明 |
|------|------|
| `category-section.gohtml` | 分类管理区整体布局 |
| `category-item.gohtml` | 分类 display 行 |
| `category-edit-form.gohtml` | 分类 inline 编辑输入框 |
| `category-add-form.gohtml` | 新建分类 inline 输入框 |
| `delete-category-dialog.gohtml` | 删除分类确认 dialog |
| `blog-edit-form.gohtml` | 扩展：添加 category 下拉 |

## CSS 样式

- 分类区使用浅蓝色背景（区别于 blog 区）
- 分类名称可点击编辑（cursor: pointer）
- 确认 dialog 与现有 delete-blog-dialog 风格一致

## 测试要点

1. 创建分类后，blog 编辑下拉立即显示新分类
2. 删除分类后，关联 blog 显示 "未分类"
3. 编辑 blog 选择分类后，保存立即生效
4. inline 编辑保存后，分类名称立即更新
5. 空分类删除无需确认 dialog（可选优化）

## Out of Scope

- 分类排序（按字母顺序显示）
- 分类颜色/图标
- 多分类（一个 blog 可属于多个分类）
- 分类嵌套（子分类）