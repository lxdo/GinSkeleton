package routers

import (
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/global/variable"
	"goskeleton/app/http/middleware/authorization"
	"goskeleton/app/http/middleware/cors"
	validatorFactory "goskeleton/app/http/validator/core/factory"
	"goskeleton/app/utils/config"
	"io"
	"net/http"
	"os"
)

// 该路由主要设置 后台管理系统等后端应用路由

func InitWebRouter() *gin.Engine {

	gin.DisableConsoleColor()
	f, _ := os.Create(variable.BASE_PATH + config.CreateYamlFactory().GetString("Logs.GinLogName"))
	gin.DefaultWriter = io.MultiWriter(f)

	router := gin.Default()

	//根据配置进行设置跨域
	if config.CreateYamlFactory().GetBool("HttpServer.AllowCrossDomain") {
		router.Use(cors.Next())
	}

	router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "HelloWorld,这是后端模块")
	})

	//处理静态资源（不建议gin框架处理静态资源，参见 Public/readme.md 说明 ）
	router.Static("/public", "./Public")             //  定义静态资源路由与实际目录映射关系
	router.StaticFS("/dir", http.Dir("./Public"))    // 将Public目录内的文件列举展示
	router.StaticFile("/abcd", "./Public/readme.md") // 可以根据文件名绑定需要返回的文件名

	//  创建一个后端接口路由组
	V_Backend := router.Group("/Admin/")
	{
		// 创建一个websocket,如果ws需要账号密码登录才能使用，就写在需要鉴权的分组，这里暂定是开放式的，不需要严格鉴权，我们简单验证一下token值
		V_Backend.GET("ws", validatorFactory.Create(consts.Validator_Prefix+"WebsocketConnect"))

		//  【不需要】中间件验证的路由  用户注册、登录
		v_noAuth := V_Backend.Group("users/")
		{
			v_noAuth.POST("register", validatorFactory.Create(consts.Validator_Prefix+"UsersRegister"))
			v_noAuth.POST("login", validatorFactory.Create(consts.Validator_Prefix+"UsersLogin"))
			v_noAuth.POST("refreshtoken", validatorFactory.Create(consts.Validator_Prefix+"RefreshToken"))
		}

		// 需要中间件验证的路由
		V_Backend.Use(authorization.CheckAuth())
		{
			// 用户组路由
			v_users := V_Backend.Group("users/")
			{
				// 查询 ，这里的验证器直接从容器获取，是因为程序启动时，将验证器注册在了容器，具体代码位置：App\Http\Validator\Web\Users\xxx
				v_users.GET("index", validatorFactory.Create(consts.Validator_Prefix+"UsersShow"))
				// 新增
				v_users.POST("create", validatorFactory.Create(consts.Validator_Prefix+"UsersStore"))
				// 更新
				v_users.POST("edit", validatorFactory.Create(consts.Validator_Prefix+"UsersUpdate"))
				// 删除
				v_users.POST("delete", validatorFactory.Create(consts.Validator_Prefix+"UsersDestroy"))
			}
			//文件上传公共路由
			v_uploadfiles := V_Backend.Group("upload/")
			{
				v_uploadfiles.POST("files", validatorFactory.Create(consts.Validator_Prefix+"UploadFiles"))
			}

		}

	}
	return router
}