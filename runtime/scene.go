package runtime

type SceneType int

const (
	SceneDev  SceneType = iota // 开发
	SceneTest                  // 测试
	SceneDemo                  // 演示
	ScenePre                   // 预生产
	SceneRel                   // 生产
)

var (
	Scene SceneType = SceneDev // 默认为开发场景
)
