// ABOUTME: 博客处理器注册入口
// ABOUTME: 通过空白导入触发各博客处理器的 init() 注册
package setup

import (
	_ "github.com/esttorhe/blogwatcher-ui/v2/internal/processor/blogs"
)
