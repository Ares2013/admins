/**********************************************************
 * Author        : Michael
 * Email         : dolotech@163.com
 * Last modified : 2016-03-18 10:16
 * Filename      : roles.go
 * Description   :
 * *******************************************************/
package admin

import (
	"basic/ssdb/gossdb"
	"basic/utils"
	"data"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

const (
	LIMIT int = 30
)

var OPTION = map[string]string{
	"30":  "30",
	"50":  "50",
	"100": "100",
	"200": "200",
}

var Roles = roles{}

type user struct {
	Userid      string // 用户id
	Nickname    string // 用户昵称
	PhoneN      string // 绑定的手机号码
	Coin        uint32 // 金币
	Diamond     uint32 // 钻石
	Vip         uint32 // Vip
	Create_ip   string // 注册账户时的IP地址
	Create_time string // 注册时间
	Sex         uint32
}
type roles struct {
	pager    pager
	selector selected
}

// 玩家信息编辑
func (this *roles) Edit(c *gin.Context) {
	userid := c.PostForm("Userid")
	nickname := c.PostForm("Nickname")
	sex := c.PostForm("Sex")
	phone := c.PostForm("Phone")
	vip := c.PostForm("Vip")
	coin := c.PostForm("Coin")
	diamond := c.PostForm("Diamond")
	pwd := c.PostForm("Password")
	pwd1 := c.PostForm("Password1")
	user := &data.User{Userid: userid}
	m := make(map[string]interface{})
	m["Nickname"] = nickname
	if sex == "1" {
		m["Sex"] = 1
	} else {
		m["Sex"] = 2
	}
	v, _ := strconv.Atoi(vip)
	m["Vip"] = v
	ci, _ := strconv.Atoi(coin)
	m["Coin"] = ci
	d, _ := strconv.Atoi(diamond)
	m["Diamond"] = d
	m["Phone"] = phone
	if pwd != "" && pwd == pwd1 {
		user.UpdatePWD(pwd)
	}
	glog.Infoln(m, pwd, pwd1)
	user.MultiHsetSave(m)
	//	c.HTML(http.StatusOK, "edit.html", gin.H{
	//		"user": user,
	//	})

}

func (this *roles) EditUser(c *gin.Context) {
	userid := c.Query("userid")
	data := &data.User{Userid: userid}
	data.Get()
	u := &user{
		Userid:      data.Userid,
		PhoneN:      data.Phone,
		Create_ip:   utils.InetTontoa(data.Create_ip).String(),
		Create_time: utils.Unix2Str(int64(data.Create_time)),
		Sex:         data.Sex,
		Nickname:    data.Nickname,
		Diamond:     data.Diamond,
		Coin:        data.Coin,
		Vip:         data.Vip,
	}
	glog.Infoln(u)
	c.HTML(http.StatusOK, "edit.html", gin.H{
		"user": u,
	})
}

// 玩家列表, 根据条件检索玩家
func (this *roles) List(c *gin.Context) {
	start_id := c.Query("start_id")
	end_id := c.Query("end_id")
	//	phone := c.Query("phone")

	page_s := c.Query("page") // string
	act_s := c.Query("act")   // string
	limit_s := c.Query("limit")

	lastID, _ := gossdb.C().Get(data.KEY_LAST_USER_ID)

	this.pager.SetPager(page_s, limit_s, act_s)

	this.selector.SetSelect("limit", limit_s, OPTION)

	glog.Infoln(start_id, end_id, act_s, limit_s, page_s, lastID.String())
	var ids []string
	if start_id != "" || end_id != "" {
		last := []rune(lastID.String())

		if start_id == "" {
			start_id = end_id
		} else if end_id == "" {
			end_id = start_id
		}
		boolean := true
		if len(end_id) <= len(last) && utils.IsNumString(start_id) && utils.IsNumString(end_id) {
			for i := 0; i < len(end_id); i++ {
				if int(end_id[i]) < int(last[i]) {
					break
				} else if int(end_id[i]) > int(last[i]) {
					boolean = false
				}
			}
		} else {
			boolean = false
		}
		if boolean {
			startidnum, _ := strconv.ParseUint(start_id, 10, 64)
			endidnum, _ := strconv.ParseUint(end_id, 10, 64)
			this.pager.SetSize(uint32(endidnum - startidnum))

			ids = utils.Between(utils.StringAddNum(start_id, this.pager.GetStart()), utils.StringAddNum(start_id, this.pager.GetEnd()))
		}
	} else {
		idnum, err := strconv.ParseUint(lastID.String(), 10, 64)
		if err == nil && idnum > 60001 {
			size := idnum - 60001
			glog.Infoln("size ", size)
			this.pager.SetSize(uint32(size))
		}
		glog.Infoln(this.pager.GetStart(), this.pager.GetEnd())
		ids = utils.Between(strconv.FormatUint(idnum-uint64(this.pager.GetEnd()), 10), strconv.FormatUint(idnum-uint64(this.pager.GetStart()), 10))
	}

	lists := data.GetMultiUser(ids)
	users := make([]*user, 0, len(lists))
	for _, v := range lists {
		u := &user{
			Userid:      v.Userid,
			Nickname:    v.Nickname,
			PhoneN:      v.Phone,
			Coin:        v.Coin,
			Diamond:     v.Diamond,
			Vip:         v.Vip,
			Create_ip:   utils.InetTontoa(v.Create_ip).String(),
			Create_time: utils.Unix2Str(int64(v.Create_time)),
			Sex:         v.Sex,
		}
		users = append(users, u)
	}
	glog.Infoln("users : ", this.pager, len(users))

	c.HTML(http.StatusOK, "lists.html", gin.H{
		"pager":    this.pager,
		"selected": this.selector,
		"users":    users,
	})
}
