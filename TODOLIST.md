帮我实现一个go版本管理工具goenv，你需要实现下述功能：

1. 执行goenv versions后：
1）支持从github.com/通过api来获取golang/go项目已经发布的所有历史tags；
2）拉取到历史tags之后，注意解析版本号，每个tag的命名格式一般是go{x}.{y}.{z}[rc{v}]，其中x、y、z、v都是数值型的版本号，非正式版本会有rc{v}候选版本号后缀；
3）解析完后，将按本号根据主版本号进行汇总，比如go1.22，将go1.22.1, go1.22.2, 以及候选版本go.1.22rc1都放置在go1.22之下，最好连changelog之类信息也拉取一下；
4）为了避免每次执行时都拉取影响查询效率，可以将已经查询过的历史版本，记录在本地的一个json配置文件中，如~/.goenv/versions.json；
5）下次行goenv versions时，优先检查该~/.goenv/versions.json是否存在，如果文件存在，则询问用户是否要检查最新版本列表，如果是则重新执行1）~4），反之直接读取并显示该json文件中的数据；

2. 执行goenv install <version>后：
1) 首先需要确定当前系统、架构，可以通过go env GOOS和go env GOARCH来获取；
2) 然后根据GOOS\GOARCH\version来确定下载链接，比如当前系统为linux，架构为amd64，安装版本为go1.22，则下载链接为https://go.dev/dl/go1.22.linux-amd64.tar.gz；
3) 下载上述压缩包到~/.goenv/downloads目录下；
4) 解压压缩包到~/.goenv/sdk/go<version>目录下；
5) 在~/.goenv/bin/下面创建一个shell脚本来调用真正的go命令行，脚本名为 go{version}，脚本内容需要考虑设置:
    - GOROOT设置为~/.goenv/sdk/go<version>
    - GOPATH=~/.goenv/go<version>
    - GOBIN设置为~/.goenv/bin/go<version>
    ps: 不要export上述设置，仅在当前go命令shell脚本中有效
    - 然后调用真正的go命令行 ~/.goenv/sdk/go<version>/bin/go <args>；
6) 在~/.goenv/bin下面创建一个shell脚本来调用真正的gofmt，脚本名为 gofmt，脚本内容调用真正的gofmt命令行 ~/.goenv/sdk/go<version>/bin/gofmt；

3. 执行特定版本的go命令行，必须显示通过go{version}来执行，所以无需通过一个统一的脚本go来动态判断需要执行哪个版本。go依赖管理依赖多种配置项设置，即使我们创建了这样的统一的脚本go能够区分不同路径下调用不同的go命令行，在IDE设置中也要额外单独设置GOPATH、GOBIN、GOROOT等一系列设置。所以不如直接简化，不做这些特殊逻辑处理。毕竟我们只是在验证不同版本的差异时才需要这么调用。
