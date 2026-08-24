// Package webui 提供纺织染色批次色差证据复核台的极简说明页面。
package webui

import (
	"html/template"
	"net/http"
)

// IndexPage 渲染一个说明页，指向核心 API 入口。
const IndexPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>纺织染色批次色差证据复核台</title>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;max-width:760px;margin:40px auto;padding:0 16px;color:#222;background:#fff}
h1{font-size:22px;border-bottom:2px solid #e0e0e0;padding-bottom:8px}
h2{font-size:16px;margin-top:24px}
p{line-height:1.6}
code{background:#f4f4f4;padding:2px 6px;border-radius:4px;font-size:13px}
li{margin:6px 0}
.tag{display:inline-block;background:#eef4ff;color:#1a56c0;border-radius:4px;padding:2px 8px;font-size:12px;margin-right:6px}
</style>
</head>
<body>
<h1>纺织染色批次色差证据复核台</h1>
<p>纺织工艺师在此核对染色批次的色差来源，判断是浴液条件、取样位置还是仪器校准造成的偏差。系统导入批次与浴液温度/pH 曲线、上传布样多点测色并校准测色仪，计算批内/批间色差，辅助剔除污点、补充工艺证据并发布复核结论。</p>

<h2>核心流程</h2>
<p><span class="tag">批次</span><span class="tag">浴液曲线</span><span class="tag">多点测色</span><span class="tag">色差计算</span><span class="tag">证据标注</span><span class="tag">结论发布</span></p>

<h2>核心 API</h2>
<ul>
<li><code>POST /api/batches</code> 创建染色批次</li>
<li><code>POST /api/batches/{id}/advance</code> 推进批次状态</li>
<li><code>POST /api/batches/{id}/bath-curves</code> 上传浴液温度/pH 曲线</li>
<li><code>POST /api/batches/{id}/measure-points</code> 上传布样测色点</li>
<li><code>POST /api/batches/{id}/measure-points/{mid}/reject</code> 剔除污点</li>
<li><code>POST /api/instruments/calibrations</code> 记录仪器校准</li>
<li><code>POST /api/batches/{id}/color-diff</code> 色差计算（cie76/cie94/cie2000）</li>
<li><code>POST /api/batches/{id}/evidences</code> 提交工艺证据</li>
<li><code>POST /api/batches/{id}/conclusion</code> 创建复核结论</li>
<li><code>POST /api/batches/{id}/conclusion/{cid}/publish</code> 发布结论</li>
<li><code>GET /api/self-check</code> 自检</li>
<li><code>POST /api/demo/import</code> 示例闭环导入</li>
</ul>
</body>
</html>`

// Handler 返回首页 handler。
func Handler() http.Handler {
	tpl := template.Must(template.New("index").Parse(IndexPage))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, nil)
	})
}
