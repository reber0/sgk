/*
 * @Author: reber
 * @Mail: reber0ask@qq.com
 * @Date: 2022-03-24 17:34:48
 * @LastEditTime: 2024-10-11 15:41:52
 */

package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/manticoresoftware/go-sdk/manticore"
	"github.com/reber0/goutils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//go:embed web
var web embed.FS

//go:embed web/sgk_index_msg.json
var sgk_index_msg []byte

type RespData struct {
	Source   string `json:"source" gorm:"column:source"`
	UID      string `json:"uid" gorm:"column:uid"`
	NickName string `json:"nickname" gorm:"column:nickname"`
	UserName string `json:"username" gorm:"column:username"`
	PassWord string `json:"password" gorm:"column:password"`
	Salt     string `json:"salt" gorm:"column:salt"`
	Secques  string `json:"secques" gorm:"column:secques"`
	Mobile   string `json:"mobile" gorm:"column:mobile"`
	Email    string `json:"email" gorm:"column:email"`
	QQ       string `json:"qq" gorm:"column:qq"`
	RealName string `json:"realname" gorm:"column:realname"`
	Gender   string `json:"gender" gorm:"column:gender"`
	Bday     string `json:"bday" gorm:"column:bday"`
	IdNo     string `json:"idno" gorm:"column:idno"`
	BankNo   string `json:"bankno" gorm:"column:bankno"`
	Address  string `json:"address" gorm:"column:address"`
	Note     string `json:"note" gorm:"column:note"`
}

var store = sessions.NewCookieStore([]byte("asdfasdf"))
var webPassWord = "111"
var mysqlURI = "root:root@tcp(127.0.0.1:3306)/mysql"
var whiteIP = []string{"127.0.0.1", "172.16.3.3"}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies([]string{"172.16.3.3"})

	templates, _ := fs.Sub(web, "web/templates")
	t, _ := template.ParseFS(templates, "*.html")
	r.SetHTMLTemplate(t)

	js, _ := fs.Sub(web, "web/static")
	r.StaticFS("/static", http.FS(js))

	r.GET("/", func(c *gin.Context) {
		remoteIP := c.ClientIP()
		if !goutils.IsInCol(remoteIP, whiteIP) {
			c.String(403, "access denied")
			c.Abort()
		} else {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"title": "SED",
			})
		}
	})
	r.POST("/login", doLogin)
	r.GET("/show", doShow)
	r.POST("/query", doQuery)

	if err := r.Run("0.0.0.0:80"); err != nil {
		fmt.Printf("startup service failed, err:%v\n", err)
	}
}

func doLogin(c *gin.Context) {
	remoteIP := c.ClientIP()
	if !goutils.IsInCol(remoteIP, whiteIP) {
		c.String(403, "access denied")
	} else {
		type PostData struct {
			PassWord string `form:"password" json:"password"`
		}

		postJson := PostData{}
		if err := c.BindJSON(&postJson); err != nil {
			c.JSON(400, gin.H{
				"code": 400,
				"msg":  "检查失败",
			})
		} else {
			password := postJson.PassWord
			if password == webPassWord {
				session, err := store.New(c.Request, "SESSIONID")
				if err != nil {
					fmt.Println("doLogin", err.Error())
				}
				session.Values["islogin"] = "1" // 在 session 中存储值
				session.Save(c.Request, c.Writer)

				c.JSON(200, gin.H{
					"redirect": "/show",
				})
			} else {
				c.JSON(200, gin.H{
					"msg": "error",
				})
			}
		}
	}
}

func doShow(c *gin.Context) {
	remoteIP := c.ClientIP()
	if !goutils.IsInCol(remoteIP, whiteIP) {
		c.String(403, "access denied")
	} else {
		// 获取客户端 cookie 并校验
		session, err := store.Get(c.Request, "SESSIONID")
		if err != nil {
			fmt.Println("doShow", err.Error())
			c.JSON(200, gin.H{
				"redirect": "/login",
			})
			c.Abort()
		}

		if session.Values["islogin"] != nil {
			c.HTML(http.StatusOK, "show.html", gin.H{
				"title": "SED",
			})
			c.Abort()
		} else {
			// c.Redirect(301, "/")
			c.JSON(200, gin.H{
				"redirect": "/login",
			})
		}
	}
}

func doQuery(c *gin.Context) {
	remoteIP := c.ClientIP()
	if !goutils.IsInCol(remoteIP, whiteIP) {
		c.String(403, "access denied")
	} else {
		// 获取客户端 cookie 并校验
		session, err := store.Get(c.Request, "SESSIONID")
		if err != nil {
			fmt.Println("doQuery", err.Error())
			c.JSON(200, gin.H{
				"redirect": "/login",
			})
			c.Abort()
		}

		if session.Values["islogin"] != nil {
			type PostData struct {
				KeyWord string `form:"keyword" json:"keyword"`
			}

			postJson := PostData{}
			if err := c.BindJSON(&postJson); err != nil {
				c.JSON(400, gin.H{
					"code": 400,
					"msg":  "检查失败",
				})
			} else {
				keyword := postJson.KeyWord

				datas := queryExec(keyword)

				c.JSON(200, gin.H{
					"code": 0,
					"data": datas,
				})
			}
		} else {
			// c.Redirect(301, "/")
			c.JSON(200, gin.H{
				"redirect": "/login",
			})
		}
	}
}

func queryExec(keyword string) []RespData {
	datas := make([]RespData, 0, 10000)

	cl := manticore.NewClient()
	cl.SetServer("127.0.0.1", 9312)
	cl.Open()
	defer cl.Close()

	dbconn, err := gorm.Open(mysql.Open(mysqlURI), &gorm.Config{})

	res, err := simplejson.NewJson(sgk_index_msg)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	for _, sgk_index := range res.MustArray() {
		sgk_index := sgk_index.(map[string]interface{})

		index := sgk_index["index"]
		db_name := sgk_index["db_name"]
		table_name := sgk_index["table_name"]
		columns := sgk_index["columns"]
		// fmt.Println(index, db_name, table_name, columns)

		res, err := cl.Query(keyword, index.(string))
		if err != nil {
			fmt.Println(err.Error())
		} else {
			var docIDSlice []string
			for _, v := range res.Matches {
				docID := strconv.Itoa(int(v.DocID))
				docIDSlice = append(docIDSlice, docID)
			}

			if len(docIDSlice) > 0 {
				ids := strings.Join(docIDSlice, ",")
				sql := fmt.Sprintf("select %s from `%s`.`%s` where id in(%s)", columns, db_name, table_name, ids)

				tmp := make([]RespData, 0, 1000)
				result := dbconn.Raw(sql).Scan(&tmp)
				if result.Error != nil {
					fmt.Println(result.Error.Error())
				}
				datas = append(datas, tmp...)
			}
		}
	}

	return datas
}
