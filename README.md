# goctl-swagger

### 1. 编译goctl-swagger插件

```
$ GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u github.com/1278651995/zero-goctl-swagger
```

### 2. 配置环境
将$GOPATH/bin中的goctl-swagger添加到环境变量

### 3. 使用姿势

* 创建api文件
    ```go
    info(
    	title: "type title here"
    	desc: "type desc here"
    	author: "type author here"
    	email: "type email here"
    	version: "type version here"
    )
    
    
    type (
    	RegisterReq {
    		Username string `json:"username"`
    		Password string `json:"password"`
    		Mobile string `json:"mobile"`
    	}
    	
    	LoginReq {
    		Username string `json:"username"`
    		Password string `json:"password"`
    	}
    	
    	UserInfoReq {
    		Id string `path:"id"`
    	}
    	
    	UserInfoReply {
    		Name string `json:"name"`
    		Age int `json:"age"`
    		Birthday string `json:"birthday"`
    		Description string `json:"description"`
    		Tag []string `json:"tag"`
    	}
    	
    	UserSearchReq {
    		KeyWord string `form:"keyWord"`
    	}
    )
    
    service user-api {
    	@doc(
    		summary: "注册"
    	)
    	@handler register
    	post /api/user/register (RegisterReq)
    	
    	@doc(
    		summary: "登录"
    	)
    	@handler login
    	post /api/user/login (LoginReq)
    	
    	@doc(
    		summary: "获取用户信息"
    	)
    	@handler getUserInfo
    	get /api/user/:id (UserInfoReq) returns (UserInfoReply)
    	
    	@doc(
    		summary: "用户搜索"
    	)
    	@handler searchUser
    	get /api/user/search (UserSearchReq) returns (UserInfoReply)
    }
    ```
* 生成swagger.json 文件
    ```shell script
    $ goctl api plugin -plugin goctl-swagger="swagger -filename user.json" -api user.api -dir .
    ```
* 指定Host，basePath [api-host-and-base-path](https://swagger.io/docs/specification/2-0/api-host-and-base-path/)
    ```shell script
    $ goctl api plugin -plugin goctl-swagger="swagger -filename user.json -host 127.0.0.2 -basepath /api" -api user.api -dir .
    ```
* swagger ui 查看生成的文档
    ```shell script
     $ docker run --rm -p 8083:8080 -e SWAGGER_JSON=/foo/user.json -v $PWD:/foo swaggerapi/swagger-ui
   ```
* Swagger Codegen 生成客户端调用代码(go,javascript,php)
  ```shell script
  for l in go javascript php; do
    docker run --rm -v "$(pwd):/go-work" swaggerapi/swagger-codegen-cli generate \
      -i "/go-work/rest.swagger.json" \
      -l "$l" \
      -o "/go-work/clients/$l"
  done
  ```
  
### 4. 结合go-zero使用自动生成接口文档

- 定义swag, swag-json接口

  ```
  	@handler swag
  	get /swag() returns()
  	@handler swagJson
  	get /swag-json() returns()
  ```

  > 注  /swag, /swag-json地址可自己定义，如需自定义，需修改zero-goctl-swagger的

- 修改生成后的Handler

   swaghandler.go

  ```go
  func swagHandler(ctx *svc.ServiceContext) http.HandlerFunc {
  	l := logic.NewSwagLogic(context.TODO(), ctx)
  	return l.Swag
  }
  
  ```

  swagjsonhandler.go

  ```go
  func swagJsonHandler(ctx *svc.ServiceContext) http.HandlerFunc {
  	return func(w http.ResponseWriter, r *http.Request) {
  		w.Header().Set("Content-Type", "application/json; charset=utf-8")
  		_, _ = w.Write(logic.SwagByte)
  	}
  }
  ```

  

  - 修改生成后的logic

    swagjsonlogic.go

    ```go
    type SwagjsonLogic struct {
    	logx.Logger
    	ctx    context.Context
    	svcCtx *svc.ServiceContext
    }
    ```

    swaglogic.go

    ```go
    type SwagLogic struct {
    	logx.Logger
    	ctx    context.Context
    	svcCtx *svc.ServiceContext
    	Swag   http.HandlerFunc
    }
    
    const swagFilePath = "app/rest/pkg/swag/swag.json"
    
    var SwagByte json.RawMessage
    
    func init() {
    	swagFile, err := os.Open(swagFilePath)
    	if err != nil {
    		fmt.Println(err)
    	}
    	defer swagFile.Close()
    	SwagByte, err = ioutil.ReadAll(swagFile)
    	if err != nil {
    		fmt.Println(err)
    	}
    }
    
    func NewSwagLogic(ctx context.Context, svcCtx *svc.ServiceContext) SwagLogic {
    	return SwagLogic{
    		Logger: logx.WithContext(ctx),
    		ctx:    ctx,
    		svcCtx: svcCtx,
    		Swag:   swag.Doc("/swag", hutils.GetEnv("env", "alpha"), SwagByte),
    	}
    }
    ```

    swag.go

    ```go
    
    type Opts func(*swaggerConfig)
    
    // SwaggerOpts configures the Doc middlewares.
    type swaggerConfig struct {
    	// SpecURL the url to find the spec for
    	SpecURL string
    	// SwaggerHost for the js that generates the swagger ui site, defaults to: http://petstore.swagger.io/
    	SwaggerHost string
    }
    
    func Doc(basePath, env string, swaggerJSON []byte, opts ...Opts) http.HandlerFunc {
    	config := &swaggerConfig{
    		SpecURL:     basePath + "-json",
    		SwaggerHost: "https://petstore.swagger.io"}
    	for _, opt := range opts {
    		opt(config)
    	}
    	// swagger json
    	responseSwaggerJSON := swaggerJSON
    	responseSwaggerJSON = responseSwaggerJSON
    
    	// swagger html
    	tmpl := template.Must(template.New("swaggerdoc").Parse(swaggerTemplateV2))
    	buf := bytes.NewBuffer(nil)
    	_ = tmpl.Execute(buf, config)
    	uiHTML := buf.Bytes()
    
    	// permission
    	needPermission := false
    	if env == "prod" {
    		needPermission = true
    		responseSwaggerJSON = []byte(strings.Replace(
    			string(swaggerJSON),
    			`"schemes": [
        "http"
      ],`,
    			`"schemes": [
        "https"
      ],`,
    			1))
    	}
    
    	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
    		if r.URL.Path == basePath {
    			if needPermission {
    				rw.WriteHeader(http.StatusOK)
    				rw.Header().Set("Content-Type", "text/plain")
    				_, _ = rw.Write([]byte("Swagger not open on prod"))
    				return
    			}
    
    			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
    			_, _ = rw.Write(uiHTML)
    
    			rw.WriteHeader(http.StatusOK)
    			return
    		}
    	})
    }
    
    const swaggerTemplateV2 = `
    	<!-- HTML for static distribution bundle build -->
    <!DOCTYPE html>
    <html lang="en">
      <head>
        <meta charset="UTF-8">
        <title>API documentation</title>
        <link rel="stylesheet" type="text/css" href="{{ .SwaggerHost }}/swagger-ui.css" >
        <link rel="icon" type="image/png" href="{{ .SwaggerHost }}/favicon-32x32.png" sizes="32x32" />
        <link rel="icon" type="image/png" href="{{ .SwaggerHost }}/favicon-16x16.png" sizes="16x16" />
        <style>
          html
          {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
          }
    
          *,
          *:before,
          *:after
          {
            box-sizing: inherit;
          }
    
          body
          {
            margin:0;
            background: #fafafa;
          }
        </style>
      </head>
    
      <body>
        <div id="swagger-ui"></div>
    
        <script src="{{ .SwaggerHost }}/swagger-ui-bundle.js"> </script>
        <script src="{{ .SwaggerHost }}/swagger-ui-standalone-preset.js"> </script>
        <script>
        window.onload = function() {
          // Begin Swagger UI call region
          const ui = SwaggerUIBundle({
            "dom_id": "#swagger-ui",
            deepLinking: true,
            presets: [
              SwaggerUIBundle.presets.apis,
              SwaggerUIStandalonePreset
            ],
            plugins: [
              SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
    		validatorUrl: null,
            url: "{{ .SpecURL }}",
          })
    
          // End Swagger UI call region
          window.ui = ui
        }
      </script>
      </body>
    </html>`
    
    
    ```

    

