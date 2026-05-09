---
phase: 17-category-management-ui
plan: 01
status: complete
completed: 2026-05-09
execution_time: 5min
---

# Plan 17-01: Category Database Methods

## Objective
实现分类系统的数据库方法，为 UI 层提供分类 CRUD 操作和数据查询能力。

## Completed Tasks

### Task 1: 实现分类 CRUD 数据库方法
- ✓ 添加 `CategoryWithBlogCount` struct（扩展 Category 带有 blog 统计）
- ✓ 实现 `CreateCategory(name string)` 方法 - 创建新分类并返回带 ID 的 Category
- ✓ 实现 `ListCategoriesWithBlogCount()` 方法 - LEFT JOIN 统计每个分类的 blog 数量，按名称排序
- ✓ 实现 `UpdateCategoryName(id int64, name string)` 方法 - 更新分类名称
- ✓ 实现 `DeleteCategory(id int64)` 方法 - 删除分类并事务性地置空关联 blog.category_id
- ✓ 实现 `UpdateBlogCategory(blogID int64, categoryID *int64)` 方法 - 更新 blog 的分类（支持 nil 表示未分类）

### Task 2: 添加数据库方法测试
- ✓ `TestCreateCategory` - 验证分类创建、ID 分配、时间戳
- ✓ `TestCreateCategoryEmptyName` - 验证空名称错误处理
- ✓ `TestListCategoriesWithBlogCount` - 验证 LEFT JOIN 统计、排序、空分类处理
- ✓ `TestUpdateCategoryName` - 验证名称更新
- ✓ `TestUpdateCategoryNameEmpty` - 验证空名称错误处理
- ✓ `TestUpdateCategoryNameNotFound` - 验证不存在分类错误处理
- ✓ `TestDeleteCategory` - 验证删除 + blog.category_id 置空（事务性）
- ✓ `TestDeleteCategoryNotFound` - 验证不存在分类错误处理
- ✓ `TestUpdateBlogCategory` - 验证 blog 分类更新（nil/un-nil）
- ✓ `TestUpdateBlogCategoryNotFound` - 验证不存在 blog 错误处理

## Additional Changes
修复现有 blog 查询方法以支持 `category_id` 字段：
- ✓ 修改 `scanBlog` 方法以扫描 `category_id`（sql.NullInt64）
- ✓ 修改 `GetBlogByID` 查询以包含 `category_id`
- ✓ 修改 `GetBlogByName` 查询以包含 `category_id`
- ✓ 修改 `GetBlogByURL` 查询以包含 `category_id`
- ✓ 修改 `ListBlogs` 查询以包含 `category_id`
- ✓ 修改 `ListBlogsWithCounts` 查询和扫描以包含 `category_id`

## Verification
- ✓ 所有测试通过（`go test ./internal/storage -v`）
- ✓ 代码编译通过（`go build ./internal/storage`）

## Files Modified
- `internal/storage/database.go` - 添加 5 个分类方法 + CategoryWithBlogCount struct + 修改 6 个 blog 查询方法
- `internal/storage/database_test.go` - 添加 10 个分类方法测试

## Key Decisions
- 使用事务处理 `DeleteCategory` 以确保原子性（先置空 blog.category_id，再删除 category）
- 使用 `sql.NullInt64` 处理 nullable `category_id` 字段
- 分类列表按名称排序（字母顺序）
- 空分类也显示在列表中（BlogCount = 0）

## Next Steps
Plan 17-02 将实现：
- 分类管理的 HTTP Handler（创建、编辑、删除）
- 分类管理的 Go templates（inline 表单、确认 dialog）
- 路由注册
