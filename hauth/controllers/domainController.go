package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego/context"
	"github.com/hzwy23/asofdate/hauth/hcache"
	"github.com/hzwy23/asofdate/hauth/models"

	"github.com/asaskevich/govalidator"
	"github.com/hzwy23/asofdate/hauth/hrpc"
	"github.com/hzwy23/asofdate/utils/hret"
	"github.com/hzwy23/asofdate/utils/i18n"
	"github.com/hzwy23/asofdate/utils/logs"
	"github.com/hzwy23/asofdate/utils/token/hjwt"
)

type domainController struct {
	models *models.ProjectMgr
}

var DomainCtl = &domainController{models: &models.ProjectMgr{}}

// 获取domain_info配置页面
func (this *domainController) GetDomainInfoPage(ctx *context.Context) {
	defer hret.HttpPanic()

	if !hrpc.BasicAuth(ctx) {
		return
	}

	rst, err := hcache.GetStaticFile("DomainPage")
	if err != nil {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 404, i18n.Get("as_of_date_page_not_exist"))
		return
	}

	ctx.ResponseWriter.Write(rst)
}

// 查询域信息
func (this *domainController) GetDomainInfo(ctx *context.Context) {

	ctx.Request.ParseForm()

	if !hrpc.BasicAuth(ctx) {
		return
	}

	offset := ctx.Request.FormValue("offset")
	limit := ctx.Request.FormValue("limit")

	rst, total, err := this.models.GetAll(offset, limit)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 312, i18n.Get("as_of_date_domain_query"))
		return
	}

	hret.WriteBootstrapTableJson(ctx.ResponseWriter, total, rst)
}

// 新增域信息
// http请求参数
// domainId     域编码,必须由数字,字母组成
// domainDesc   域名称,不能为空
// domainStatus 域状态,必须是0或者1中的一个
func (this *domainController) PostDomainInfo(ctx *context.Context) {
	ctx.Request.ParseForm()

	if !hrpc.BasicAuth(ctx) {
		return
	}

	domainId := ctx.Request.FormValue("domainId")
	domainDesc := ctx.Request.FormValue("domainDesc")
	domainStatus := ctx.Request.FormValue("domainStatus")
	//校验
	if !govalidator.IsWord(domainId) {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 419, i18n.Get("as_of_date_domain_id_check"))
		return
	}

	//
	if !govalidator.IsIn(domainStatus, "0", "1") {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 419, i18n.Get("as_of_date_domain_status_check"))
		return
	}

	if govalidator.IsEmpty(domainDesc) {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 419, i18n.Get("as_of_date_domain_isempty"))
		return
	}

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 403, i18n.Disconnect())
		return
	}

	err = this.models.Post(domainId, domainDesc, domainStatus, jclaim.User_id, jclaim.Domain_id)

	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, i18n.Get("as_of_date_domain_add_failed"), err)
		return
	}

	hret.WriteHttpOkMsgs(ctx.ResponseWriter, i18n.Get("success"))
}

// 删除域信息
func (this *domainController) DeleteDomainInfo(ctx *context.Context) {
	ctx.Request.ParseForm()
	if !hrpc.BasicAuth(ctx) {
		return
	}

	ijs := []byte(ctx.Request.FormValue("JSON"))
	var js []models.ProjectMgr
	err := json.Unmarshal(ijs, &js)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, i18n.Get("as_of_date_domain_delete"))
		return
	}

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 403, i18n.Get("as_of_date_disconnect"))
		return
	}

	err = this.models.Delete(js, jclaim.User_id, jclaim.Domain_id)
	if err != nil {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, err.Error())
		return
	}

	hret.WriteHttpOkMsgs(ctx.ResponseWriter, i18n.Get("success"))
}

// 更新域信息
func (this *domainController) UpdateDomainInfo(ctx *context.Context) {
	ctx.Request.ParseForm()

	if !hrpc.BasicAuth(ctx) {
		return
	}

	domainId := ctx.Request.FormValue("domainId")
	domainDesc := ctx.Request.FormValue("domainDesc")
	domainStatus := ctx.Request.FormValue("domainStatus")

	// 校验域名称,不能为空
	if govalidator.IsEmpty(domainDesc) {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, i18n.Get("as_of_date_domain_isempty"))
		return
	}

	// 校验域状态编码,必须是0或者1
	if !govalidator.IsIn(domainStatus, "0", "1") {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, i18n.Get("as_of_date_domain_status_check"))
		return
	}

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 403, i18n.Disconnect())
		return
	}

	if !hrpc.CheckDomain(ctx, domainId, "w") {
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 403, i18n.Get("as_of_date_domain_permission_denied_modify"))
		return
	}

	err = this.models.Update(domainDesc, domainStatus, jclaim.User_id, domainId)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 421, i18n.Get("as_of_date_domain_update"))
		return
	}

	hret.WriteHttpOkMsgs(ctx.ResponseWriter, i18n.Get("success"))
}


// 获取指定域详细信息
func (this *domainController) GetDetails(ctx *context.Context) {
	ctx.Request.ParseForm()
	var domain_id = ctx.Request.FormValue("domain_id")

	rst, err := this.models.GetRow(domain_id)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 419, i18n.Get("as_of_date_domain_details"))
		return
	}
	hret.WriteJson(ctx.ResponseWriter, rst)
}

// 获取用户自己所属域的编码
func (this *domainController) GetDomainId(ctx *context.Context) {
	ctx.Request.ParseForm()

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 403, i18n.Disconnect())
		return
	}

	hret.WriteJson(ctx.ResponseWriter, jclaim.Domain_id)
}

func init() {
	hcache.Register("DomainPage", "./views/hauth/domain_info.tpl")
}
