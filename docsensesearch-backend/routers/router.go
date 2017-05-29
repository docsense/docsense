package routers

import (
	"docsensesearch/controllers"
	"docsensesearch/search"
	"docsensesearch/upload"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/auth"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/api/search", &search.Controller{})
	beego.Router("/api/searchbyfile", &search.ByFileController{})
	beego.Router("/api/upload", &upload.Controller{})
	beego.Router("/api/managesplists", &controllers.ManageSpListsController{})
	beego.Router("/api/download", &controllers.DownloadController{})
	beego.Router("/api/_refreshsearch", &controllers.RefreshSearchController{})

	beego.InsertFilter("/api/managesplists", beego.BeforeRouter, auth.Basic("a", "b"))
	beego.InsertFilter("/api/_refreshsearch", beego.BeforeRouter, auth.Basic("a", "b"))
}
