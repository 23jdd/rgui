# rgui

rgui 是一个基于 Go 和 raylib 的 GUI 库。目前提供一个可运行的 Windows 中文计数器示例。

## 运行第一个示例

在安装了 Go 1.27.1 和 Windows 等线字体（`%WINDIR%\Fonts\Deng.ttf`）的环境中，从项目根目录运行：

```sh
go run ./examples/counter
```

点击「加一」更新计数，按 Esc 关闭窗口。`task.toml` 提供相同的运行命令，构建任务生成 `demo.exe`。

## 窗口配置

`rgui.Run` 接收 `WindowConfig`。窗口默认以 480 × 300 逻辑像素打开，最小 240 × 160，支持缩放和高 DPI，帧率上限为 60 FPS。

| 字段 | 用法 |
| --- | --- |
| `Width`、`Height` | 初始尺寸，自动限制在设置的最小和最大尺寸内。 |
| `MinWidth`、`MinHeight`、`MaxWidth`、`MaxHeight` | 分轴设置缩放范围。最小值为零时采用默认值，最大值为零时该轴无上限。 |
| `FixedSize`、`Borderless`、`AlwaysOnTop` | 分别控制禁止手动缩放、隐藏原生边框和标题栏、窗口置顶。 |
| `VSync`、`TargetFPS` | 启用垂直同步并设置帧率上限。上限为零时使用默认 60 FPS，设为 -1 时取消帧率上限。 |
| `Background` | 可选的 `*color.RGBA`，用于清除每帧背景。 |

运行 `go run ./examples/layout` 可查看 560 × 340 的初始窗口、缩放范围和自定义背景。无边框窗口仍可用 Esc 关闭。

## 界面模型

采用保留模式（retained mode）的组件树，状态更新采用类似 SolidJS 的细粒度响应式模型。每次挂载视图时执行一次 `View`，创建组件树和该视图的局部状态。库在帧之间保留节点及其交互状态，绘制和鼠标事件都使用同一棵树。

rgui 管理 raylib 窗口和绘制循环。应用从主 goroutine 调用 `rgui.Run(config, View)`。每帧根据窗口尺寸计算组件位置，再处理鼠标输入、统一求值待更新的绑定并绘制。关闭窗口时卸载组件树并释放窗口与字体资源。

### 首版 API 示例

以下是计数器的视图代码，省略了库函数和类型的 `rgui.` 前缀。完整示例见 `examples/counter/main.go`。

```go
func View() Node {
    count := NewSignal(0)

    return Stack(
        Text(func() string {
            return fmt.Sprintf("点击次数：%d", count.Get())
        }),
        Button("加一", func() {
            count.Set(count.Get() + 1)
        }),
    )
}
```

- `Text` 在挂载时求值，记录回调读取的 Signal，得到初始文本 `点击次数：0`。
- 在 UI 线程调用 `Set` 后，`Get` 立即读到新值。受影响的绑定进入待更新队列，在下一次绘制前统一求值。同一绑定在一帧内收到多次写入时只求值一次，使用最后写入的值，并更新原有节点。
- 每次重新求值按本次实际读取的 Signal 更新订阅。节点卸载时解除它的全部订阅。
- `Stack` 默认按子节点的声明顺序竖排，使用固定间距。Windows 计数器示例加载系统中文字体，用于标签和按钮。

## 首个里程碑

先做一个可导入的最小 GUI 库，用计数器示例验证绘制、输入和状态更新能连起来。

| 需求 | 验收方式 |
| --- | --- |
| 库和示例分离 | 示例通过导入库使用公开接口，项目可以从源码构建。 |
| 组件树跨帧保留 | 每次挂载调用 `View` 一次。连续点击时复用同一棵树和原有标签节点。重新挂载后计数从 0 开始。 |
| 绘制标签和按钮 | 在 Windows 上运行示例，能读到「点击次数：0」和「加一」。 |
| 鼠标输入驱动响应式更新 | 鼠标左键在同一个按钮内按下和释放时，计数加一。持续按住、在按钮外按下或在按钮外释放时，计数保持原值。 |
| 同帧合并更新 | 同一帧内连续两次 `Set`，每次 `Get` 都能读到刚写入的值。绘制前，依赖绑定只重新求值一次并使用最终值。 |
| 更新范围和生命周期 | 写入 Signal 后，下一次绘制显示新值。无关绑定的求值次数保持不变。卸载节点后，再写入原 Signal 时，该节点的绑定不再执行。 |
| 窗口缩放 | 在 Windows 上调整窗口大小后，按钮的显示位置和可点击区域仍然一致。 |

### 首版范围

首版以 Windows 鼠标计数器为验收环境。`Stack` 的固定间距为 12 个逻辑像素。键盘导航、通用布局和其他平台验证列入后续范围。

## 首版实施顺序

1. **建立构建基线**。把窗口原型迁入 `examples/counter`，根包改为可导入的 `rgui`。处理未完成的 `CreateSignal`，同步更新 `task.toml`。完成后 `go test ./...` 和 `go build ./...` 都能通过。
2. **实现 Signal 和文本绑定**。实现 `NewSignal`、`Get`、`Set`、更新队列以及绑定的挂载和卸载。用无需打开窗口的测试检查同帧合并、依赖隔离、条件读取变化和订阅清理。
3. **接通组件树和计数器**。由库入口管理窗口、输入、绑定更新、布局和绘制，接通 `Stack`、`Text`、`Button` 和 Windows 系统中文字体。用 `go run ./examples/counter` 手动验收中文初值、连续点击、长按、按钮外释放和窗口缩放。

## 绘制图元

`rgui/draw` 提供在 raylib 绘制帧中使用的图元。`Button` 负责交互，用 `draw.Rect` 和 `draw.Text` 绘制。`rgui.Text(func() string)` 负责响应式绑定，复用同一套文字测量和绘制参数。

```sh
go run ./examples/primitives
```

| 批次 | 范围 | 验收条件 |
| --- | --- | --- |
| 1. 基础形状和文字 | `Rect`、`Circle`、`Line`、`Text`；纯色填充、描边、矩形圆角 | 同一个示例展示填充、描边和圆角。计数器改用共享图元后，中文仍居中，窗口缩放后按钮绘制与点击区域一致。 |
| 2. 复杂几何 | `Ellipse`、`Polyline`、`Polygon`、`Arc`、`Ring`、`Path` | 图元展示页能区分开放折线与闭合多边形、不同起止角度的弧和圆环，以及包含直线、二次和三次贝塞尔曲线的开放与闭合路径。 |
| 3. 文本、图像与组合 | `TextBlock`、`Image`、`NinePatch`、`Clip`、`Transform` | 文本按给定宽度换行，测量结果与绘制使用同一排版。图像可指定源区域和目标尺寸，九宫格缩放保留四角。嵌套裁剪取交集，平移、旋转和缩放按父子层级组合，后续兄弟节点恢复各自的绘制状态。 |
| 4. 进阶样式 | 线性渐变、径向渐变、阴影、虚线描边 | 展示页能辨认线性方向与径向中心，阴影可配置偏移和模糊，虚线间距沿连续路径保持一致。 |

`task.toml` 的 `primitives` 任务也可启动图元展示页。展示页覆盖表中的四组能力，计数器可用于验证控件的绘制和点击。

- `Polygon` 支持简单凹多边形。每个 `Path` 绘制一条开放或闭合轮廓，多轮廓、孔洞和自交填充留到后续。
- `Canvas.Clip` 接收屏幕坐标，裁剪范围相交。`Canvas.Transform` 按父子层级组合平移、旋转和缩放，横纵缩放因子均取正值。这两个作用域目前用于绘制。接入交互组件时，命中检测也需要应用相同规则。
- 线性渐变用于矩形，径向渐变用于圆形，软阴影用于圆角矩形。虚线沿描边路径连续计算。
- `Image` 和 `NinePatch` 接收调用方持有的 raylib 纹理，调用方负责加载和释放。

## 布局系统

参考 WPF 的两阶段模型。父容器用 `Measure(available)` 询问子节点在给定空间下的期望尺寸，再用 `Arrange(slot)` 分配最终位置和大小。绘制与鼠标命中使用同一份最终布局结果。[WPF 布局文档](https://learn.microsoft.com/en-us/dotnet/desktop/wpf/advanced/layout)

| 批次 | 范围 | 验收条件 |
| --- | --- | --- |
| 1. 布局基础与 StackPanel | 横排和竖排、`Margin`、`Padding`、`Gap`、宽高与对齐 | 同一组节点切换排列方向后顺序正确。文字按字体实际尺寸测量，显式宽高优先，居中和拉伸使用父容器分配的区域。现有计数器通过兼容的 `Stack` 入口保持原布局。 |
| 2. Grid | 行列支持固定逻辑像素、`Auto` 和带权重的 `Star` | 工具栏用 `Auto / 1* / 固定` 三列验证内容列、弹性空白和固定按钮区。另用 `1* / 2*` 验证剩余空间按 1:2 分配。缩放窗口后点击位置跟随按钮。 |
| 3. 补充面板 | `DockPanel`、`WrapPanel`，按示例验证 | Dock 布局中的工具栏贴边，主区域填满剩余空间。Wrap 布局随窗口变窄换行。 |

- **StackPanel** 在排列轴上按内容测量子节点，横轴按父容器可用空间安排。带权重的剩余空间分配放在 Grid。[WPF StackPanel 与 DockPanel 比较](https://learn.microsoft.com/en-us/dotnet/desktop/wpf/controls/how-to-choose-between-stackpanel-and-dockpanel)
- **Grid** 支持 `RowSpan` 和 `ColumnSpan`。固定轨道先预留，`Auto` 轨道按内容增长，`Star` 轨道按权重分配非负剩余空间。[WPF Grid 自动布局](https://learn.microsoft.com/en-us/dotnet/desktop/wpf/advanced/how-to-use-a-grid-for-automatic-layout)
- **Signal 与布局**。文本绑定更新导致期望尺寸变化时，在绘制前重新测量和排列受影响的布局。窗口尺寸变化时，在处理鼠标命中前刷新布局。首版允许重算整棵树，按节点缓存测量结果留到后续。

第一批提供 `StackPanel(Vertical, gap, children...)` 和 `StackPanel(Horizontal, gap, children...)`。用 `WithLayout(node, Box{...})` 设置宽高、外边距、内边距和两个方向的对齐。宽高为零时按内容测量，默认对齐为拉伸。`Stack` 保留原有的固定行高和根容器宽度。

第二批用 `Grid(rows, columns, cells...)` 指定行列轨道和子节点，`Cell.Row` 和 `Cell.Column` 从零开始。轨道用 `Fixed(pixels)`、`Auto()` 或 `Star(weight)` 定义。沿未限制尺寸的轴，正权重的 `Star` 轨道按内容测量。

第三批提供 `DockPanel(lastChildFill, children...)` 和 `WrapPanel(direction, gap, children...)`。Dock 子节点按声明顺序占用指定边缘，启用 `lastChildFill` 后最后一个子节点占用剩余矩形。Wrap 子节点沿主轴排列，空间不足时换到下一行或列。

布局算法用无需打开窗口的测试验证。运行 `go run ./examples/layout` 查看可缩放工具栏、1:2 权重分配、贴边栏和自动换行。改变窗口宽度后点击重新排布的 `Item` 按钮，工具栏计数会更新为该项的编号。

## 焦点和键盘输入

- 点击可用按钮时取得焦点。按 Tab 或 Shift+Tab 按组件声明顺序循环切换，跳过禁用和完全位于窗口外的按钮。焦点按钮显示金色边框。
- 聚焦按钮在释放 Space 或 Enter 时执行同一个点击回调。按住时不会连发，切换焦点或窗口失焦会取消未完成的按键操作。窗口失焦也会取消未完成的鼠标点击。
- 使用 `WithDisabled(button, true)` 禁用按钮。布局示例中的 `Item 6` 用于验证禁用行为。

运行 `go run ./examples/layout`。按两次 Tab 选中「+」，按 Space 加一，再用 Tab 切到 `Reset`，按 Enter 清零。

## 自定义节点

`rgui.Node` 有三个阶段。自定义节点可放进现有面板，也可直接从 `View` 返回。

```go
type Node interface {
    Measure(MeasureContext) rl.Vector2
    Place(rl.Rectangle)
    Primitives(PaintContext) []draw.Primitive
}
```

1. `Measure` 接收可用空间和默认字体的文字测量函数，返回内容的期望尺寸。可用宽高可能是正无穷。一次布局可能测量同一节点多次。
2. `Place` 接收最终的内容矩形，对应 WPF 的 Arrange。节点可以保存矩形或在此计算内部几何。
3. `Primitives` 每帧根据最近一次 `Place` 的结果生成图元。框架按切片顺序绘制，并将自定义节点的绘制裁剪到内容区域。

可以在自定义结构体中嵌入 `rgui.BaseNode`，然后用该结构体的指针作为 `Node`。框架自动读取其 `Box`，包括 `Width` / `Height`、`MinWidth` / `MinHeight`、`MaxWidth` / `MaxHeight`、`HorizontalAlignment` / `VerticalAlignment`、`Margin` 和 `Padding`。在 UI 线程修改属性后，框架会重新布局。

默认宽高按内容测量，最大尺寸为零时表示无上限。尺寸冲突时最小值优先，显式宽高受有效的最小值和最大值限制。`Stretch` 尽量填满可用空间，受最大尺寸限制时在剩余空间内居中。[WPF 尺寸约束](https://learn.microsoft.com/en-us/dotnet/desktop/wpf/controls/how-to-set-the-width-properties-of-an-element)

`DesiredSize()` 返回含外边距的期望尺寸，`LayoutSlot()` 返回父容器分配的原始矩形。默认 `Place` 保存最终内容矩形，绘制时通过 `Bounds()` 读取。默认内容测量返回零尺寸，绘制返回空列表。进度条示例直接设置嵌入的 `Box`，只覆盖测量和绘制。

- 内置控件、面板和自定义节点统一使用 `Node`。布局属性通过 `WithLayout(node, box)` 设置，由框架应用到父容器分配的区域，再将内容矩形传给自定义节点的 `Place`。[WPF 布局文档](https://learn.microsoft.com/en-us/dotnet/desktop/wpf/advanced/layout)
- 绘制时可以读取 Signal 的最新值。`Measure` 应依据传入的约束与稳定的尺寸配置计算，读取 Signal 不会单独让布局失效。自定义节点接口当前提供测量和绘制能力。
- `draw.RectPrimitive` 和 `draw.TextPrimitive` 可直接返回，其他现有图元可通过 `draw.PrimitiveFunc` 加入列表。

运行 `go run ./examples/customnode` 查看自定义进度条。按下 `Advance` 后，文本和进度条读取同一个 Signal。

## 后续路线图

参考本地 `../ui` 的实现，将以下六项接入 rgui 的 `Node`、Signal 和 raylib 绘制循环。每个阶段完成验收后单独提交。

| 阶段 | 范围 | 验收条件 |
| --- | --- | --- |
| 1. 子树生命周期 | 自定义容器暴露子节点，框架管理父子归属、挂载、卸载和订阅清理。 | 同一个节点只能有一个父节点，重复挂载和成环可检测。移除子树后，其绑定不再执行，焦点和按下状态被清理，其他节点的状态保留。 |
| 2. 通用输入与事件路由 | 自定义节点可接收指针、键盘和焦点事件。支持预览、冒泡、指针捕获、`TabIndex` 和命中开关。 | 测试根到目标、目标到根的顺序及事件取消。拖出控件后仍能向捕获节点发送移动和释放，窗口失焦或节点卸载时释放捕获。现有按钮的鼠标和键盘行为保持一致。 |
| 3. 布局失效与缓存 | 分开记录测量和放置失效，把子节点的布局变化传给祖先。复用约束和布局状态均未变化的分支。 | 同一帧多次修改只做一次必要的布局更新。改动一个叶节点时，干净的兄弟节点不重复测量。保留现有直接修改 `BaseNode.Box` 后可重新布局的行为。 |
| 4. 扩展面板 | 增加支持背景、边框和内边距的 `Border`，支持边缘锚点的 `Canvas`，以及 `Grid` 跨行跨列和容器裁剪。 | 示例覆盖四边锚点、跨越混合轨道的单元格和嵌套裁剪。绘制和命中使用相同的裁剪边界。 |
| 5. 高 DPI 布局 | 统一逻辑坐标、物理像素和鼠标坐标，按物理像素对齐共享边缘。 | 在 125% 和 150% 下验证相邻区域无缝、窗口四角完整覆盖，以及缩放、最大化和恢复后点击仍与绘制对齐。 |
| 6. 帧统计与按需重绘 | 先统计布局、绘制和呈现耗时，再验证静止帧复用和脏区重绘。 | 静止帧不重新生成场景图元。局部重绘和整帧重绘的像素结果一致，覆盖透明叠加、节点移动、窗口缩放和 DPI 变化。 |

实现从子树生命周期开始。事件路由和布局失效都需要稳定的父子关系。面板裁剪完成后，再验证 DPI 下绘制、裁剪和命中共用坐标。按需重绘以同一场景的整帧结果和帧统计作为对照。

自定义容器使用 `WithChildren(custom, children...)`。子节点按声明顺序叠放在内容区域。保留返回的 `Node`，在 UI 线程用 `SetChildren(container, children...)` 更新。仍在列表中的节点保留状态，移除节点释放订阅。

用 `BaseNode.Input` 或 `WithInput(node, options)` 注册输入处理。事件先预览再冒泡，设置 `Handled` 可停止后续处理和默认动作。鼠标捕获通过 `InputEvent.CapturePointer()` 请求，释放鼠标、移除节点或窗口失焦时自动结束。Tab 按 `TabIndex` 和声明顺序导航，负值跳过 Tab。

布局复用约束和属性未变化的节点结果。对齐变化只使放置失效，尺寸变化同时使测量失效。直接修改 `Box` 仍会被框架检测；自定义节点的内容尺寸改变时可调用 `BaseNode.InvalidateMeasure()`。

运行 `go run ./examples/panels` 查看 `Border`、边缘锚点 `Canvas`、混合轨道跨列和嵌套裁剪。设置 `Box.ClipToBounds` 会用容器的内容矩形裁剪子节点的绘制和命中；圆角属于边框外观。

窗口的逻辑视口由实际帧缓冲尺寸除以 DPI 比例得到。布局按物理像素对齐绝对边缘，裁剪采用同一坐标映射，鼠标使用 raylib 提供的逻辑坐标。

## 每次改动的工作流

1. 选一条验收条件，写清触发动作和预期结果。涉及公开 API 时，先写期望的调用代码。
2. 实现这一条行为，把可独立验证的状态和输入逻辑留在窗口循环之外。
3. 在同一次改动里同步更新受影响的 API 示例、运行命令和 `task.toml`。
4. 运行 `go test ./...` 和 `go build ./...`。涉及窗口、输入或绘制时，再运行示例并记录对应的手动验证结果。
5. 验证通过后，用 `git status --short -- .` 检查项目改动，逐个暂存本次任务涉及的文件。用 `git diff --cached` 检查整笔暂存内容，确认只包含本次任务后再执行 `git commit -m "简述本次改动"`。每笔提交对应一项独立、可验证的改动。
