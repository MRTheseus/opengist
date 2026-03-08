# 二维码按钮和私有gist显示问题详细分析

## 问题1：二维码按钮无反应

### 代码分析

查看 `all.html` 第59行：
```html
<script src="https://cdn.jsdelivr.net/npm/qrcode@1.5.3/build/qrcode.min.js"></script>
```

### 根本原因

**QRCode库的CDN链接可能无法访问！**

- `cdn.jsdelivr.net` 是国外CDN
- 如果用户的网络无法访问这个CDN（比如没有科学上网），QRCode库就不会加载
- 当点击二维码按钮时，`new QRCode()` 会报错：`QRCode is not defined`
- JavaScript报错后，整个 `showQR` 函数执行失败，modal不会显示

### 验证方法

打开浏览器开发者工具（F12），查看Console面板，应该能看到类似错误：
```
Uncaught ReferenceError: QRCode is not defined
```

---

## 问题2：私有gist显示公开

### 代码分析

查看 `db.go` 第23-28行：
```go
type Visibility int

const (
    PublicVisibility  Visibility = 0
    PrivateVisibility Visibility = 1
)
```

查看 `template.go` 第37-50行的 `toInt` 函数：
```go
"toInt": func(v interface{}) int {
    switch val := v.(type) {
    case int:
        return val
    case int64:
        return int(val)
    case int32:
        return int(val)
    case uint:
        return int(val)
    default:
        return 0  // 关键问题！
    }
},
```

### 根本原因

**`toInt` 函数没有处理自定义类型 `Visibility`！**

- `Visibility` 是自定义类型，底层是 `int`
- 但在Go的类型断言中，`Visibility` 不等于 `int`
- 当 `Visibility` 类型传入 `toInt` 函数时，不匹配任何 case
- 进入 `default` 分支，返回 `0`
- 所以 `toInt(PrivateVisibility)` 返回 `0`，而不是 `1`
- `0` 等于 `PublicVisibility`，所以显示"公开"

---

## 解决方案

### 问题1：二维码按钮

**方案A**：使用国内CDN
```html
<script src="https://cdn.bootcdn.net/ajax/libs/qrcodejs/1.0.0/qrcode.min.js"></script>
```

**方案B**：使用unpkg CDN（更稳定）
```html
<script src="https://unpkg.com/qrcode@1.5.3/build/qrcode.min.js"></script>
```

### 问题2：私有gist显示公开

**方案**：在 `toInt` 函数中添加对 `Visibility` 类型的处理

修改 `template.go`：
```go
"toInt": func(v interface{}) int {
    switch val := v.(type) {
    case int:
        return val
    case int64:
        return int(val)
    case int32:
        return int(val)
    case uint:
        return int(val)
    case db.Visibility:  // 添加这一行
        return int(val)
    default:
        return 0
    }
},
```

或者更简单的方案：直接在模板中比较，不使用 `toInt`：
```html
{{if eq .Visibility 1}}私有{{else}}公开{{end}}
```

但这样可能还是会有类型问题，因为 `1` 是 `int` 类型，而 `.Visibility` 是 `Visibility` 类型。

**最佳方案**：修改 `toInt` 函数，添加 `db.Visibility` 类型的处理。

---

## 实施步骤

1. 修改 `internal/template/template.go`：
   - 添加 `db.Visibility` 类型的处理
   - 需要import `db` 包

2. 修改 `cmd/web/templates/all.html`：
   - 更换QRCode库的CDN链接

3. 重新构建并测试
