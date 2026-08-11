# MatchOption

用于从命令输出或 HTTP 响应体中提取 IP 地址的共享规则。字段位于数据源的 `match:` 键下。

```yaml
match:
  jq: ".ip"
  regex: "IP:\\s+(\\S+)"
```

两个字段都是可选项。省略 `match` 时，Lightddns 仍会尝试纯文本解析。

## `jq`

`jq` 指的是 [jq manual](https://jqlang.org/manual/) 描述的查询语言。Lightddns 自身不实现这套语言；当前使用 Go 实现 [gojq](https://github.com/itchyny/gojq) 执行该字段。因此写法上按 jq 过滤器理解，但语法解析和运行时兼容性以 `gojq` 为准。

```yaml
match:
  jq: ".ip"
```

表达式可以产出一个或多个值：

```yaml
match:
  jq: ".addresses[]"
```

只有字符串结果会被继续解析为 IP 地址。对象、数组、数字、布尔值和 `null` 会被忽略。如果字符串结果不是合法 IP 字面量，本次 jq 提取会失败。

## `regex`

使用正则表达式提取 IP 地址。

```yaml
match:
  regex: "Current IP:\\s+(\\S+)"
```

表达式必须包含捕获组。每次正则匹配后，Lightddns 只解析第一个捕获组作为 IP 地址，不会解析整个匹配结果。

```yaml
# 正确：第一个捕获组只有地址本身。
match:
  regex: "IP=(\\S+)"

# 错误：没有可供 Lightddns 解析的捕获组。
match:
  regex: "IP=\\S+"
```

## 纯文本

当之前的规则没有返回地址时，Lightddns 会去掉首尾空白，再按 Unicode 空白字符切分输入，并把每个片段解析为 IP 地址。

这种输出可以直接解析：

```text
203.0.113.10
2001:db8::10
```

这种输出需要 `regex` 或 `jq`：

```text
ip=203.0.113.10
"203.0.113.10"
```

## 顺序

使用默认匹配器的数据源（例如 `command`）按下面顺序提取：

1. 若配置了 `jq`，先把输入按 JSON 解析并执行 jq 表达式。只有 jq 返回至少一个 IP 地址时才立即返回。
2. 若配置了 `regex` 且 jq 没有返回地址，则对每次正则匹配解析第一个捕获组。只有正则返回至少一个 IP 地址时才立即返回。
3. 若仍未找到地址，则使用纯文本解析。
4. 若所有启用的策略都没有得到地址，则返回收集到的提取错误。

HTTP 数据源额外参考响应的 `Content-Type`：

1. 如果响应 `Content-Type` 是 JSON 且配置了 `match.jq`，HTTP 只使用 jq；不会继续回退到正则或纯文本。
2. 否则，HTTP 会先尝试 `match.regex`，再尝试纯文本。

JSON 端点应使用 `match.jq`。带有说明文字的文本页面应使用 `match.regex`。如果响应体只有一个或多个按空白分隔的 IP 字面量，可以省略 `match`。
